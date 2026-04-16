# REST Client Statements

Statements for creating and managing consumed REST services (REST clients).

A REST client document defines a third-party HTTP API — its base URL, authentication, and operations — once. Microflows then call its operations using `SEND REST REQUEST` without repeating connection details.

## Import from OpenAPI

| Statement | Description |
|-----------|-------------|
| [IMPORT REST CLIENT FROM OPENAPI](import-rest-client.md) | Create a consumed REST service from an OpenAPI 3.0 JSON spec |
| [DESCRIBE OPENAPI FILE](import-rest-client.md#describe-openapi-file) | Preview what an OpenAPI spec would generate — no project needed |

## Create / Modify / Drop

| Statement | Description |
|-----------|-------------|
| `CREATE REST CLIENT Module.Name ( ... ) { ... }` | Create a REST client manually |
| `CREATE OR MODIFY REST CLIENT Module.Name ( ... ) { ... }` | Idempotent create/update |
| `DROP REST CLIENT Module.Name` | Remove a REST client |

## Show / Describe

| Statement | Description |
|-----------|-------------|
| `SHOW REST CLIENTS [IN Module]` | List all consumed REST services |
| `DESCRIBE REST CLIENT Module.Name` | Show full MDL definition (re-executable) |

## Catalog Tables

After `REFRESH CATALOG`, the following tables are available:

| Table | Contents |
|-------|----------|
| `CATALOG.REST_CLIENTS` | Consumed REST services |
| `CATALOG.REST_OPERATIONS` | Operations per service (method, path, parameters) |

## Requirements

REST client documents require **Mendix 10.1+**.
