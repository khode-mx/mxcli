# SQL GENERATE CONNECTOR

## Synopsis

    SQL alias GENERATE CONNECTOR INTO module [ TABLES ( table [, ...] ) ] [ VIEWS ( view [, ...] ) ] [ EXEC ]

## Description

Generates MDL statements that create a Mendix Database Connector module from the schema of an external database. The generated MDL includes entities, attributes, and associations that mirror the external database tables and views, configured for use with the Mendix Database Connector module.

Without `EXEC`, the command outputs the generated MDL to the console for review. With `EXEC`, it executes the MDL immediately against the open project.

By default, all user tables are included. Use `TABLES (...)` and `VIEWS (...)` to limit generation to specific tables and views.

The type mapping from SQL types to Mendix types is handled automatically (e.g., `VARCHAR` becomes `String`, `INTEGER` becomes `Integer`, `TIMESTAMP` becomes `DateTime`).

## Parameters

**alias**
: The connection alias established with `SQL CONNECT`.

**module**
: The target Mendix module where the connector entities will be created.

**table**
: One or more table names to include. If omitted, all user tables are included.

**view**
: One or more view names to include. If omitted, no views are included unless `TABLES` is also omitted (in which case all tables are included by default).

**EXEC**
: Execute the generated MDL immediately instead of printing it. Requires an open Mendix project.

## Examples

### Preview generated MDL for all tables

```sql
SQL source GENERATE CONNECTOR INTO HRModule;
```

### Generate for specific tables

```sql
SQL source GENERATE CONNECTOR INTO HRModule TABLES (employees, departments);
```

### Generate for tables and views

```sql
SQL source GENERATE CONNECTOR INTO HRModule TABLES (employees) VIEWS (employee_summary);
```

### Generate and execute immediately

```sql
SQL source GENERATE CONNECTOR INTO HRModule TABLES (employees, departments) EXEC;
```

### Full workflow

```sql
SQL CONNECT postgres 'postgres://user:pass@localhost:5432/hrdb' AS source;
SQL source SHOW TABLES;
SQL source DESCRIBE employees;
SQL source GENERATE CONNECTOR INTO HRModule TABLES (employees, departments) EXEC;
SQL DISCONNECT source;
```

## See Also

[SQL CONNECT](sql-connect.md), [SQL (query)](sql-query.md), [IMPORT FROM](import-from.md)
