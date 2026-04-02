# Proposal: Version-Aware Agent Support

**Status:** Draft
**Date:** 2026-04-02

## Problem Statement

Three use cases require mxcli to be version-aware at the MDL level:

1. **Generate**: An AI coding agent creates MDL for a Mendix project but doesn't know which features are available in the project's version. It writes `CREATE VIEW ENTITY` for a 9.x project and gets a cryptic BSON error.

2. **Validate**: A user runs an MDL script that uses 11.0+ syntax on a 10.24 project. The script executes, writes corrupt BSON, and the error only surfaces when Studio Pro tries to open the project.

3. **Upgrade**: A customer wants to migrate from 10.24 to 11.6 and leverage new capabilities. They need to know what patterns can be modernized and what new features become available.

Today, version knowledge is scattered across:
- `sdk/mpr/version/version.go` (6 hardcoded features)
- Comments in MDL example scripts (`-- @version: 11.0+`)
- CLAUDE.md documentation (prose, not machine-readable)
- Implicit knowledge in the BSON writers (version-conditional serialization)

There is no way for an agent or user to **query** what's available, no **pre-flight check** before writing, and no **upgrade advisor** for migration.

## Design Principles

1. **Single source of truth** -- One structured data format defines version capabilities. Everything else reads from it.
2. **Machine-readable first** -- An AI agent should be able to query capabilities, not read prose.
3. **Queryable at runtime** -- `SHOW FEATURES` returns capabilities for the connected project's version.
4. **Fail-fast with actionable messages** -- Error before writing BSON, not after Studio Pro crashes.
5. **Incremental updates** -- Adding a new version's capabilities should be a data change, not a code change.
6. **Reuse existing infrastructure** -- Linter rules, skills, executor commands follow established patterns.

## Architecture

```
                    Version Feature Registry
                    (sdk/versions/*.yaml)
                           |
            +--------------+--------------+
            |              |              |
       SHOW FEATURES   Executor       Linter
       (MDL command)   Pre-checks    Rules (VER0xx)
            |              |              |
            v              v              v
       Agent queries   Error before   Upgrade
       capabilities    BSON write     recommendations
```

### Layer 1: Version Feature Registry

A structured YAML file per major version, embedded via `go:embed`:

```
sdk/versions/
  mendix-9.yaml
  mendix-10.yaml
  mendix-11.yaml
```

Each file defines features, their introduction version, syntax, deprecations, and upgrade hints:

