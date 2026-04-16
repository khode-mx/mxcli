// SPDX-License-Identifier: Apache-2.0

// Package executor - Commands for agent-editor Agent documents.
//
// Handles `SHOW AGENTS [IN module]` and `DESCRIBE AGENT Module.Name`.
// The underlying BSON is a CustomBlobDocuments$CustomBlobDocument with
// CustomDocumentType = "agenteditor.agent".
package executor

import (
	"fmt"
	"strings"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/sdk/agenteditor"
)

// showAgentEditorAgents handles SHOW AGENTS [IN module].
func (e *Executor) showAgentEditorAgents(moduleName string) error {
	if e.reader == nil {
		return fmt.Errorf("not connected to a project")
	}

	agents, err := e.reader.ListAgentEditorAgents()
	if err != nil {
		return fmt.Errorf("failed to list agents: %w", err)
	}

	h, err := e.getHierarchy()
	if err != nil {
		return err
	}

	result := &TableResult{
		Columns: []string{"Qualified Name", "Module", "Name", "Usage", "Model", "Tools", "KBs"},
	}

	for _, a := range agents {
		modID := h.FindModuleID(a.ContainerID)
		modName := h.GetModuleName(modID)
		if moduleName != "" && modName != moduleName {
			continue
		}
		modelName := ""
		if a.Model != nil {
			modelName = a.Model.QualifiedName
		}
		result.Rows = append(result.Rows, []any{
			fmt.Sprintf("%s.%s", modName, a.Name),
			modName,
			a.Name,
			a.UsageType,
			modelName,
			len(a.Tools),
			len(a.KBTools),
		})
	}

	result.Summary = fmt.Sprintf("(%d agent(s))", len(result.Rows))
	return e.writeResult(result)
}

// describeAgentEditorAgent handles DESCRIBE AGENT Module.Name. Emits a
// round-trippable CREATE AGENT statement reflecting the Contents JSON.
func (e *Executor) describeAgentEditorAgent(name ast.QualifiedName) error {
	if e.reader == nil {
		return fmt.Errorf("not connected to a project")
	}

	a := e.findAgentEditorAgent(name.Module, name.Name)
	if a == nil {
		return fmt.Errorf("agent not found: %s", name)
	}

	h, err := e.getHierarchy()
	if err != nil {
		return err
	}
	modID := h.FindModuleID(a.ContainerID)
	modName := h.GetModuleName(modID)
	qualifiedName := fmt.Sprintf("%s.%s", modName, a.Name)

	if a.Documentation != "" {
		fmt.Fprintf(e.output, "/**\n * %s\n */\n", a.Documentation)
	}

	fmt.Fprintf(e.output, "CREATE AGENT %s (\n", qualifiedName)

	// Build property lines. User-set properties are emitted in a stable
	// order; empty values are omitted.
	var lines []string
	if a.UsageType != "" {
		lines = append(lines, fmt.Sprintf("  UsageType: %s", a.UsageType))
	}
	if a.Description != "" {
		lines = append(lines, fmt.Sprintf("  Description: '%s'", escapeSQLString(a.Description)))
	}
	if a.Model != nil && a.Model.QualifiedName != "" {
		lines = append(lines, fmt.Sprintf("  Model: %s", a.Model.QualifiedName))
	}
	if a.Entity != nil && a.Entity.QualifiedName != "" {
		lines = append(lines, fmt.Sprintf("  Entity: %s", a.Entity.QualifiedName))
	}
	if len(a.Variables) > 0 {
		var parts []string
		for _, v := range a.Variables {
			kind := "String"
			if v.IsAttributeInEntity {
				kind = "EntityAttribute"
			}
			parts = append(parts, fmt.Sprintf("\"%s\": %s", v.Key, kind))
		}
		lines = append(lines, fmt.Sprintf("  Variables: (%s)", strings.Join(parts, ", ")))
	}
	if a.MaxTokens != nil {
		lines = append(lines, fmt.Sprintf("  MaxTokens: %d", *a.MaxTokens))
	}
	if a.ToolChoice != "" {
		lines = append(lines, fmt.Sprintf("  ToolChoice: %s", a.ToolChoice))
	}
	if a.Temperature != nil {
		lines = append(lines, fmt.Sprintf("  Temperature: %g", *a.Temperature))
	}
	if a.TopP != nil {
		lines = append(lines, fmt.Sprintf("  TopP: %g", *a.TopP))
	}
	if a.SystemPrompt != "" {
		lines = append(lines, fmt.Sprintf("  SystemPrompt: '%s'", escapeSQLString(a.SystemPrompt)))
	}
	if a.UserPrompt != "" {
		lines = append(lines, fmt.Sprintf("  UserPrompt: '%s'", escapeSQLString(a.UserPrompt)))
	}

	for i, line := range lines {
		if i < len(lines)-1 {
			fmt.Fprintln(e.output, line+",")
		} else {
			fmt.Fprintln(e.output, line)
		}
	}

	// Body with TOOL / MCP SERVICE / KNOWLEDGE BASE blocks if present.
	hasBody := len(a.Tools) > 0 || len(a.KBTools) > 0
	if hasBody {
		fmt.Fprintln(e.output, ")")
		fmt.Fprintln(e.output, "{")

		for i, t := range a.Tools {
			emitToolBlock(e, t)
			if i < len(a.Tools)-1 || len(a.KBTools) > 0 {
				fmt.Fprintln(e.output)
			}
		}
		for i, kb := range a.KBTools {
			emitKBBlock(e, kb)
			if i < len(a.KBTools)-1 {
				fmt.Fprintln(e.output)
			}
		}

		fmt.Fprintln(e.output, "};")
	} else {
		fmt.Fprintln(e.output, ");")
	}
	fmt.Fprintln(e.output, "/")
	return nil
}

