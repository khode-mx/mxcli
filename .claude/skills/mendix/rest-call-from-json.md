# REST Call from JSON Payload — End-to-End Skill

Use this skill to generate the full stack of Mendix integration artifacts from a JSON payload:
JSON Structure → Non-persistent entities → Import Mapping → REST CALL microflow.

## Overview — Four Steps

1. **CREATE JSON STRUCTURE** — store the raw payload and derive the element tree
2. **CREATE ENTITY** (non-persistent) — one per JSON object type, with attributes per JSON field
3. **CREATE IMPORT MAPPING** — link JSON structure elements to entities and attributes
4. **CREATE MICROFLOW** — inline REST CALL that invokes the import mapping

---

## Step 1 — JSON Structure

```sql
CREATE JSON STRUCTURE Module.JSON_MyStructure
  FROM '{"key": "value", "count": 1}';
```

- The executor **formats** the snippet (pretty-print) then **refreshes** (derives element tree) automatically.
- The snippet must be valid JSON; use single quotes around it in MDL.
- Escape single quotes inside the snippet by doubling them: `''`.
- The derived element tree must stay consistent with the snippet — the executor sorts JSON object keys alphabetically to match `json.MarshalIndent` output.

**Verify** after creation:
```sql
DESCRIBE JSON STRUCTURE Module.JSON_MyStructure;
-- Should show: element tree under "-- Element tree:" comment
```

---

## Step 2 — Non-Persistent Entities

Derive one entity per JSON object type. Name them after what they represent (not after JSON keys).

```sql
CREATE OR REPLACE NON-PERSISTENT ENTITY Module.MyRootObject (
  stringField: String(0),
  intField: Integer,
  decimalField: Decimal,
  boolField: Boolean DEFAULT false
);

CREATE OR REPLACE NON-PERSISTENT ENTITY Module.MyNestedObject (
  name: String(0),
  code: String(0)
);

CREATE ASSOCIATION Module.MyRootObject_MyNestedObject
  FROM Module.MyRootObject
  TO Module.MyNestedObject;
```

**Rules:**
- All string fields: `String(0)` (unlimited). **Do NOT write `String(unlimited)`** — the grammar only accepts number literals. `String(0)` stores as unlimited and DESCRIBE outputs `String(unlimited)`.
- All number fields: `Integer`, `Decimal`, or `Long` — remove defaults for optional fields
- Boolean fields **require** `DEFAULT true|false`
- `NON-PERSISTENT` (hyphen) before `ENTITY` — not `(NON_PERSISTENT)` after the name
- Attributes go inside `(...)` comma-separated with colon separator: `name: Type`
- One association per parent→child relationship; name it `Parent_Child`

---

## Step 3 — Import Mapping

> **Full reference**: See [json-structures-and-mappings.md](json-structures-and-mappings.md) for complete import/export mapping syntax, domain model patterns, and common mistakes.

```sql
CREATE IMPORT MAPPING Module.IMM_MyMapping
  WITH JSON STRUCTURE Module.JSON_MyStructure
{
  CREATE Module.MyRootObject {
    stringField = stringField,
    intField    = intField,
    CREATE Module.MyRootObject_MyNestedObject/Module.MyNestedObject = nestedKey {
      name = name,
      code = code
    }
  }
};
```

**Syntax rules:**
- Root object: `CREATE Module.Entity { ... }` — always starts with handling keyword
- Value mappings: `AttributeName = jsonFieldName` — entity attribute on the left, JSON field on the right
- Nested objects: `CREATE Association/Entity = jsonKey { ... }` — association path + JSON key
- Object handling: `CREATE` (default), `FIND` (requires KEY), `FIND OR CREATE`
- KEY marker: `Attr = jsonField KEY` — marks the attribute as a matching key
- Value transforms: `Attr = Module.Microflow(jsonField)` — call a microflow to transform the value

**Verify** after creation — check Schema elements are ticked in Studio Pro:
- Open the import mapping in Studio Pro
- All JSON structure elements should appear ticked in the Schema elements panel
- If not ticked: JsonPath mismatch between import mapping and JSON structure elements

