# Proposal: Generate Database Connector from External Schema

**Status**: Draft
**Date**: 2026-03-05
**Branch**: `sql`
**Depends on**: Phase 1-3 (SQL CONNECT, schema discovery, IMPORT) — all implemented

## Motivation

When using Mendix as a **frontend for existing applications**, developers need to:

1. Connect to the existing database
2. Create non-persistent entities matching the database tables/views
3. Set up constants for JDBC connection credentials
4. Create a Database Connection document with query definitions
5. Build pages to display the data

Steps 2–4 are tedious and mechanical: discover columns, map SQL types to Mendix types, generate boilerplate MDL. With `SQL CONNECT` and `DESCRIBE` already implemented, mxcli can automate the entire process with a single command.

## Use Case

```
-- Connect to the external database
SQL CONNECT postgres 'postgres://user:pass@host:5432/legacydb' AS source;

-- Discover the schema
SQL source SHOW TABLES;
SQL source DESCRIBE employees;

-- Generate everything: constants + non-persistent entities + database connection + queries
SQL source GENERATE CONNECTOR INTO HRModule;

-- Or generate for specific tables only
SQL source GENERATE CONNECTOR INTO HRModule TABLES (employees, departments);

-- Or generate for views
SQL source GENERATE CONNECTOR INTO HRModule VIEWS (active_employees_v);
```

Output: a complete MDL script that can be piped to `mxcli exec`, printed for review, or executed directly.

## Generated MDL

For a PostgreSQL database with table `employees(id serial, name varchar(100), email varchar(200), salary numeric(10,2), active boolean, hired_at timestamp)`, the command generates:

```sql
-- Constants (names must follow Database Connector convention: {ConnectionName}_DB{Suffix})
CREATE CONSTANT HRModule.SourceDatabase_DBSource TYPE String
  DEFAULT 'jdbc:postgresql://host:5432/legacydb'
  COMMENT 'JDBC connection string for SourceDatabase';

CREATE CONSTANT HRModule.SourceDatabase_DBUsername TYPE String
  DEFAULT 'user'
  COMMENT 'Database username for SourceDatabase';

CREATE CONSTANT HRModule.SourceDatabase_DBPassword TYPE String
  DEFAULT ''
  COMMENT 'Database password for SourceDatabase';

-- Non-persistent entity for employees table
CREATE NON-PERSISTENT ENTITY HRModule.Employee (
  EmployeeId: Integer,
  Name: String(100),
  Email: String(200),
  Salary: Decimal,
  Active: Boolean,
  HiredAt: DateTime
);

-- Database Connection with queries
CREATE DATABASE CONNECTION HRModule.SourceDatabase
TYPE 'PostgreSQL'
CONNECTION STRING @HRModule.SourceDatabase_DBSource
USERNAME @HRModule.SourceDatabase_DBUsername
PASSWORD @HRModule.SourceDatabase_DBPassword
BEGIN
  QUERY GetAllEmployees
    SQL 'SELECT id, name, email, salary, active, hired_at FROM employees'
    RETURNS HRModule.Employee
    MAP (
      id AS EmployeeId,
      name AS Name,
      email AS Email,
      salary AS Salary,
      active AS Active,
      hired_at AS HiredAt
    );
END;
```

## Design Decisions

### SQL Type → Mendix Type Mapping

