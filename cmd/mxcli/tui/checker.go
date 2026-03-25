package tui

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mendixlabs/mxcli/cmd/mxcli/docker"
)

// CheckError represents a single mx check diagnostic.
type CheckError struct {
	Severity     string // "ERROR" or "WARNING"
	Code         string // e.g. "CE0001"
	Message      string
	DocumentName string // e.g. "Page 'P_ComboBox_Enum'" (from JSON output)
	ElementName  string // e.g. "Property 'Association' of combo box 'cmbPriority'"
	ModuleName   string // e.g. "MyFirstModule"
}

// MxCheckResultMsg carries the result of an async mx check run.
type MxCheckResultMsg struct {
	Errors []CheckError
	Err    error
}

// MxCheckStartMsg signals that a check run has started.
type MxCheckStartMsg struct{}

// MxCheckRerunMsg requests a manual re-run of mx check (e.g. from overlay "r" key).
type MxCheckRerunMsg struct{}

// mxCheckJSON mirrors the JSON structure produced by `mx check -j`.
type mxCheckJSON struct {
	Errors   []mxCheckEntry `json:"errors"`
	Warnings []mxCheckEntry `json:"warnings"`
}

type mxCheckEntry struct {
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Locations []mxCheckLocation `json:"locations"`
}

type mxCheckLocation struct {
	ModuleName   string `json:"module-name"`
	DocumentName string `json:"document-name"`
	ElementName  string `json:"element-name"`
}

// runMxCheck returns a tea.Cmd that runs mx check asynchronously.
// Uses `-j` for JSON output to get document-level location information.
func runMxCheck(projectPath string) tea.Cmd {
	return tea.Batch(
		func() tea.Msg { return MxCheckStartMsg{} },
		func() tea.Msg {
			mxPath, err := docker.ResolveMx("")
			if err != nil {
				Trace("checker: mx not found: %v", err)
				return MxCheckResultMsg{Err: err}
			}

			jsonFile, err := os.CreateTemp("", "mx-check-*.json")
			if err != nil {
				return MxCheckResultMsg{Err: err}
			}
			jsonPath := jsonFile.Name()
			jsonFile.Close()
			defer os.Remove(jsonPath)

			Trace("checker: running %s check %s -j %s", mxPath, projectPath, jsonPath)
			cmd := exec.Command(mxPath, "check", projectPath, "-j", jsonPath)
			_, runErr := cmd.CombinedOutput()

			checkErrors, parseErr := parseCheckJSON(jsonPath)
			if parseErr != nil {
				Trace("checker: JSON parse error: %v", parseErr)
				// If JSON parsing fails, return the run error
				if runErr != nil {
					return MxCheckResultMsg{Err: runErr}
				}
				return MxCheckResultMsg{Err: parseErr}
			}

			Trace("checker: done, %d diagnostics", len(checkErrors))

			// mx check returns non-zero exit code when there are errors,
			// but we still want to show the parsed errors.
			if runErr != nil && len(checkErrors) == 0 {
				return MxCheckResultMsg{Err: runErr}
			}
			return MxCheckResultMsg{Errors: checkErrors}
		},
	)
}

// parseCheckJSON reads the JSON file produced by `mx check -j` and converts to CheckError slice.
func parseCheckJSON(jsonPath string) ([]CheckError, error) {
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, err
	}

	var result mxCheckJSON
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	var checkErrors []CheckError
	for _, entry := range result.Errors {
		checkErrors = append(checkErrors, entryToCheckError(entry, "ERROR"))
	}
	for _, entry := range result.Warnings {
		checkErrors = append(checkErrors, entryToCheckError(entry, "WARNING"))
	}
	return checkErrors, nil
}

func entryToCheckError(entry mxCheckEntry, severity string) CheckError {
	ce := CheckError{
		Severity: severity,
		Code:     entry.Code,
		Message:  entry.Message,
	}
	if len(entry.Locations) > 0 {
		loc := entry.Locations[0]
		ce.ModuleName = loc.ModuleName
		ce.DocumentName = loc.DocumentName
		ce.ElementName = loc.ElementName
	}
	return ce
}

// renderCheckResults formats check errors for display in an overlay.
func renderCheckResults(errors []CheckError) string {
	if errors == nil {
		return "No check has been run yet. Changes to the project will trigger an automatic check."
	}
	if len(errors) == 0 {
		return CheckPassStyle.Render("✓ Project check passed — no errors or warnings")
	}

	var sb strings.Builder
	var errorCount, warningCount int
	for _, e := range errors {
		if e.Severity == "ERROR" {
			errorCount++
		} else {
			warningCount++
		}
	}

	// Summary header
	sb.WriteString(CheckHeaderStyle.Render("mx check Results"))
	sb.WriteString("\n")
	var summaryParts []string
	if errorCount > 0 {
		summaryParts = append(summaryParts, CheckErrorStyle.Render("● "+itoa(errorCount)+" errors"))
	}
	if warningCount > 0 {
		summaryParts = append(summaryParts, CheckWarnStyle.Render("● "+itoa(warningCount)+" warnings"))
	}
	sb.WriteString(strings.Join(summaryParts, "  "))
	sb.WriteString("\n\n")

	// Detail lines
	for _, e := range errors {
		var label string
		if e.Severity == "ERROR" {
			label = CheckErrorStyle.Render(e.Severity + " " + e.Code)
		} else {
			label = CheckWarnStyle.Render(e.Severity + " " + e.Code)
		}
		sb.WriteString(label + "\n")
		sb.WriteString("  " + e.Message + "\n")
		if e.DocumentName != "" {
			loc := formatDocLocation(e.ModuleName, e.DocumentName)
			if e.ElementName != "" {
				loc += " > " + e.ElementName
			}
			sb.WriteString("  " + CheckLocStyle.Render(loc) + "\n")
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// formatDocLocation converts mx JSON document-name (e.g. "Page 'P_ComboBox'")
// into a qualified name like "MyModule.P_ComboBox (Page)".
func formatDocLocation(moduleName, documentName string) string {
	// documentName format: "Type 'Name'" — extract type and name
	if idx := strings.Index(documentName, " '"); idx > 0 {
		docType := documentName[:idx]
		docName := strings.TrimSuffix(documentName[idx+2:], "'")
		qname := docName
		if moduleName != "" {
			qname = moduleName + "." + docName
		}
		return qname + " (" + docType + ")"
	}
	// Fallback: just prefix with module
	if moduleName != "" {
		return moduleName + "." + documentName
	}
	return documentName
}

// formatCheckBadge returns a compact badge string for the status bar.
func formatCheckBadge(errors []CheckError, running bool) string {
	if running {
		return CheckRunningStyle.Render("⟳ checking")
	}
	if errors == nil {
		return "" // no check has run yet
	}
	if len(errors) == 0 {
		return CheckPassStyle.Render("✓")
	}
	var errorCount, warningCount int
	for _, e := range errors {
		if e.Severity == "ERROR" {
			errorCount++
		} else {
			warningCount++
		}
	}
	var parts []string
	if errorCount > 0 {
		parts = append(parts, CheckErrorStyle.Render("✗ "+itoa(errorCount)+"E"))
	}
	if warningCount > 0 {
		parts = append(parts, CheckWarnStyle.Render(itoa(warningCount)+"W"))
	}
	return strings.Join(parts, " ")
}
