# SHOW MODULES, SHOW ENTITIES

The `SHOW` family of commands lists project elements by type. They are the fastest way to see what a project contains.

## Listing modules

Every Mendix project is organized into modules. Start by listing them:

```sql
SHOW MODULES;
```

Example output:

```
MyFirstModule
Administration
Atlas_Core
System
```

By default, system and marketplace modules (like `System` and `Atlas_Core`) are included. The modules appear in the order they are defined in the project.

## Listing entities

To see all entities across all modules:

```sql
SHOW ENTITIES;
```

Example output:

```
Administration.Account
MyFirstModule.Customer
MyFirstModule.Order
MyFirstModule.OrderLine
```

Each entity is shown as a **qualified name** -- the module name and entity name separated by a dot.

### Filtering by module

Most SHOW commands accept an `IN` clause to filter results to a single module:

```sql
SHOW ENTITIES IN MyFirstModule;
```

```
MyFirstModule.Customer
MyFirstModule.Order
MyFirstModule.OrderLine
```

This is typically what you want when working on a specific module.

## Listing microflows

```sql
SHOW MICROFLOWS;
```

```
Administration.ChangeMyPassword
MyFirstModule.ACT_Customer_Save
MyFirstModule.ACT_Order_Process
MyFirstModule.DS_Customer_GetAll
```

Filter to a module:

```sql
SHOW MICROFLOWS IN MyFirstModule;
```

```
MyFirstModule.ACT_Customer_Save
MyFirstModule.ACT_Order_Process
MyFirstModule.DS_Customer_GetAll
```

## Listing pages

```sql
SHOW PAGES;
```

```
Administration.Account_Overview
Administration.Login
MyFirstModule.Customer_Overview
MyFirstModule.Customer_Edit
MyFirstModule.Order_Detail
```

Filter to a module:

```sql
SHOW PAGES IN MyFirstModule;
```

```
MyFirstModule.Customer_Overview
MyFirstModule.Customer_Edit
MyFirstModule.Order_Detail
```

## Other SHOW commands

The same pattern works for all major element types:

```sql
SHOW ENUMERATIONS;
SHOW ENUMERATIONS IN MyFirstModule;

SHOW ASSOCIATIONS;
SHOW ASSOCIATIONS IN MyFirstModule;

SHOW WORKFLOWS;
SHOW WORKFLOWS IN MyFirstModule;

SHOW NANOFLOWS;
SHOW NANOFLOWS IN MyFirstModule;

SHOW CONSTANTS;
SHOW CONSTANTS IN MyFirstModule;

SHOW SNIPPETS;
SHOW SNIPPETS IN MyFirstModule;
```

You can also list security-related elements:

```sql
SHOW MODULE ROLES;
SHOW MODULE ROLES IN MyFirstModule;

SHOW USER ROLES;

SHOW DEMO USERS;
```

And navigation:

```sql
SHOW NAVIGATION;
```

## Using SHOW from the command line

Every SHOW command works as a CLI one-liner with `-c`:

```bash
mxcli -p app.mpr -c "SHOW ENTITIES"
mxcli -p app.mpr -c "SHOW MICROFLOWS IN MyFirstModule"
mxcli -p app.mpr -c "SHOW PAGES"
```

This is useful for quick lookups without entering the REPL, and for piping output to other tools:

```bash
# Count entities per module
mxcli -p app.mpr -c "SHOW ENTITIES" | cut -d. -f1 | sort | uniq -c

# Find all microflows with "Save" in the name
mxcli -p app.mpr -c "SHOW MICROFLOWS" | grep -i save
```

## Summary of SHOW commands

| Command | Description |
|---------|-------------|
| `SHOW MODULES` | List all modules |
| `SHOW ENTITIES [IN Module]` | List entities |
| `SHOW MICROFLOWS [IN Module]` | List microflows |
| `SHOW NANOFLOWS [IN Module]` | List nanoflows |
| `SHOW PAGES [IN Module]` | List pages |
| `SHOW SNIPPETS [IN Module]` | List snippets |
| `SHOW ENUMERATIONS [IN Module]` | List enumerations |
| `SHOW ASSOCIATIONS [IN Module]` | List associations |
| `SHOW CONSTANTS [IN Module]` | List constants |
| `SHOW WORKFLOWS [IN Module]` | List workflows |
| `SHOW BUSINESS EVENTS [IN Module]` | List business event services |
| `SHOW JAVA ACTIONS [IN Module]` | List Java actions |
| `SHOW MODULE ROLES [IN Module]` | List module roles |
| `SHOW USER ROLES` | List user roles |
| `SHOW DEMO USERS` | List demo users |
| `SHOW NAVIGATION` | Show navigation profiles |

Now that you can list elements, the next step is inspecting individual elements in detail with [DESCRIBE and SEARCH](describe-search.md).
