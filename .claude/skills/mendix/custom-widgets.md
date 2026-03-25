---
name: mendix-custom-widgets
description: Use when writing MDL for GALLERY, COMBOBOX, or third-party pluggable widgets in CREATE PAGE / ALTER PAGE statements. Covers built-in widget syntax, child slots (TEMPLATE/FILTER), and adding new custom widgets via .def.json.
---

# Custom & Pluggable Widgets in MDL

## Built-in Pluggable Widgets

### GALLERY

Card-layout list with optional template content and filters.

```sql
GALLERY galleryName (
  DataSource: DATABASE FROM Module.Entity SORT BY Name ASC,
  Selection: Single | Multiple | None
) {
  TEMPLATE template1 {
    DYNAMICTEXT title (Content: '{1}', ContentParams: [{1} = Name], RenderMode: H4)
    DYNAMICTEXT info  (Content: '{1}', ContentParams: [{1} = Email])
  }
  FILTER filter1 {
    TEXTFILTER   searchName  (Attribute: Name)
    NUMBERFILTER searchScore (Attribute: Score)
    DROPDOWNFILTER searchStatus (Attribute: Status)
    DATEFILTER   searchDate  (Attribute: CreatedAt)
  }
}
```

- `TEMPLATE` block → mapped to `content` property (child widgets rendered per row)
- `FILTER` block → mapped to `filtersPlaceholder` property (shown above list)
- `Selection: None` omits the selection property (default if omitted)

### COMBOBOX

Two modes depending on the attribute type:

```sql
-- Enumeration mode (Attribute is an enum)
COMBOBOX cbStatus (Label: 'Status', Attribute: Status)

-- Association mode (Attribute is an association)
COMBOBOX cmbCustomer (
  Label: 'Customer',
  Attribute: Order_Customer,
  DataSource: DATABASE Module.Customer,
  CaptionAttribute: Name
)
```

- Engine detects association mode when `DataSource` or `CaptionAttribute` is present
- `CaptionAttribute` is the display attribute on the **target** entity

## Adding a Third-Party Widget

### Step 1 — Extract .def.json from .mpk

```bash
mxcli widget extract --mpk widgets/MyWidget.mpk
# Output: .mxcli/widgets/mywidget.def.json

# Override MDL keyword
mxcli widget extract --mpk widgets/MyWidget.mpk --mdl-name MYWIDGET
```

Extraction auto-infers operations from XML property types:

| XML Type | Operation | MDL Source Key |
|----------|-----------|----------------|
| attribute | attribute | `Attribute` |
| association | association | `Association` |
| datasource | datasource | `DataSource` |
| selection | selection | `Selection` |
| widgets | widgets (child slot) | container name |
| boolean/string/enumeration | primitive | hardcoded `Value` |

### Step 2 — Place .def.json

```
project/.mxcli/widgets/mywidget.def.json   ← project scope
~/.mxcli/widgets/mywidget.def.json         ← global scope
```

Project definitions override global ones with the same MDL name.

### Step 3 — Add template JSON

Copy a Studio Pro-created widget JSON to:
```
project/.mxcli/widgets/mywidget.json
```

Then set `"templateFile": "mywidget.json"` in the .def.json.

**CRITICAL**: Template must include both `type` (PropertyTypes) and `object` (default WidgetObject). Extract from a real Studio Pro MPR — do NOT generate programmatically. Mismatched structure causes CE0463.

### Step 4 — Use in MDL

```sql
MYWIDGET myWidget1 (DataSource: DATABASE Module.Entity, Attribute: Name)
```

## .def.json Reference

```json
{
  "widgetId":        "com.vendor.widget.web.mywidget.MyWidget",
  "mdlName":         "MYWIDGET",
  "templateFile":    "mywidget.json",
  "defaultEditable": "Always",
  "propertyMappings": [
    {"propertyKey": "datasource",  "source": "DataSource", "operation": "datasource"},
    {"propertyKey": "attribute",   "source": "Attribute",  "operation": "attribute"},
    {"propertyKey": "someFlag",    "value":  "true",       "operation": "primitive"}
  ],
  "childSlots": [
    {"propertyKey": "content", "mdlContainer": "TEMPLATE", "operation": "widgets"}
  ],
  "modes": [
    {
      "name": "association",
      "condition": "hasDataSource",
      "propertyMappings": [
        {"propertyKey": "optionsSource", "value": "association", "operation": "primitive"},
        {"propertyKey": "assoc",         "source": "Attribute",   "operation": "association"},
        {"propertyKey": "assocDS",       "source": "DataSource",  "operation": "datasource"}
      ]
    },
    {
      "name": "default",
      "propertyMappings": [
        {"propertyKey": "attr", "source": "Attribute", "operation": "attribute"}
      ]
    }
  ]
}
```

**Mode conditions**: `hasDataSource` | `hasProp:PropertyKey`
Modes are evaluated in order — first match wins; no condition = default fallback.

## Verify & Debug

```bash
# List registered widgets
mxcli widget list -p App.mpr

# Check after creating a page
mxcli check script.mdl -p App.mpr --references

# Full mx check (catches CE0463)
~/.mxcli/mxbuild/*/modeler/mx check App.mpr

# Debug CE0463 — compare NDSL dumps
mxcli bson dump -p App.mpr --type page --object "Module.PageName" --format ndsl
```

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| CE0463 after page creation | Template version mismatch — extract fresh template from Studio Pro MPR |
| Widget not recognized | Check `mxcli widget list`; .def.json MDL name must match grammar keyword |
| TEMPLATE content missing | Widget needs `childSlots` entry with `"mdlContainer": "TEMPLATE"` |
| Association COMBOBOX shows enum behavior | Add `DataSource` or `CaptionAttribute` to trigger association mode |
| COMBOBOX CE1613 after page creation | Known engine bug — ComboBox BSON serialization writes wrong pointer type; track issue separately |
| Custom widget not found | Place .def.json in `.mxcli/widgets/` inside the project directory |
