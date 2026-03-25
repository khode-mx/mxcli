// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mendixlabs/mxcli/sdk/widgets/definitions"
)

func TestRegistryLoadsAllEmbeddedDefinitions(t *testing.T) {
	reg, err := NewWidgetRegistry()
	if err != nil {
		t.Fatalf("NewWidgetRegistry() error: %v", err)
	}

	// We expect 7 embedded definitions
	if got := reg.Count(); got != 7 {
		t.Errorf("registry count = %d, want 7", got)
	}
}

func TestRegistryGetByMDLName(t *testing.T) {
	reg, err := NewWidgetRegistry()
	if err != nil {
		t.Fatalf("NewWidgetRegistry() error: %v", err)
	}

	tests := []struct {
		mdlName  string
		widgetID string
	}{
		{"COMBOBOX", "com.mendix.widget.web.combobox.Combobox"},
		{"GALLERY", "com.mendix.widget.web.gallery.Gallery"},
		{"DATAGRID", "com.mendix.widget.web.datagrid.Datagrid"},
		{"TEXTFILTER", "com.mendix.widget.web.datagridtextfilter.DatagridTextFilter"},
		{"NUMBERFILTER", "com.mendix.widget.web.datagridnumberfilter.DatagridNumberFilter"},
		{"DROPDOWNFILTER", "com.mendix.widget.web.datagriddropdownfilter.DatagridDropdownFilter"},
		{"DATEFILTER", "com.mendix.widget.web.datagriddatefilter.DatagridDateFilter"},
	}

	for _, tt := range tests {
		t.Run(tt.mdlName, func(t *testing.T) {
			def, ok := reg.Get(tt.mdlName)
			if !ok {
				t.Fatalf("Get(%q) not found", tt.mdlName)
			}
			if def.WidgetID != tt.widgetID {
				t.Errorf("WidgetID = %q, want %q", def.WidgetID, tt.widgetID)
			}
		})
	}
}

func TestRegistryGetCaseInsensitive(t *testing.T) {
	reg, err := NewWidgetRegistry()
	if err != nil {
		t.Fatalf("NewWidgetRegistry() error: %v", err)
	}

	// Should work with any case
	for _, name := range []string{"combobox", "ComboBox", "COMBOBOX", "Combobox"} {
		def, ok := reg.Get(name)
		if !ok {
			t.Errorf("Get(%q) not found", name)
			continue
		}
		if def.MDLName != "COMBOBOX" {
			t.Errorf("Get(%q).MDLName = %q, want COMBOBOX", name, def.MDLName)
		}
	}
}

func TestRegistryGetUnknownWidget(t *testing.T) {
	reg, err := NewWidgetRegistry()
	if err != nil {
		t.Fatalf("NewWidgetRegistry() error: %v", err)
	}

	_, ok := reg.Get("NONEXISTENT")
	if ok {
		t.Error("Get(NONEXISTENT) should return false")
	}
}

func TestRegistryGetByWidgetID(t *testing.T) {
	reg, err := NewWidgetRegistry()
	if err != nil {
		t.Fatalf("NewWidgetRegistry() error: %v", err)
	}

	def, ok := reg.GetByWidgetID("com.mendix.widget.web.gallery.Gallery")
	if !ok {
		t.Fatal("GetByWidgetID(Gallery) not found")
	}
	if def.MDLName != "GALLERY" {
		t.Errorf("MDLName = %q, want GALLERY", def.MDLName)
	}
}

func TestAllEmbeddedDefinitionsAreValidJSON(t *testing.T) {
	entries, err := definitions.EmbeddedFS.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir error: %v", err)
	}

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".def.json") {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			data, err := definitions.EmbeddedFS.ReadFile(entry.Name())
			if err != nil {
				t.Fatalf("ReadFile error: %v", err)
			}

			var def WidgetDefinition
			if err := json.Unmarshal(data, &def); err != nil {
				t.Fatalf("JSON unmarshal error: %v", err)
			}

			// Validate required fields
			if def.WidgetID == "" {
				t.Error("widgetId is empty")
			}
			if def.MDLName == "" {
				t.Error("mdlName is empty")
			}
			if def.TemplateFile == "" {
				t.Error("templateFile is empty")
			}

			// Must have either propertyMappings, modes, or childSlots
			hasMappings := len(def.PropertyMappings) > 0
			hasModes := len(def.Modes) > 0
			hasSlots := len(def.ChildSlots) > 0
			if !hasMappings && !hasModes && !hasSlots {
				t.Error("definition has no propertyMappings, modes, or childSlots")
			}
		})
	}
}