```yaml
# sdk/versions/mendix-10.yaml
major: 10
supported_range: "10.0..10.24"
lts_versions: ["10.24"]
mts_versions: ["10.6", "10.12", "10.18"]

features:
  # --- Domain Model ---
  domain_model:
    entities:
      introduced: "10.0"
      mdl: "CREATE PERSISTENT ENTITY Module.Name (...)"
    non_persistent_entities:
      introduced: "10.0"
      mdl: "CREATE NON-PERSISTENT ENTITY Module.Name (...)"
    view_entities:
      introduced: "10.18"
      mdl: "CREATE VIEW ENTITY Module.Name (...) AS SELECT ..."
      bson_notes: "10.x: OQL stored inline on OqlViewEntitySource; 11.0+: removed"
    calculated_attributes:
      introduced: "10.0"
      mdl: "CALCULATED BY Module.Microflow"
    entity_generalization:
      introduced: "10.0"
      mdl: "EXTENDS Module.ParentEntity"
    alter_entity:
      introduced: "10.0"
      mdl: "ALTER ENTITY Module.Name ADD ATTRIBUTE ..."

  # --- OQL Functions (used in VIEW ENTITY queries) ---
  oql_functions:
    basic_select:
      introduced: "10.18"
      mdl: "SELECT, FROM, WHERE, AS, AND, OR, NOT"
    aggregate_functions:
      introduced: "10.18"
      mdl: "COUNT, SUM, AVG, MIN, MAX, GROUP BY"
    subqueries:
      introduced: "10.18"
      mdl: "Inline subqueries in SELECT and WHERE"
    join_types:
      introduced: "10.18"
      mdl: "INNER JOIN, LEFT JOIN, RIGHT JOIN, FULL JOIN"
    # New OQL functions added per minor release:
    # string_length: { introduced: "10.20" }
    # date_diff: { introduced: "10.21" }
    # (to be populated from release notes)

  # --- Microflows ---
  microflows:
    basic:
      introduced: "10.0"
      mdl: "CREATE MICROFLOW Module.Name (...) BEGIN ... END"
    show_page_with_params:
      introduced: null
      available_in: "11.0+"
      workaround: "Pass data via a non-persistent entity or microflow parameter"
    send_rest_request:
      introduced: "10.1"
      mdl: "SEND REST REQUEST ..."
    send_rest_query_params:
      introduced: null
      available_in: "11.0+"
      notes: "Query parameters in REST requests"
    execute_database_query:
      introduced: "10.6"
      mdl: "EXECUTE DATABASE QUERY ..."
    execute_database_query_runtime_connection:
      introduced: null
      available_in: "11.0+"
      notes: "Runtime connection override for database queries"
    loop_in_branch:
      introduced: "10.0"
      notes: "LOOP inside IF/ELSE branches"

  # --- Pages ---
  pages:
    basic:
      introduced: "10.0"
      mdl: "CREATE PAGE Module.Name (...) { ... }"
    page_parameters:
      introduced: null
      available_in: "11.0+"
      workaround: "Use non-persistent entity or microflow parameter"
    page_variables:
      introduced: null
      available_in: "11.0+"
    pluggable_widgets:
      introduced: "10.0"
      widgets: [ComboBox, DataGrid2, Gallery, Image, TextFilter, NumberFilter, DateFilter, DropdownFilter]
      notes: "Widget templates are version-specific; MPK augmentation handles drift"
    design_properties_v3:
      introduced: null
      available_in: "11.0+"
      notes: "Atlas v3 design properties (Card style, Disable row wrap)"
    alter_page:
      introduced: "10.0"
      mdl: "ALTER PAGE Module.Name SET/INSERT/DROP/REPLACE ..."

  # --- Security ---
  security:
    module_roles:
      introduced: "10.0"
    user_roles:
      introduced: "10.0"
    entity_access:
      introduced: "10.0"
    demo_users:
      introduced: "10.0"

  # --- Integration ---
  integration:
    rest_client_basic:
      introduced: "10.1"
      mdl: "CREATE REST CLIENT Module.Name ..."
    rest_client_query_params:
      introduced: null
      available_in: "11.0+"
    rest_client_headers:
      introduced: "10.4"
    database_connector_basic:
      introduced: "10.6"
      mdl: "CREATE DATABASE CONNECTION Module.Name ..."
    database_connector_execute:
      introduced: "10.6"
      mdl: "EXECUTE DATABASE QUERY in microflows"
      bson_notes: "Full BSON format requires 11.0+"
    business_events:
      introduced: "10.0"
      mdl: "CREATE BUSINESS EVENT SERVICE Module.Name ..."
    odata_client:
      introduced: "10.0"

  # --- Workflows ---
  workflows:
    basic:
      introduced: "9.0"
      mdl: "CREATE WORKFLOW Module.Name ..."
    parallel_split:
      introduced: "9.0"
    user_task:
      introduced: "9.0"

  # --- Navigation ---
  navigation:
    profiles:
      introduced: "10.0"
    menu_items:
      introduced: "10.0"
    home_pages:
      introduced: "10.0"

deprecated:
  - id: "DEP001"
    pattern: "Persistable: false on view entities"
    replaced_by: "Persistable: true (auto-set)"
    since: "10.18"
    severity: "info"

upgrade_opportunities:
  from_10_to_11:
    - feature: "page_parameters"
      description: "Replace non-persistent entity parameter passing with direct page parameters"
      effort: "low"
    - feature: "design_properties_v3"
      description: "Atlas v3 design properties available for richer styling"
      effort: "low"
    - feature: "rest_client_query_params"
      description: "REST clients can now define query parameters directly"
      effort: "low"
    - feature: "database_connector_runtime_connection"
      description: "Database queries can override connection at runtime"
      effort: "low"
    - feature: "association_storage"
      description: "New association storage format (automatic on project upgrade)"
      effort: "none"
```

