# Proposal: Bulk External Action Support from OData Contracts

## Problem

Issue [#143](https://github.com/mendixlabs/mxcli/issues/143) requests importing all entities **and actions** from a consumed OData service. `CREATE EXTERNAL ENTITIES FROM Module.Service` now handles bulk entity creation. This proposal covers the action side: generating the artifacts needed to call external OData actions.

## Current State

| Capability | Status |
|------------|--------|
| Browse contract actions | ✅ `SHOW CONTRACT ACTIONS FROM Module.Service` |
| Describe contract action | ✅ `DESCRIBE CONTRACT ACTION Module.Service.Action` |
| Call external action in microflow | ✅ `CALL EXTERNAL ACTION Module.Service.Action(...)` |
| Bulk-create external entities | ✅ `CREATE EXTERNAL ENTITIES FROM Module.Service` |
| Bulk-create action support artifacts | ❌ Not implemented |
| Parse complex types from $metadata | ❌ Not implemented |
| Response tree depth handling | ❌ Not modeled |

## What Calling an External Action Requires

### Artifacts per action

1. **Parameter entities** — for each action parameter with a complex type (non-primitive), a non-persistent entity (NPE) or external entity must exist in the domain model
2. **Return type entity** — if the action returns a complex type, the corresponding entity must exist
3. **The microflow** — with `CALL EXTERNAL ACTION` wired up with parameter mappings

### Type classification

OData `$metadata` defines several type kinds that map differently to Mendix:

| OData type | Mendix equivalent | Example |
|------------|-------------------|---------|
| `Edm.String`, `Edm.Int32`, etc. | Primitive attribute types | `Edm.String` → `String(200)` |
| Entity type (has entity set) | External entity | `TripPin.Person` → external entity |
| Complex type (no entity set) | Non-persistent entity | `TripPin.City` → NPE with `CityName`, `Region` |
| `Collection(Namespace.Type)` | List of entity | `Collection(TripPin.Trip)` → list parameter |
| Enum type | Enumeration | `TripPin.PersonGender` → enumeration |

### Response tree depth (key complexity)

When Studio Pro configures an external action call, it allows the user to choose **how deep** the response should be deserialized:

- **Top-level only** — only the returned entity's own attributes
- **With associations** — also deserialize associated complex objects / navigation properties

This is controlled in the BSON by the `VariableDataType` field on `CallExternalAction`, which specifies both the return type and what parts of the object graph to materialize. The exact mechanism needs investigation by creating reference examples in Studio Pro.

**Open questions:**
- How is the depth/scope stored in BSON? Is it a separate field or encoded in `VariableDataType`?
- Can it reference navigation properties selectively (like `$expand` in OData)?
- Does it affect which NPEs need to exist? (i.e., if you only request top-level, do you still need NPEs for nested complex types?)

## Gaps to Address

### 1. Complex type parsing in EDMX

`sdk/mpr/edmx.go` currently parses entity types, enum types, and actions from `$metadata` XML. It does **not** parse `<ComplexType>` elements. These are required because:

- Action parameters often use complex types as input
- Action return types may be complex types
- Navigation properties on entity types may reference complex types

**Work needed:**
- Add `EdmComplexType` struct (similar to `EdmEntityType` but without key properties or entity set)
- Parse `<ComplexType>` in `ParseEdmx()`
- Store on `EdmSchema.ComplexTypes`

### 2. Type resolution for action parameters

`EdmAction.Parameters[].Type` can be:
- A primitive (`Edm.String`) — no entity needed
- A qualified entity type (`Namespace.Customer`) — needs external entity
- A qualified complex type (`Namespace.Address`) — needs NPE
- A collection (`Collection(Namespace.Item)`) — needs entity + list parameter type
- An enum (`Namespace.Status`) — needs enumeration

Resolution requires looking up whether a type name refers to an entity type (has entity set), complex type, or enum type in the schema.

### 3. Response tree depth

The core design question: when generating action call scaffolding, how do we handle the response tree?

**Option A: Top-level only (simplest)**
Generate only the immediate return type entity. Users manually expand if needed.

**Option B: Full tree (Studio Pro default)**
Walk navigation properties of the return type, recursively create NPEs/external entities for all reachable types.

**Option C: User-controlled depth**
```sql
CREATE EXTERNAL ACTIONS FROM Module.Service DEPTH 1;  -- top-level only
CREATE EXTERNAL ACTIONS FROM Module.Service DEPTH 2;  -- include direct associations
```

**Recommendation:** Start with Option A, document that users can add depth manually. Investigate Studio Pro's BSON to understand the exact storage before implementing deeper options.

### 4. CallExternalAction BSON completeness

Current parser/writer for `CallExternalAction` may be missing fields that Studio Pro writes. Specifically:
- `VariableDataType` — not currently parsed (controls return type inference)
- Response depth/scope fields — unknown, needs BSON investigation

## Proposed Syntax

### Phase 1: Entity/type scaffolding only

```sql
-- Create NPEs and enumerations for all action parameter/return types
-- that don't already exist as entities in the project
CREATE EXTERNAL ACTIONS FROM Module.Service;

-- Filter to specific actions
CREATE EXTERNAL ACTIONS FROM Module.Service ACTIONS (GetTrips, CreateTrip);

-- Into a different module
CREATE EXTERNAL ACTIONS FROM Module.Service INTO Integration;

-- Idempotent
CREATE OR MODIFY EXTERNAL ACTIONS FROM Module.Service;
```

This would:
1. Parse all actions from cached `$metadata`
2. For each action, resolve parameter types and return type
3. Create NPEs for complex types that don't have entity sets
4. Create external entities for entity types that aren't already imported
5. Create enumerations for enum types
6. Output a summary of what was created

It would **not** generate microflows — that's the user's job (or a Phase 2 feature).

### Phase 2: Microflow generation (future)

```sql
-- Generate stub microflows that call each action
CREATE EXTERNAL ACTION MICROFLOWS FROM Module.Service;
```

### Phase 3: DESCRIBE FORMAT mdl for actions (future)

```sql
-- Generate a complete MDL script for calling an action
DESCRIBE CONTRACT ACTION Module.Service.GetTrips FORMAT mdl;
```

Would output something like:

```sql
-- Required NPE for return type
CREATE NON-PERSISTENT ENTITY Module.Trip (
    TripId: Long,
    Name: String(200),
    Description: String(500)
);

-- Microflow to call the action
CREATE MICROFLOW Module.ACT_GetTrips($PersonId: String) RETURNS List of Module.Trip
BEGIN
    $Result = CALL EXTERNAL ACTION Module.Service.GetTrips(personId = $PersonId);
    RETURN $Result;
END;
```

## Investigation Required Before Implementation

Before coding, these questions need answers from Studio Pro reference examples:

1. **Create a reference project** with a consumed OData service (e.g., TripPin) that has actions with complex type parameters and return types
2. **Inspect the BSON** for `CallExternalAction` to understand:
   - How `VariableDataType` encodes the return type
   - Whether there's a depth/expand field for response tree
   - How bound actions differ in storage
3. **Inspect the NPEs** that Studio Pro creates for complex types:
   - Are they plain `DomainModels$EntityImpl` with `Persistable: false`?
   - Do they have a special `Source` field (like external entities have `Rest$ODataRemoteEntitySource`)?
   - How are associations between NPEs and external entities stored?
4. **Inspect enumerations** from OData:
   - Do they use `Rest$ODataRemoteEnumerationSource`?
   - How are they linked back to the consumed service?

## Implementation Order

1. **Parse complex types** from `$metadata` (`sdk/mpr/edmx.go`)
2. **Type resolver** — given a qualified type name, determine if it's entity/complex/enum/primitive
3. **Phase 1 executor** — `CREATE EXTERNAL ACTIONS FROM` creates NPEs, enums, external entities
4. **BSON investigation** — Studio Pro reference project for CallExternalAction fields
5. **Phase 2** — `DESCRIBE CONTRACT ACTION ... FORMAT mdl` generates complete MDL
6. **Phase 3** — microflow generation

## Related Files

- `sdk/mpr/edmx.go` — EDMX parsing (needs ComplexType support)
- `sdk/mpr/parser_microflow_actions.go` — CallExternalAction parser
- `sdk/mpr/writer_microflow_actions.go` — CallExternalAction writer
- `sdk/microflows/microflows_actions.go` — CallExternalAction struct
- `mdl/executor/cmd_contract.go` — contract browsing + CREATE EXTERNAL ENTITIES
- `mdl/executor/cmd_odata.go` — OData CRUD commands
- `mdl/ast/ast_odata.go` — OData AST nodes
- `docs/11-proposals/odata-services-proposal.md` — original OData proposal (Phase 3 section)
