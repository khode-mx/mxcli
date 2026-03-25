# mxcli Code Search Extension - Implementation Proposal

## Executive Summary

This proposal extends mxcli with code search capabilities to help Claude Code understand and navigate Mendix projects. Based on analysis of the existing codebase, this can be implemented by:

1. Adding a `catalog.references` table to track cross-element relationships
2. Extending the existing catalog builder to extract references during indexing
3. Adding new MDL statements for search queries
4. Adding CLI subcommands for direct search access
5. Implementing context assembly for LLM consumption

## Current State Analysis

### Existing Infrastructure (What We Already Have)

**Catalog System** (`mdl/catalog/`):
- In-memory SQLite database with 13 tables
- Fast mode (metadata only) and full mode (activities/widgets)
- SQL query interface via `SELECT ... FROM CATALOG.xxx`
- Snapshot system for point-in-time comparisons

**Current Reference Tracking**:
- `activities.MicroflowId` - links activities to their containing microflow
- `activities.EntityRef` - entity referenced by CreateObjectAction
- `widgets.ContainerId` - links widgets to pages/snippets
- `widgets.EntityRef`, `widgets.AttributeRef` - data binding references
- `entities.Generalization` - parent entity reference

**What's Missing**:
- Microflow-to-microflow call references
- Page-to-layout references
- Entity event handlers → microflow references
- Association → target entity references
- Widget action → page/microflow references
- Unified reference query interface

### Key Files

| Component | Location | Purpose |
|-----------|----------|---------|
| Catalog schema | `mdl/catalog/tables.go` | Table definitions |
| Catalog builder | `mdl/catalog/builder*.go` | Populates tables |
| MDL grammar | `mdl/grammar/MDLParser.g4`, `MDLLexer.g4` | Syntax definition |
| AST nodes | `mdl/ast/ast_query.go` | Statement types |
| Executor | `mdl/executor/executor.go` | Statement dispatch |
| CLI | `cmd/mxcli/main.go` | Cobra commands |
| Microflow actions | `sdk/microflows/microflows_actions.go` | Action types with references |
| Page widgets | `sdk/pages/pages_widgets*.go` | Widget types with references |

---

## Proposed Implementation

### 1. Schema Changes

#### New References Table

Add to `mdl/catalog/tables.go`:

```sql
CREATE TABLE IF NOT EXISTS references (
    Id TEXT PRIMARY KEY,
    SourceType TEXT NOT NULL,       -- 'microflow', 'page', 'entity', 'widget', etc.
    SourceId TEXT NOT NULL,         -- UUID of the source element
    SourceName TEXT NOT NULL,       -- Qualified name (e.g., 'Module.Microflow')
    SourceLocation TEXT,            -- Optional: activity/widget ID within document
    TargetType TEXT NOT NULL,       -- 'microflow', 'entity', 'page', 'attribute', etc.
    TargetId TEXT,                  -- UUID of target (if available)
    TargetName TEXT NOT NULL,       -- Qualified name of referenced thing
    RefKind TEXT NOT NULL,          -- 'call', 'retrieve', 'change', 'create', etc.
    -- Snapshot metadata (same as other tables)
    ProjectId TEXT,
    ProjectName TEXT,
    SnapshotId TEXT,
    SnapshotDate TEXT,
    SnapshotSource TEXT,
    SourceId TEXT,
    SourceBranch TEXT,
    SourceRevision TEXT
);

CREATE INDEX IF NOT EXISTS idx_refs_target ON references(TargetType, TargetName);
CREATE INDEX IF NOT EXISTS idx_refs_source ON references(SourceType, SourceName);
CREATE INDEX IF NOT EXISTS idx_refs_kind ON references(RefKind);
CREATE INDEX IF NOT EXISTS idx_refs_source_id ON references(SourceId);
CREATE INDEX IF NOT EXISTS idx_refs_target_id ON references(TargetId);
```

#### Convenience Views

