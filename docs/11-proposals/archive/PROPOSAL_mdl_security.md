# Proposal: MDL Security Support

## Overview

This document proposes MDL syntax for managing Mendix security configuration. Security in Mendix spans three levels:

1. **Project Security** — Security level, admin credentials, application user roles, demo users
2. **Module Security** — Module roles (defined per module)
3. **Artifact Access** — Per-entity, per-microflow, and per-page access rules tied to module roles

## BSON Storage Reference

The following BSON types are involved:

| BSON `$Type` | Containment | Purpose |
|---|---|---|
| `Security$ProjectSecurity` | Project root | Security level, user roles, demo users, password policy |
| `Security$UserRole` | Inside ProjectSecurity | Application-level user role combining module roles |
| `Security$DemoUserImpl` | Inside ProjectSecurity | Demo user for testing |
| `Security$ModuleSecurity` | Per module (`ModuleSecurity`) | Container for module roles |
| `Security$ModuleRole` | Inside ModuleSecurity | Module-level role |
| `DomainModels$AccessRule` | Inside Entity | Entity access: create/delete, member rights, XPath |
| `DomainModels$MemberAccess` | Inside AccessRule | Per-attribute/association rights |
| `Microflows$Microflow.AllowedModuleRoles` | Microflow field | Which module roles can execute |
| `Microflows$Microflow.ApplyEntityAccess` | Microflow field | Whether entity access rules apply |
| `Forms$Page.AllowedModuleRoles` | Page field | Which module roles can view the page |

### Key BSON Structures

**Project Security:**
```json
{
  "$Type": "Security$ProjectSecurity",
  "SecurityLevel": "CheckEverything",  // CheckNothing | CheckFormsAndMicroflows | CheckEverything
  "AdminUserName": "MxAdmin",
  "AdminPassword": "1",
  "AdminUserRole": "Administrator",
  "CheckSecurity": true,
  "StrictMode": false,
  "StrictPageUrlCheck": true,
  "EnableDemoUsers": true,
  "EnableGuestAccess": false,
  "GuestUserRole": "",
  "UserRoles": [2, { "$Type": "Security$UserRole", "Name": "Administrator", "ModuleRoles": [1, "Administration.Administrator", "Shop.ShopUser", ...], "ManageAllRoles": true, ... }],
  "DemoUsers": [2, { "$Type": "Security$DemoUserImpl", "UserName": "demo_administrator", "Password": "...", "UserRoles": [1, "Administrator"], "Entity": "Administration.Account" }],
  "PasswordPolicySettings": { "MinimumLength": 1, "RequireDigit": false, "RequireMixedCase": false, "RequireSymbol": false }
}
```

**Module Security:**
```json
{
  "$Type": "Security$ModuleSecurity",
  "ModuleRoles": [3, { "$Type": "Security$ModuleRole", "Name": "ShopUser", "Description": "" }]
}
```

