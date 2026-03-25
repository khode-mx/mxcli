# Proposal: Catalog Tables for Go mxcli

## Overview

Add support for catalog tables that enable flexible SQL querying of Mendix project metadata. This feature provides an in-memory SQLite database populated from MPR data, allowing users to run arbitrary SQL queries for analysis, searching, and agentic workflows.

## Motivation

Current MDL commands like `SHOW ENTITIES`, `SHOW MICROFLOWS`, etc. provide fixed-format output. Catalog tables enable:

1. **Flexible querying** - Filter, join, aggregate data using standard SQL
2. **Cross-referencing** - Find relationships between objects (e.g., which microflows use entity X)
3. **Agentic search** - AI agents can explore project structure dynamically
4. **Impact analysis** - Track dependencies and usages across the model
5. **Comparison** - Snapshot support for comparing project states over time

## User Experience

### Building the Catalog

```
mdl> SHOW CATALOG TABLES;
Building catalog...
✓ Modules: 12
✓ Entities: 45
✓ Microflows: 120
✓ Nanoflows: 15
✓ Pages: 65
✓ Snippets: 8
✓ Enumerations: 12
✓ Activities: 450
✓ Widgets: 890
✓ XPath Expressions: 78
✓ Catalog ready (1.8s)

Found 13 catalog table(s)
| Table                     |
|---------------------------|
| CATALOG.MODULES           |
| CATALOG.ENTITIES          |
| CATALOG.MICROFLOWS        |
| CATALOG.NANOFLOWS         |
| CATALOG.PAGES             |
| CATALOG.SNIPPETS          |
| CATALOG.ENUMERATIONS      |
| CATALOG.ACTIVITIES        |
| CATALOG.WIDGETS           |
| CATALOG.XPATH_EXPRESSIONS |
| CATALOG.PROJECTS          |
| CATALOG.SNAPSHOTS         |
| CATALOG.OBJECTS           |
```

### Querying Catalog Tables

```sql
-- Find all entities in a specific module
SELECT Name, QualifiedName, EntityType
FROM CATALOG.ENTITIES
WHERE ModuleName = 'MyModule';

-- Find microflows that contain Java actions
SELECT DISTINCT m.QualifiedName, m.Description
FROM CATALOG.MICROFLOWS m
JOIN CATALOG.ACTIVITIES a ON m.Id = a.MicroflowId
WHERE a.ActivityType = 'JavaActionCallAction';

-- Find pages using a specific entity
SELECT p.QualifiedName, w.WidgetType
FROM CATALOG.PAGES p
JOIN CATALOG.WIDGETS w ON p.Id = w.ContainerId
WHERE w.EntityRef LIKE '%Customer%';

-- Count activities by type across project
SELECT ActivityType, COUNT(*) as Count
FROM CATALOG.ACTIVITIES
GROUP BY ActivityType
ORDER BY Count DESC;

-- Find complex XPath expressions
SELECT DocumentQualifiedName, XPathExpression
FROM CATALOG.XPATH_EXPRESSIONS
WHERE LENGTH(XPathExpression) > 50;
```

## Catalog Table Schema

### CATALOG.MODULES

| Column           | Type    | Description                        |
|------------------|---------|------------------------------------|
| Id               | TEXT    | Module UUID                        |
| Name             | TEXT    | Module name                        |
| QualifiedName    | TEXT    | Same as Name for modules           |
| Description      | TEXT    | Module documentation               |
| IsSystemModule   | INTEGER | 1 if system/marketplace module     |
| AppStoreVersion  | TEXT    | Marketplace version if applicable  |
| AppStoreGuid     | TEXT    | Marketplace GUID                   |
| ProjectId        | TEXT    | Parent project ID                  |
| SnapshotId       | TEXT    | Snapshot this data belongs to      |
| SnapshotDate     | TEXT    | ISO timestamp of snapshot          |

### CATALOG.ENTITIES

