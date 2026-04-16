# Proposal: Import REST Client from OpenAPI Contract

## Overview

**Document type:** `Rest$ConsumedRestService`
**Priority:** High — eliminates the most tedious part of REST integration setup
**Depends on:** Consumed REST Services (fully implemented — SHOW/DESCRIBE/CREATE/DROP)

When a third-party REST API provides an OpenAPI 3.0 specification, Mendix Studio Pro can import it directly to create a Consumed REST Service document. mxcli has no equivalent: users must write `CREATE REST CLIENT` statements by hand, transcribing paths, methods, parameters, and response types one operation at a time.

This proposal adds `IMPORT REST CLIENT ... FROM OPENAPI` — a command that parses an OpenAPI 3.0 JSON or YAML file and generates the equivalent `ConsumedRestService` BSON document in one step.

Studio Pro also stores the raw OpenAPI spec in the `OpenApiFile.Content` BSON field on the service document. This proposal preserves that behavior so the resulting documents are indistinguishable from those created by Studio Pro.

## What Already Exists

| Layer | Status | Location |
|-------|--------|----------|
| **Generated metamodel** | Yes — `OpenApiFile` field defined | `generated/metamodel/types.go:7956` — `RestConsumedRestService.OpenAPIFile *RestOpenApiFile`; `RestOpenApiFile.Content string` at line 8154 |
| **Go model type** | Partial — no `OpenApiContent` field | `model/types.go:618` — `ConsumedRestService` struct complete; `OpenApiFile` not mapped |
| **BSON parser** | Partial — `OpenApiFile` not parsed | `sdk/mpr/parser_rest.go` — parses all other REST service fields |
| **BSON writer** | Partial — `OpenApiFile` not written | `sdk/mpr/writer_rest.go` — writes all other REST service fields |
| **AST** | No import stmt | `mdl/ast/ast_rest.go` — `CreateRestClientStmt` and `DropRestClientStmt` only |
| **Grammar** | No import rule | `mdl/grammar/MDLParser.g4:2395` — `createRestClientStatement` exists; no `importRestClientStatement` |
| **Visitor** | No import handler | `mdl/visitor/visitor_rest.go` — handles CREATE/DROP only |
| **Executor** | No import handler | `mdl/executor/cmd_rest_clients.go` — handles CREATE/SHOW/DESCRIBE/DROP |
| **OpenAPI parser** | Does not exist | No package in tree; stdlib `encoding/json`/`gopkg.in/yaml.v3` sufficient |

## BSON Structure

The `OpenApiFile` field on `Rest$ConsumedRestService` (from `generated/metamodel/types.go:7956`):

```
Rest$ConsumedRestService
  ...
  OpenApiFile: Rest$OpenApiFile?   // null when created by hand; set when imported from spec
    Content: string                // full raw OpenAPI JSON or YAML text
```

When this field is present Studio Pro shows "View OpenAPI" in the service editor and uses the spec to validate operations. Writing it is optional for functional correctness but required for full Studio Pro compatibility.

## Proposed MDL Syntax

### IMPORT REST CLIENT

```sql
IMPORT REST CLIENT Module.Name FROM OPENAPI '/path/to/openapi.json';

-- With base URL override (takes precedence over servers[0].url in the spec)
IMPORT REST CLIENT Module.Name FROM OPENAPI '/path/to/openapi.json'
  BASE URL 'https://api.example.com/v2';

-- Place into a module folder
IMPORT REST CLIENT Module.Name FROM OPENAPI '/path/to/openapi.json'
  BASE URL 'https://api.example.com/v2'
  FOLDER 'Module/Integrations';

-- Overwrite an existing service (equivalent to DELETE + recreate)
IMPORT OR REPLACE REST CLIENT Module.Name FROM OPENAPI '/path/to/openapi.json';
```

The module qualifier in the name is required. `IMPORT OR REPLACE` mirrors the `CREATE OR MODIFY` pattern already used in `CREATE REST CLIENT`.

### DESCRIBE OPENAPI FILE (read-only preview)