| SQL Type (PostgreSQL/Oracle/MSSQL) | Mendix Type | Notes |
|---|---|---|
| `integer`, `int`, `int4`, `NUMBER(n,0)` n≤10 | `Integer` | |
| `bigint`, `int8`, `NUMBER(n,0)` n>10 | `Long` | |
| `smallint`, `int2`, `tinyint` | `Integer` | |
| `serial`, `bigserial`, `identity` | `AutoNumber` | Primary key columns only; `Integer` otherwise |
| `numeric`, `decimal`, `NUMBER(p,s)` s>0, `money` | `Decimal` | |
| `real`, `float4`, `float`, `double`, `BINARY_FLOAT` | `Decimal` | Mendix has no float type |
| `varchar(n)`, `character varying(n)`, `VARCHAR2(n)`, `nvarchar(n)` | `String(n)` | Length preserved |
| `text`, `clob`, `ntext`, `nvarchar(max)` | `String(unlimited)` | No length limit |
| `char(n)`, `nchar(n)` | `String(n)` | |
| `boolean`, `bit` (MSSQL) | `Boolean` | |
| `date` | `DateTime` | Mendix has no date-only type |
| `timestamp`, `datetime`, `datetime2`, `DATE` (Oracle) | `DateTime` | |
| `timestamp with time zone`, `datetimeoffset` | `DateTime` | |
| `bytea`, `blob`, `varbinary`, `RAW` | `— skipped —` | Binary not mappable to Mendix type |
| `uuid`, `uniqueidentifier` | `String(36)` | |
| `json`, `jsonb`, `xml` | `String(unlimited)` | |

Unmappable types are skipped with a warning comment in the output.

### Naming Conventions

- **Entity name**: Table name → PascalCase singular. `employees` → `Employee`, `order_items` → `OrderItem`
- **Attribute name**: Column name → PascalCase. `first_name` → `FirstName`, `employee_id` → `EmployeeId`
- **Constants**: `{Module}.{ConnectionName}_DBSource`, `{Module}.{ConnectionName}_DBUsername`, `{Module}.{ConnectionName}_DBPassword` (matches Mendix Database Connector convention)
- **Connection name**: `{Module}.{Alias}Database` (e.g. alias `f1` → `F1Database`)
- **Query name**: `GetAll{EntityName}` (one per table/view)

### DSN → JDBC URL Conversion

The mxcli connection uses Go driver DSNs (`postgres://...`), but the Mendix Database Connector needs JDBC URLs. Convert automatically:

| Go DSN | JDBC URL |
|--------|----------|
| `postgres://user:pass@host:5432/db` | `jdbc:postgresql://host:5432/db` |
| `oracle://user:pass@host:1521/service` | `jdbc:oracle:thin:@//host:1521/service` |
| `sqlserver://user:pass@host:1433?database=db` | `jdbc:sqlserver://host:1433;databaseName=db` |

### Output Modes

| Flag | Behavior |
|------|----------|
| (default) | Print generated MDL to stdout |
| `--exec` / `EXEC` | Execute the generated MDL against the open project |
| `--file <path>` | Write MDL to file |

In REPL mode (MDL syntax), use suffix keyword:

```sql
SQL source GENERATE CONNECTOR INTO HRModule;           -- print to stdout
SQL source GENERATE CONNECTOR INTO HRModule EXEC;      -- execute immediately
```

## MDL Syntax

```
SQL <alias> GENERATE CONNECTOR INTO <module>
  [TABLES (<table1>, <table2>, ...)]
  [VIEWS (<view1>, <view2>, ...)]
  [EXEC];
```

- Without `TABLES`/`VIEWS`: generates for **all** user tables (excluding system tables)
- With `TABLES`: only the listed tables
- With `VIEWS`: only the listed views
- Can combine `TABLES` and `VIEWS`
- `EXEC`: execute the generated MDL immediately (requires open project)

## Implementation Plan

### Step 1: SQL Type Mapping (`sql/typemap.go` — new)

```go
type MendixType struct {
    TypeName string // "String", "Integer", "Decimal", "Boolean", "DateTime", "Long", "AutoNumber"
    Length   int    // for String(n); 0 = unlimited
}

// MapSQLType maps a SQL data_type string to a Mendix type.
func MapSQLType(driver DriverName, sqlType string, length int, isPK bool) *MendixType

// GoDriverDSNToJDBC converts a Go driver DSN to a JDBC URL.
func GoDriverDSNToJDBC(driver DriverName, dsn string) (string, error)
```

### Step 2: Name Conventions (`sql/naming.go` — new)

```go
// TableToEntityName converts "order_items" → "OrderItem" (PascalCase, singular)
func TableToEntityName(tableName string) string

// ColumnToAttributeName converts "first_name" → "FirstName" (PascalCase)
func ColumnToAttributeName(colName string) string

// TableToQueryName returns "GetAll{EntityName}" for a table
func TableToQueryName(tableName string) string
```

