// SPDX-License-Identifier: Apache-2.0

// Package executor - JavaScript Action commands (SHOW/DESCRIBE JAVASCRIPT ACTIONS)
package executor

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/sdk/javaactions"
)

// showJavaScriptActions handles SHOW JAVASCRIPT ACTIONS command.
func (e *Executor) showJavaScriptActions(moduleName string) error {
	h, err := e.getHierarchy()
	if err != nil {
		return fmt.Errorf("failed to build hierarchy: %w", err)
	}

	jsActions, err := e.reader.ListJavaScriptActions()
	if err != nil {
		return fmt.Errorf("failed to list javascript actions: %w", err)
	}

	type row struct {
		qualifiedName string
		module        string
		name          string
		platform      string
		folderPath    string
	}
	var rows []row

	for _, jsa := range jsActions {
		modID := h.FindModuleID(jsa.ContainerID)
		modName := h.GetModuleName(modID)
		if moduleName == "" || modName == moduleName {
			qualifiedName := modName + "." + jsa.Name
			folderPath := h.BuildFolderPath(jsa.ContainerID)
			platform := jsa.Platform
			if platform == "" {
				platform = "All"
			}
			rows = append(rows, row{qualifiedName, modName, jsa.Name, platform, folderPath})
		}
	}

	sort.Slice(rows, func(i, j int) bool {
		return strings.ToLower(rows[i].qualifiedName) < strings.ToLower(rows[j].qualifiedName)
	})

	result := &TableResult{
		Columns: []string{"Qualified Name", "Module", "Name", "Platform", "Folder"},
		Summary: fmt.Sprintf("(%d javascript actions)", len(rows)),
	}
	for _, r := range rows {
		result.Rows = append(result.Rows, []any{r.qualifiedName, r.module, r.name, r.platform, r.folderPath})
	}
	return e.writeResult(result)
}

// describeJavaScriptAction handles DESCRIBE JAVASCRIPT ACTION command.
func (e *Executor) describeJavaScriptAction(name ast.QualifiedName) error {
	qualifiedName := name.Module + "." + name.Name
	jsa, err := e.reader.ReadJavaScriptActionByName(qualifiedName)
	if err != nil {
		return fmt.Errorf("javascript action not found: %s", qualifiedName)
	}

	var sb strings.Builder

	// Documentation comment
	doc := strings.ReplaceAll(jsa.Documentation, "\r\n", "\n")
	doc = strings.ReplaceAll(doc, "\r", "\n")
	if doc != "" {
		sb.WriteString("/**\n")
		for line := range strings.SplitSeq(doc, "\n") {
			sb.WriteString(" * ")
			sb.WriteString(line)
			sb.WriteString("\n")
		}
		sb.WriteString(" */\n")
	}

	// Type parameters
	sb.WriteString("CREATE JAVASCRIPT ACTION ")
	sb.WriteString(qualifiedName)
	if len(jsa.TypeParameters) > 0 {
		sb.WriteString("<")
		for i, tp := range jsa.TypeParameters {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(tp.Name)
		}
		sb.WriteString(">")
	}
	sb.WriteString("(")

	// Parameters
	hasParamDescriptions := false
	for _, p := range jsa.Parameters {
		if p.Description != "" {
			hasParamDescriptions = true
			break
		}
	}

	for i, param := range jsa.Parameters {
		if i > 0 {
			sb.WriteString(", ")
		}
		if hasParamDescriptions {
			sb.WriteString("\n    ")
		}
		sb.WriteString(param.Name)
		sb.WriteString(": ")
		if param.ParameterType != nil {
			sb.WriteString(formatJavaScriptActionType(param.ParameterType))
		} else {
			sb.WriteString("Object")
		}
		if param.IsRequired {
			sb.WriteString(" NOT NULL")
		}
		if param.Description != "" {
			paramDoc := strings.ReplaceAll(param.Description, "\r\n", "\n")
			paramDoc = strings.ReplaceAll(paramDoc, "\r", "\n")
			firstLine, _, _ := strings.Cut(paramDoc, "\n")
			sb.WriteString("  -- ")
			sb.WriteString(firstLine)
		}
	}
	if hasParamDescriptions {
		sb.WriteString("\n")
	}
	sb.WriteString(")")

	// Platform
	platform := jsa.Platform
	if platform == "" {
		platform = "All"
	}
	if platform != "All" {
		sb.WriteString("\n  PLATFORM ")
		sb.WriteString(platform)
	}

	// Return type
	if jsa.ReturnType != nil {
		sb.WriteString("\n  RETURNS ")
		sb.WriteString(formatJavaActionReturnType(jsa.ReturnType))
	}

	// RETURN NAME metadata
	if jsa.ActionDefaultReturnName != "" {
		sb.WriteString("\n-- RETURN NAME: '")
		sb.WriteString(jsa.ActionDefaultReturnName)
		sb.WriteString("'")
	}

	// EXPOSED AS clause
	if jsa.MicroflowActionInfo != nil && jsa.MicroflowActionInfo.Caption != "" {
		sb.WriteString("\n  EXPOSED AS '")
		sb.WriteString(jsa.MicroflowActionInfo.Caption)
		sb.WriteString("' IN '")
		sb.WriteString(jsa.MicroflowActionInfo.Category)
		sb.WriteString("'")
		if jsa.MicroflowActionInfo.Icon != "" {
			sb.WriteString("\n-- ICON: ")
			sb.WriteString(jsa.MicroflowActionInfo.Icon)
		}
	}

	// JavaScript source code
	userCode, extraCode := e.readJavaScriptActionSource(name.Module, name.Name)
	if userCode != "" {
		sb.WriteString("\nAS $$\n")
		sb.WriteString(userCode)
		sb.WriteString("\n$$")
	}

	sb.WriteString(";")

	fmt.Fprintln(e.output, sb.String())

	// Additional info as comments
	if jsa.ExportLevel != "" && jsa.ExportLevel != "Hidden" {
		fmt.Fprintf(e.output, "-- EXPORT LEVEL: %s\n", jsa.ExportLevel)
	}
	if jsa.Excluded {
		fmt.Fprintln(e.output, "-- EXCLUDED: true")
	}
	if platform != "All" {
		// already shown inline
	} else {
		fmt.Fprintln(e.output, "-- PLATFORM: All")
	}
	if extraCode != "" {
		fmt.Fprintln(e.output, "-- EXTRA CODE:")
		for line := range strings.SplitSeq(extraCode, "\n") {
			fmt.Fprintf(e.output, "-- %s\n", line)
		}
	}

	return nil
}