---

## Step 4 — REST CALL Microflow

Place the microflow in the `[Pages]/Operations/` folder or `Private/` depending on whether it is public.

```sql
CREATE MICROFLOW Module.GET_MyData ()
BEGIN
  @position(-5, 200)
  DECLARE $baseUrl String = 'https://api.example.com';
  @position(185, 200)
  DECLARE $endpoint String = $baseUrl + '/path';
  @position(375, 200)
  $Result = REST CALL GET '{1}' WITH ({1} = $endpoint)
    HEADER 'Accept' = 'application/json'
    TIMEOUT 300
    RETURNS MAPPING Module.IMM_MyMapping AS Module.MyRootObject ON ERROR ROLLBACK;
  @position(565, 200)
  LOG INFO NODE 'Integration' 'Retrieved result' WITH ();
END;
/
```

**Key points:**
- `@position` annotations control the canvas layout — StartEvent is auto-placed 150px to the left of the first annotated activity
- The output variable name is **automatically derived** from the entity name in `AS Module.MyEntity` — do NOT hardcode it on the left side; the executor overrides it
- Single vs list result is **automatically detected**: if the JSON structure's root element is an Object, the variable type is `ObjectType` (single); if Array, `ListType` (list)
- `ON ERROR ROLLBACK` — standard error handling for integration calls

**For list responses** (JSON root is an array):
```sql
  $Results = REST CALL GET '{1}' WITH ({1} = $endpoint)
    HEADER 'Accept' = 'application/json'
    TIMEOUT 300
    RETURNS MAPPING Module.IMM_MyMapping AS Module.MyItem ON ERROR ROLLBACK;
  @position(565, 200)
  $Count = COUNT($MyItem);
```

---

## Step 5 — Import/Export Mapping in Microflows (Optional)

Instead of using `RETURNS MAPPING` on a REST CALL, you can use standalone import/export mapping actions. This is useful when you already have a JSON string and want to map it to entities, or when you want to serialize entities back to JSON.

### Import from mapping

Applies an import mapping to a string variable (JSON content) to produce entity objects:

```sql
-- With assignment (non-persistent entities, need the result in the flow)
$OrderResponse = IMPORT FROM MAPPING Module.IMM_Order($JsonContent);

-- Without assignment (persistent entities, just stores to DB)
IMPORT FROM MAPPING Module.IMM_Order($JsonContent);
```

### Export to mapping

Applies an export mapping to an entity object to produce a JSON string:

```sql
$JsonOutput = EXPORT TO MAPPING Module.EMM_Order($OrderResponse);
```

### Complete import → process → export microflow

```sql
CREATE MICROFLOW Module.ProcessOrderData ()
BEGIN
  DECLARE $ResponseContent String = $latestHttpResponse/Content;
  $OrderResponse = IMPORT FROM MAPPING Module.IMM_Order($ResponseContent);
  -- Process the imported data...
  $JsonOutput = EXPORT TO MAPPING Module.EMM_Order($OrderResponse);
  LOG INFO NODE 'Integration' 'Exported: ' + $JsonOutput;
END;
/
```

---

## Complete Example — Generic Nested API

JSON shape: a root object containing metadata fields and a nested `item` object.