### Layer 2: MDL Commands

#### `SHOW FEATURES`

Lists all features available for the connected project's Mendix version, or for a specified version without a project connection:

```sql
-- When connected to a project (uses project's Mendix version)
SHOW FEATURES;

-- Without a project connection (query any version)
SHOW FEATURES FOR VERSION 10.24;

-- Filter by area
SHOW FEATURES IN integration;

-- Show only features available in the project's version
SHOW FEATURES WHERE available = true;
```

Output:
```
| Feature                | Available | Since  | Notes                                    |
|------------------------|-----------|--------|------------------------------------------|
| Persistent entities    | Yes       | 10.0   |                                          |
| View entities          | Yes       | 10.18  | OQL stored inline on source object       |
| Page parameters        | No        | 11.0+  | Use non-persistent entity workaround     |
| Pluggable widgets      | Yes       | 10.0   | ComboBox, DataGrid2, Gallery, Image      |
| Design properties v3   | No        | 11.0+  | Atlas v3 required                        |
| REST client            | Partial   | 10.1   | Query parameters require 11.0+           |
| Database connector     | Partial   | 10.6   | EXECUTE DATABASE QUERY requires 11.0+    |
| Business events        | Yes       | 10.0   |                                          |
| Workflows              | Yes       | 9.0    |                                          |
```

#### `SHOW FEATURES ADDED SINCE <version>`

Shows what becomes available when upgrading:

```sql
SHOW FEATURES ADDED SINCE 10.24;
```

Output:
```
| Feature              | Available In | Description                              | Effort |
|----------------------|-------------|------------------------------------------|--------|
| Page parameters      | 11.0        | Direct page parameter passing            | Low    |
| Design properties v3 | 11.0        | Atlas v3 Card style, Disable row wrap    | Low    |
| REST query params    | 11.0        | Query parameter support in REST clients  | Low    |
| Portable app format  | 11.6        | New deployment format                    | None   |
```

#### `SHOW UPGRADE OPPORTUNITIES`

When connected, analyzes the current project for patterns that could benefit from the project's own version or a target version:

```sql
-- What can I improve using my current version's capabilities?
SHOW UPGRADE OPPORTUNITIES;

-- What would upgrading to 11.6 enable?
SHOW UPGRADE OPPORTUNITIES TO 11.6;
```

Output:
```
| Opportunity                | Target  | Description                                          | Effort |
|----------------------------|---------|------------------------------------------------------|--------|
| Page parameters            | 11.0    | 12 pages use NP entity workaround for parameter passing | Low  |
| Design properties v3       | 11.0    | 8 containers could use Card style / row wrap         | Low    |
| REST query parameters      | 11.0    | 3 REST clients build query strings manually          | Low    |
| Database runtime connection| 11.0    | 2 microflows hardcode DB connection                  | Low    |
```

#### `SHOW DEPRECATED`

Lists deprecated patterns in the current project:

```sql
SHOW DEPRECATED;
```

### Layer 3: Executor Pre-Checks

Before writing BSON, the executor checks version compatibility and produces actionable errors:

```go
// In cmd_entities.go, before creating a view entity:
if s.IsViewEntity {
    pv := e.reader.ProjectVersion()
    if !pv.IsAtLeast(10, 18) {
        return fmt.Errorf(
            "CREATE VIEW ENTITY requires Mendix 10.18+ (project is %s)\n"+
            "  hint: upgrade your project or use a regular entity with a microflow data source",
            pv.ProductVersion,
        )
    }
}
```

