// SPDX-License-Identifier: Apache-2.0

package catalog

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"strings"

	"github.com/mendixlabs/mxcli/sdk/microflows"
	"github.com/mendixlabs/mxcli/sdk/pages"
)

// XPath usage types
const (
	XPathUsageRetrieve   = "RETRIEVE"
	XPathUsageSecurity   = "SECURITY"
	XPathUsageDatasource = "DATASOURCE"
)

// buildXPathExpressions extracts all XPath constraint expressions from microflows,
// domain model access rules, and page data sources into the xpath_expressions table.
// Only runs in full mode.
func (b *Builder) buildXPathExpressions() error {
	if !b.fullMode {
		return nil
	}

	stmt, err := b.tx.Prepare(`
		INSERT INTO xpath_expressions (
			Id, DocumentType, DocumentId, DocumentQualifiedName,
			ComponentType, ComponentId, ComponentName,
			XPathExpression, TargetEntity, ReferencedEntities,
			IsParameterized, UsageType, ModuleName,
			ProjectId, ProjectName, SnapshotId, SnapshotDate,
			SnapshotSource, SourceId, SourceBranch, SourceRevision
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	projectID, projectName, snapshotID, snapshotDate, snapshotSource, sourceID, sourceBranch, sourceRevision := b.snapshotMeta()
	count := 0

	// 1. Extract from microflow/nanoflow retrieve actions
	mfs, err := b.cachedMicroflows()
	if err == nil {
		for _, mf := range mfs {
			moduleID := b.hierarchy.findModuleID(mf.ContainerID)
			moduleName := b.hierarchy.getModuleName(moduleID)
			sourceQN := moduleName + "." + mf.Name

			count += b.extractMicroflowXPath(stmt, mf.ObjectCollection, "MICROFLOW", string(mf.ID), sourceQN, moduleName,
				projectID, projectName, snapshotID, snapshotDate, snapshotSource, sourceID, sourceBranch, sourceRevision)
		}
	}

	nfs, err := b.cachedNanoflows()
	if err == nil {
		for _, nf := range nfs {
			moduleID := b.hierarchy.findModuleID(nf.ContainerID)
			moduleName := b.hierarchy.getModuleName(moduleID)
			sourceQN := moduleName + "." + nf.Name

			count += b.extractMicroflowXPath(stmt, nf.ObjectCollection, "NANOFLOW", string(nf.ID), sourceQN, moduleName,
				projectID, projectName, snapshotID, snapshotDate, snapshotSource, sourceID, sourceBranch, sourceRevision)
		}
	}

	// 2. Extract from entity access rules
	dms, err := b.cachedDomainModels()
	if err == nil {
		for _, dm := range dms {
			moduleID := b.hierarchy.findModuleID(dm.ContainerID)
			moduleName := b.hierarchy.getModuleName(moduleID)

			for _, ent := range dm.Entities {
				entityQN := moduleName + "." + ent.Name

				for _, rule := range ent.AccessRules {
					if rule.XPathConstraint == "" {
						continue
					}

					id := xpathID(string(ent.ID), string(rule.ID), rule.XPathConstraint)
					isParam := boolToInt(containsVariable(rule.XPathConstraint))
					refs := extractReferencedEntities(rule.XPathConstraint)

					stmt.Exec(id, "DOMAIN_MODEL", string(dm.ID), entityQN,
						"ACCESS_RULE", string(rule.ID), "",
						rule.XPathConstraint, entityQN, refs,
						isParam, XPathUsageSecurity, moduleName,
						projectID, projectName, snapshotID, snapshotDate,
						snapshotSource, sourceID, sourceBranch, sourceRevision)
					count++
				}
			}
		}
	}

	// 3. Extract from page/widget data sources
	pageList, err := b.cachedPages()
	if err == nil {
		for _, pg := range pageList {
			moduleID := b.hierarchy.findModuleID(pg.ContainerID)
			moduleName := b.hierarchy.getModuleName(moduleID)
			sourceQN := moduleName + "." + pg.Name

			if pg.LayoutCall != nil {
				for _, arg := range pg.LayoutCall.Arguments {
					if arg.Widget != nil {
						count += b.extractWidgetXPath(stmt, arg.Widget, "PAGE", string(pg.ID), sourceQN, moduleName,
							projectID, projectName, snapshotID, snapshotDate, snapshotSource, sourceID, sourceBranch, sourceRevision)
					}
				}
			}
		}
	}

	b.report("XPath Expressions", count)
	return nil
}

// extractMicroflowXPath extracts XPath constraints from microflow/nanoflow actions.
func (b *Builder) extractMicroflowXPath(stmt *sql.Stmt,
	oc *microflows.MicroflowObjectCollection, docType, docID, docQN, moduleName,
	projectID, projectName, snapshotID, snapshotDate, snapshotSource, sourceID, sourceBranch, sourceRevision string) int {

	if oc == nil {
		return 0
	}

	count := 0
	for _, obj := range oc.Objects {
		act, ok := obj.(*microflows.ActionActivity)
		if !ok || act.Action == nil {
			continue
		}

		var xpath, entityQN, componentName string
		var componentID string

		switch a := act.Action.(type) {
		case *microflows.RetrieveAction:
			if a.Source == nil {
				continue
			}
			dbSrc, ok := a.Source.(*microflows.DatabaseRetrieveSource)
			if !ok || dbSrc.XPathConstraint == "" {
				continue
			}
			xpath = dbSrc.XPathConstraint
			entityQN = dbSrc.EntityQualifiedName
			componentID = string(act.ID)
			componentName = act.Caption
		default:
			continue
		}

		id := xpathID(docID, componentID, xpath)
		isParam := boolToInt(containsVariable(xpath))
		refs := extractReferencedEntities(xpath)

		stmt.Exec(id, docType, docID, docQN,
			"RETRIEVE_ACTION", componentID, componentName,
			xpath, entityQN, refs,
			isParam, XPathUsageRetrieve, moduleName,
			projectID, projectName, snapshotID, snapshotDate,
			snapshotSource, sourceID, sourceBranch, sourceRevision)
		count++
	}
	return count
}

// extractWidgetXPath recursively extracts XPath constraints from page widgets.
func (b *Builder) extractWidgetXPath(stmt *sql.Stmt,
	w pages.Widget, docType, docID, docQN, moduleName,
	projectID, projectName, snapshotID, snapshotDate, snapshotSource, sourceID, sourceBranch, sourceRevision string) int {

	if w == nil {
		return 0
	}

	count := 0

	// Helper to recurse into children
	recurse := func(widgets []pages.Widget) int {
		n := 0
		for _, child := range widgets {
			n += b.extractWidgetXPath(stmt, child, docType, docID, docQN, moduleName,
				projectID, projectName, snapshotID, snapshotDate, snapshotSource, sourceID, sourceBranch, sourceRevision)
		}
		return n
	}

	// Extract XPath from DatabaseSource
	extractFromDS := func(ds pages.DataSource, widgetName string) {
		if ds == nil {
			return
		}
		dbSrc, ok := ds.(*pages.DatabaseSource)
		if !ok || dbSrc.XPathConstraint == "" {
			return
		}

		entityQN := dbSrc.EntityName
		if entityQN == "" && dbSrc.EntityID != "" {
			entityQN = b.resolveEntityID(dbSrc.EntityID)
		}

		id := xpathID(docID, string(dbSrc.ID), dbSrc.XPathConstraint)
		isParam := boolToInt(containsVariable(dbSrc.XPathConstraint))
		refs := extractReferencedEntities(dbSrc.XPathConstraint)

		stmt.Exec(id, docType, docID, docQN,
			"WIDGET", string(dbSrc.ID), widgetName,
			dbSrc.XPathConstraint, entityQN, refs,
			isParam, XPathUsageDatasource, moduleName,
			projectID, projectName, snapshotID, snapshotDate,
			snapshotSource, sourceID, sourceBranch, sourceRevision)
		count++
	}

	switch widget := w.(type) {
	case *pages.DataView:
		extractFromDS(widget.DataSource, widget.Name)
		count += recurse(widget.Widgets)
		count += recurse(widget.FooterWidgets)

	case *pages.ListView:
		extractFromDS(widget.DataSource, widget.Name)
		count += recurse(widget.Widgets)

	case *pages.DataGrid:
		extractFromDS(widget.DataSource, widget.Name)
		count += recurse(widget.ControlBarWidgets)

	case *pages.TemplateGrid:
		extractFromDS(widget.DataSource, widget.Name)
		count += recurse(widget.Widgets)
		count += recurse(widget.ControlBarWidgets)

	case *pages.Gallery:
		extractFromDS(widget.DataSource, widget.Name)
		if widget.ContentWidget != nil {
			count += b.extractWidgetXPath(stmt, widget.ContentWidget, docType, docID, docQN, moduleName,
				projectID, projectName, snapshotID, snapshotDate, snapshotSource, sourceID, sourceBranch, sourceRevision)
		}
		count += recurse(widget.FilterWidgets)

	case *pages.Container:
		count += recurse(widget.Widgets)

	case *pages.LayoutGrid:
		for _, row := range widget.Rows {
			for _, col := range row.Columns {
				count += recurse(col.Widgets)
			}
		}

	case *pages.CustomWidget:
		if widget.WidgetObject != nil {
			count += b.extractWidgetObjectXPath(stmt, widget.WidgetObject, docType, docID, docQN, moduleName,
				projectID, projectName, snapshotID, snapshotDate, snapshotSource, sourceID, sourceBranch, sourceRevision)
		}
	}

	return count
}

// extractWidgetObjectXPath extracts XPath from pluggable widget property objects.
func (b *Builder) extractWidgetObjectXPath(stmt *sql.Stmt,
	obj *pages.WidgetObject, docType, docID, docQN, moduleName,
	projectID, projectName, snapshotID, snapshotDate, snapshotSource, sourceID, sourceBranch, sourceRevision string) int {

	if obj == nil {
		return 0
	}

	count := 0
	for _, prop := range obj.Properties {
		if prop.Value == nil {
			continue
		}
		val := prop.Value

		// Check datasource-typed properties
		if val.DataSource != nil {
			dbSrc, ok := val.DataSource.(*pages.DatabaseSource)
			if ok && dbSrc.XPathConstraint != "" {
				entityQN := dbSrc.EntityName
				if entityQN == "" && dbSrc.EntityID != "" {
					entityQN = b.resolveEntityID(dbSrc.EntityID)
				}

				id := xpathID(docID, string(dbSrc.ID), dbSrc.XPathConstraint)
				isParam := boolToInt(containsVariable(dbSrc.XPathConstraint))
				refs := extractReferencedEntities(dbSrc.XPathConstraint)

				stmt.Exec(id, docType, docID, docQN,
					"WIDGET", string(dbSrc.ID), "",
					dbSrc.XPathConstraint, entityQN, refs,
					isParam, XPathUsageDatasource, moduleName,
					projectID, projectName, snapshotID, snapshotDate,
					snapshotSource, sourceID, sourceBranch, sourceRevision)
				count++
			}
		}

		// Recurse into nested widget objects
		for _, child := range val.Objects {
			count += b.extractWidgetObjectXPath(stmt, child, docType, docID, docQN, moduleName,
				projectID, projectName, snapshotID, snapshotDate, snapshotSource, sourceID, sourceBranch, sourceRevision)
		}
		for _, child := range val.Widgets {
			count += b.extractWidgetXPath(stmt, child, docType, docID, docQN, moduleName,
				projectID, projectName, snapshotID, snapshotDate, snapshotSource, sourceID, sourceBranch, sourceRevision)
		}
	}
	return count
}

// xpathID generates a deterministic ID from the document, component, and expression.
func xpathID(docID, componentID, xpath string) string {
	h := sha256.Sum256([]byte(docID + "|" + componentID + "|" + xpath))
	return fmt.Sprintf("%x", h[:16])
}

// containsVariable checks if an XPath expression uses $variable references.
func containsVariable(xpath string) bool {
	return strings.Contains(xpath, "$")
}

// extractReferencedEntities extracts qualified entity/association names from XPath.
// Returns a comma-separated list of Module.Name patterns found.
func extractReferencedEntities(xpath string) string {
	var refs []string
	seen := make(map[string]bool)

	// Scan for qualified names (Module.Name pattern) that aren't string literals
	inString := false
	for i := 0; i < len(xpath); i++ {
		if xpath[i] == '\'' {
			inString = !inString
			continue
		}
		if inString {
			continue
		}

		// Look for uppercase letter starting a potential qualified name
		if isUpperLetter(xpath[i]) {
			// Scan ahead for Module.Name pattern
			j := i
			for j < len(xpath) && (isIdentChar(xpath[j]) || xpath[j] == '.') {
				j++
			}
			name := xpath[i:j]
			if strings.Contains(name, ".") && !seen[name] {
				refs = append(refs, name)
				seen[name] = true
			}
			i = j - 1
		}
	}

	return strings.Join(refs, ",")
}

func isUpperLetter(c byte) bool {
	return c >= 'A' && c <= 'Z'
}

func isIdentChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
