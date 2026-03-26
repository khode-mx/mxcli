# SET

## Synopsis

    SET variable = value

## Description

Sets a session variable that controls REPL behavior. Session variables persist for the duration of the session and are reset when the REPL exits. Variable names are case-insensitive.

## Parameters

**variable**
: The name of the session variable to set.

**value**
: The value to assign. Can be a string (in single quotes), a boolean (`TRUE` or `FALSE`), or an unquoted identifier.

### Recognized Variables

| Variable | Values | Description |
|----------|--------|-------------|
| `output_format` / `FORMAT` | `'json'`, `'table'` | Controls output format for query results |
| `verbose` | `TRUE`, `FALSE` | Enable verbose output for debugging |
| `AUTOCOMMIT` | `TRUE`, `FALSE` | Automatically save changes after each mutation |

## Examples

### Set output format to JSON

```sql
SET output_format = 'json';
```

### Enable verbose mode

```sql
SET verbose = TRUE;
```

### Enable autocommit

```sql
SET AUTOCOMMIT = TRUE;
```

## See Also

[SHOW STATUS](show-status.md), [OPEN PROJECT](../connection/open-project.md)
