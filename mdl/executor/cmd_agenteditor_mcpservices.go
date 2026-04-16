// SPDX-License-Identifier: Apache-2.0

// Package executor - Commands for agent-editor Consumed MCP Service documents.
//
// Handles `SHOW CONSUMED MCP SERVICES [IN module]` and
// `DESCRIBE CONSUMED MCP SERVICE Module.Name`. The underlying BSON is a
// CustomBlobDocuments$CustomBlobDocument with CustomDocumentType =
// "agenteditor.consumedMCPService".
package executor

import (
	"fmt"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/sdk/agenteditor"
)

// showAgentEditorConsumedMCPServices handles SHOW CONSUMED MCP SERVICES [IN module].
func (e *Executor) showAgentEditorConsumedMCPServices(moduleName string) error {
	if e.reader == nil {
		return fmt.Errorf("not connected to a project")
	}

	svcs, err := e.reader.ListAgentEditorConsumedMCPServices()
	if err != nil {
		return fmt.Errorf("failed to list consumed MCP services: %w", err)
	}

	h, err := e.getHierarchy()
	if err != nil {
		return err
	}

	result := &TableResult{
		Columns: []string{"Qualified Name", "Module", "Name", "Protocol", "Version", "Timeout"},
	}

	for _, c := range svcs {
		modID := h.FindModuleID(c.ContainerID)
		modName := h.GetModuleName(modID)
		if moduleName != "" && modName != moduleName {
			continue
		}
		result.Rows = append(result.Rows, []any{
			fmt.Sprintf("%s.%s", modName, c.Name),
			modName,
			c.Name,
			c.ProtocolVersion,
			c.Version,
			c.ConnectionTimeoutSeconds,
		})
	}

	result.Summary = fmt.Sprintf("(%d consumed MCP service(s))", len(result.Rows))
	return e.writeResult(result)
}

// describeAgentEditorConsumedMCPService handles DESCRIBE CONSUMED MCP SERVICE Module.Name.
func (e *Executor) describeAgentEditorConsumedMCPService(name ast.QualifiedName) error {
	if e.reader == nil {
		return fmt.Errorf("not connected to a project")
	}

	c := e.findAgentEditorConsumedMCPService(name.Module, name.Name)
	if c == nil {
		return fmt.Errorf("consumed MCP service not found: %s", name)
	}

	h, err := e.getHierarchy()
	if err != nil {
		return err
	}
	modID := h.FindModuleID(c.ContainerID)
	modName := h.GetModuleName(modID)
	qualifiedName := fmt.Sprintf("%s.%s", modName, c.Name)

	if c.Documentation != "" {
		fmt.Fprintf(e.output, "/**\n * %s\n */\n", c.Documentation)
	}

	fmt.Fprintf(e.output, "CREATE CONSUMED MCP SERVICE %s (\n", qualifiedName)

	var lines []string
	if c.ProtocolVersion != "" {
		lines = append(lines, fmt.Sprintf("  ProtocolVersion: %s", c.ProtocolVersion))
	}
	if c.Version != "" {
		lines = append(lines, fmt.Sprintf("  Version: '%s'", escapeSQLString(c.Version)))
	}
	if c.ConnectionTimeoutSeconds != 0 {
		lines = append(lines, fmt.Sprintf("  ConnectionTimeoutSeconds: %d", c.ConnectionTimeoutSeconds))
	}
	if c.InnerDocumentation != "" {
		lines = append(lines, fmt.Sprintf("  Documentation: '%s'", escapeSQLString(c.InnerDocumentation)))
	}

	for i, line := range lines {
		if i < len(lines)-1 {
			fmt.Fprintln(e.output, line+",")
		} else {
			fmt.Fprintln(e.output, line)
		}
	}

	fmt.Fprintln(e.output, ");")
	fmt.Fprintln(e.output, "/")
	return nil
}

// findAgentEditorConsumedMCPService looks up an MCP service by module and name.
func (e *Executor) findAgentEditorConsumedMCPService(moduleName, svcName string) *agenteditor.ConsumedMCPService {
	svcs, err := e.reader.ListAgentEditorConsumedMCPServices()
	if err != nil {
		return nil
	}
	h, err := e.getHierarchy()
	if err != nil {
		return nil
	}
	for _, c := range svcs {
		modID := h.FindModuleID(c.ContainerID)
		modName := h.GetModuleName(modID)
		if c.Name == svcName && modName == moduleName {
			return c
		}
	}
	return nil
}