```sql
-- Microflow call graph
CREATE VIEW IF NOT EXISTS call_graph AS
SELECT
    SourceName as Caller,
    TargetName as Callee
FROM references
WHERE SourceType = 'microflow'
  AND TargetType = 'microflow'
  AND RefKind = 'call';

-- Entity usage view
CREATE VIEW IF NOT EXISTS entity_usage AS
SELECT
    TargetName as Entity,
    SourceType,
    SourceName,
    RefKind
FROM references
WHERE TargetType = 'entity';

-- Page references view
CREATE VIEW IF NOT EXISTS page_refs AS
SELECT
    SourceName as Source,
    TargetName as Page,
    RefKind
FROM references
WHERE TargetType = 'page';
```

### 2. Reference Extraction

#### Reference Kinds to Extract

| Source Type | Reference Kind | Target Type | Example |
|-------------|----------------|-------------|---------|
| microflow | call | microflow | MicroflowCallAction |
| microflow | create | entity | CreateObjectAction |
| microflow | change | entity | ChangeObjectAction |
| microflow | retrieve | entity | RetrieveAction |
| microflow | delete | entity | DeleteObjectAction |
| microflow | show_page | page | ShowPageAction |
| microflow | close_page | page | ClosePageAction |
| microflow | parameter_type | entity | Parameter with Object/List type |
| microflow | return_type | entity | Return type Object/List |
| page | layout | layout | LayoutCall |
| page | data_source | entity | DataView/ListView entity |
| page | data_source | microflow | MicroflowSource |
| page | action | microflow | MicroflowClientAction |
| page | action | page | PageClientAction |
| widget | attribute | attribute | TextBox attribute binding |
| widget | action | microflow | Button on-click action |
| entity | generalization | entity | Parent entity |
| entity | event_handler | microflow | EventHandler |
| entity | association | entity | Association target |
| enumeration | attribute_type | entity | EnumerationAttributeType |

#### Builder Extensions

**New file: `mdl/catalog/builder_references.go`**

```go
package catalog

import (
    "database/sql"
    "github.com/mendixlabs/mxcli/model"
    "github.com/mendixlabs/mxcli/sdk/microflows"
    "github.com/mendixlabs/mxcli/sdk/pages"
)

// addReference adds a reference to the database
func (b *Builder) addReference(stmt *sql.Stmt,
    sourceType, sourceID, sourceName, sourceLocation,
    targetType, targetID, targetName, refKind string) error {

    projectID, projectName, snapshotID, snapshotDate, snapshotSource,
        srcID, srcBranch, srcRevision := b.snapshotMeta()

    _, err := stmt.Exec(
        model.GenerateID(), // unique ID for reference record
        sourceType, sourceID, sourceName, sourceLocation,
        targetType, targetID, targetName, refKind,
        projectID, projectName, snapshotID, snapshotDate, snapshotSource,
        srcID, srcBranch, srcRevision,
    )
    return err
}

// extractMicroflowReferences extracts all references from a microflow
func (b *Builder) extractMicroflowReferences(stmt *sql.Stmt,
    mf *microflows.Microflow, qualifiedName string) error {

    sourceID := string(mf.ID)

    // Parameter types
    for _, param := range mf.Parameters {
        if entityName := getEntityFromDataType(param.ParameterType); entityName != "" {
            b.addReference(stmt, "microflow", sourceID, qualifiedName, "",
                "entity", "", entityName, "parameter_type")
        }
    }

    // Return type
    if entityName := getEntityFromDataType(mf.ReturnType); entityName != "" {
        b.addReference(stmt, "microflow", sourceID, qualifiedName, "",
            "entity", "", entityName, "return_type")
    }

    // Actions
    if mf.ObjectCollection != nil {
        for _, obj := range mf.ObjectCollection.Objects {
            b.extractActionReferences(stmt, obj, sourceID, qualifiedName)
        }
    }

    return nil
}

// extractActionReferences extracts references from microflow actions
func (b *Builder) extractActionReferences(stmt *sql.Stmt,
    obj microflows.MicroflowObject, sourceID, sourceName string) {

    activityID := string(obj.GetID())

    if act, ok := obj.(*microflows.ActionActivity); ok && act.Action != nil {
        switch a := act.Action.(type) {
        case *microflows.MicroflowCallAction:
            // Microflow call reference
            if a.MicroflowQualifiedName != "" {
                b.addReference(stmt, "microflow", sourceID, sourceName, activityID,
                    "microflow", string(a.MicroflowID), a.MicroflowQualifiedName, "call")
            }

        case *microflows.CreateObjectAction:
            if a.EntityQualifiedName != "" {
                b.addReference(stmt, "microflow", sourceID, sourceName, activityID,
                    "entity", string(a.EntityID), a.EntityQualifiedName, "create")
            }

        case *microflows.ChangeObjectAction:
            if a.EntityQualifiedName != "" {
                b.addReference(stmt, "microflow", sourceID, sourceName, activityID,
                    "entity", "", a.EntityQualifiedName, "change")
            }

        case *microflows.RetrieveAction:
            if src, ok := a.RetrieveSource.(*microflows.DatabaseRetrieveSource); ok {
                if src.EntityQualifiedName != "" {
                    b.addReference(stmt, "microflow", sourceID, sourceName, activityID,
                        "entity", string(src.EntityID), src.EntityQualifiedName, "retrieve")
                }
            }

        case *microflows.ShowPageAction:
            if a.PageQualifiedName != "" {
                b.addReference(stmt, "microflow", sourceID, sourceName, activityID,
                    "page", string(a.PageID), a.PageQualifiedName, "show_page")
            }

        // Add more action types as needed...
        }
    }
}

// getEntityFromDataType extracts entity name from a data type
func getEntityFromDataType(dt microflows.DataType) string {
    switch t := dt.(type) {
    case *microflows.ObjectType:
        return t.EntityQualifiedName
    case *microflows.ListType:
        return t.EntityQualifiedName
    }
    return ""
}
```