// readJavaScriptActionSource reads the JavaScript source file and extracts user code and extra code.
func (e *Executor) readJavaScriptActionSource(moduleName, actionName string) (userCode, extraCode string) {
	if e.mprPath == "" {
		return "", ""
	}

	projectRoot := filepath.Dir(e.mprPath)
	// JavaScript source uses original module name casing (not lowercased like javasource)
	jsPath := filepath.Join(projectRoot, "javascriptsource", moduleName, "actions", actionName+".js")

	content, err := os.ReadFile(jsPath)
	if err != nil {
		// Try lowercase module name as fallback
		jsPath = filepath.Join(projectRoot, "javascriptsource", strings.ToLower(moduleName), "actions", actionName+".js")
		content, err = os.ReadFile(jsPath)
		if err != nil {
			return "", ""
		}
	}

	source := string(content)

	// Extract user code
	beginUserCode := "// BEGIN USER CODE"
	endUserCode := "// END USER CODE"
	if beginIdx := strings.Index(source, beginUserCode); beginIdx != -1 {
		if endIdx := strings.Index(source, endUserCode); endIdx != -1 && endIdx > beginIdx {
			uc := source[beginIdx+len(beginUserCode) : endIdx]
			uc = strings.TrimPrefix(uc, "\n")
			uc = strings.TrimSuffix(uc, "\n")
			uc = strings.TrimRight(uc, " \t")
			userCode = uc
		}
	}

	// Extract extra code
	beginExtraCode := "// BEGIN EXTRA CODE"
	endExtraCode := "// END EXTRA CODE"
	if beginIdx := strings.Index(source, beginExtraCode); beginIdx != -1 {
		if endIdx := strings.Index(source, endExtraCode); endIdx != -1 && endIdx > beginIdx {
			ec := source[beginIdx+len(beginExtraCode) : endIdx]
			ec = strings.TrimSpace(ec)
			if ec != "" {
				extraCode = ec
			}
		}
	}

	return userCode, extraCode
}

// formatJavaScriptActionType formats a JavaScript action parameter type for MDL output.
func formatJavaScriptActionType(t javaactions.CodeActionParameterType) string {
	if t == nil {
		return "Object"
	}
	// EntityTypeParameterType → ENTITY <name> syntax
	if etp, ok := t.(*javaactions.EntityTypeParameterType); ok {
		if etp.TypeParameterName != "" {
			return "ENTITY <" + etp.TypeParameterName + ">"
		}
		return "ENTITY <>"
	}
	return t.TypeString()
}