// emitToolBlock writes one TOOL or MCP SERVICE block for the agent body.
func emitToolBlock(e *Executor, t agenteditor.AgentTool) {
	switch t.ToolType {
	case "MCP":
		if t.Document == nil {
			// malformed — skip
			return
		}
		fmt.Fprintf(e.output, "  MCP SERVICE %s {\n", t.Document.QualifiedName)
		fmt.Fprintf(e.output, "    Enabled: %t\n", t.Enabled)
		if t.Description != "" {
			fmt.Fprintf(e.output, "    Description: '%s'\n", escapeSQLString(t.Description))
		}
		fmt.Fprintln(e.output, "  }")
	default:
		// Microflow or unknown tool type — emit generic TOOL block.
		name := t.Name
		if name == "" {
			name = "Tool_" + strings.ReplaceAll(t.ID, "-", "")[:8]
		}
		fmt.Fprintf(e.output, "  TOOL %s {\n", name)
		if t.ToolType != "" {
			fmt.Fprintf(e.output, "    ToolType: %s,\n", t.ToolType)
		}
		if t.Document != nil && t.Document.QualifiedName != "" {
			fmt.Fprintf(e.output, "    Document: %s,\n", t.Document.QualifiedName)
		}
		fmt.Fprintf(e.output, "    Enabled: %t", t.Enabled)
		if t.Description != "" {
			fmt.Fprintln(e.output, ",")
			fmt.Fprintf(e.output, "    Description: '%s'\n", escapeSQLString(t.Description))
		} else {
			fmt.Fprintln(e.output)
		}
		fmt.Fprintln(e.output, "  }")
	}
}

// emitKBBlock writes one KNOWLEDGE BASE block for the agent body.
func emitKBBlock(e *Executor, kb agenteditor.AgentKBTool) {
	name := kb.Name
	if name == "" {
		name = "KB_" + strings.ReplaceAll(kb.ID, "-", "")[:8]
	}
	fmt.Fprintf(e.output, "  KNOWLEDGE BASE %s {\n", name)
	if kb.Document != nil && kb.Document.QualifiedName != "" {
		fmt.Fprintf(e.output, "    Source: %s,\n", kb.Document.QualifiedName)
	}
	if kb.CollectionIdentifier != "" {
		fmt.Fprintf(e.output, "    Collection: '%s',\n", escapeSQLString(kb.CollectionIdentifier))
	}
	if kb.MaxResults != 0 {
		fmt.Fprintf(e.output, "    MaxResults: %d,\n", kb.MaxResults)
	}
	if kb.Description != "" {
		fmt.Fprintf(e.output, "    Description: '%s',\n", escapeSQLString(kb.Description))
	}
	fmt.Fprintf(e.output, "    Enabled: %t\n", kb.Enabled)
	fmt.Fprintln(e.output, "  }")
}

// findAgentEditorAgent looks up an agent by module and name.
func (e *Executor) findAgentEditorAgent(moduleName, agentName string) *agenteditor.Agent {
	agents, err := e.reader.ListAgentEditorAgents()
	if err != nil {
		return nil
	}
	h, err := e.getHierarchy()
	if err != nil {
		return nil
	}
	for _, a := range agents {
		modID := h.FindModuleID(a.ContainerID)
		modName := h.GetModuleName(modID)
		if a.Name == agentName && modName == moduleName {
			return a
		}
	}
	return nil
}
