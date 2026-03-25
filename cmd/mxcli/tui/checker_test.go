package tui

import (
	"os"
	"strings"
	"testing"
)

func TestParseCheckJSON(t *testing.T) {
	jsonContent := `{
		"serialization_version": 1,
		"errors": [
			{
				"code": "CE1613",
				"message": "The selected association 'MyModule.Priority' no longer exists.",
				"locations": [
					{
						"module-name": "MyModule",
						"document-name": "Page 'P_ComboBox'",
						"element-name": "Property 'Association' of combo box 'cmbPriority'"
					}
				]
			},
			{
				"code": "CE0463",
				"message": "Widget definition changed for DataGrid2",
				"locations": [
					{
						"module-name": "MyModule",
						"document-name": "Page 'CustomerList'",
						"element-name": "DataGrid2 widget"
					}
				]
			}
		],
		"warnings": [
			{
				"code": "CW0001",
				"message": "Unused variable '$var' in microflow",
				"locations": [
					{
						"module-name": "MyModule",
						"document-name": "Microflow 'DoSomething'",
						"element-name": "Variable '$var'"
					}
				]
			}
		]
	}`

	tmpFile, err := os.CreateTemp("", "mx-check-test-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString(jsonContent)
	tmpFile.Close()

	errors, err := parseCheckJSON(tmpFile.Name())
	if err != nil {
		t.Fatalf("parseCheckJSON: %v", err)
	}
	if len(errors) != 3 {
		t.Fatalf("expected 3 errors, got %d", len(errors))
	}

	// First error
	if errors[0].Severity != "ERROR" {
		t.Errorf("expected ERROR, got %q", errors[0].Severity)
	}
	if errors[0].Code != "CE1613" {
		t.Errorf("expected CE1613, got %q", errors[0].Code)
	}
	if errors[0].DocumentName != "Page 'P_ComboBox'" {
		t.Errorf("unexpected document: %q", errors[0].DocumentName)
	}
	if errors[0].ElementName == "" {
		t.Error("expected non-empty element name")
	}
	if errors[0].ModuleName != "MyModule" {
		t.Errorf("expected MyModule, got %q", errors[0].ModuleName)
	}

	// Second error
	if errors[1].Code != "CE0463" {
		t.Errorf("expected CE0463, got %q", errors[1].Code)
	}

	// Third: warning
	if errors[2].Severity != "WARNING" {
		t.Errorf("expected WARNING, got %q", errors[2].Severity)
	}
	if errors[2].Code != "CW0001" {
		t.Errorf("expected CW0001, got %q", errors[2].Code)
	}
}

func TestParseCheckJSONEmpty(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "mx-check-test-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString(`{"serialization_version": 1}`)
	tmpFile.Close()

	errors, err := parseCheckJSON(tmpFile.Name())
	if err != nil {
		t.Fatalf("parseCheckJSON: %v", err)
	}
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors, got %d", len(errors))
	}
}

func TestParseCheckJSONNoLocations(t *testing.T) {
	jsonContent := `{
		"serialization_version": 1,
		"errors": [
			{
				"code": "CE9999",
				"message": "Some error without location",
				"locations": []
			}
		]
	}`

	tmpFile, err := os.CreateTemp("", "mx-check-test-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString(jsonContent)
	tmpFile.Close()

	errors, err := parseCheckJSON(tmpFile.Name())
	if err != nil {
		t.Fatalf("parseCheckJSON: %v", err)
	}
	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}
	if errors[0].DocumentName != "" {
		t.Errorf("expected empty document name, got %q", errors[0].DocumentName)
	}
}

func TestRenderCheckResultsNilVsEmpty(t *testing.T) {
	// nil = no check has run yet
	result := renderCheckResults(nil)
	if result == "" {
		t.Error("expected non-empty result for nil errors")
	}
	if strings.Contains(result, "passed") {
		t.Error("nil errors should NOT show 'passed' — no check has run yet")
	}

	// empty = check ran, no errors found
	result = renderCheckResults([]CheckError{})
	if !strings.Contains(result, "passed") {
		t.Error("empty errors should show 'passed'")
	}
}

func TestRenderCheckResultsWithDocLocation(t *testing.T) {
	errors := []CheckError{
		{
			Severity:     "ERROR",
			Code:         "CE1613",
			Message:      "Association no longer exists",
			DocumentName: "Page 'P_ComboBox'",
			ElementName:  "combo box 'cmbPriority'",
			ModuleName:   "MyModule",
		},
	}
	result := renderCheckResults(errors)
	if !strings.Contains(result, "MyModule.P_ComboBox (Page)") {
		t.Errorf("expected qualified doc location, got: %s", result)
	}
	if !strings.Contains(result, "combo box 'cmbPriority'") {
		t.Error("expected element name in rendered output")
	}
}

func TestFormatCheckBadge(t *testing.T) {
	// No check run yet
	badge := formatCheckBadge(nil, false)
	if badge != "" {
		t.Errorf("expected empty badge, got %q", badge)
	}

	// Running
	badge = formatCheckBadge(nil, true)
	if badge == "" {
		t.Error("expected non-empty badge for running state")
	}

	// Pass
	badge = formatCheckBadge([]CheckError{}, false)
	if badge == "" {
		t.Error("expected non-empty badge for pass state")
	}

	// Errors
	errors := []CheckError{
		{Severity: "ERROR", Code: "CE0001"},
		{Severity: "WARNING", Code: "CW0001"},
		{Severity: "ERROR", Code: "CE0002"},
	}
	badge = formatCheckBadge(errors, false)
	if badge == "" {
		t.Error("expected non-empty badge with errors")
	}
}