This pattern already exists informally in the codebase (version-conditional BSON writing). The proposal formalizes it with:

1. A `CheckFeature(feature, version)` function that returns a user-friendly error
2. Pre-checks at the start of each executor command
3. Consistent error format with hints

### Layer 4: Linter Rules (VER prefix)

New linter rule category `VER` for version-related checks:

| Rule | Name | Description |
|------|------|-------------|
| VER001 | UnsupportedFeature | Feature used that's not available in project version |
| VER002 | DeprecatedPattern | Deprecated pattern that has a modern replacement |
| VER003 | UpgradeOpportunity | Pattern that can be simplified on a newer version |

**VER001** runs during `mxcli check` and `mxcli lint`:
```
[VER001] CREATE VIEW ENTITY requires Mendix 10.18+ (project is 10.12.0)
  at line 42 in script.mdl
  hint: upgrade to 10.18+ or use a microflow data source
```

**VER003** runs during `mxcli lint --upgrade-hints`:
```
[VER003] Page MyModule.EditCustomer uses non-persistent entity for parameter passing
  This pattern can be replaced with page parameters in Mendix 11.0+
  effort: low
```

### Layer 5: Skills (AI Agent Guidance)

One skill file: `.claude/skills/version-awareness.md`

```markdown
# Version Awareness

## Before Generating MDL

Always check the project's Mendix version before writing MDL:

    SHOW STATUS;           -- shows connected project version
    SHOW FEATURES;         -- shows available features

## Version-Conditional Patterns

If a feature is not available, use the documented workaround:

    SHOW FEATURES WHERE name = 'page_parameters';
    -- If not available, use non-persistent entity pattern instead

## Upgrade Workflow

When migrating to a newer version:

    SHOW FEATURES ADDED SINCE 10.24;    -- what's new
    SHOW DEPRECATED;                     -- what to update
    mxcli lint --upgrade-hints -p app.mpr  -- automated suggestions
```

This skill is small and stable -- it teaches the agent to **query** mxcli rather than embedding version knowledge in the skill itself. The version data lives in the registry.

### Layer 6: Keeping Data Current

The version registry needs updates when Mendix releases new versions. Proposed pipeline:

1. **Automated**: `mxcli diff-schemas 11.5 11.6` compares reflection data between versions, outputs added/removed types and properties as a diff report.

2. **Semi-automated**: An agent reads the diff report + Mendix release notes and proposes updates to the YAML registry. Human reviews and merges.

3. **On-demand**: `mxcli update-features` downloads the latest registry from a central source (GitHub release asset), similar to how `mxcli setup mxbuild` downloads tooling.

4. **Community**: The `-- @version:` directives in MDL test scripts serve as executable documentation. If a test fails on a version, the directive gets updated — and that update feeds back into the registry.

## Implementation Plan

### Phase 1: Version Feature Registry + SHOW FEATURES (foundation)

1. Create `sdk/versions/` package with YAML loader and `go:embed`
2. Create YAML files for Mendix 9, 10, 11 (initial feature set from existing knowledge)
3. Implement `SHOW FEATURES` command in executor
4. Implement `SHOW FEATURES ADDED SINCE <version>` variant
5. Wire into AST/grammar: add `FEATURES` keyword to MDLParser.g4

**Deliverable**: Agent can query `SHOW FEATURES` and get machine-readable output.

### Phase 2: Executor Pre-Checks (fail-fast)

1. Add `CheckFeatureAvailable(feature string)` method to Executor
2. Add version checks to CREATE VIEW ENTITY, CREATE REST CLIENT, CREATE PAGE (with Params), EXECUTE DATABASE QUERY
3. Produce error messages with version requirement, current version, and workaround hint
4. Test: run MDL scripts with version-gated features on older projects

**Deliverable**: Unsupported features fail immediately with actionable error instead of corrupting BSON.

### Phase 3: Linter Rules (VER category)