| Column        | Type    | Description                              |
|---------------|---------|------------------------------------------|
| Id            | TEXT    | Entity UUID                              |
| Name          | TEXT    | Entity name                              |
| QualifiedName | TEXT    | Module.Entity                            |
| ModuleName    | TEXT    | Parent module name                       |
| Folder        | TEXT    | Folder path within module                |
| EntityType    | TEXT    | PERSISTENT, NON_PERSISTENT, EXTERNAL     |
| Description   | TEXT    | Entity documentation                     |
| Generalization| TEXT    | Parent entity qualified name             |
| AttributeCount| INTEGER | Number of attributes                     |
| AssociationCount| INTEGER | Number of associations                 |
| ProjectId     | TEXT    | Parent project ID                        |
| SnapshotId    | TEXT    | Snapshot ID                              |
| SnapshotDate  | TEXT    | Snapshot timestamp                       |

### CATALOG.MICROFLOWS

| Column         | Type    | Description                             |
|----------------|---------|-----------------------------------------|
| Id             | TEXT    | Microflow UUID                          |
| Name           | TEXT    | Microflow name                          |
| QualifiedName  | TEXT    | Module.Microflow                        |
| ModuleName     | TEXT    | Parent module name                      |
| Folder         | TEXT    | Folder path within module               |
| MicroflowType  | TEXT    | MICROFLOW or NANOFLOW                   |
| Description    | TEXT    | Documentation                           |
| ReturnType     | TEXT    | Return type (Entity, Boolean, etc.)     |
| ParameterCount | INTEGER | Number of parameters                    |
| ActivityCount  | INTEGER | Number of activities                    |
| ProjectId      | TEXT    | Parent project ID                       |
| SnapshotId     | TEXT    | Snapshot ID                             |
| SnapshotDate   | TEXT    | Snapshot timestamp                      |

### CATALOG.NANOFLOWS

Same schema as MICROFLOWS, filtered for nanoflows only.

### CATALOG.PAGES

| Column        | Type    | Description                              |
|---------------|---------|------------------------------------------|
| Id            | TEXT    | Page UUID                                |
| Name          | TEXT    | Page name                                |
| QualifiedName | TEXT    | Module.Page                              |
| ModuleName    | TEXT    | Parent module name                       |
| Folder        | TEXT    | Folder path within module                |
| Title         | TEXT    | Page title                               |
| URL           | TEXT    | Page URL                                 |
| LayoutRef     | TEXT    | Layout qualified name                    |
| Description   | TEXT    | Page documentation                       |
| ParameterCount| INTEGER | Number of page parameters                |
| WidgetCount   | INTEGER | Number of widgets                        |
| ProjectId     | TEXT    | Parent project ID                        |
| SnapshotId    | TEXT    | Snapshot ID                              |
| SnapshotDate  | TEXT    | Snapshot timestamp                       |

### CATALOG.SNIPPETS

| Column        | Type    | Description                              |
|---------------|---------|------------------------------------------|
| Id            | TEXT    | Snippet UUID                             |
| Name          | TEXT    | Snippet name                             |
| QualifiedName | TEXT    | Module.Snippet                           |
| ModuleName    | TEXT    | Parent module name                       |
| Folder        | TEXT    | Folder path within module                |
| Description   | TEXT    | Snippet documentation                    |
| ParameterCount| INTEGER | Number of snippet parameters             |
| WidgetCount   | INTEGER | Number of widgets                        |
| ProjectId     | TEXT    | Parent project ID                        |
| SnapshotId    | TEXT    | Snapshot ID                              |
| SnapshotDate  | TEXT    | Snapshot timestamp                       |

### CATALOG.ENUMERATIONS

| Column        | Type    | Description                              |
|---------------|---------|------------------------------------------|
| Id            | TEXT    | Enumeration UUID                         |
| Name          | TEXT    | Enumeration name                         |
| QualifiedName | TEXT    | Module.Enumeration                       |
| ModuleName    | TEXT    | Parent module name                       |
| Folder        | TEXT    | Folder path within module                |
| Description   | TEXT    | Enumeration documentation                |
| ValueCount    | INTEGER | Number of enumeration values             |
| ProjectId     | TEXT    | Parent project ID                        |
| SnapshotId    | TEXT    | Snapshot ID                              |
| SnapshotDate  | TEXT    | Snapshot timestamp                       |