```sql
DESCRIBE OPENAPI FILE '/path/to/openapi.json';
```

Outputs the `CREATE REST CLIENT` MDL that would be generated — no project connection required. Useful for inspecting a spec before committing it to an MPR.

Example output:

```sql
/**
 * Swagger Petstore
 * A sample API that uses a petstore as an example.
 */
CREATE REST CLIENT MyModule.PetStore (
  BaseUrl: 'https://petstore.swagger.io/v2',
  Authentication: NONE
)
{
  OPERATION findPetsByStatus {
    Method: GET,
    Path: '/pet/findByStatus',
    Query: ($status: String),
    Response: JSON
  }

  OPERATION addPet {
    Method: POST,
    Path: '/pet',
    Headers: ('Content-Type' = 'application/json'),
    Body: JSON FROM $body,
    Response: JSON
  }

  OPERATION getPetById {
    Method: GET,
    Path: '/pet/{petId}',
    Parameters: ($petId: Integer),
    Response: JSON
  }

  OPERATION deletePet {
    Method: DELETE,
    Path: '/pet/{petId}',
    Parameters: ($petId: Integer),
    Response: NONE
  }
};
```

## OpenAPI → MDL Mapping

| OpenAPI Field | MDL Field | Notes |
|---|---|---|
| `info.title` + `info.description` | doc comment | Combined as javadoc on the service |
| `servers[0].url` | `BaseUrl` | First server entry; overridden by `BASE URL` clause |
| `paths.{path}.{method}.operationId` | `OPERATION name` | CamelCase; falls back to `{Method}_{sanitized_path}` if absent |
| `paths.{path}.{method}.summary` + `.description` | operation doc comment | |
| `paths.{path}` | `Path` | Preserved as-is including `{param}` placeholders |
| `paths.{path}.{method}` | `Method` | Uppercased: `GET`, `POST`, `PUT`, `PATCH`, `DELETE` |
| `parameters[?].in = "path"` | `Parameters: ($name: Type)` | Extracts type from `schema.type` |
| `parameters[?].in = "query"` | `Query: ($name: Type)` | Required params → `QUERY`; optional → `QUERY` with comment |
| `parameters[?].in = "header"` | `Headers: ('Name' = '')` | Static value only; dynamic values require manual update |
| `requestBody.content["application/json"]` | `Body: JSON FROM $body` | Other content types mapped to `Body: TEMPLATE` |
| `responses["200"].content["application/json"]` | `Response: JSON` | Non-JSON success responses → `Response: STRING` |
| `responses["200"]` (no body) | `Response: NONE` | |
| `responses["204"]` | `Response: NONE` | |
| `securitySchemes.{name}.type = "http" scheme = "basic"` | `Authentication: BASIC` | Credentials left blank — set via microflow or ALTER |
| `securitySchemes.{name}.type = "apiKey"` | Header or Query entry | Added as a static-value header or query parameter |
| `securitySchemes.{name}.type = "oauth2"` / `"openIdConnect"` | `Authentication: NONE` + warning | Not natively supported; logged as a warning |
| `x-timeout` (extension) | `Timeout: N` | Non-standard; ignored if absent |

### Type Mapping (OpenAPI schema → MDL parameter type)

| OpenAPI `schema.type` | OpenAPI `schema.format` | MDL Type |
|---|---|---|
| `string` | — | `String` |
| `integer` | `int32` | `Integer` |
| `integer` | `int64` | `Long` |
| `number` | `float` / `double` | `Decimal` |
| `boolean` | — | `Boolean` |
| any array / object | — | `String` (serialised; user refines) |

## Implementation Plan

### Phase 1: Extend Model Type (`model/types.go`)

Add `OpenApiContent` to `ConsumedRestService`:

```go
type ConsumedRestService struct {
    // ... existing fields ...
    OpenApiContent string `json:"openApiContent,omitempty"` // raw spec text (stored in OpenApiFile.Content BSON field)
}
```

### Phase 2: Extend BSON Parser (`sdk/mpr/parser_rest.go`)