```sql
-- Step 1: JSON Structure
CREATE JSON STRUCTURE MyModule.JSON_MyApi
  FROM '{"item":{"code":"ABC","label":"Example","count":42},"status":"OK","version":"1.0"}';

-- Step 2: Entities
CREATE OR REPLACE NON-PERSISTENT ENTITY MyModule.MyApiResponse (
  ApiStatus: String(0),
  Version: String(0)
);

CREATE OR REPLACE NON-PERSISTENT ENTITY MyModule.MyApiItem (
  Code: String(0),
  Label: String(0),
  Count: Integer
);

CREATE ASSOCIATION MyModule.MyApiResponse_MyApiItem
  FROM MyModule.MyApiResponse
  TO MyModule.MyApiItem;

-- Step 3: Import Mapping
CREATE IMPORT MAPPING MyModule.IMM_MyApi
  WITH JSON STRUCTURE MyModule.JSON_MyApi
{
  CREATE MyModule.MyApiResponse {
    ApiStatus = status,
    Version = version,
    CREATE MyModule.MyApiResponse_MyApiItem/MyModule.MyApiItem = item {
      Code  = code,
      Label = label,
      Count = count
    }
  }
};

-- Step 4: Microflow
CREATE MICROFLOW MyModule.GET_MyApi_Item ()
BEGIN
  @position(-5, 200)
  DECLARE $url String = 'https://api.example.com/item';
  @position(185, 200)
  $MyApiResponse = REST CALL GET '{1}' WITH ({1} = $url)
    HEADER 'Accept' = 'application/json'
    TIMEOUT 300
    RETURNS MAPPING MyModule.IMM_MyApi AS MyModule.MyApiResponse ON ERROR ROLLBACK;
  @position(375, 200)
  RETRIEVE $MyApiItem FROM $MyApiResponse/MyModule.MyApiResponse_MyApiItem;
  @position(565, 200)
  LOG INFO NODE 'Integration' 'Status={1} Code={2} Count={3}'
    WITH ({1} = $MyApiResponse/ApiStatus,
          {2} = $MyApiItem/Code,
          {3} = toString($MyApiItem/Count));
END;
/
```

> **Note on association retrieve:** `RETRIEVE $Child FROM $Parent/Module.Assoc` on a Reference-type association traversed forward (from FK-owner to referenced entity) returns a **single entity**, not a list. `LIMIT 1` is redundant and will be dropped from the BSON roundtrip.

---

## Gotchas and Common Errors

| Symptom | Cause | Fix |
|---------|-------|-----|
| Studio Pro "not consistent with snippet" | JSON element tree keys not in alphabetical order | Executor sorts keys; re-derive from snippet |
| Schema elements not ticked in import mapping | JsonPath mismatch | Named object elements use `(Object)\|key`, NOT `(Object)\|key\|(Object)` |
| Import mapping not linked in REST call | Wrong BSON field name | Use `ReturnValueMapping`, not `Mapping` |
| Studio Pro shows "List of X" but mapping returns single X | `ForceSingleOccurrence` not set | Executor auto-detects from JSON structure root element type |
| StartEvent behind first activities | Default posX=200 vs @position(-5,...) | Fixed: executor pre-scans for first @position and shifts StartEvent left |
| `TypeCacheUnknownTypeException` | Wrong BSON `$Type` names | `ImportMappings$ObjectMappingElement` / `ImportMappings$ValueMappingElement` (no `Import` prefix) |
| `mxcli check` rejects `WITH JSON STRUCTURE` | Stale binary | Rebuild: `go build -o ./bin/mxcli ./cmd/mxcli` then retry |
| `mismatched input 'unlimited'` | `String(unlimited)` in CREATE | Use `String(0)` for unlimited strings; `unlimited` is DESCRIBE output only |
| `mismatched input ')'` on entity | Wrong entity form `CREATE ENTITY X (NON_PERSISTENT)` | Use `CREATE OR REPLACE NON-PERSISTENT ENTITY X (attrs)` |
| Attribute not found in Studio Pro | Attribute not fully qualified | Must be `Module.Entity.AttributeName` in the BSON |

---

## Naming Conventions (MES)

| Artifact | Pattern | Example |
|----------|---------|---------|
| JSON Structure | `JSON_<ApiName>` | `JSON_OrderApi` |
| Import Mapping | `IMM_<ApiName>` | `IMM_OrderApi` |
| Root entity | Describes the API response | `OrderApiResponse` |
| Nested entities | Describes the domain concept | `OrderItem`, `OrderAddress` |
| Microflow | `METHOD_Resource_Operation` | `GET_Order_ById` |
| Folder | `Private/` for mappings/structures, `Operations/` for public microflows | — |
