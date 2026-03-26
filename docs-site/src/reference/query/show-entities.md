# SHOW ENTITIES

## Synopsis

    SHOW ENTITIES [IN <module>]

    SHOW ENTITY <qualified_name>

## Description

Lists entities in the project. Without the `IN` clause, lists all entities across all modules. With `IN <module>`, restricts the listing to entities in the specified module.

The `SHOW ENTITY` variant displays a summary of a single entity by its qualified name, including its type (persistent/non-persistent) and attribute count.

## Parameters

*module*
: The name of the module to filter by. Only entities belonging to this module are shown.

*qualified_name*
: A `Module.EntityName` reference identifying a specific entity.

## Examples

List all entities in the project:

```sql
SHOW ENTITIES
```

List entities in a specific module:

```sql
SHOW ENTITIES IN Sales
```

Show summary of a single entity:

```sql
SHOW ENTITY Sales.Customer
```

## See Also

[DESCRIBE ENTITY](describe-entity.md), [SHOW ASSOCIATIONS](show-associations.md), [SHOW MODULES](show-modules.md)
