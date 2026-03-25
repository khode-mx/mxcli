# Page Syntax V3 - Implementation Reference

## Status: Implemented ✅

> **Note:** This is now the recommended syntax for creating pages in MDL.
> V1 (BEGIN/END) and V2 (->) syntaxes are still supported for backward compatibility.

---

**Refactor MDL Page Grammar for Agent-Friendly Syntax**

Refactor the MDL page grammar to use a more consistent, explicit syntax that is optimized for agentic coding (LLM code generation). The goal is to eliminate syntax variations and make the language predictable with a single pattern for all constructs.

## Core Principle

All widgets follow this pattern:
```
WIDGETTYPE name (Prop: Value, Prop: Value) { children }
```

## Changes to Implement

### 1. Unified Page Header Block

All page metadata goes in a single `()` block after the qualified name:

**Before:**
```
CREATE PAGE PgTest.Example (
  $Order: PgTest.Order
)
  TITLE 'Example'
  LAYOUT Atlas_Core.Atlas_Default
  URL 'example'
{ ... }
```

**After:**
```
CREATE PAGE PgTest.Example
(
  Params: {
    $Order: PgTest.Order,
    $Customer: PgTest.Customer
  },
  Title: 'Example',
  Layout: Atlas_Core.Atlas_Default,
  Url: 'example'
)
{ ... }
```

- `Params:` block is optional, only present when page has parameters
- All metadata properties use consistent `Key: Value` syntax

### 2. DataSource Property

Replace `->` and `-> DATABASE` with explicit `DataSource:` property. Always include the source type keyword.

**Before:**
```
DATAVIEW OrderForm -> $Order { ... }
DATAGRID ProductGrid -> DATABASE PgTest.Product { ... }
GALLERY productGallery -> DATABASE PgTest.Product { ... }
```

**After:**
```
DATAVIEW OrderForm (DataSource: $Order) { ... }
DATAGRID ProductGrid (DataSource: DATABASE PgTest.Product) { ... }
GALLERY productGallery (DataSource: DATABASE PgTest.Product) { ... }
```

DataSource type keywords:
- `DATABASE <Entity>` — database source
- `MICROFLOW <QualifiedName>` — microflow source
- `NANOFLOW <QualifiedName>` — nanoflow source
- `ASSOCIATION <Path>` — association source
- `$ParamName` — parameter reference (no keyword needed, already typed)

### 3. Attribute Property for Attribute Bindings

Replace `->` with explicit `Attribute:` property for input widgets.

**Before:**
```
TEXTBOX txtName 'Product Name' -> Name
CHECKBOX cbIsActive 'Is Active' -> IsActive
DATEPICKER dpDate 'Created On' -> CreateDate
```

**After:**
```
TEXTBOX txtName (Label: 'Product Name', Attribute: Name)
CHECKBOX cbIsActive (Label: 'Is Active', Attribute: IsActive)
DATEPICKER dpDate (Label: 'Created On', Attribute: CreateDate)
```

### 4. Action Property for Buttons and Navigation

Replace `->` with explicit `Action:` property. Same syntax for ACTIONBUTTON and navigation ITEM.

**Before:**
```
ACTIONBUTTON btnSave 'Save' -> SAVE_CHANGES (ButtonStyle: Success)
ACTIONBUTTON btnRun 'Run' -> MICROFLOW PgTest.DoSomething
ITEM -> SHOW_PAGE 'PageTemplates.Customer_Overview' { ... }
```

**After:**
```
ACTIONBUTTON btnSave (Caption: 'Save', Action: SAVE_CHANGES, ButtonStyle: Success)
ACTIONBUTTON btnRun (Caption: 'Run', Action: MICROFLOW PgTest.DoSomething)
ITEM item1 (Action: SHOW_PAGE PageTemplates.Customer_Overview) { ... }
```

Action keywords:
- `SHOW_PAGE <PageRef>`
- `MICROFLOW <QualifiedName>`
- `NANOFLOW <QualifiedName>`
- `OPEN_LINK <Url>`
- `SAVE_CHANGES`
- `CANCEL_CHANGES`
- `CLOSE_PAGE`
- `DELETE_OBJECT`
- `CREATE_OBJECT`
- `SIGN_OUT`
- `CALL_WORKFLOW`
- `SYNCHRONIZE`
- (and others as defined in Mendix)

### 5. Consistent Property Syntax

All properties use `(Prop: Value, Prop: Value)` syntax. No bare strings or keyword-value pairs outside parentheses.

**Before:**
```
DYNAMICTEXT text1 'Hello' (RenderMode: H2)
TITLE 'Page Title'
LAYOUT Atlas_Core.Atlas_Default
```

**After:**
```
DYNAMICTEXT text1 (Content: 'Hello', RenderMode: H2)
Title: 'Page Title'   // inside header block
Layout: Atlas_Core.Atlas_Default   // inside header block
```

### 6. DATAGRID Column Syntax

Columns use `Attribute:` for attribute binding, `Caption:` for header, and support capability properties and nested children (filters, custom content).

**Before:**
```
DATAGRID ProductGrid -> DATABASE PgTest.Product
  WHERE [Stock < 10]
  ORDER BY Stock ASC
{
  Columns: {
    COLUMN -> Name
    COLUMN -> Stock
  }
}
```

