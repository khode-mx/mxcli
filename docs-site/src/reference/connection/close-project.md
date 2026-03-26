# CLOSE PROJECT

## Synopsis

    DISCONNECT

## Description

Closes the currently open project and releases the file handle. Any uncommitted changes are discarded. After disconnecting, query and mutation statements are unavailable until a new project is opened with `CONNECT LOCAL`.

## Parameters

This statement takes no parameters.

## Examples

### Close the current project

```sql
DISCONNECT;
```

### Open a different project

```sql
DISCONNECT;
CONNECT LOCAL '/path/to/other-project.mpr';
```

## See Also

[OPEN PROJECT](open-project.md), [SHOW STATUS](../session/show-status.md)