In `parseConsumedRestService()`, read the `openApiFile` subdocument and populate `svc.OpenApiContent`:

```go
if openApiFile, ok := doc.Lookup("openApiFile").DocumentOK(); ok {
    svc.OpenApiContent = stringField(openApiFile, "content")
}
```

### Phase 3: Extend BSON Writer (`sdk/mpr/writer_rest.go`)

In `CreateConsumedRestService()` (or equivalent writer function), if `svc.OpenApiContent != ""`, serialize the `OpenApiFile` subdocument:

```go
if svc.OpenApiContent != "" {
    openApiFile := bson.D{
        {"$Type", "Rest$OpenApiFile"},
        {"content", svc.OpenApiContent},
    }
    doc = append(doc, bson.E{Key: "openApiFile", Value: openApiFile})
}
```

### Phase 4: Add AST Type (`mdl/ast/ast_rest.go`)

```go
// ImportRestClientFromOpenAPIStmt represents:
//   IMPORT [OR REPLACE] REST CLIENT Module.Name FROM OPENAPI '/path/to/spec.json'
//   [BASE URL 'https://...']
//   [FOLDER 'Module/Subfolder']
type ImportRestClientFromOpenAPIStmt struct {
    Name      QualifiedName
    SpecPath  string // path or URL to the OpenAPI file
    BaseUrl   string // overrides servers[0].url; empty = use spec value
    Folder    string // folder path within module; empty = module root
    OrReplace bool   // true if IMPORT OR REPLACE was used
}

func (s *ImportRestClientFromOpenAPIStmt) isStatement() {}
```

Also add to the `describeStatement` alternatives:

```go
// DescribeOpenAPIFile represents: DESCRIBE OPENAPI FILE '/path/to/spec.json'
type DescribeOpenAPIFileStmt struct {
    SpecPath string
}

func (s *DescribeOpenAPIFileStmt) isStatement() {}
```

### Phase 5: Update Grammar (`mdl/grammar/MDLParser.g4`)

Add a new branch to `statement`:

```antlr
| importRestClientStatement
```

Add the rule:

```antlr
importRestClientStatement
    : IMPORT (OR REPLACE)? REST CLIENT qualifiedName
      FROM OPENAPI STRING_LITERAL
      (BASE URL STRING_LITERAL)?
      (FOLDER STRING_LITERAL)?
    ;
```

Extend `describeStatement`:

```antlr
| DESCRIBE OPENAPI FILE STRING_LITERAL    // DESCRIBE OPENAPI FILE '/path/to/spec.json'
```

New tokens needed (add to `MDLLexer.g4` if not already present): `OPENAPI`, `FILE`. Check that `REPLACE` and `OR` tokens exist (they are used by `CREATE OR REPLACE` in published REST service rules).

After grammar changes: **run `make grammar`** to regenerate the ANTLR parser.

### Phase 6: Add Visitor Handler (`mdl/visitor/visitor_rest.go`)

Add `ExitImportRestClientStatement()` and `ExitDescribeOpenAPIFileStatement()`:

```go
func (b *ASTBuilder) ExitImportRestClientStatement(ctx *parser.ImportRestClientStatementContext) {
    stmt := &ast.ImportRestClientFromOpenAPIStmt{
        Name:      buildQualifiedName(ctx.QualifiedName()),
        SpecPath:  unquote(ctx.STRING_LITERAL(0).GetText()),
        OrReplace: ctx.REPLACE() != nil,
    }
    // Optional BASE URL clause
    if ctx.BASE() != nil {
        stmt.BaseUrl = unquote(ctx.STRING_LITERAL(1).GetText())
    }
    // Optional FOLDER clause
    if ctx.FOLDER() != nil {
        stmt.Folder = unquote(ctx.GetStop().GetText()) // last STRING_LITERAL
    }
    b.push(stmt)
}
```

### Phase 7: Add OpenAPI Parser (`mdl/openapi/parser.go` — new package)

