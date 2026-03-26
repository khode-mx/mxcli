# SQL (query)

## Synopsis

    SQL alias query_text

    SQL alias SHOW TABLES
    SQL alias SHOW VIEWS
    SQL alias SHOW FUNCTIONS
    SQL alias DESCRIBE table_name

## Description

Executes a SQL query or schema discovery command against an external database connection. The alias identifies which connection to use (established with `SQL CONNECT`).

**Raw SQL passthrough**: Any SQL text after the alias is sent directly to the external database. The results are displayed as a formatted table. This supports SELECT, INSERT, UPDATE, DELETE, DDL, and any other SQL the target database accepts.

**Schema discovery commands** provide a portable way to explore the database schema without writing database-specific SQL:

| Command | Description |
|---------|-------------|
| `SHOW TABLES` | Lists user tables in the database |
| `SHOW VIEWS` | Lists user views in the database |
| `SHOW FUNCTIONS` | Lists functions and stored procedures |
| `DESCRIBE table_name` | Shows columns, data types, and nullability for a table |

The schema discovery commands query `information_schema` internally and work consistently across PostgreSQL, Oracle, and SQL Server.

## Parameters

**alias**
: The connection alias established with `SQL CONNECT`.

**query_text**
: Any valid SQL for the target database. Sent as-is to the database engine.

**table_name** (DESCRIBE only)
: The name of the table or view to describe.

## Examples

### Query data

```sql
SQL source SELECT * FROM users WHERE active = true LIMIT 10;
```

### List tables

```sql
SQL source SHOW TABLES;
```

### List views

```sql
SQL source SHOW VIEWS;
```

### Describe a table

```sql
SQL source DESCRIBE users;
```

### Insert data

```sql
SQL source INSERT INTO audit_log (action, timestamp) VALUES ('export', NOW());
```

### Run aggregate queries

```sql
SQL source SELECT department, COUNT(*) AS headcount
  FROM employees
  GROUP BY department
  ORDER BY headcount DESC;
```

### From the command line

```sql
-- Shell command:
-- mxcli sql --driver postgres --dsn 'postgres://...' "SELECT * FROM users"
```

## See Also

[SQL CONNECT](sql-connect.md), [SQL DISCONNECT](sql-disconnect.md), [SQL GENERATE CONNECTOR](sql-generate-connector.md), [IMPORT FROM](import-from.md)
