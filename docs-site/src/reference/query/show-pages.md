# SHOW PAGES / SHOW SNIPPETS

## Synopsis

    SHOW PAGES [IN <module>]

    SHOW SNIPPETS [IN <module>]

## Description

Lists pages or snippets in the project. Without the `IN` clause, lists all pages (or snippets) across all modules. With `IN <module>`, restricts the listing to the specified module.

Pages are the user interface screens of a Mendix application. Snippets are reusable page fragments that can be embedded in multiple pages via `SNIPPETCALL` widgets.

## Parameters

*module*
: The name of the module to filter by. Only pages or snippets belonging to this module are shown.

## Examples

List all pages in the project:

```sql
SHOW PAGES
```

List pages in a specific module:

```sql
SHOW PAGES IN Sales
```

List all snippets:

```sql
SHOW SNIPPETS
```

List snippets in a specific module:

```sql
SHOW SNIPPETS IN Common
```

## See Also

[DESCRIBE PAGE](describe-page.md), [SHOW WIDGETS](show-widgets.md), [SHOW MODULES](show-modules.md)
