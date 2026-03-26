# MOVE

## Synopsis

    MOVE document_type qualified_name TO module_name
    MOVE document_type qualified_name TO FOLDER 'folder_path'
    MOVE document_type qualified_name TO FOLDER 'folder_path' IN module_name

## Description

Moves a document to a different folder or module. Supported document types are `PAGE`, `MICROFLOW`, `NANOFLOW`, `SNIPPET`, `ENUMERATION`, and `ENTITY`. When moving to a folder, missing intermediate folders are created automatically. Cross-module moves change the qualified name of the document, which may break by-name references elsewhere in the project.

Entity moves only support moving to a module (not to a folder), because entities are embedded in domain model documents.

## Parameters

**document_type**
: The type of document to move. One of: `PAGE`, `MICROFLOW`, `NANOFLOW`, `SNIPPET`, `ENUMERATION`, `ENTITY`.

**qualified_name**
: The current `Module.Name` of the document to move.

**module_name**
: The target module. When specified without `FOLDER`, moves the document to the module root.

**folder_path**
: The target folder path within the module. Use `/` for nested paths (e.g., `'Orders/Processing'`). Enclosed in single quotes.

## Examples

### Move a page to a folder in the same module

```sql
MOVE PAGE MyModule.CustomerEdit TO FOLDER 'Customers';
```

### Move a microflow to a nested folder

```sql
MOVE MICROFLOW MyModule.ACT_ProcessOrder TO FOLDER 'Orders/Processing';
```

### Move a snippet to a different module

```sql
MOVE SNIPPET OldModule.NavigationMenu TO Common;
```

### Move an entity to a different module

```sql
MOVE ENTITY OldModule.Customer TO NewModule;
```

### Move a page to a folder in a different module

```sql
MOVE PAGE OldModule.CustomerPage TO FOLDER 'Screens' IN NewModule;
```

### Check impact before a cross-module move

```sql
SHOW IMPACT OF OldModule.Customer;
MOVE ENTITY OldModule.Customer TO NewModule;
```

## See Also

[CREATE MODULE](create-module.md), [CREATE FOLDER](create-folder.md)