A minimal, dependency-free OpenAPI 3.0 parser using only `encoding/json` and `gopkg.in/yaml.v3` (already in go.mod via other dependencies — verify before adding):

```go
package openapi

// Spec is a minimal OpenAPI 3.0 representation covering fields needed for REST client generation.
type Spec struct {
    Info       Info                       `json:"info" yaml:"info"`
    Servers    []Server                   `json:"servers" yaml:"servers"`
    Paths      map[string]PathItem        `json:"paths" yaml:"paths"`
    Components Components                 `json:"components" yaml:"components"`
    Security   []map[string][]string      `json:"security" yaml:"security"`
}

type Info struct {
    Title       string `json:"title" yaml:"title"`
    Description string `json:"description" yaml:"description"`
    Version     string `json:"version" yaml:"version"`
}

type Server struct {
    URL         string `json:"url" yaml:"url"`
    Description string `json:"description" yaml:"description"`
}

type PathItem struct {
    Get     *Operation `json:"get" yaml:"get"`
    Post    *Operation `json:"post" yaml:"post"`
    Put     *Operation `json:"put" yaml:"put"`
    Patch   *Operation `json:"patch" yaml:"patch"`
    Delete  *Operation `json:"delete" yaml:"delete"`
    Head    *Operation `json:"head" yaml:"head"`
    Options *Operation `json:"options" yaml:"options"`
}

// ... Operation, Parameter, RequestBody, Response, Components, SecurityScheme types

// ParseFile reads an OpenAPI spec from a file path (JSON or YAML, detected by extension).
func ParseFile(path string) (*Spec, error)

// ToRestClientModel converts a parsed spec to a model.ConsumedRestService.
// baseUrlOverride replaces servers[0].url when non-empty.
func ToRestClientModel(spec *Spec, name model.QualifiedName, baseUrlOverride string) (*model.ConsumedRestService, []string, error)
// Returns: service, warnings (e.g. unsupported auth schemes), error
```

Key conversion logic lives in `ToRestClientModel()`:
- Iterate `spec.Paths` deterministically (sort keys)
- For each path × method, call `operationToRestClientOp()`
- Map `schema.type` to MDL data type via a lookup table
- Sanitize `operationId` to a valid MDL identifier (replace non-alphanumeric with `_`)
- Collect warnings for unsupported features (OAuth2, multipart bodies, complex schemas)

### Phase 8: Add Executor Handler (`mdl/executor/cmd_rest_clients.go`)

```go
// importRestClientFromOpenAPI handles IMPORT REST CLIENT ... FROM OPENAPI command.
func (e *Executor) importRestClientFromOpenAPI(stmt *ast.ImportRestClientFromOpenAPIStmt) error {
    if e.writer == nil {
        return fmt.Errorf("not connected to a project (read-only mode)")
    }
    if err := e.checkFeature("integration", "rest_client_basic",
        "IMPORT REST CLIENT", "upgrade your project to 10.1+"); err != nil {
        return err
    }

    spec, err := openapi.ParseFile(stmt.SpecPath)
    if err != nil {
        return fmt.Errorf("failed to parse OpenAPI spec: %w", err)
    }

    svc, warnings, err := openapi.ToRestClientModel(spec, stmt.Name, stmt.BaseUrl)
    if err != nil {
        return fmt.Errorf("failed to convert OpenAPI spec: %w", err)
    }
    for _, w := range warnings {
        fmt.Fprintf(e.output, "Warning: %s\n", w)
    }

    // Store raw spec content (match Studio Pro behavior)
    rawBytes, _ := os.ReadFile(stmt.SpecPath)
    svc.OpenApiContent = string(rawBytes)

    // Resolve container (module / folder)
    module, err := e.findModule(stmt.Name.Module)
    if err != nil {
        return fmt.Errorf("module not found: %s", stmt.Name.Module)
    }
    svc.ContainerID = module.ID
    if stmt.Folder != "" {
        folderID, err := e.resolveFolder(module.ID, stmt.Folder)
        if err != nil {
            return fmt.Errorf("failed to resolve folder '%s': %w", stmt.Folder, err)
        }
        svc.ContainerID = folderID
    }

    // Handle OR REPLACE
    if stmt.OrReplace {
        existing, _ := e.reader.ListConsumedRestServices()
        h, _ := e.getHierarchy()
        for _, ex := range existing {
            modID := h.FindModuleID(ex.ContainerID)
            modName := h.GetModuleName(modID)
            if strings.EqualFold(modName, stmt.Name.Module) && strings.EqualFold(ex.Name, stmt.Name.Name) {
                _ = e.writer.DeleteConsumedRestService(ex.ID)
                break
            }
        }
    }

    if err := e.writer.CreateConsumedRestService(svc); err != nil {
        return fmt.Errorf("failed to create REST client: %w", err)
    }

    fmt.Fprintf(e.output, "Imported REST client: %s.%s (%d operations)\n",
        stmt.Name.Module, stmt.Name.Name, len(svc.Operations))
    return nil
}
```

