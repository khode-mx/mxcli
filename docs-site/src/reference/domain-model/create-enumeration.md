# CREATE ENUMERATION

## Synopsis

    CREATE [ OR MODIFY ] ENUMERATION module.name (
        value_name 'caption'
        [, ...]
    )

## Description

`CREATE ENUMERATION` defines a new enumeration type in a module. An enumeration is a fixed set of named values that can be used as an attribute type on entities (via `Enumeration(Module.EnumName)`).

Each enumeration value has two parts: a **value name** (an identifier used in code and expressions) and a **caption** (a display string shown in the UI). The value name must be a valid identifier. The caption is a single-quoted string.

If `OR MODIFY` is specified, the statement is idempotent. If the enumeration already exists, its values are updated to match the definition.

A documentation comment (`/** ... */`) placed before the statement is preserved as the enumeration's documentation.

## Parameters

**OR MODIFY**
: Makes the statement idempotent. If the enumeration already exists, its values are replaced with the new definition. Without this clause, creating a duplicate enumeration is an error.

**module.name**
: The qualified name of the enumeration in the form `Module.EnumerationName`. The module must already exist.

**value_name**
: An identifier for the enumeration value. Used in expressions and microflow logic.

**caption**
: A single-quoted display string for the value. This is what end users see in the UI.

## Examples

### Basic enumeration

```sql
CREATE ENUMERATION Sales.OrderStatus (
    Draft 'Draft',
    Pending 'Pending Approval',
    Approved 'Approved',
    Shipped 'Shipped',
    Delivered 'Delivered',
    Cancelled 'Cancelled'
);
```

### Enumeration with documentation

```sql
/** Priority levels for support tickets */
CREATE ENUMERATION Support.TicketPriority (
    Low 'Low',
    Medium 'Medium',
    High 'High',
    Critical 'Critical'
);
```

### Idempotent with OR MODIFY

```sql
CREATE OR MODIFY ENUMERATION Sales.OrderStatus (
    Draft 'Draft',
    Pending 'Pending Approval',
    Approved 'Approved',
    Shipped 'Shipped',
    Delivered 'Delivered',
    Returned 'Returned',
    Cancelled 'Cancelled'
);
```

### Using an enumeration as an attribute type

```sql
CREATE PERSISTENT ENTITY Sales.Order (
    OrderNumber: String(50) NOT NULL,
    Status: Enumeration(Sales.OrderStatus) DEFAULT 'Draft'
);
```

## Notes

- Enumeration values can also be managed incrementally with `ALTER ENUMERATION ... ADD VALUE` and `ALTER ENUMERATION ... REMOVE VALUE`.
- The `DEFAULT` value on an entity attribute references the enumeration value name as a quoted string (e.g., `DEFAULT 'Draft'`).

## See Also

[DROP ENUMERATION](drop-enumeration.md), [CREATE ENTITY](create-entity.md)
