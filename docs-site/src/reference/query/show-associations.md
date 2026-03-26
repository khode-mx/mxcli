# SHOW ASSOCIATIONS

## Synopsis

    SHOW ASSOCIATIONS [IN <module>]

    SHOW ASSOCIATION <qualified_name>

## Description

Lists associations in the project. Without the `IN` clause, lists all associations across all modules. With `IN <module>`, restricts the listing to associations in the specified module.

The `SHOW ASSOCIATION` variant displays a summary of a single association by its qualified name, including the FROM entity, TO entity, and association type.

## Parameters

*module*
: The name of the module to filter by. Only associations belonging to this module are shown.

*qualified_name*
: A `Module.AssociationName` reference identifying a specific association.

## Examples

List all associations in the project:

```sql
SHOW ASSOCIATIONS
```

List associations in a specific module:

```sql
SHOW ASSOCIATIONS IN Sales
```

Show a single association:

```sql
SHOW ASSOCIATION Sales.Order_Customer
```

## See Also

[DESCRIBE ASSOCIATION](describe-association.md), [SHOW ENTITIES](show-entities.md), [SHOW MODULES](show-modules.md)