Singularization: simple rules only (strip trailing `s`, `es`, `ies` → `y`). No NLP dependency.

### Step 3: Schema Reader (`sql/schema.go` — new)

```go
type TableSchema struct {
    Schema  string
    Name    string
    Columns []ColumnSchema
    IsView  bool
}

type ColumnSchema struct {
    Name       string
    DataType   string
    Nullable   bool
    Length     int    // from character_maximum_length
    IsPK       bool
}

// ReadTableSchema reads column metadata with primary key detection.
func ReadTableSchema(ctx context.Context, conn *Connection, tableName string) (*TableSchema, error)

// ReadAllTableSchemas reads schemas for all user tables.
func ReadAllTableSchemas(ctx context.Context, conn *Connection) ([]*TableSchema, error)
```

This extends `meta.go`'s `DescribeTable` with PK detection and parsed length values (the current `DescribeTable` returns raw `information_schema` rows).

### Step 4: MDL Generator (`sql/generate.go` — new)

```go
type GenerateConfig struct {
    Conn       *Connection
    Module     string        // target Mendix module
    Alias      string        // connection alias (for naming)
    Tables     []string      // nil = all tables
    Views      []string      // nil = no views; empty = all views
    DSN        string        // original DSN for JDBC conversion
}

type GenerateResult struct {
    MDL           string       // complete MDL script
    TablesCount   int
    ViewsCount    int
    SkippedCols   []string     // columns with unmappable types
}

// GenerateConnector discovers schema and generates complete MDL.
func GenerateConnector(ctx context.Context, cfg *GenerateConfig) (*GenerateResult, error)
```

### Step 5: Grammar — add to `sqlPassthroughStatement`

The `SQL <alias> GENERATE CONNECTOR INTO <module>` can be parsed within the existing `sqlPassthroughStatement` rule (the freeform SQL portion). The executor detects `GENERATE CONNECTOR` as the query prefix and dispatches to the generator instead of executing SQL.

Alternatively, add a dedicated grammar rule:

```antlr
sqlGenerateStatement
    : SQL identifierOrKeyword GENERATE CONNECTOR INTO identifierOrKeyword
      (TABLES LPAREN identifierOrKeyword (COMMA identifierOrKeyword)* RPAREN)?
      (VIEWS LPAREN identifierOrKeyword (COMMA identifierOrKeyword)* RPAREN)?
      (EXEC)?
    ;
```

Add `GENERATE` and `CONNECTOR` tokens to `MDLLexer.g4`.

**Recommendation**: Dedicated grammar rule — gives better error messages and cleaner AST.

### Step 6: AST Node (`mdl/ast/ast_sql.go`)

```go
type SQLGenerateConnectorStmt struct {
    Alias      string
    Module     string
    Tables     []string // nil = all
    Views      []string // nil = none
    Exec       bool     // execute immediately
}
func (s *SQLGenerateConnectorStmt) isStatement() {}
```

### Step 7: Visitor (`mdl/visitor/visitor_sql.go`)

Add `ExitSqlGenerateStatement` to extract alias, module, table/view lists, exec flag.

### Step 8: Executor (`mdl/executor/cmd_sql.go`)

Add `execSQLGenerateConnector(s *ast.SQLGenerateConnectorStmt)`:
1. Get connection from manager
2. Call `sqllib.GenerateConnector(ctx, cfg)`
3. If `Exec` and project is open: parse + execute the generated MDL
4. Otherwise: print to stdout

### Step 9: CLI Flag (`cmd/mxcli/cmd_sql.go`)

Add `--generate` flag to `mxcli sql` subcommand:
```bash
mxcli sql --alias source --generate --module HRModule
mxcli sql --alias source --generate --module HRModule --tables employees,departments
```

### Step 10: Update DATABASE CONNECTION Grammar