**Extend `builder_microflows.go`**:
- Add reference extraction after processing each microflow
- Only in full mode (references require deep parsing)

**New file: `mdl/catalog/builder_pages_refs.go`**:
- Extract page → layout references
- Extract widget → entity/microflow references
- Extract action → page/microflow references

**Extend `builder_modules.go`**:
- Extract entity → generalization references
- Extract entity → event handler microflow references
- Extract association → target entity references

### 3. MDL Statement Extensions

#### New SEARCH Statement

**Grammar additions** (`MDLParser.g4`):

```antlr
searchStatement
    : SEARCH (searchType)? searchPattern (searchOptions)*
    ;

searchType
    : ENTITY | MICROFLOW | PAGE | REFERENCES | CALLERS | CALLEES
    ;

searchPattern
    : STRING_LITERAL
    | qualifiedName
    ;

searchOptions
    : IN qualifiedName
    | WHERE expression
    | TRANSITIVE
    | LIMIT INTEGER_LITERAL
    ;

// Keywords
SEARCH: S E A R C H;
CALLERS: C A L L E R S;
CALLEES: C A L L E E S;
REFERENCES: R E F E R E N C E S;
TRANSITIVE: T R A N S I T I V E;
```

**AST types** (`mdl/ast/ast_search.go`):

```go
package ast

type SearchType int

const (
    SearchAll SearchType = iota
    SearchEntity
    SearchMicroflow
    SearchPage
    SearchReferences
    SearchCallers
    SearchCallees
)

type SearchStmt struct {
    SearchType   SearchType
    Pattern      string
    InModule     string
    WhereClause  string
    Transitive   bool
    Limit        int
}

func (s *SearchStmt) isStatement() {}
```

**Executor** (`mdl/executor/cmd_search.go`):