**After:**
```
DATAGRID ProductGrid (
  DataSource: DATABASE FROM PgTest.Product WHERE [Stock < 10] SORT BY Stock ASC
) {
  COLUMN colName (Attribute: Name, Caption: 'Name')
  COLUMN colStock (Attribute: Stock, Caption: 'Stock', CanSort: true) {
    NUMBERFILTER filterStock (Attribute: Stock)
  }
  COLUMN colActions (Caption: 'Actions') {
    ACTIONBUTTON btnEdit (Caption: 'Edit', Action: SHOW_PAGE PgTest.Product_Edit)
  }
}
```

- `WHERE` and `SORT BY` are inline in the `DataSource:` expression, matching RETRIEVE syntax
- Columns are direct children, no `Columns: { }` wrapper
- Columns can have nested filter widgets or custom content

### 7. Filter Bindings

Filters use `Attribute:` property.

**Before:**
```
TEXTFILTER searchName -> Name
```

**After:**
```
TEXTFILTER searchName (Attribute: Name)
NUMBERFILTER filterStock (Attribute: Stock)
DATEFILTER filterDate (Attribute: OrderDate)
```

### 8. GALLERY Syntax

**After:**
```
GALLERY productGallery (DataSource: DATABASE PgTest.Product, Selection: Single) {
  FILTER {
    TEXTFILTER searchName (Attribute: Name)
  }
  TEMPLATE {
    DYNAMICTEXT prodName (Content: '{Name}', RenderMode: H4)
    DYNAMICTEXT prodCode (Content: 'SKU: {Code}')
    DYNAMICTEXT prodPrice (Content: 'Price: {Price}')
  }
}
```

### 9. Widget Names Always Required

Every widget must have a name identifier. Do not make names optional, as widgets may be referenced from microflows, nanoflows, or other external locations.

## Complete Example

**Before:**
```
CREATE PAGE PgTest.P015_Product_EditFull (
  $Product: PgTest.Product
)
  TITLE 'Edit Product Details'
  LAYOUT Atlas_Core.PopupLayout
{
  DATAVIEW ProductForm -> $Product {
    TEXTBOX txtName 'Product Name' -> Name
    TEXTBOX txtPrice 'Unit Price' -> Price
    CHECKBOX cbIsActive 'Product is Active' -> IsActive
    DATEPICKER dpCreatedDate 'Created On' -> CreateDate
    FOOTER {
      ACTIONBUTTON btnSave 'Save' -> SAVE_CHANGES (ButtonStyle: Success)
      ACTIONBUTTON btnCancel 'Cancel' -> CANCEL_CHANGES (ButtonStyle: Default)
    }
  }
}
```

**After:**
```
CREATE PAGE PgTest.P015_Product_EditFull
(
  Params: {
    $Product: PgTest.Product
  },
  Title: 'Edit Product Details',
  Layout: Atlas_Core.PopupLayout
)
{
  DATAVIEW ProductForm (DataSource: $Product) {
    TEXTBOX txtName (Label: 'Product Name', Attribute: Name)
    TEXTBOX txtPrice (Label: 'Unit Price', Attribute: Price)
    CHECKBOX cbIsActive (Label: 'Product is Active', Attribute: IsActive)
    DATEPICKER dpCreatedDate (Label: 'Created On', Attribute: CreateDate)
    FOOTER {
      ACTIONBUTTON btnSave (Caption: 'Save', Action: SAVE_CHANGES, ButtonStyle: Success)
      ACTIONBUTTON btnCancel (Caption: 'Cancel', Action: CANCEL_CHANGES)
    }
  }
}
```

## Implementation Notes

- Update the grammar/parser to support the new syntax ✅
- Update any code generation or serialization logic ✅
- Ensure the linter validates the new syntax ✅

## Implementation Files

| File | Purpose |
|------|---------|
| `mdl/grammar/MDLLexer.g4` | Tokens for V3 keywords |
| `mdl/grammar/MDLParser.g4` | V3 widget and page grammar rules |
| `mdl/ast/ast_page_v3.go` | AST types for V3 pages and widgets |
| `mdl/visitor/visitor_page_v3.go` | Parser to AST conversion for V3 |
| `mdl/executor/cmd_pages_builder_v3.go` | V3 widget builders |
| `mdl/executor/cmd_pages_describe.go` | DESCRIBE output in V3 format |

## Verification

```bash
# Parse V3 syntax
./bin/mxcli check mdl-examples/doctype-tests/04-page-examples-v3.mdl

# Execute against project
./bin/mxcli -p app.mpr -c "execute script 'mdl-examples/doctype-tests/04-page-examples-v3.mdl'"

# Verify in Studio Pro
reference/mxbuild/modeler/mx check app.mpr
```

## Example Files

See `mdl-examples/doctype-tests/04-page-examples-v3.mdl` for comprehensive examples.

## Skill Documentation

See `.claude/skills/mendix/`:
- `create-page.md` - V3 syntax reference
- `overview-pages.md` - CRUD page patterns
- `master-detail-pages.md` - Master-detail patterns
