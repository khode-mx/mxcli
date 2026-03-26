# SHOW ENUMERATIONS

## Synopsis

    SHOW ENUMERATIONS [IN <module>]

## Description

Lists enumerations in the project. Without the `IN` clause, lists all enumerations across all modules. With `IN <module>`, restricts the listing to enumerations in the specified module.

Enumerations define a fixed set of named values that can be used as attribute types on entities.

## Parameters

*module*
: The name of the module to filter by. Only enumerations belonging to this module are shown.

## Examples

List all enumerations in the project:

```sql
SHOW ENUMERATIONS
```

List enumerations in a specific module:

```sql
SHOW ENUMERATIONS IN Sales
```

## See Also

[DESCRIBE ENUMERATION](describe-enumeration.md), [SHOW ENTITIES](show-entities.md), [SHOW MODULES](show-modules.md)