### CATALOG.ACTIVITIES

| Column                 | Type    | Description                        |
|------------------------|---------|------------------------------------|
| Id                     | TEXT    | Activity UUID                      |
| Name                   | TEXT    | Activity type name                 |
| Caption                | TEXT    | Display caption                    |
| ActivityType           | TEXT    | TypeName (e.g., ActionActivity)    |
| MicroflowId            | TEXT    | Parent microflow UUID              |
| MicroflowQualifiedName | TEXT    | Parent microflow qualified name    |
| ModuleName             | TEXT    | Module name                        |
| Folder                 | TEXT    | Folder path                        |
| EntityRef              | TEXT    | Referenced entity (if applicable)  |
| ActionType             | TEXT    | Specific action (Create, Retrieve) |
| Description            | TEXT    | Activity documentation             |
| ProjectId              | TEXT    | Parent project ID                  |
| SnapshotId             | TEXT    | Snapshot ID                        |
| SnapshotDate           | TEXT    | Snapshot timestamp                 |

### CATALOG.WIDGETS

| Column                 | Type    | Description                        |
|------------------------|---------|------------------------------------|
| Id                     | TEXT    | Widget UUID                        |
| Name                   | TEXT    | Widget name                        |
| WidgetType             | TEXT    | Widget type (DataGrid, TextBox)    |
| ContainerId            | TEXT    | Parent page/snippet UUID           |
| ContainerQualifiedName | TEXT    | Parent qualified name              |
| ContainerType          | TEXT    | PAGE or SNIPPET                    |
| ModuleName             | TEXT    | Module name                        |
| Folder                 | TEXT    | Folder path                        |
| EntityRef              | TEXT    | Data source entity                 |
| AttributeRef           | TEXT    | Bound attribute                    |
| Description            | TEXT    | Widget documentation               |
| ProjectId              | TEXT    | Parent project ID                  |
| SnapshotId             | TEXT    | Snapshot ID                        |
| SnapshotDate           | TEXT    | Snapshot timestamp                 |

### CATALOG.XPATH_EXPRESSIONS

| Column                 | Type    | Description                        |
|------------------------|---------|------------------------------------|
| Id                     | TEXT    | Generated ID                       |
| DocumentType           | TEXT    | MICROFLOW, PAGE, etc.              |
| DocumentId             | TEXT    | Parent document UUID               |
| DocumentQualifiedName  | TEXT    | Parent document qualified name     |
| ComponentType          | TEXT    | RETRIEVE_ACTION, DATASOURCE, etc.  |
| ComponentId            | TEXT    | Component UUID                     |
| ComponentName          | TEXT    | Component name                     |
| XPathExpression        | TEXT    | Raw XPath string                   |
| XPathAST               | TEXT    | Parsed XPath as JSON               |
| TargetEntity           | TEXT    | Entity being queried               |
| ReferencedEntities     | TEXT    | JSON array of referenced entities  |
| IsParameterized        | INTEGER | 1 if contains parameters           |
| UsageType              | TEXT    | RETRIEVE, CONSTRAINT, FILTER       |
| ModuleName             | TEXT    | Module name                        |
| Folder                 | TEXT    | Folder path                        |
| ProjectId              | TEXT    | Parent project ID                  |
| SnapshotId             | TEXT    | Snapshot ID                        |
| SnapshotDate           | TEXT    | Snapshot timestamp                 |

### CATALOG.PROJECTS

| Column           | Type    | Description                        |
|------------------|---------|------------------------------------|
| ProjectId        | TEXT    | Project identifier                 |
| ProjectName      | TEXT    | Project name                       |
| MendixVersion    | TEXT    | Mendix version                     |
| CreatedDate      | TEXT    | Catalog creation timestamp         |
| LastSnapshotDate | TEXT    | Most recent snapshot               |
| SnapshotCount    | INTEGER | Number of snapshots                |

