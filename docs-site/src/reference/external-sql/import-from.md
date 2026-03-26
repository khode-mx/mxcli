# IMPORT FROM

## Synopsis

    IMPORT FROM alias QUERY 'sql'
      INTO module.Entity
      MAP ( source_col AS AttrName [, ...] )
      [ LINK ( source_col TO AssocName ON MatchAttr ) [, ...] ]
      [ BATCH size ]
      [ LIMIT count ]

## Description

Imports data from an external database into a Mendix application's PostgreSQL database. The command executes the specified SQL query against the external connection, maps the result columns to entity attributes, and inserts rows in batches using direct database insertion.

The import pipeline handles:

- **Mendix ID generation**: Automatically generates unique Mendix object IDs for each inserted row, following the Mendix ID format and sequence tracking.
- **Batch insertion**: Rows are inserted in configurable batch sizes for performance.
- **Association linking**: The `LINK` clause matches imported rows to existing objects via an association, looking up the target by a matching attribute value.

The target Mendix application database is auto-detected from the project's database configuration. The import connects directly to the application's PostgreSQL database.

## Parameters

**alias**
: The external database connection alias established with `SQL CONNECT`.

**sql**
: A SQL query string (in single quotes) to execute against the external database. The result set columns are available for mapping.

**module.Entity**
: The fully qualified Mendix entity to import into (e.g., `HR.Employee`).

**source_col**
: A column name from the SQL query result set.

**AttrName**
: The target attribute name on the Mendix entity.

**AssocName** (LINK only)
: The name of the association to set. The source column value is matched against the `MatchAttr` on the associated entity to find the target object.

**MatchAttr** (LINK only)
: The attribute on the associated entity used to look up the target object for linking.

**size** (BATCH only)
: Number of rows per insert batch. Defaults to a reasonable batch size if not specified.

**count** (LIMIT only)
: Maximum number of rows to import from the query result.

## Examples

### Basic import

```sql
IMPORT FROM source QUERY 'SELECT name, email FROM employees'
  INTO HR.Employee
  MAP (name AS Name, email AS Email);
```

### Import with association linking

```sql
IMPORT FROM source QUERY 'SELECT name, email, dept_name FROM employees'
  INTO HR.Employee
  MAP (name AS Name, email AS Email)
  LINK (dept_name TO Employee_Department ON Name);
```

### Import with batch size and limit

```sql
IMPORT FROM source QUERY 'SELECT name, email, dept_name FROM employees'
  INTO HR.Employee
  MAP (name AS Name, email AS Email)
  LINK (dept_name TO Employee_Department ON Name)
  BATCH 500
  LIMIT 1000;
```

### Full workflow: connect, explore, import

```sql
SQL CONNECT postgres 'postgres://user:pass@legacydb:5432/hr' AS source;
SQL source SHOW TABLES;
SQL source DESCRIBE employees;

IMPORT FROM source QUERY 'SELECT name, email FROM employees WHERE active = true'
  INTO HR.Employee
  MAP (name AS Name, email AS Email)
  BATCH 500;

SQL DISCONNECT source;
```

## See Also

[SQL CONNECT](sql-connect.md), [SQL (query)](sql-query.md), [SQL GENERATE CONNECTOR](sql-generate-connector.md)