func TestRegistryLoadUserDefinitions(t *testing.T) {
	reg, err := NewWidgetRegistry()
	if err != nil {
		t.Fatalf("NewWidgetRegistry() error: %v", err)
	}

	// Create a temp directory with a custom definition
	tmpDir := t.TempDir()
	widgetsDir := filepath.Join(tmpDir, ".mxcli", "widgets")
	if err := os.MkdirAll(widgetsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll error: %v", err)
	}

	customDef := `{
		"widgetId": "com.example.custom.MyWidget",
		"mdlName": "MYWIDGET",
		"templateFile": "mywidget.json",
		"defaultEditable": "Always",
		"propertyMappings": [
			{"propertyKey": "value", "source": "Attribute", "operation": "attribute"}
		]
	}`

	defPath := filepath.Join(widgetsDir, "mywidget.def.json")
	if err := os.WriteFile(defPath, []byte(customDef), 0o644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	// Create a fake project file in the temp directory
	projectPath := filepath.Join(tmpDir, "App.mpr")

	// Load user definitions
	if err := reg.LoadUserDefinitions(projectPath); err != nil {
		t.Fatalf("LoadUserDefinitions error: %v", err)
	}

	// The custom widget should now be found
	def, ok := reg.Get("MYWIDGET")
	if !ok {
		t.Fatal("custom widget MYWIDGET not found after LoadUserDefinitions")
	}
	if def.WidgetID != "com.example.custom.MyWidget" {
		t.Errorf("WidgetID = %q, want com.example.custom.MyWidget", def.WidgetID)
	}

	// Built-in widgets should still be available
	_, ok = reg.Get("COMBOBOX")
	if !ok {
		t.Error("built-in COMBOBOX lost after LoadUserDefinitions")
	}
}

func TestRegistryComboboxModes(t *testing.T) {
	reg, err := NewWidgetRegistry()
	if err != nil {
		t.Fatalf("NewWidgetRegistry() error: %v", err)
	}

	def, ok := reg.Get("COMBOBOX")
	if !ok {
		t.Fatal("COMBOBOX not found")
	}

	if len(def.Modes) != 2 {
		t.Fatalf("modes count = %d, want 2", len(def.Modes))
	}

	defaultMode, ok := def.Modes["default"]
	if !ok {
		t.Fatal("default mode not found")
	}
	if len(defaultMode.PropertyMappings) != 1 {
		t.Errorf("default mode mappings = %d, want 1", len(defaultMode.PropertyMappings))
	}

	assocMode, ok := def.Modes["association"]
	if !ok {
		t.Fatal("association mode not found")
	}
	if assocMode.Condition != "hasDataSource" {
		t.Errorf("association mode condition = %q, want hasDataSource", assocMode.Condition)
	}
	if len(assocMode.PropertyMappings) != 4 {
		t.Errorf("association mode mappings = %d, want 4", len(assocMode.PropertyMappings))
	}
}

func TestRegistryGalleryChildSlots(t *testing.T) {
	reg, err := NewWidgetRegistry()
	if err != nil {
		t.Fatalf("NewWidgetRegistry() error: %v", err)
	}

	def, ok := reg.Get("GALLERY")
	if !ok {
		t.Fatal("GALLERY not found")
	}

	if len(def.ChildSlots) != 2 {
		t.Fatalf("childSlots count = %d, want 2", len(def.ChildSlots))
	}

	// Verify slot mappings
	slotsByContainer := make(map[string]ChildSlotMapping)
	for _, slot := range def.ChildSlots {
		slotsByContainer[slot.MDLContainer] = slot
	}

	contentSlot, ok := slotsByContainer["TEMPLATE"]
	if !ok {
		t.Fatal("TEMPLATE slot not found")
	}
	if contentSlot.PropertyKey != "content" {
		t.Errorf("TEMPLATE slot propertyKey = %q, want content", contentSlot.PropertyKey)
	}

	filterSlot, ok := slotsByContainer["FILTER"]
	if !ok {
		t.Fatal("FILTER slot not found")
	}
	if filterSlot.PropertyKey != "filtersPlaceholder" {
		t.Errorf("FILTER slot propertyKey = %q, want filtersPlaceholder", filterSlot.PropertyKey)
	}
}
