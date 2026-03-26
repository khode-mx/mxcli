# SHOW NAVIGATION

## Synopsis

```sql
SHOW NAVIGATION
SHOW NAVIGATION MENU [ profile ]
SHOW NAVIGATION HOMES
DESCRIBE NAVIGATION [ profile ]
```

## Description

Displays navigation configuration for the current project.

### SHOW NAVIGATION

Displays a summary of all configured navigation profiles, showing the profile type and whether a home page, login page, and menu are configured.

### SHOW NAVIGATION MENU

Displays the menu tree for one or all navigation profiles. When a profile is specified, only that profile's menu is shown. Without a profile, all profiles' menus are displayed.

The menu tree is rendered as an indented hierarchy showing menu labels and their target pages.

### SHOW NAVIGATION HOMES

Displays home page assignments across all navigation profiles, including role-specific overrides.

### DESCRIBE NAVIGATION

Outputs the full MDL representation of one or all navigation profiles. The output is round-trippable -- it can be re-executed with `CREATE OR REPLACE NAVIGATION` to recreate the profile.

## Parameters

`profile`
:   Optional. One of: `Responsive`, `Tablet`, `Phone`, `NativePhone`. When omitted, all profiles are shown.

## Examples

Show a summary of all profiles:

```sql
SHOW NAVIGATION;
```

Show the menu tree for the responsive profile:

```sql
SHOW NAVIGATION MENU Responsive;
```

Show all home page assignments:

```sql
SHOW NAVIGATION HOMES;
```

Export the responsive profile as MDL:

```sql
DESCRIBE NAVIGATION Responsive;
```

## See Also

[ALTER NAVIGATION](alter-navigation.md)
