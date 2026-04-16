// SPDX-License-Identifier: Apache-2.0

// Package executor - Commands for agent-editor Knowledge Base documents.
//
// Handles `SHOW KNOWLEDGE BASES [IN module]` and
// `DESCRIBE KNOWLEDGE BASE Module.Name`. The underlying BSON is a
// CustomBlobDocuments$CustomBlobDocument with CustomDocumentType =
// "agenteditor.knowledgebase".
package executor

import (
	"fmt"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/sdk/agenteditor"
)

// showAgentEditorKnowledgeBases handles SHOW KNOWLEDGE BASES [IN module].
func (e *Executor) showAgentEditorKnowledgeBases(moduleName string) error {
	if e.reader == nil {
		return fmt.Errorf("not connected to a project")
	}

	kbs, err := e.reader.ListAgentEditorKnowledgeBases()
	if err != nil {
		return fmt.Errorf("failed to list knowledge bases: %w", err)
	}

	h, err := e.getHierarchy()
	if err != nil {
		return err
	}

	result := &TableResult{
		Columns: []string{"Qualified Name", "Module", "Name", "Provider", "Key Constant", "Embedding Model"},
	}

	for _, k := range kbs {
		modID := h.FindModuleID(k.ContainerID)
		modName := h.GetModuleName(modID)
		if moduleName != "" && modName != moduleName {
			continue
		}
		keyConstant := ""
		if k.Key != nil {
			keyConstant = k.Key.QualifiedName
		}
		result.Rows = append(result.Rows, []any{
			fmt.Sprintf("%s.%s", modName, k.Name),
			modName,
			k.Name,
			k.Provider,
			keyConstant,
			k.ModelDisplayName,
		})
	}

	result.Summary = fmt.Sprintf("(%d knowledge base(s))", len(result.Rows))
	return e.writeResult(result)
}

// describeAgentEditorKnowledgeBase handles DESCRIBE KNOWLEDGE BASE Module.Name.
func (e *Executor) describeAgentEditorKnowledgeBase(name ast.QualifiedName) error {
	if e.reader == nil {
		return fmt.Errorf("not connected to a project")
	}

	k := e.findAgentEditorKnowledgeBase(name.Module, name.Name)
	if k == nil {
		return fmt.Errorf("knowledge base not found: %s", name)
	}

	h, err := e.getHierarchy()
	if err != nil {
		return err
	}
	modID := h.FindModuleID(k.ContainerID)
	modName := h.GetModuleName(modID)
	qualifiedName := fmt.Sprintf("%s.%s", modName, k.Name)

	if k.Documentation != "" {
		fmt.Fprintf(e.output, "/**\n * %s\n */\n", k.Documentation)
	}

	fmt.Fprintf(e.output, "CREATE KNOWLEDGE BASE %s (\n", qualifiedName)

	var lines []string
	if k.Provider != "" {
		lines = append(lines, fmt.Sprintf("  Provider: %s", k.Provider))
	}
	if k.Key != nil && k.Key.QualifiedName != "" {
		lines = append(lines, fmt.Sprintf("  Key: %s", k.Key.QualifiedName))
	}
	if k.ModelDisplayName != "" {
		lines = append(lines, fmt.Sprintf("  ModelDisplayName: '%s'", escapeSQLString(k.ModelDisplayName)))
	}
	if k.ModelName != "" {
		lines = append(lines, fmt.Sprintf("  ModelName: '%s'", escapeSQLString(k.ModelName)))
	}
	if k.KeyName != "" {
		lines = append(lines, fmt.Sprintf("  KeyName: '%s'", escapeSQLString(k.KeyName)))
	}
	if k.KeyID != "" {
		lines = append(lines, fmt.Sprintf("  KeyId: '%s'", escapeSQLString(k.KeyID)))
	}
	if k.Environment != "" {
		lines = append(lines, fmt.Sprintf("  Environment: '%s'", escapeSQLString(k.Environment)))
	}
	if k.DeepLinkURL != "" {
		lines = append(lines, fmt.Sprintf("  DeepLinkURL: '%s'", escapeSQLString(k.DeepLinkURL)))
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

// findAgentEditorKnowledgeBase looks up a KB by module and name.
func (e *Executor) findAgentEditorKnowledgeBase(moduleName, kbName string) *agenteditor.KnowledgeBase {
	kbs, err := e.reader.ListAgentEditorKnowledgeBases()
	if err != nil {
		return nil
	}
	h, err := e.getHierarchy()
	if err != nil {
		return nil
	}
	for _, k := range kbs {
		modID := h.FindModuleID(k.ContainerID)
		modName := h.GetModuleName(modID)
		if k.Name == kbName && modName == moduleName {
			return k
		}
	}
	return nil
}