The current grammar only supports `HOST`/`PORT`/`DATABASE` options. Update to also support:
- `CONNECTION STRING` (string literal or `@Module.Constant` reference)
- `BEGIN ... END` block with `QUERY` definitions
- `QUERY` with `SQL`, `PARAMETER`, `RETURNS`, `MAP` clauses

This is needed so the generated MDL can be round-tripped (parsed back and executed).

```antlr
createDatabaseConnectionStatement
    : DATABASE CONNECTION qualifiedName
      databaseConnectionOptions
      (BEGIN databaseQuery* END)?
    ;

databaseConnectionOption
    : TYPE STRING_LITERAL
    | CONNECTION STRING (STRING_LITERAL | AT qualifiedName)
    | HOST STRING_LITERAL
    | PORT NUMBER_LITERAL
    | DATABASE STRING_LITERAL
    | USERNAME (STRING_LITERAL | AT qualifiedName)
    | PASSWORD (STRING_LITERAL | AT qualifiedName)
    ;

databaseQuery
    : QUERY IDENTIFIER
      SQL (STRING_LITERAL | DOLLAR_STRING)
      (PARAMETER IDENTIFIER COLON dataType)*
      RETURNS qualifiedName
      (MAP LPAREN queryColumnMapping (COMMA queryColumnMapping)* RPAREN)?
      SEMICOLON
    ;

queryColumnMapping
    : identifierOrKeyword AS identifierOrKeyword
    ;
```

## Files Summary

| Action | File | Description |
|--------|------|-------------|
| CREATE | `sql/typemap.go` | SQL → Mendix type mapping, DSN → JDBC conversion |
| CREATE | `sql/typemap_test.go` | Type mapping tests |
| CREATE | `sql/naming.go` | Table/column → entity/attribute naming conventions |
| CREATE | `sql/naming_test.go` | Naming convention tests |
| CREATE | `sql/schema.go` | Structured schema reader (extends meta.go) |
| CREATE | `sql/generate.go` | MDL generation from schema |
| CREATE | `sql/generate_test.go` | Generator tests |
| MODIFY | `mdl/grammar/MDLLexer.g4` | Add `GENERATE`, `CONNECTOR` tokens |
| MODIFY | `mdl/grammar/MDLParser.g4` | Add `sqlGenerateStatement` rule; extend `createDatabaseConnectionStatement` with `CONNECTION STRING`, `BEGIN/END`, `QUERY` |
| MODIFY | `mdl/ast/ast_sql.go` | Add `SQLGenerateConnectorStmt` |
| MODIFY | `mdl/visitor/visitor_sql.go` | Parse generate statement |
| MODIFY | `mdl/executor/cmd_sql.go` | Execute generate (print or exec) |
| MODIFY | `mdl/executor/executor.go` | Add dispatch case |
| MODIFY | `mdl/executor/stmt_summary.go` | Add summary case |
| MODIFY | `cmd/mxcli/cmd_sql.go` | Add `--generate` and `--module` flags |
| MODIFY | `cmd/mxcli/help_topics/sql.txt` | Document GENERATE CONNECTOR |
| MODIFY | `docs/01-project/MDL_QUICK_REFERENCE.md` | Add syntax row |

## Key Design Notes

1. **No new dependencies** — all type mapping and naming logic is pure Go string manipulation
2. **Roundtrip-safe** — generated MDL can be parsed back by extending the DATABASE CONNECTION grammar
3. **Incremental** — running GENERATE again for the same module is safe (CREATE statements use create-or-update semantics in the executor)
4. **Constants use constant references** (`@Module.Constant`) — credentials are not embedded in the connection definition
5. **One query per table/view** — generates `GetAll{Entity}` as the baseline; users add parameterized queries manually

## Future Extensions (Not in This Phase)

- **Parameterized queries**: Generate `GetByPK` queries using primary key columns as parameters
- **Pages**: Generate overview pages (DataGrid2) for each entity after generating the connector
- **Sync mode**: Generate persistent entities + scheduled microflow for periodic data sync instead of non-persistent entities + Database Connector
- **OData mode**: Generate an OData service exposing the external data via External Entities
- **Diff/update**: Compare existing connector against current schema and generate ALTER statements