**Entity Access Rule** (inside entity's `AccessRules` array):
```json
{
  "$Type": "DomainModels$AccessRule",
  "AllowCreate": false,
  "AllowDelete": true,
  "AllowedModuleRoles": [1, "Shop.ShopUser"],
  "DefaultMemberAccessRights": "ReadWrite",
  "MemberAccesses": [3,
    { "$Type": "DomainModels$MemberAccess", "AccessRights": "ReadWrite", "Attribute": "Shop.Customer.FirstName", "Association": "" },
    { "$Type": "DomainModels$MemberAccess", "AccessRights": "ReadWrite", "Association": "Shop.Order_Customer", "Attribute": "" }
  ],
  "XPathConstraint": "",
  "XPathConstraintCaption": ""
}
```

**Microflow Access** (fields on `Microflows$Microflow`):
```json
{
  "AllowedModuleRoles": [1, "Shop.ShopUser"],
  "ApplyEntityAccess": false
}
```

**Page Access** (field on `Forms$Page`):
```json
{
  "AllowedRoles": [1, "Shop.ShopUser"]
}
```

## Proposed MDL Syntax

### 1. Project Security Level

```sql
-- Set security level
ALTER PROJECT SECURITY LEVEL PRODUCTION;     -- CheckEverything
ALTER PROJECT SECURITY LEVEL PROTOTYPE;      -- CheckFormsAndMicroflows
ALTER PROJECT SECURITY LEVEL OFF;            -- CheckNothing

-- Show current security level
SHOW PROJECT SECURITY;
```

### 2. Module Roles

```sql
-- Create module roles
CREATE MODULE ROLE Shop.ShopUser;
CREATE MODULE ROLE Shop.ShopUser DESCRIPTION 'Regular shop user with read access';
CREATE MODULE ROLE Shop.Admin DESCRIPTION 'Shop administrator';

-- Show module roles
SHOW MODULE ROLES;
SHOW MODULE ROLES IN Shop;

-- Describe a module role (shows access rules across entities/microflows/pages)
DESCRIBE MODULE ROLE Shop.ShopUser;

-- Drop a module role
DROP MODULE ROLE Shop.Admin;
```

### 3. Application User Roles

```sql
-- Create user roles that combine module roles
CREATE USER ROLE Administrator (
    Shop.ShopUser,
    Shop.Admin,
    Administration.Administrator,
    System.Administrator
)
MANAGE ALL ROLES;

CREATE USER ROLE User (
    Shop.ShopUser,
    Administration.User,
    System.User
);

CREATE USER ROLE Api (
    System.User,
    ShopSvc.Api
);

-- Show user roles
SHOW USER ROLES;
DESCRIBE USER ROLE Administrator;

-- Modify user role
ALTER USER ROLE User ADD MODULE ROLES (
    NewModule.NewRole
);

ALTER USER ROLE User REMOVE MODULE ROLES (
    OldModule.OldRole
);

-- Drop user role
DROP USER ROLE Api;
```

### 4. Entity Access Rules

Entity access rules define which module roles can create, read, write, and delete entity instances.

```sql
-- Grant access to an entity for a module role
GRANT Shop.ShopUser ON Shop.Customer (
    CREATE,
    DELETE,
    READ *,           -- Read all attributes/associations
    WRITE *           -- Write all attributes/associations
);

-- Fine-grained member access
GRANT Shop.ShopUser ON Shop.Customer (
    READ (FirstName, LastName, EmailAddress),
    WRITE (FirstName, LastName),
    READ ASSOCIATION (Shop.Order_Customer)
);

-- With XPath constraint (row-level security)
GRANT Shop.ShopUser ON Shop.Customer (
    READ *,
    WRITE (FirstName, LastName)
)
WHERE '[%CurrentUser%/Shop.Customer_Account = System.owner]';

-- Multiple roles with same access
GRANT Shop.ShopUser, Shop.Admin ON Shop.Order (
    CREATE,
    DELETE,
    READ *,
    WRITE *
);

-- Revoke access
REVOKE Shop.ShopUser ON Shop.Customer;

-- Show entity access
SHOW ACCESS ON Shop.Customer;
DESCRIBE ACCESS ON Shop.Customer;
```

### 5. Microflow Access

```sql
-- Grant module roles access to execute a microflow
GRANT EXECUTE ON MICROFLOW Shop.ACT_CreateOrder TO Shop.ShopUser;
GRANT EXECUTE ON MICROFLOW Shop.ACT_CreateOrder TO Shop.ShopUser, Shop.Admin;

-- Apply entity access (microflow respects entity-level security)
ALTER MICROFLOW Shop.ACT_CreateOrder APPLY ENTITY ACCESS;
ALTER MICROFLOW Shop.ACT_CreateOrder NO ENTITY ACCESS;

-- Revoke microflow access
REVOKE EXECUTE ON MICROFLOW Shop.ACT_CreateOrder FROM Shop.ShopUser;

-- Show microflow access
SHOW ACCESS ON MICROFLOW Shop.ACT_CreateOrder;
```

### 6. Page Access

```sql
-- Grant module roles access to view a page
GRANT VIEW ON PAGE Shop.CustomerOverview TO Shop.ShopUser;
GRANT VIEW ON PAGE Shop.CustomerOverview TO Shop.ShopUser, Shop.Admin;

-- Revoke page access
REVOKE VIEW ON PAGE Shop.CustomerOverview FROM Shop.ShopUser;

-- Show page access
SHOW ACCESS ON PAGE Shop.CustomerOverview;
```

### 7. Demo Users

```sql
-- Create demo users (for development/testing)
CREATE DEMO USER 'demo_admin' PASSWORD '1' (
    Administrator
);

CREATE DEMO USER 'demo_user' PASSWORD '1' (
    User,
    Api
);

-- Enable/disable demo users
ALTER PROJECT SECURITY DEMO USERS ON;
ALTER PROJECT SECURITY DEMO USERS OFF;

-- Show demo users
SHOW DEMO USERS;

-- Drop demo user
DROP DEMO USER 'demo_admin';
```

### 8. Inline Security in CREATE Statements

For convenience, security can be declared inline when creating entities, microflows, and pages:

```sql
-- Entity with inline access rules
CREATE PERSISTENT ENTITY Shop.Customer (
    FirstName: String(200),
    LastName: String(200),
    Email: String(200)
)
ACCESS (
    GRANT Shop.ShopUser (CREATE, DELETE, READ *, WRITE *),
    GRANT Shop.Admin (CREATE, DELETE, READ *, WRITE *)
        WHERE '[%CurrentUser%/Shop.Customer_Account = System.owner]'
);

-- Microflow with inline access
CREATE MICROFLOW Shop.ACT_CreateOrder ($Customer: Shop.Customer)
RETURNS Shop.Order AS $Order
ACCESS (Shop.ShopUser, Shop.Admin)
APPLY ENTITY ACCESS
BEGIN
    -- ...
END;

-- Page with inline access
CREATE PAGE Shop.CustomerOverview
(
    Title: 'Customers',
    Layout: Atlas_Core.Atlas_Default
)
ACCESS (Shop.ShopUser, Shop.Admin)
{
    -- widgets ...
};
```

### 9. Bulk Security Operations

```sql
-- Grant a role access to all microflows in a module
GRANT EXECUTE ON ALL MICROFLOWS IN Shop TO Shop.ShopUser;

-- Grant a role access to all pages in a module
GRANT VIEW ON ALL PAGES IN Shop TO Shop.ShopUser;

-- Grant default entity access for all entities in a module
GRANT Shop.ShopUser ON ALL ENTITIES IN Shop (
    READ *
);
```

### 10. DESCRIBE and SHOW for Security Audit

```sql
-- Comprehensive security overview
SHOW PROJECT SECURITY;

-- What can a module role do?
DESCRIBE MODULE ROLE Shop.ShopUser;
-- Output:
--   Module Role: Shop.ShopUser
--   User Roles: Administrator, User
--   Entity Access:
--     Shop.Customer: CREATE, DELETE, READ *, WRITE *
--     Shop.Order: READ *, WRITE (Status)
--   Microflow Access:
--     Shop.ACT_CreateOrder: EXECUTE
--     Shop.ACT_DeleteOrder: EXECUTE
--   Page Access:
--     Shop.CustomerOverview: VIEW
--     Shop.OrderDetail: VIEW

-- What roles can access a specific entity?
SHOW ACCESS ON Shop.Customer;

-- Security matrix for a module
SHOW SECURITY MATRIX IN Shop;
```

## Implementation Phases

### Phase 1: Read Security (SHOW/DESCRIBE)

Read-only commands to inspect existing security configuration. Requires:

- Parse `Security$ProjectSecurity` from MPR
- Parse `Security$ModuleSecurity` from each module
- Read `AccessRules` from entities (already partially implemented in `parser_domainmodel.go`)
- Read `AllowedModuleRoles` from microflows
- Read `AllowedRoles` from pages
- New executor commands: `SHOW PROJECT SECURITY`, `SHOW MODULE ROLES`, `SHOW ACCESS ON`, `DESCRIBE MODULE ROLE`
- Add security data to the catalog for cross-referencing

**Existing code to build on:**
- `sdk/mpr/parser_domainmodel.go:parseAccessRule()` — already parses `AllowCreate`, `AllowDelete`, `XPathConstraint`, module role IDs
- `sdk/domainmodel/domainmodel.go` — `AccessRule` and `MemberAccess` structs defined
- `generated/metamodel/types.go` — `SecurityModuleRole`, `SecurityProjectSecurity`, etc. defined
- `generated/metamodel/enums.go` — `SecuritySecurityLevel` enum defined

**Implementation gaps:**
- No parsing for `Security$ProjectSecurity` or `Security$ModuleSecurity` units
- No parsing of `AllowedModuleRoles` on microflows (field exists in BSON but is not read)
- No parsing of `AllowedRoles` on pages
- Entity access rules are parsed but role IDs are not resolved to names
- Writer serializes `AccessRules: [3]` (empty) — access rules are lost on entity modification

### Phase 2: Module Roles & User Roles (CREATE/ALTER/DROP)

Write support for module and user role management:

- `CREATE MODULE ROLE` — add role to `Security$ModuleSecurity`
- `DROP MODULE ROLE` — remove role (with impact check)
- `CREATE USER ROLE` — add to `Security$ProjectSecurity`
- `ALTER USER ROLE` — modify module role assignments
- `ALTER PROJECT SECURITY LEVEL` — change security level

### Phase 3: Access Rule Management (GRANT/REVOKE)

Write support for access control:

- `GRANT ... ON entity` — add/modify `DomainModels$AccessRule` inside entity
- `REVOKE ... ON entity` — remove access rule
- `GRANT EXECUTE ON MICROFLOW` — set `AllowedModuleRoles` on microflow
- `GRANT VIEW ON PAGE` — set `AllowedRoles` on page
- Inline `ACCESS (...)` in `CREATE ENTITY`, `CREATE MICROFLOW`, `CREATE PAGE`
- Fix writer to preserve existing access rules when modifying entities

### Phase 4: Bulk Operations and Audit

- `GRANT ... ON ALL MICROFLOWS/PAGES/ENTITIES IN module`
- `SHOW SECURITY MATRIX`
- Catalog integration for security queries
- Lint rules for security (e.g., "microflow has no access rules", "entity accessible to all roles")

## Grammar Additions

New tokens for `MDLLexer.g4`:

```antlr
// Security keywords
SECURITY: S E C U R I T Y;
ROLE: R O L E;
ROLES: R O L E S;
GRANT: G R A N T;
REVOKE: R E V O K E;
EXECUTE: E X E C U T E;  // already exists
ACCESS: A C C E S S;
PRODUCTION: P R O D U C T I O N;
PROTOTYPE: P R O T O T Y P E;
MANAGE: M A N A G E;
DEMO: D E M O;
MATRIX: M A T R I X;
APPLY: A P P L Y;
```

New parser rules for `MDLParser.g4`:

```antlr
// Security statements
securityStatement
    : createModuleRoleStmt
    | dropModuleRoleStmt
    | createUserRoleStmt
    | alterUserRoleStmt
    | dropUserRoleStmt
    | grantEntityAccessStmt
    | revokeEntityAccessStmt
    | grantMicroflowAccessStmt
    | revokeMicroflowAccessStmt
    | grantPageAccessStmt
    | revokePageAccessStmt
    | alterProjectSecurityStmt
    | showProjectSecurityStmt
    | showModuleRolesStmt
    | showAccessStmt
    | describeModuleRoleStmt
    | createDemoUserStmt
    | dropDemoUserStmt
    ;

createModuleRoleStmt
    : CREATE MODULE ROLE qualifiedName (DESCRIPTION STRING_LITERAL)?
    ;

grantEntityAccessStmt
    : GRANT moduleRoleList ON qualifiedName LPAREN entityAccessRights RPAREN
      (WHERE STRING_LITERAL)?
    ;

entityAccessRights
    : entityAccessRight (COMMA entityAccessRight)*
    ;

entityAccessRight
    : CREATE
    | DELETE
    | READ STAR
    | READ LPAREN identifierList RPAREN
    | WRITE STAR
    | WRITE LPAREN identifierList RPAREN
    | READ ASSOCIATION LPAREN qualifiedNameList RPAREN
    ;
```

## Design Decisions

### SQL-style GRANT/REVOKE

The `GRANT`/`REVOKE` syntax was chosen because:
1. It's familiar to anyone who knows SQL security
2. It maps cleanly to the Mendix access model (roles, permissions, objects)
3. It reads naturally: "GRANT ShopUser ON Customer (READ *, WRITE *)"
4. It's the standard way to express authorization in declarative languages

### Inline ACCESS Clause

The inline `ACCESS (...)` clause on CREATE statements enables one-shot provisioning:
- Define and secure an entity in one statement
- No separate GRANT needed for new entities
- Consistent with how Studio Pro presents security alongside entity configuration

### Module Roles as First-Class Citizens

Module roles are the fundamental unit of access in Mendix:
- Entity access rules reference module roles
- Microflow access references module roles
- Page access references module roles
- User roles are composites of module roles

MDL reflects this by making `DESCRIBE MODULE ROLE` the primary audit tool.

### XPath Constraints as Strings

Entity access XPath constraints are kept as string literals (`WHERE '...'`) rather than parsed. This matches how Mendix stores them and avoids needing an XPath parser in MDL.

## Example: Complete Security Setup

```sql
-- 1. Set security level
ALTER PROJECT SECURITY LEVEL PRODUCTION;

-- 2. Create module roles
CREATE MODULE ROLE Shop.Customer DESCRIPTION 'Customer self-service access';
CREATE MODULE ROLE Shop.Employee DESCRIPTION 'Shop employee with full access';
CREATE MODULE ROLE Shop.Manager DESCRIPTION 'Shop manager with admin access';

-- 3. Create user roles
CREATE USER ROLE Customer (
    Shop.Customer,
    System.User
);

CREATE USER ROLE Employee (
    Shop.Customer,
    Shop.Employee,
    Administration.User,
    System.User
);

CREATE USER ROLE Manager (
    Shop.Customer,
    Shop.Employee,
    Shop.Manager,
    Administration.Administrator,
    System.Administrator
)
MANAGE ALL ROLES;

-- 4. Entity access
GRANT Shop.Customer ON Shop.Order (
    CREATE,
    READ *,
    WRITE (Status, Notes)
)
WHERE '[%CurrentUser%/Shop.Order_Customer = System.owner]';

GRANT Shop.Employee ON Shop.Order (
    CREATE,
    DELETE,
    READ *,
    WRITE *
);

GRANT Shop.Manager ON Shop.Order (
    CREATE,
    DELETE,
    READ *,
    WRITE *
);

-- 5. Microflow access
GRANT EXECUTE ON MICROFLOW Shop.ACT_PlaceOrder TO Shop.Customer, Shop.Employee;
GRANT EXECUTE ON MICROFLOW Shop.ACT_CancelOrder TO Shop.Employee, Shop.Manager;
GRANT EXECUTE ON MICROFLOW Shop.ACT_DeleteOrder TO Shop.Manager;

-- 6. Page access
GRANT VIEW ON PAGE Shop.MyOrders TO Shop.Customer;
GRANT VIEW ON PAGE Shop.AllOrders TO Shop.Employee, Shop.Manager;
GRANT VIEW ON PAGE Shop.AdminDashboard TO Shop.Manager;

-- 7. Demo users
CREATE DEMO USER 'demo_customer' PASSWORD 'customer1' (Customer);
CREATE DEMO USER 'demo_employee' PASSWORD 'employee1' (Employee);
CREATE DEMO USER 'demo_manager' PASSWORD 'manager1' (Manager);
ALTER PROJECT SECURITY DEMO USERS ON;
```
