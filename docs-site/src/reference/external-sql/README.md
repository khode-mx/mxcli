# External SQL Statements

Statements for connecting to and querying external databases (PostgreSQL, Oracle, SQL Server). Credentials are isolated from session output and logs -- the DSN is never displayed after the initial connection.

External SQL connections are independent of the Mendix project connection. You can have multiple external database connections open simultaneously, each identified by an alias.

| Statement | Description |
|-----------|-------------|
| [SQL CONNECT](sql-connect.md) | Open a connection to an external database |
| [SQL DISCONNECT](sql-disconnect.md) | Close an external database connection |
| [SQL (query)](sql-query.md) | Execute SQL queries and schema discovery commands |
| [SQL GENERATE CONNECTOR](sql-generate-connector.md) | Generate Database Connector MDL from an external schema |
| [IMPORT FROM](import-from.md) | Import data from an external database into a Mendix app database |