```go
package executor

import (
    "fmt"
    "strings"
    "github.com/mendixlabs/mxcli/mdl/ast"
)

func (e *Executor) execSearch(s *ast.SearchStmt) error {
    switch s.SearchType {
    case ast.SearchCallers:
        return e.searchCallers(s.Pattern, s.Transitive, s.Limit)
    case ast.SearchCallees:
        return e.searchCallees(s.Pattern, s.Transitive, s.Limit)
    case ast.SearchReferences:
        return e.searchReferences(s.Pattern, s.InModule, s.Limit)
    default:
        return e.searchSymbols(s)
    }
}

func (e *Executor) searchCallers(target string, transitive bool, limit int) error {
    if transitive {
        // Recursive CTE for transitive callers
        query := `
            WITH RECURSIVE callers_cte AS (
                SELECT Caller, 1 as Depth
                FROM call_graph
                WHERE Callee = ?
                UNION ALL
                SELECT cg.Caller, c.Depth + 1
                FROM call_graph cg
                JOIN callers_cte c ON cg.Callee = c.Caller
                WHERE c.Depth < 10
            )
            SELECT DISTINCT Caller, MIN(Depth) as Depth
            FROM callers_cte
            GROUP BY Caller
            ORDER BY Depth, Caller
        `
        if limit > 0 {
            query += fmt.Sprintf(" LIMIT %d", limit)
        }
        return e.execCatalogQuery(query, target)
    }

    // Direct callers only
    query := `SELECT Caller FROM call_graph WHERE Callee = ?`
    if limit > 0 {
        query += fmt.Sprintf(" LIMIT %d", limit)
    }
    return e.execCatalogQuery(query, target)
}

func (e *Executor) searchCallees(source string, transitive bool, limit int) error {
    // Similar to searchCallers but reversed
    // ...
}

func (e *Executor) searchReferences(target, module string, limit int) error {
    query := `
        SELECT SourceType, SourceName, RefKind
        FROM references
        WHERE TargetName = ? OR TargetName LIKE ?
    `
    args := []interface{}{target, "%" + target}

    if module != "" {
        query += " AND SourceName LIKE ?"
        args = append(args, module+".%")
    }

    query += " ORDER BY RefKind, SourceType, SourceName"
    if limit > 0 {
        query += fmt.Sprintf(" LIMIT %d", limit)
    }

    return e.execCatalogQueryWithArgs(query, args...)
}
```

#### MDL Function Syntax (Alternative)

For REPL usage, add function-style queries:

```ruby
# Find callers
SHOW CALLERS OF Module.MicroflowName
SHOW CALLERS OF Module.MicroflowName TRANSITIVE

# Find callees
SHOW CALLEES OF Module.MicroflowName

# Find references
SHOW REFERENCES TO Module.EntityName
SHOW REFERENCES TO Module.EntityName.AttributeName

# Impact analysis
SHOW IMPACT OF Module.EntityName
```

### 4. CLI Commands

**Add to `cmd/mxcli/main.go`**:

```go
var searchCmd = &cobra.Command{
    Use:   "search <pattern>",
    Short: "Search for elements in the project",
    Long: `Search for modules, entities, microflows, pages, etc.

Examples:
  mxcli search -p app.mpr Customer
  mxcli search -p app.mpr --type entity Order
  mxcli search -p app.mpr --callers Module.ValidateOrder
  mxcli search -p app.mpr --callees Module.ProcessOrder --transitive
  mxcli search -p app.mpr --refs Module.Customer
`,
    Args: cobra.MinimumNArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        // Implementation
    },
}

func init() {
    searchCmd.Flags().StringP("type", "t", "", "Element type: entity, microflow, page")
    searchCmd.Flags().Bool("callers", false, "Find callers of a microflow")
    searchCmd.Flags().Bool("callees", false, "Find callees of a microflow")
    searchCmd.Flags().Bool("refs", false, "Find references to an element")
    searchCmd.Flags().Bool("transitive", false, "Include transitive references")
    searchCmd.Flags().IntP("limit", "l", 50, "Maximum results")

    rootCmd.AddCommand(searchCmd)
}
```

**Additional search subcommands**:

```go
// mxcli callers <microflow> [--transitive]
var callersCmd = &cobra.Command{...}

// mxcli callees <microflow> [--transitive]
var calleesCmd = &cobra.Command{...}

// mxcli refs <element> [--type entity|microflow|page]
var refsCmd = &cobra.Command{...}

// mxcli impact <element>
var impactCmd = &cobra.Command{...}

// mxcli context <element> [--depth N] [--max-tokens N]
var contextCmd = &cobra.Command{...}
```

### 5. Context Assembly Feature

The `context` command assembles relevant information for LLM consumption.

**New file: `mdl/executor/cmd_context.go`**:

```go
package executor

import (
    "fmt"
    "strings"
)

type ContextAssembler struct {
    executor  *Executor
    maxTokens int
    depth     int
    visited   map[string]bool
    output    strings.Builder
}

func (e *Executor) assembleContext(target string, depth, maxTokens int) (string, error) {
    ca := &ContextAssembler{
        executor:  e,
        maxTokens: maxTokens,
        depth:     depth,
        visited:   make(map[string]bool),
    }

    return ca.assemble(target)
}

func (ca *ContextAssembler) assemble(target string) (string, error) {
    // 1. Determine target type (microflow, entity, page)
    targetType, err := ca.detectType(target)
    if err != nil {
        return "", err
    }

    // 2. Get the target element definition
    ca.addSection("Target: " + target)
    ca.addDefinition(target, targetType)

    // 3. Get related elements based on type
    switch targetType {
    case "microflow":
        ca.assembleMicroflowContext(target)
    case "entity":
        ca.assembleEntityContext(target)
    case "page":
        ca.assemblePageContext(target)
    }

    return ca.output.String(), nil
}

func (ca *ContextAssembler) assembleMicroflowContext(name string) {
    // Add entities used (create, change, retrieve)
    ca.addSection("Entities Used")
    ca.addQueryResults(`
        SELECT DISTINCT TargetName, RefKind
        FROM references
        WHERE SourceName = ? AND TargetType = 'entity'
    `, name)

    // Add called microflows (up to depth)
    ca.addSection("Called Microflows")
    ca.addCallees(name, ca.depth)

    // Add direct callers (limited)
    ca.addSection("Direct Callers (sample)")
    ca.addQueryResults(`
        SELECT SourceName FROM references
        WHERE TargetName = ? AND RefKind = 'call'
        LIMIT 5
    `, name)

    // Add parameter/return types
    ca.addSection("Signature")
    ca.addMicroflowSignature(name)
}

func (ca *ContextAssembler) assembleEntityContext(name string) {
    // Add entity definition with attributes
    ca.addSection("Entity Definition")
    ca.addEntityDefinition(name)

    // Add microflows that use this entity
    ca.addSection("Microflows Using This Entity")
    ca.addQueryResults(`
        SELECT DISTINCT SourceName, RefKind
        FROM references
        WHERE TargetName = ? AND SourceType = 'microflow'
        ORDER BY RefKind, SourceName
        LIMIT 20
    `, name)

    // Add pages that display this entity
    ca.addSection("Pages Displaying This Entity")
    ca.addQueryResults(`
        SELECT DISTINCT SourceName
        FROM references
        WHERE TargetName = ? AND SourceType = 'page'
        LIMIT 10
    `, name)
}
```

**CLI command**:

```bash
# Assemble context for a microflow
mxcli context -p app.mpr Module.ProcessOrder --depth 2 --max-tokens 4000

# Output format (markdown):
# ## Target: Module.ProcessOrder
#
# ### Definition
# MICROFLOW Module.ProcessOrder(Order: MyModule.Order) RETURNS Boolean
#   ...microflow definition...
#
# ### Entities Used
# | Entity | Usage |
# |--------|-------|
# | MyModule.Order | retrieve, change |
# | MyModule.OrderLine | create |
#
# ### Called Microflows
# - Module.ValidateOrder (depth 1)
# - Module.CalculateTotal (depth 1)
# - Module.SendNotification (depth 2)
#
# ### Direct Callers
# - Module.ACT_SubmitOrder
# - Module.ACT_ReprocessOrder
```

### 6. Optional LSP Server (Phase 2)

For IDE integration, add an LSP server mode:

**New package: `mdl/lsp/`**

```go
package lsp

import (
    "github.com/sourcegraph/go-lsp"
    "github.com/sourcegraph/jsonrpc2"
)

type Server struct {
    catalog *catalog.Catalog
}

// Supported LSP methods:
// - textDocument/definition
// - textDocument/references
// - textDocument/hover
// - workspace/symbol
```

**CLI command**:

```bash
mxcli lsp              # stdio mode
mxcli lsp --tcp :9257  # TCP mode
```

This is deferred to Phase 2 as it requires additional complexity and VS Code extension development.

---

## Implementation Plan

### Phase 1: Core Reference Tracking (Estimated: Medium Complexity)

1. **Schema changes** - Add references table and indexes
   - File: `mdl/catalog/tables.go`
   - Add table creation SQL
   - Add view definitions

2. **Reference extraction for microflows**
   - File: `mdl/catalog/builder_references.go` (new)
   - File: `mdl/catalog/builder_microflows.go` (extend)
   - Extract: call, create, change, retrieve, show_page references