### CATALOG.SNAPSHOTS

| Column          | Type    | Description                         |
|-----------------|---------|-------------------------------------|
| SnapshotId      | TEXT    | Snapshot identifier                 |
| SnapshotName    | TEXT    | Optional name                       |
| ProjectId       | TEXT    | Parent project ID                   |
| ProjectName     | TEXT    | Project name                        |
| SnapshotDate    | TEXT    | Snapshot timestamp                  |
| SnapshotSource  | TEXT    | LIVE, GIT, IMPORT                   |
| SourceId        | TEXT    | Git commit or import ID             |
| SourceBranch    | TEXT    | Git branch name                     |
| SourceRevision  | TEXT    | Git revision                        |
| ObjectCount     | INTEGER | Total objects in snapshot           |
| IsActive        | INTEGER | 1 if current active snapshot        |

### CATALOG.OBJECTS

Union view of all object types for generic querying:

| Column        | Type    | Description                              |
|---------------|---------|------------------------------------------|
| Id            | TEXT    | Object UUID                              |
| ObjectType    | TEXT    | MODULE, ENTITY, MICROFLOW, PAGE, etc.    |
| Name          | TEXT    | Object name                              |
| QualifiedName | TEXT    | Full qualified name                      |
| ModuleName    | TEXT    | Parent module (empty for modules)        |
| Folder        | TEXT    | Folder path                              |
| Description   | TEXT    | Documentation                            |
| ProjectId     | TEXT    | Parent project ID                        |
| SnapshotId    | TEXT    | Snapshot ID                              |
| SnapshotDate  | TEXT    | Snapshot timestamp                       |

## Architecture

### Component Structure

```
mdl/
├── catalog/
│   ├── catalog.go           # Main Catalog struct and interface
│   ├── builder.go           # Builds catalog from MPR data
│   ├── tables.go            # Table definitions and schemas
│   ├── query.go             # SQL query execution
│   └── snapshot.go          # Snapshot management
│
├── executor/
│   ├── cmd_catalog.go       # SHOW CATALOG TABLES, SELECT handlers
│   └── ...
│
└── grammar/
    └── MDLParser.g4         # Add SELECT statement grammar
```

### Catalog Interface

```go
package catalog

import (
    "database/sql"
    "time"
)

// Catalog provides SQL querying over Mendix project metadata.
type Catalog struct {
    db          *sql.DB        // In-memory SQLite
    projectID   string
    projectName string
    snapshots   map[string]*Snapshot
    activeSnap  string
}

// Snapshot represents a point-in-time view of project data.
type Snapshot struct {
    ID         string
    Name       string
    Date       time.Time
    Source     SnapshotSource
    SourceID   string
    Branch     string
    Revision   string
    ObjectCount int
}

type SnapshotSource string

const (
    SnapshotSourceLive    SnapshotSource = "LIVE"
    SnapshotSourceGit     SnapshotSource = "GIT"
    SnapshotSourceImport  SnapshotSource = "IMPORT"
)

// New creates a new catalog with an in-memory SQLite database.
func New() (*Catalog, error)

// Build populates the catalog from MPR reader data.
func (c *Catalog) Build(reader *mpr.Reader, progress func(table string, count int)) error

// Query executes a SQL query and returns results.
func (c *Catalog) Query(sql string) (*QueryResult, error)

// Tables returns the list of available catalog tables.
func (c *Catalog) Tables() []string

// CreateSnapshot creates a named snapshot of current state.
func (c *Catalog) CreateSnapshot(name string, source SnapshotSource) (*Snapshot, error)

// Close releases catalog resources.
func (c *Catalog) Close() error

// QueryResult holds query results.
type QueryResult struct {
    Columns []string
    Rows    [][]interface{}
    Count   int
}
```

### Builder Implementation

