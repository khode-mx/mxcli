# SHOW REFERENCES / IMPACT / CONTEXT

## Synopsis

    SHOW REFERENCES TO qualified_name

    SHOW IMPACT OF qualified_name

    SHOW CONTEXT OF qualified_name [ DEPTH n ]

## Description

These commands provide different views of cross-reference information for a given element. All three require `REFRESH CATALOG FULL` to have been run beforehand.

**SHOW REFERENCES TO** lists all elements that reference the specified element. This includes microflows that use an entity, pages that display it, associations that connect to it, and any other form of reference.

**SHOW IMPACT OF** performs an impact analysis showing what would be affected if the specified element were changed or removed. This is broader than `SHOW REFERENCES` as it considers transitive dependencies and indirect effects.

**SHOW CONTEXT OF** assembles the surrounding context of an element -- its definition, its callers, callees, and related elements -- suitable for providing to an LLM or for understanding an element in its broader project context. The optional `DEPTH` parameter controls how many levels of related elements to include.

## Parameters

**qualified_name**
: The fully qualified name of the element to analyze (e.g., `Module.EntityName`, `Module.MicroflowName`).

**n** (CONTEXT only)
: The number of levels of related elements to include. Defaults to 1 if not specified. Higher values include more surrounding context but produce more output.

## Examples

### Find all references to an entity

```sql
REFRESH CATALOG FULL;
SHOW REFERENCES TO Sales.Customer;
```

### Analyze impact before making changes

```sql
SHOW IMPACT OF Sales.Customer;
```

### Gather context for a microflow

```sql
SHOW CONTEXT OF Sales.ACT_CreateOrder;
```

### Gather deeper context

```sql
SHOW CONTEXT OF Sales.ACT_CreateOrder DEPTH 3;
```

### Check impact before moving an element

```sql
SHOW IMPACT OF Sales.CustomerEdit;
MOVE PAGE Sales.CustomerEdit TO NewModule;
```

### From the command line

```sql
-- Shell commands:
-- mxcli refs -p app.mpr Sales.Customer
-- mxcli impact -p app.mpr Sales.Customer
-- mxcli context -p app.mpr Sales.ACT_CreateOrder --depth 3
```

## See Also

[SHOW CALLERS / CALLEES](show-callers-callees.md), [REFRESH CATALOG](refresh-catalog.md), [SELECT FROM CATALOG](select-from-catalog.md)