3. **Reference extraction for pages**
   - File: `mdl/catalog/builder_pages_refs.go` (new)
   - File: `mdl/catalog/builder_pages.go` (extend)
   - Extract: layout, data_source, action references

4. **Reference extraction for entities**
   - File: `mdl/catalog/builder_modules.go` (extend)
   - Extract: generalization, event_handler, association references

5. **Basic query commands**
   - File: `mdl/executor/cmd_search.go` (new)
   - Add: `SHOW CALLERS`, `SHOW CALLEES`, `SHOW REFERENCES`

### Phase 2: Search Commands & CLI (Estimated: Low-Medium Complexity)

6. **Grammar extensions**
   - Files: `mdl/grammar/MDLParser.g4`, `MDLLexer.g4`
   - Add: SEARCH statement, CALLERS/CALLEES keywords

7. **AST and visitor**
   - Files: `mdl/ast/ast_search.go` (new), `mdl/visitor/visitor_search.go` (new)

8. **CLI commands**
   - File: `cmd/mxcli/main.go`
   - Add: search, callers, callees, refs subcommands

### Phase 3: Context Assembly (Estimated: Medium Complexity)

9. **Context assembler**
   - File: `mdl/executor/cmd_context.go` (new)
   - Implement context assembly logic

10. **CLI integration**
    - Add: context command with depth/token options

### Phase 4: LSP Server (Estimated: High Complexity, Optional)

11. **LSP implementation**
    - Package: `mdl/lsp/` (new)
    - Implement: definition, references, hover, symbol

12. **CLI integration**
    - Add: lsp command

---

## Testing Strategy

1. **Unit tests** for reference extraction
   - Test each action type extracts correct references
   - Test edge cases (nil values, missing qualified names)

2. **Integration tests** for search queries
   - Test transitive caller/callee queries
   - Test reference filtering by type/module

3. **End-to-end tests** with sample projects
   - Verify expected references are found
   - Verify context assembly produces useful output

---

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Performance with large projects | High | Use indexes, limit results, lazy loading |
| Missing action type handlers | Medium | Implement common types first, log unknowns |
| Qualified name resolution | Medium | Use existing hierarchy system |
| Full mode requirement | Low | Document that references need REFRESH CATALOG FULL |

---

## Open Questions

1. **Should references be cached or computed on-demand?**
   - Current proposal: Part of catalog build (cached)
   - Alternative: Compute on query (slower but always fresh)

2. **How to handle cross-module references?**
   - Qualified names handle this naturally
   - Consider: Module filtering options

3. **What's the priority for LSP support?**
   - Could be deferred if CLI/REPL is sufficient for Claude Code

4. **Should we add a `SHOW IMPACT` command?**
   - Would show everything affected by changing an element
   - Could be expensive for highly-connected elements

---

## Appendix: Current Catalog Schema Reference

```sql
-- Existing tables (from mdl/catalog/tables.go)
modules (Id, Name, QualifiedName, ModuleName, Folder, Description, ...)
entities (Id, Name, QualifiedName, ModuleName, Folder, EntityType, Generalization, ...)
microflows (Id, Name, QualifiedName, ModuleName, Folder, MicroflowType, ReturnType, ...)
nanoflows -- View: filtered microflows
pages (Id, Name, QualifiedName, ModuleName, Folder, Title, URL, LayoutRef, ...)
snippets (Id, Name, QualifiedName, ModuleName, Folder, ...)
enumerations (Id, Name, QualifiedName, ModuleName, Folder, ValueCount, ...)
activities (Id, Name, Caption, ActivityType, MicroflowId, EntityRef, ActionType, ...)
widgets (Id, Name, WidgetType, ContainerId, EntityRef, AttributeRef, ...)
xpath_expressions (...)  -- Schema exists but not populated
projects (...)
snapshots (...)
objects -- View: union of all element types

-- Proposed new table
references (Id, SourceType, SourceId, SourceName, SourceLocation,
            TargetType, TargetId, TargetName, RefKind, ...)

-- Proposed new views
call_graph (Caller, Callee)
entity_usage (Entity, SourceType, SourceName, RefKind)
page_refs (Source, Page, RefKind)
```
