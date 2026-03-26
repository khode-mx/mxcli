# SQL DISCONNECT

## Synopsis

    SQL DISCONNECT alias

## Description

Closes an external database connection identified by the given alias. After disconnecting, the alias is no longer valid and any subsequent `SQL alias ...` commands will fail.

Use `SQL CONNECTIONS` to see which connections are currently active before disconnecting.

## Parameters

**alias**
: The alias of the connection to close. Must match an active connection established with `SQL CONNECT`.

## Examples

### Disconnect a connection

```sql
SQL DISCONNECT source;
```

### Full connection lifecycle

```sql
SQL CONNECT postgres 'postgres://user:pass@localhost:5432/mydb' AS source;
SQL source SHOW TABLES;
SQL source SELECT COUNT(*) FROM users;
SQL DISCONNECT source;
```

## See Also

[SQL CONNECT](sql-connect.md), [SQL (query)](sql-query.md)
