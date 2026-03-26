# REVOKE

## Synopsis

```sql
-- Entity access
REVOKE module.Role ON module.Entity

-- Microflow access
REVOKE EXECUTE ON MICROFLOW module.Name FROM module.Role [, ...]

-- Page access
REVOKE VIEW ON PAGE module.Name FROM module.Role [, ...]

-- Nanoflow access
REVOKE EXECUTE ON NANOFLOW module.Name FROM module.Role [, ...]
```

## Description

Removes previously granted access rights from module roles. Each form is the counterpart to the corresponding GRANT statement.

### Entity Access

Removes the entire entity access rule for the specified module role on the entity. Unlike GRANT, there is no way to partially revoke (e.g., remove only WRITE while keeping READ). The entire rule is removed.

### Microflow Access

Removes execute permission on a microflow from one or more module roles.

### Page Access

Removes view permission on a page from one or more module roles.

### Nanoflow Access

Removes execute permission on a nanoflow from one or more module roles.

## Parameters

`module.Role`
:   The module role losing access. Must be a qualified name (`Module.RoleName`).

`module.Entity`
:   The entity whose access rule is removed.

`module.Name`
:   The target microflow, nanoflow, or page.

`FROM module.Role [, ...]`
:   One or more module roles losing access (for microflow, page, and nanoflow forms).

## Examples

Remove entity access for a role:

```sql
REVOKE Shop.Viewer ON Shop.Customer;
```

Remove microflow execution from multiple roles:

```sql
REVOKE EXECUTE ON MICROFLOW Shop.ACT_Order_Process FROM Shop.Viewer;
```

Remove page visibility:

```sql
REVOKE VIEW ON PAGE Shop.Admin_Dashboard FROM Shop.User;
```

Remove nanoflow execution:

```sql
REVOKE EXECUTE ON NANOFLOW Shop.NAV_ValidateInput FROM Shop.Viewer;
```

## See Also

[GRANT](grant.md), [CREATE MODULE ROLE](create-module-role.md)
