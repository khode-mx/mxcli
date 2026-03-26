# SHOW MICROFLOWS / SHOW NANOFLOWS

## Synopsis

    SHOW MICROFLOWS [IN <module>]

    SHOW NANOFLOWS [IN <module>]

## Description

Lists microflows or nanoflows in the project. Without the `IN` clause, lists all microflows (or nanoflows) across all modules. With `IN <module>`, restricts the listing to the specified module.

Microflows run on the server side and can perform database operations, call external services, and execute Java actions. Nanoflows run on the client side and are used for offline-capable and low-latency logic.

## Parameters

*module*
: The name of the module to filter by. Only microflows or nanoflows belonging to this module are shown.

## Examples

List all microflows in the project:

```sql
SHOW MICROFLOWS
```

List microflows in a specific module:

```sql
SHOW MICROFLOWS IN Administration
```

List all nanoflows:

```sql
SHOW NANOFLOWS
```

List nanoflows in a specific module:

```sql
SHOW NANOFLOWS IN MyFirstModule
```

## See Also

[DESCRIBE MICROFLOW](describe-microflow.md), [SHOW PAGES](show-pages.md), [SHOW MODULES](show-modules.md)
