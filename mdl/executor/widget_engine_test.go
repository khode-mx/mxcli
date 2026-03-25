// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"encoding/json"
	"testing"

	"github.com/mendixlabs/mxcli/sdk/pages"
	"go.mongodb.org/mongo-driver/bson"
)

func TestWidgetDefinitionJSONRoundTrip(t *testing.T) {
	original := WidgetDefinition{
		WidgetID:         "com.mendix.widget.web.combobox.Combobox",
		MDLName:          "COMBOBOX",
		TemplateFile:     "combobox.json",
		DefaultEditable:  "Always",
		DefaultSelection: "Single",
		PropertyMappings: []PropertyMapping{
			{PropertyKey: "attributeEnumeration", Source: "Attribute", Operation: "attribute"},
			{PropertyKey: "optionsSourceType", Value: "enumeration", Operation: "primitive"},
		},
		ChildSlots: []ChildSlotMapping{
			{PropertyKey: "content", MDLContainer: "TEMPLATE", Operation: "widgets"},
		},
		Modes: map[string]WidgetMode{
			"association": {
				Condition:   "DataSource != nil",
				Description: "Association-based ComboBox with datasource",
				PropertyMappings: []PropertyMapping{
					{PropertyKey: "attributeAssociation", Source: "Attribute", Operation: "association"},
					{PropertyKey: "optionsSourceType", Value: "association", Operation: "primitive"},
				},
				ChildSlots: []ChildSlotMapping{
					{PropertyKey: "menuContent", MDLContainer: "MENU", Operation: "widgets"},
				},
			},
		},
	}

	encoded, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal WidgetDefinition: %v", err)
	}

	var decoded WidgetDefinition
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("failed to unmarshal WidgetDefinition: %v", err)
	}

	// Verify top-level fields
	if decoded.WidgetID != original.WidgetID {
		t.Errorf("WidgetID: got %q, want %q", decoded.WidgetID, original.WidgetID)
	}
	if decoded.MDLName != original.MDLName {
		t.Errorf("MDLName: got %q, want %q", decoded.MDLName, original.MDLName)
	}
	if decoded.DefaultEditable != original.DefaultEditable {
		t.Errorf("DefaultEditable: got %q, want %q", decoded.DefaultEditable, original.DefaultEditable)
	}
	if decoded.DefaultSelection != original.DefaultSelection {
		t.Errorf("DefaultSelection: got %q, want %q", decoded.DefaultSelection, original.DefaultSelection)
	}

	// Verify property mappings
	if len(decoded.PropertyMappings) != len(original.PropertyMappings) {
		t.Fatalf("PropertyMappings count: got %d, want %d", len(decoded.PropertyMappings), len(original.PropertyMappings))
	}
	if decoded.PropertyMappings[0].Operation != "attribute" {
		t.Errorf("PropertyMappings[0].Operation: got %q, want %q", decoded.PropertyMappings[0].Operation, "attribute")
	}

	// Verify child slots
	if len(decoded.ChildSlots) != 1 {
		t.Fatalf("ChildSlots count: got %d, want 1", len(decoded.ChildSlots))
	}
	if decoded.ChildSlots[0].MDLContainer != "TEMPLATE" {
		t.Errorf("ChildSlots[0].MDLContainer: got %q, want %q", decoded.ChildSlots[0].MDLContainer, "TEMPLATE")
	}

	// Verify modes
	assocMode, ok := decoded.Modes["association"]
	if !ok {
		t.Fatal("Modes[\"association\"] not found")
	}
	if assocMode.Condition != "DataSource != nil" {
		t.Errorf("Mode condition: got %q, want %q", assocMode.Condition, "DataSource != nil")
	}
	if len(assocMode.PropertyMappings) != 2 {
		t.Errorf("Mode PropertyMappings count: got %d, want 2", len(assocMode.PropertyMappings))
	}
	if len(assocMode.ChildSlots) != 1 {
		t.Errorf("Mode ChildSlots count: got %d, want 1", len(assocMode.ChildSlots))
	}
}

func TestWidgetDefinitionJSONOmitsEmptyOptionalFields(t *testing.T) {
	minimal := WidgetDefinition{
		WidgetID:        "com.example.Widget",
		MDLName:         "MYWIDGET",
		TemplateFile:    "mywidget.json",
		DefaultEditable: "Always",
	}

	encoded, err := json.Marshal(minimal)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &raw); err != nil {
		t.Fatalf("failed to unmarshal to map: %v", err)
	}

	// defaultSelection should be omitted when empty
	if _, exists := raw["defaultSelection"]; exists {
		t.Error("defaultSelection should be omitted when empty")
	}
}

func TestOperationRegistryLookupFound(t *testing.T) {
	reg := NewOperationRegistry()

	builtinOps := []string{"attribute", "association", "primitive", "datasource", "widgets"}
	for _, name := range builtinOps {
		fn := reg.Lookup(name)
		if fn == nil {
			t.Errorf("Lookup(%q) returned nil, want non-nil", name)
		}
	}
}

func TestOperationRegistryLookupNotFound(t *testing.T) {
	reg := NewOperationRegistry()

	fn := reg.Lookup("nonexistent")
	if fn != nil {
		t.Error("Lookup(\"nonexistent\") should return nil")
	}
}

func TestOperationRegistryCustomRegistration(t *testing.T) {
	reg := NewOperationRegistry()

	called := false
	reg.Register("custom", func(obj bson.D, propTypeIDs map[string]pages.PropertyTypeIDEntry, propertyKey string, ctx *BuildContext) bson.D {
		called = true
		return obj
	})

	fn := reg.Lookup("custom")
	if fn == nil {
		t.Fatal("Lookup(\"custom\") returned nil after Register")
	}

	fn(bson.D{}, nil, "test", &BuildContext{})
	if !called {
		t.Error("custom operation was not called")
	}
}

func TestSetChildWidgets(t *testing.T) {
	val := bson.D{
		{Key: "PrimitiveValue", Value: ""},
		{Key: "Widgets", Value: bson.A{int32(2)}},
		{Key: "XPathConstraint", Value: ""},
	}

	childWidgets := []bson.D{
		{{Key: "$Type", Value: "Forms$TextBox"}, {Key: "Name", Value: "textBox1"}},
		{{Key: "$Type", Value: "Forms$TextBox"}, {Key: "Name", Value: "textBox2"}},
	}

	updated := setChildWidgets(val, childWidgets)

	// Find Widgets field
	for _, elem := range updated {
		if elem.Key == "Widgets" {
			arr, ok := elem.Value.(bson.A)
			if !ok {
				t.Fatal("Widgets value is not bson.A")
			}
			// Should have version marker + 2 widgets
			if len(arr) != 3 {
				t.Errorf("Widgets array length: got %d, want 3", len(arr))
			}
			// First element should be version marker
			if marker, ok := arr[0].(int32); !ok || marker != 2 {
				t.Errorf("Widgets[0]: got %v, want int32(2)", arr[0])
			}
			return
		}
	}
	t.Error("Widgets field not found in result")
}