1. Implement VER001 (UnsupportedFeature) -- reads from version registry
2. Implement VER002 (DeprecatedPattern) -- reads deprecated list from registry
3. Wire into `mxcli lint` and `mxcli check`
4. Add SARIF output support for CI integration

**Deliverable**: `mxcli lint -p app.mpr` reports version issues.

### Phase 4: Upgrade Advisor

1. Implement VER003 (UpgradeOpportunity) linter rule
2. Implement `SHOW DEPRECATED` command
3. Implement `SHOW FEATURES ADDED SINCE` with effort estimates
4. Implement `mxcli lint --upgrade-hints --target-version 11.6`

**Deliverable**: Migration planning from any version to any newer version.

### Phase 5: Skills + Agent Integration

1. Create `.claude/skills/version-awareness.md`
2. Update `.claude/skills/check-syntax.md` to include version pre-check
3. Update `mxcli init` to include version-awareness skill in project setup
4. Test: AI agent generates valid MDL for both 10.24 and 11.6 projects

**Deliverable**: AI agents automatically adapt to project version.

### Phase 6: Automated Registry Updates

1. Implement `mxcli diff-schemas <from> <to>` using reflection data
2. Create agent workflow: diff-schemas output + release notes -> YAML update PR
3. Implement `mxcli update-features` for on-demand downloads
4. Add to nightly CI: verify registry matches reflection data

**Deliverable**: Registry stays current with minimal manual effort.

## Relationship to BSON Schema Registry Proposal

This proposal complements `BSON_SCHEMA_REGISTRY_PROPOSAL.md`:

- **Schema Registry** handles **structural** version differences (field names, defaults, encoding) at the BSON level
- **This proposal** handles **feature-level** version differences (what MDL commands are available) at the user/agent level

The version feature registry (YAML) is simpler and more immediately useful than the full schema registry. It can be built first and later integrated with the schema registry as that matures.

```
User/Agent Layer     This Proposal        "What can I do?"
                          |
MDL Layer            Executor pre-checks  "Will this work?"
                          |
BSON Layer           Schema Registry      "How do I serialize this?"
```

## Design Decisions

1. **YAML as source of truth.** Human-readable, easy to edit, reviewed in PRs. Compiled to embedded Go structs at build time via `go:embed` + YAML parser.

2. **Fine-grained features.** Mendix adds capabilities per minor release across all areas (new OQL functions, database connector features, REST enhancements, page properties). The registry must track at the **per-capability** level, not just per-concept. Example: `oql_functions.string_length` introduced in 10.8, `oql_functions.date_diff` in 10.12, `rest_client.query_parameters` in 11.0. Grouping by area with individual capability entries:

    ```yaml
    oql_functions:
      string_length: { introduced: "10.8" }
      date_diff: { introduced: "10.12" }
      coalesce: { introduced: "10.14" }
    database_connector:
      basic_query: { introduced: "10.6" }
      parameterized_query: { introduced: "10.12" }
      runtime_connection: { introduced: "11.0" }
    ```

3. **SHOW FEATURES works without a project.** `SHOW FEATURES FOR VERSION 10.24` queries the registry directly without an MPR connection. When connected, it defaults to the project's version. Useful for upgrade planning: `SHOW FEATURES ADDED SINCE 10.24` shows what upgrading to the latest would gain; `SHOW FEATURES FOR VERSION 11.6` shows the full capability set of a target version. When connected, also supports: `SHOW UPGRADE OPPORTUNITIES` to identify patterns in the current project that could benefit from the project's own version capabilities (e.g., detecting workarounds that are no longer needed).

4. **Minor version granularity by default.** Metamodel changes happen at the minor release level. Patch-level overrides supported where needed but expected to be rare.

5. **Non-interactive upgrade advisor for now.** `mxcli lint --upgrade-hints --target-version 11.6` outputs a report. Interactive `mxcli upgrade` can be added later as the tooling matures.
