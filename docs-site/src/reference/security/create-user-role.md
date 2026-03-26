# CREATE USER ROLE

## Synopsis

```sql
CREATE USER ROLE Name ( module.Role [, ...] ) [ MANAGE ALL ROLES ]
```

## Description

Creates a project-level user role that aggregates one or more module roles. User roles are assigned to end users (either directly or via demo users) and determine the combined set of permissions across all modules.

A user role name must be unique at the project level. It is not qualified with a module name because user roles span multiple modules.

The optional `MANAGE ALL ROLES` clause grants the user role the ability to manage all other user roles in the application. This is typically reserved for administrator roles.

## Parameters

`Name`
:   The name of the user role. Not module-qualified. Must be unique among all user roles in the project.

`module.Role [, ...]`
:   A comma-separated list of module roles to include. Each module role is specified as a qualified name (`Module.RoleName`). The module roles must already exist.

`MANAGE ALL ROLES`
:   Optional. When specified, users with this role can manage (assign/unassign) all user roles in the application.

## Examples

Create an administrator role with management privileges:

```sql
CREATE USER ROLE AppAdmin (Shop.Admin, System.Administrator) MANAGE ALL ROLES;
```

Create a regular user role:

```sql
CREATE USER ROLE AppUser (Shop.User, Notifications.Viewer);
```

Create a role that spans multiple modules:

```sql
CREATE USER ROLE SalesManager (
    Sales.Manager,
    Inventory.Viewer,
    Reports.User,
    Administration.User
);
```

## See Also

[CREATE MODULE ROLE](create-module-role.md), [GRANT](grant.md), [CREATE DEMO USER](create-demo-user.md)
