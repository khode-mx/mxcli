# SQL CONNECT

## Synopsis

    SQL CONNECT driver 'dsn' AS alias

## Description

Opens a connection to an external database. The connection is identified by the alias, which is used in all subsequent SQL commands to route queries to the correct database.

Multiple connections can be open simultaneously with different aliases. Use `SQL CONNECTIONS` to list active connections (shows alias and driver only -- the DSN is never displayed for security).

Supported database drivers:

| Driver | Aliases | Example DSN |
|--------|---------|-------------|
| `postgres` | `pg`, `postgresql` | `postgres://user:pass@host:5432/dbname` |
| `oracle` | `ora` | `oracle://user:pass@host:1521/service` |
| `sqlserver` | `mssql` | `sqlserver://user:pass@host:1433?database=dbname` |

## Parameters

**driver**
: The database driver to use. One of `postgres` (or `pg`, `postgresql`), `oracle` (or `ora`), or `sqlserver` (or `mssql`).

**dsn**
: The data source name (connection string) enclosed in single quotes. The format depends on the driver. The DSN is stored in memory only and never appears in session output or logs.

**alias**
: A short identifier for this connection. Used in all subsequent `SQL alias ...` commands. Must be unique among active connections.

## Examples

### Connect to PostgreSQL

```sql
SQL CONNECT postgres 'postgres://user:pass@localhost:5432/mydb' AS source;
```

### Connect to Oracle

```sql
SQL CONNECT oracle 'oracle://scott:tiger@dbhost:1521/ORCL' AS erp;
```

### Connect to SQL Server

```sql
SQL CONNECT sqlserver 'sqlserver://sa:Password1@localhost:1433?database=northwind' AS legacy;
```

### List active connections

```sql
SQL CONNECTIONS;
```

### Connect using driver aliases

```sql
SQL CONNECT pg 'postgres://user:pass@localhost:5432/mydb' AS src;
SQL CONNECT mssql 'sqlserver://sa:Pass@host:1433?database=db' AS dst;
```

## See Also

[SQL DISCONNECT](sql-disconnect.md), [SQL (query)](sql-query.md), [SQL GENERATE CONNECTOR](sql-generate-connector.md)
