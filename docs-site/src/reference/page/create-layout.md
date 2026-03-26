# CREATE LAYOUT

## Synopsis

```sql
CREATE LAYOUT module.Name
{
    widget_tree
}
```

## Description

Creates a page layout in the specified module. Layouts define the overall structure of pages -- they typically include a header, navigation, content placeholder, and footer. Pages reference a layout via the `Layout` property.

Layout creation in MDL has limited support. Most Mendix projects use layouts provided by the Atlas UI module (e.g., `Atlas_Core.Atlas_Default`, `Atlas_Core.PopupLayout`) rather than creating custom layouts through MDL. For advanced layout customization, use Mendix Studio Pro.

### Common Atlas Layouts

These layouts are available in most Mendix projects that include Atlas Core:

| Layout | Description |
|--------|-------------|
| `Atlas_Core.Atlas_Default` | Standard responsive page with sidebar navigation |
| `Atlas_Core.Atlas_TopBar` | Page with top navigation bar |
| `Atlas_Core.PopupLayout` | Modal popup dialog |
| `Atlas_Core.Atlas_Default_NativePhone` | Native mobile layout |

## Parameters

`module.Name`
:   The qualified name of the layout (`Module.LayoutName`).

## Examples

Reference an existing layout when creating a page:

```sql
CREATE PAGE MyModule.Dashboard
(
    Title: 'Dashboard',
    Layout: Atlas_Core.Atlas_Default
)
{
    CONTAINER cntMain {
        DYNAMICTEXT txtWelcome (Attribute: WelcomeMessage)
    }
};
```

Reference a popup layout for a dialog page:

```sql
CREATE PAGE MyModule.ConfirmDelete
(
    Params: { $Item: MyModule.Item },
    Title: 'Confirm Delete',
    Layout: Atlas_Core.PopupLayout
)
{
    DATAVIEW dvItem (DataSource: $Item) {
        DYNAMICTEXT txtMessage (Attribute: Name)
        FOOTER footer1 {
            ACTIONBUTTON btnDelete (Caption: 'Delete', Action: DELETE, ButtonStyle: Danger)
            ACTIONBUTTON btnCancel (Caption: 'Cancel', Action: CANCEL_CHANGES)
        }
    }
};
```

## See Also

[CREATE PAGE](create-page.md), [SHOW PAGES](/reference/query/show-pages.md)