```go
package catalog

// Builder populates catalog tables from MPR data.
type Builder struct {
    catalog *Catalog
    reader  *mpr.Reader
    snapshot *Snapshot
}

func (b *Builder) Build(progress func(table string, count int)) error {
    // Build tables in dependency order
    if err := b.buildModules(progress); err != nil {
        return err
    }
    if err := b.buildEntities(progress); err != nil {
        return err
    }
    if err := b.buildMicroflows(progress); err != nil {
        return err
    }
    if err := b.buildNanoflows(progress); err != nil {
        return err
    }
    if err := b.buildPages(progress); err != nil {
        return err
    }
    if err := b.buildSnippets(progress); err != nil {
        return err
    }
    if err := b.buildEnumerations(progress); err != nil {
        return err
    }
    if err := b.buildActivities(progress); err != nil {
        return err
    }
    if err := b.buildWidgets(progress); err != nil {
        return err
    }
    if err := b.buildXPathExpressions(progress); err != nil {
        return err
    }
    if err := b.buildObjectsView(progress); err != nil {
        return err
    }
    return nil
}

func (b *Builder) buildModules(progress func(string, int)) error {
    modules, err := b.reader.ListModules()
    if err != nil {
        return err
    }

    stmt, err := b.catalog.db.Prepare(`
        INSERT INTO modules (Id, Name, QualifiedName, Description,
            IsSystemModule, AppStoreVersion, AppStoreGuid,
            ProjectId, SnapshotId, SnapshotDate)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `)
    if err != nil {
        return err
    }
    defer stmt.Close()

    for _, m := range modules {
        isSystem := 0
        if m.FromAppStore {
            isSystem = 1
        }
        _, err := stmt.Exec(
            m.ID, m.Name, m.Name, m.Documentation,
            isSystem, m.AppStoreVersion, m.AppStoreGuid,
            b.snapshot.ProjectID, b.snapshot.ID, b.snapshot.Date,
        )
        if err != nil {
            return err
        }
    }

    progress("Modules", len(modules))
    return nil
}

// Similar methods for other tables...
```

### Grammar Extensions

Add to MDLParser.g4:

```antlr
// Catalog statements
catalogStatement
    : SHOW CATALOG TABLES
    | selectStatement
    ;

selectStatement
    : SELECT selectColumns FROM catalogTable
      (WHERE whereClause)?
      (GROUP BY groupByClause)?
      (ORDER BY orderByClause)?
      (LIMIT limitClause)?
    ;

catalogTable
    : CATALOG DOT IDENTIFIER
    ;

selectColumns
    : STAR
    | selectColumn (COMMA selectColumn)*
    ;

selectColumn
    : expression (AS? IDENTIFIER)?
    ;

// SQL expressions (subset)
expression
    : IDENTIFIER
    | IDENTIFIER DOT IDENTIFIER
    | STRING_LITERAL
    | NUMBER
    | expression (PLUS | MINUS | STAR | SLASH) expression
    | expression (EQ | NE | LT | GT | LE | GE) expression
    | expression (AND | OR) expression
    | expression LIKE expression
    | expression IN LPAREN expressionList RPAREN
    | functionCall
    | LPAREN expression RPAREN
    ;

functionCall
    : IDENTIFIER LPAREN expressionList? RPAREN
    ;
```

### Executor Integration

```go
// cmd_catalog.go
package executor

func (e *Executor) execShowCatalogTables() error {
    // Build catalog if not already built
    if e.catalog == nil {
        cat, err := catalog.New()
        if err != nil {
            return err
        }
        e.catalog = cat

        fmt.Fprintln(e.output, "Building catalog...")
        err = e.catalog.Build(e.reader, func(table string, count int) {
            fmt.Fprintf(e.output, "✓ %s: %d\n", table, count)
        })
        if err != nil {
            return err
        }
    }

    tables := e.catalog.Tables()
    fmt.Fprintf(e.output, "\nFound %d catalog table(s)\n", len(tables))

    // Output table list
    e.outputTable([]string{"Table"}, tables)
    return nil
}

func (e *Executor) execSelect(stmt *ast.SelectStmt) error {
    // Ensure catalog is built
    if e.catalog == nil {
        return fmt.Errorf("catalog not built - run SHOW CATALOG TABLES first")
    }

    // Convert AST to SQL string
    sql := e.selectToSQL(stmt)

    // Execute query
    result, err := e.catalog.Query(sql)
    if err != nil {
        return fmt.Errorf("failed to execute catalog query: %w", err)
    }

    // Output results
    fmt.Fprintf(e.output, "Found %d result(s)\n", result.Count)
    if result.Count == 0 {
        fmt.Fprintln(e.output, "(no results)")
        return nil
    }

    e.outputQueryResults(result)
    return nil
}
```