Add `describeOpenAPIFile()` for the read-only preview path — parses the spec and calls the existing `outputConsumedRestServiceMDL()` without touching an MPR file.

### Phase 9: Wire into Executor Dispatch

Add cases in `mdl/executor/executor_dispatch.go` (or `executor.go`, wherever the main switch lives):

```go
case *ast.ImportRestClientFromOpenAPIStmt:
    return e.importRestClientFromOpenAPI(s)
case *ast.DescribeOpenAPIFileStmt:
    return e.describeOpenAPIFile(s)
```

### Phase 10: Help Text and Examples

- Add entry to `cmd/mxcli/help_topics/rest.txt`
- Create `mdl-examples/doctype-tests/20-openapi-import-examples.mdl` with roundtrip examples using the Petstore spec

## Version Requirements

No new version gate required. `IMPORT REST CLIENT` creates a `Rest$ConsumedRestService` document, which already requires Mendix 10.1.0+ (gated by `rest_client_basic` in `sdk/versions/mendix-10.yaml:111`). The executor pre-check in Phase 8 reuses that existing gate.

## Complexity

**Medium** — The OpenAPI parsing and field-mapping logic is the core of the work. The AST/grammar/visitor/executor plumbing follows established patterns. The main risks are:

1. **Spec diversity** — real-world OpenAPI specs vary widely. The parser should be tolerant of missing optional fields and emit warnings rather than errors for unsupported features.
2. **YAML dependency** — if `gopkg.in/yaml.v3` is not already in `go.sum`, it adds a new dependency. Confirm with `grep 'gopkg.in/yaml' go.sum` before assuming it's available; alternatively accept JSON only in Phase 1 and add YAML support later.
3. **Identifier sanitization** — `operationId` values in real specs often contain characters that are invalid as MDL identifiers.

## Testing

- Create `mdl-examples/doctype-tests/20-openapi-import-examples.mdl`
- Use `https://petstore.swagger.io/v2/swagger.json` (Swagger 2.0) and `https://petstore3.swagger.io/api/v3/openapi.json` (OpenAPI 3.0) as reference inputs
- Roundtrip: `IMPORT REST CLIENT` → `DESCRIBE REST CLIENT` → compare operation count and paths
- Verify `DESCRIBE OPENAPI FILE` works without an open project (`mxcli check` path)
- Verify `IMPORT OR REPLACE` overwrites an existing service without leaving orphaned documents

## Related

- `docs/11-proposals/show-describe-consumed-rest-services.md` — SHOW/DESCRIBE/CREATE/DROP for the same document type (fully implemented)
- `docs/11-proposals/integration-pane-proposal.md` — Phase 2 notes that `OpenApiFile` field "may" be present in REST services; this proposal confirms it and implements the import path
- `mdl/executor/cmd_rest_clients.go` — existing executor to extend
- `mdl/ast/ast_rest.go` — existing AST types to extend
- `sdk/mpr/parser_rest.go` — existing BSON parser to extend
- `sdk/mpr/writer_rest.go` — existing BSON writer to extend
