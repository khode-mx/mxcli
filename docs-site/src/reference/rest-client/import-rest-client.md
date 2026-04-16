# IMPORT REST CLIENT FROM OPENAPI

Import a consumed REST service from an OpenAPI 3.0 JSON specification in one step. The spec is parsed automatically and stored in the service document — exactly as Mendix Studio Pro does when you use its built-in OpenAPI import.

**Requires Mendix 10.1+.**

## Syntax

```sql
IMPORT REST CLIENT Module.Name FROM OPENAPI '/path/to/openapi.json';

-- Idempotent: replace an existing service with the same name
IMPORT OR REPLACE REST CLIENT Module.Name FROM OPENAPI '/path/to/openapi.json';

-- Override base URL (use when spec has no servers array)
IMPORT REST CLIENT Module.Name FROM OPENAPI '/path/to/openapi.json'
  SET BaseUrl: 'https://api.example.com';

-- Override base URL and authentication
IMPORT OR REPLACE REST CLIENT Module.Name FROM OPENAPI '/path/to/openapi.json'
  SET BaseUrl: 'https://api.example.com',
      Authentication: BASIC (Username: '$Module.ApiUser', Password: '$Module.ApiPass');
```

## Parameters

| Parameter | Description |
|-----------|-------------|
| `Module.Name` | Qualified name for the new REST client. Module must already exist. |
| `'/path/to/openapi.json'` | Absolute or relative path to the OpenAPI 3.0 JSON spec file. |
| `OR REPLACE` | If a REST client with the same name already exists, replace it. Without this clause the command fails if the service already exists. |
| `SET BaseUrl` | Override the base URL from the spec's `servers[0].url`. Required when the spec has no `servers` array. |
| `SET Authentication` | Override the authentication. Supported values: `NONE`, `BASIC (Username: '...', Password: '...')`. |

## What Gets Imported

| OpenAPI Field | REST Client Field |
|---------------|-------------------|
| `servers[0].url` | `BaseUrl` (trailing slash stripped) |
| `paths[path][method].operationId` | Operation name (sanitized to valid identifier) |
| `paths[path]` | `Path` (same `{param}` placeholder format as Mendix) |
| `paths[path][method]` | `Method` (uppercased: `GET`, `POST`, `PUT`, `PATCH`, `DELETE`) |
| `parameters[*].in = "path"` | `Parameters` |
| `parameters[*].in = "query"` | `Query` |
| `requestBody` (non-DELETE/HEAD) | `Body: JSON` |
| `responses["200"/"201"/"2XX"]` | `Response: JSON` |
| `responses["204"]` / no 2xx | `Response: NONE` |
| Raw spec bytes | `OpenApiFile.Content` (stored as-is, enables "View OpenAPI" in Studio Pro) |

**Note:** duplicate parameter names within a single operation (a spec authoring error) are preserved as-is, matching Studio Pro behaviour. Studio Pro will surface these as warnings when you open the project.

### Type Mapping

| OpenAPI Type | Format | Mendix Type |
|--------------|--------|-------------|
| `string` | — | `String` |
| `integer` | `int32` (or none) | `Integer` |
| `integer` | `int64` | `Long` |
| `number` | — | `Decimal` |
| `boolean` | — | `Boolean` |

## Examples

### Basic import

```sql
-- Create the module first
CREATE MODULE PetStoreIntegration;

-- Import all endpoints from the spec
IMPORT REST CLIENT PetStoreIntegration.PetStoreAPI FROM OPENAPI '/specs/petstore.json';
```

### Spec without servers array

```sql
IMPORT REST CLIENT MyModule.YourAPI FROM OPENAPI '/specs/openapi.json'
  SET BaseUrl: 'https://api.example.com/v1';
```

### Idempotent re-import after spec update

```sql
IMPORT OR REPLACE REST CLIENT MyModule.YourAPI FROM OPENAPI '/specs/openapi.json';
```

## DESCRIBE OPENAPI FILE

Preview what `IMPORT REST CLIENT` would generate without modifying the project. No project connection is needed.

```sql
DESCRIBE OPENAPI FILE '/path/to/openapi.json';
```

Output is a complete `CREATE REST CLIENT` statement — re-executable MDL showing all operations that would be created.

## Studio Pro Compatibility

The raw spec is stored in the service document's `OpenApiFile.Content` field. When you open the project in Studio Pro, the REST client shows a **View OpenAPI** button and validates operations against the spec, exactly as if you had used Studio Pro's built-in import.

## Related

- [REST Client Statements](README.md) — overview of all REST client commands
- [SEND REST REQUEST](../../language/microflow/send-rest-request.md) — call an operation from a microflow
- `SHOW REST CLIENTS` / `DESCRIBE REST CLIENT Module.Name` — list and inspect imported services