## Implementation Plan

### Phase 1: Core Infrastructure
- Create `mdl/catalog/` package structure
- Implement in-memory SQLite setup with table schemas
- Add basic `SHOW CATALOG TABLES` command

### Phase 2: Table Builders
- Implement builders for each table type
- Add progress reporting during build
- Handle cross-references (activities→microflows, widgets→pages)

### Phase 3: Query Support
- Add SELECT statement to grammar
- Implement SQL generation from AST
- Add query execution and result formatting

### Phase 4: Advanced Features
- Snapshot support for versioning
- CATALOG.OBJECTS union view
- XPath expression parsing and AST storage

### Phase 5: Optimizations
- Lazy loading of detailed data
- Caching of frequently accessed data
- Index optimization for common queries

## Use Cases

### 1. Find All Usages of an Entity

```sql
-- Find all microflows that use Customer entity
SELECT DISTINCT m.QualifiedName, a.ActivityType, a.Caption
FROM CATALOG.MICROFLOWS m
JOIN CATALOG.ACTIVITIES a ON m.Id = a.MicroflowId
WHERE a.EntityRef LIKE '%Customer%'
ORDER BY m.QualifiedName;

-- Find all pages with Customer data
SELECT p.QualifiedName, w.WidgetType, w.Name
FROM CATALOG.PAGES p
JOIN CATALOG.WIDGETS w ON p.Id = w.ContainerId
WHERE w.EntityRef LIKE '%Customer%';
```

### 2. Code Quality Analysis

```sql
-- Find microflows with many activities (complexity)
SELECT QualifiedName, ActivityCount
FROM CATALOG.MICROFLOWS
WHERE ActivityCount > 20
ORDER BY ActivityCount DESC;

-- Find pages with many widgets
SELECT QualifiedName, WidgetCount
FROM CATALOG.PAGES
WHERE WidgetCount > 50
ORDER BY WidgetCount DESC;
```

### 3. Documentation Coverage

```sql
-- Find undocumented microflows
SELECT QualifiedName
FROM CATALOG.MICROFLOWS
WHERE Description IS NULL OR Description = '';

-- Find documented entities
SELECT QualifiedName, Description
FROM CATALOG.ENTITIES
WHERE Description IS NOT NULL AND Description != '';
```

### 4. Agentic Exploration

AI agents can use catalog queries to understand project structure:

```sql
-- Get project overview
SELECT ObjectType, COUNT(*) as Count
FROM CATALOG.OBJECTS
GROUP BY ObjectType;

-- Find entry points (pages without parameters)
SELECT QualifiedName, Title, URL
FROM CATALOG.PAGES
WHERE ParameterCount = 0 AND URL IS NOT NULL;

-- Find external integrations
SELECT QualifiedName, ActivityType
FROM CATALOG.ACTIVITIES
WHERE ActivityType IN ('RestCallAction', 'WebServiceCallAction', 'JavaActionCallAction');
```

## Dependencies

- `github.com/mattn/go-sqlite3` - Already a project dependency (used for MPR v1)
- No additional dependencies required

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Memory usage for large projects | Lazy loading, streaming results |
| Build time for 1000+ objects | Progress indicators, parallel building |
| SQL injection in queries | Parameterized queries where possible |
| Schema changes breaking queries | Version catalog schema, migration support |

## Success Criteria

1. All 13 catalog tables populated correctly
2. Standard SQL queries execute successfully
3. Build time < 5 seconds for typical projects
4. Memory usage < 100MB for large projects
5. Query results match expected TypeScript output
