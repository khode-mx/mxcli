// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/sdk/mpr"
)

// showContractEntities handles SHOW CONTRACT ENTITIES FROM Module.Service.
func (e *Executor) showContractEntities(name *ast.QualifiedName) error {
	if name == nil {
		return fmt.Errorf("service name required: SHOW CONTRACT ENTITIES FROM Module.Service")
	}

	doc, svcQN, err := e.parseServiceContract(*name)
	if err != nil {
		return err
	}

	type row struct {
		entitySet  string
		entityType string
		key        string
		props      int
		navs       int
		summary    string
	}

	// Build entity set lookup
	esMap := make(map[string]string) // entity type qualified name → entity set name
	for _, es := range doc.EntitySets {
		esMap[es.EntityType] = es.Name
	}

	var rows []row

	for _, s := range doc.Schemas {
		for _, et := range s.EntityTypes {
			entitySetName := esMap[s.Namespace+"."+et.Name]
			key := strings.Join(et.KeyProperties, ", ")
			summary := et.Summary
			if len(summary) > 60 {
				summary = summary[:57] + "..."
			}

			rows = append(rows, row{entitySetName, et.Name, key, len(et.Properties), len(et.NavigationProperties), summary})
		}
	}

	if len(rows) == 0 {
		fmt.Fprintf(e.output, "No entity types found in contract for %s.\n", svcQN)
		return nil
	}

	sort.Slice(rows, func(i, j int) bool {
		return strings.ToLower(rows[i].entityType) < strings.ToLower(rows[j].entityType)
	})

	result := &TableResult{
		Columns: []string{"EntitySet", "EntityType", "Key", "Props", "Navs", "Summary"},
		Summary: fmt.Sprintf("(%d entity types in %s contract)", len(rows), svcQN),
	}
	for _, r := range rows {
		result.Rows = append(result.Rows, []any{r.entitySet, r.entityType, r.key, r.props, r.navs, r.summary})
	}
	return e.writeResult(result)
}

// showContractActions handles SHOW CONTRACT ACTIONS FROM Module.Service.
func (e *Executor) showContractActions(name *ast.QualifiedName) error {
	if name == nil {
		return fmt.Errorf("service name required: SHOW CONTRACT ACTIONS FROM Module.Service")
	}

	doc, svcQN, err := e.parseServiceContract(*name)
	if err != nil {
		return err
	}

	if len(doc.Actions) == 0 {
		fmt.Fprintf(e.output, "No actions/functions found in contract for %s.\n", svcQN)
		return nil
	}

	type row struct {
		name       string
		params     int
		returnType string
		bound      string
	}

	var rows []row

	for _, a := range doc.Actions {
		ret := a.ReturnType
		if ret == "" {
			ret = "(void)"
		}
		// Shorten namespace prefix
		if idx := strings.LastIndex(ret, "."); idx >= 0 {
			ret = ret[idx+1:]
		}
		if strings.HasPrefix(ret, "Collection(") {
			inner := ret[len("Collection(") : len(ret)-1]
			if idx := strings.LastIndex(inner, "."); idx >= 0 {
				inner = inner[idx+1:]
			}
			ret = "Collection(" + inner + ")"
		}

		bound := "No"
		if a.IsBound {
			bound = "Yes"
		}

		rows = append(rows, row{a.Name, len(a.Parameters), ret, bound})
	}

	sort.Slice(rows, func(i, j int) bool {
		return strings.ToLower(rows[i].name) < strings.ToLower(rows[j].name)
	})

	result := &TableResult{
		Columns: []string{"Action", "Params", "ReturnType", "Bound"},
		Summary: fmt.Sprintf("(%d actions/functions in %s contract)", len(rows), svcQN),
	}
	for _, r := range rows {
		result.Rows = append(result.Rows, []any{r.name, r.params, r.returnType, r.bound})
	}
	return e.writeResult(result)
}

// describeContractEntity handles DESCRIBE CONTRACT ENTITY Service.EntityName [FORMAT mdl].
func (e *Executor) describeContractEntity(name ast.QualifiedName, format string) error {
	// Name is Module.Service.EntityName — split into service ref and entity name
	// or Module.Service (list all) — but DESCRIBE should have a specific entity
	svcName, entityName, err := splitContractRef(name)
	if err != nil {
		return err
	}

	doc, svcQN, err := e.parseServiceContract(svcName)
	if err != nil {
		return err
	}

	et := doc.FindEntityType(entityName)
	if et == nil {
		return fmt.Errorf("entity type %q not found in contract for %s", entityName, svcQN)
	}

	if strings.EqualFold(format, "mdl") {
		return e.outputContractEntityMDL(et, svcQN, doc)
	}

	// Default: human-readable format
	fmt.Fprintf(e.output, "%s (Key: %s)\n", et.Name, strings.Join(et.KeyProperties, ", "))
	if et.Summary != "" {
		fmt.Fprintf(e.output, "  Summary: %s\n", et.Summary)
	}
	if et.Description != "" {
		fmt.Fprintf(e.output, "  Description: %s\n", et.Description)
	}
	fmt.Fprintln(e.output)

	// Properties
	nameWidth := len("Property")
	typeWidth := len("Type")
	for _, p := range et.Properties {
		if len(p.Name) > nameWidth {
			nameWidth = len(p.Name)
		}
		typeStr := formatEdmType(p)
		if len(typeStr) > typeWidth {
			typeWidth = len(typeStr)
		}
	}

	fmt.Fprintf(e.output, "  %-*s  %-*s  %s\n", nameWidth, "Property", typeWidth, "Type", "Nullable")
	fmt.Fprintf(e.output, "  %s  %s  %s\n", strings.Repeat("-", nameWidth), strings.Repeat("-", typeWidth), "--------")
	for _, p := range et.Properties {
		nullable := "Yes"
		if p.Nullable != nil && !*p.Nullable {
			nullable = "No"
		}
		fmt.Fprintf(e.output, "  %-*s  %-*s  %s\n", nameWidth, p.Name, typeWidth, formatEdmType(p), nullable)
	}

	// Navigation properties
	if len(et.NavigationProperties) > 0 {
		fmt.Fprintln(e.output)
		fmt.Fprintln(e.output, "  Navigation Properties:")
		for _, nav := range et.NavigationProperties {
			multiplicity := "0..1"
			if nav.IsMany {
				multiplicity = "*"
			}
			target := nav.TargetType
			if target == "" && nav.ToRole != "" {
				target = nav.ToRole
			}
			fmt.Fprintf(e.output, "    → %-20s  (%s %s)\n", nav.Name, target, multiplicity)
		}
	}

	return nil
}

// describeContractAction handles DESCRIBE CONTRACT ACTION Service.ActionName [FORMAT mdl].
func (e *Executor) describeContractAction(name ast.QualifiedName, format string) error {
	svcName, actionName, err := splitContractRef(name)
	if err != nil {
		return err
	}

	doc, svcQN, err := e.parseServiceContract(svcName)
	if err != nil {
		return err
	}

	var action *mpr.EdmAction
	for _, a := range doc.Actions {
		if strings.EqualFold(a.Name, actionName) {
			action = a
			break
		}
	}
	if action == nil {
		return fmt.Errorf("action %q not found in contract for %s", actionName, svcQN)
	}

	fmt.Fprintf(e.output, "%s\n", action.Name)
	if action.IsBound {
		fmt.Fprintln(e.output, "  Bound: Yes")
	}

	if len(action.Parameters) > 0 {
		fmt.Fprintln(e.output, "  Parameters:")
		for _, p := range action.Parameters {
			nullable := ""
			if p.Nullable != nil && !*p.Nullable {
				nullable = " NOT NULL"
			}
			fmt.Fprintf(e.output, "    %-20s  %s%s\n", p.Name, shortenEdmType(p.Type), nullable)
		}
	}

	if action.ReturnType != "" {
		fmt.Fprintf(e.output, "  Returns: %s\n", shortenEdmType(action.ReturnType))
	} else {
		fmt.Fprintln(e.output, "  Returns: (void)")
	}

	return nil
}

// outputContractEntityMDL outputs a CREATE EXTERNAL ENTITY statement from contract metadata.
func (e *Executor) outputContractEntityMDL(et *mpr.EdmEntityType, svcQN string, doc *mpr.EdmxDocument) error {
	// Find entity set name
	entitySetName := et.Name + "s" // fallback
	for _, es := range doc.EntitySets {
		if strings.HasSuffix(es.EntityType, "."+et.Name) || es.EntityType == et.Name {
			entitySetName = es.Name
			break
		}
	}

	// Extract module from service qualified name
	module := svcQN
	if idx := strings.Index(svcQN, "."); idx >= 0 {
		module = svcQN[:idx]
	}

	fmt.Fprintf(e.output, "CREATE EXTERNAL ENTITY %s.%s\n", module, et.Name)
	fmt.Fprintf(e.output, "FROM ODATA CLIENT %s (\n", svcQN)
	fmt.Fprintf(e.output, "    EntitySet: '%s',\n", entitySetName)
	fmt.Fprintf(e.output, "    RemoteName: '%s',\n", et.Name)
	fmt.Fprintf(e.output, "    Countable: Yes\n")
	fmt.Fprintln(e.output, ")")
	fmt.Fprintln(e.output, "(")

	for i, p := range et.Properties {
		// Skip ID properties that are not real attributes
		isKey := false
		for _, k := range et.KeyProperties {
			if p.Name == k {
				isKey = true
				break
			}
		}
		if isKey && p.Name == "ID" {
			continue
		}

		mendixType := edmToMendixType(p)
		comma := ","
		if i == len(et.Properties)-1 {
			comma = ""
		}
		fmt.Fprintf(e.output, "    %s: %s%s\n", p.Name, mendixType, comma)
	}

	fmt.Fprintln(e.output, ");")
	fmt.Fprintln(e.output, "/")

	return nil
}

// parseServiceContract finds a consumed OData service by name and parses its cached $metadata.
func (e *Executor) parseServiceContract(name ast.QualifiedName) (*mpr.EdmxDocument, string, error) {
	services, err := e.reader.ListConsumedODataServices()
	if err != nil {
		return nil, "", fmt.Errorf("failed to list consumed OData services: %w", err)
	}

	h, err := e.getHierarchy()
	if err != nil {
		return nil, "", fmt.Errorf("failed to build hierarchy: %w", err)
	}

	for _, svc := range services {
		modID := h.FindModuleID(svc.ContainerID)
		modName := h.GetModuleName(modID)

		if !strings.EqualFold(modName, name.Module) || !strings.EqualFold(svc.Name, name.Name) {
			continue
		}

		svcQN := modName + "." + svc.Name

		if svc.Metadata == "" {
			return nil, svcQN, fmt.Errorf("no cached contract metadata for %s (MetadataUrl: %s). The service metadata has not been downloaded yet", svcQN, svc.MetadataUrl)
		}

		doc, err := mpr.ParseEdmx(svc.Metadata)
		if err != nil {
			return nil, svcQN, fmt.Errorf("failed to parse contract metadata for %s: %w", svcQN, err)
		}

		return doc, svcQN, nil
	}

	return nil, "", fmt.Errorf("consumed OData service not found: %s.%s", name.Module, name.Name)
}

// splitContractRef splits Module.Service.EntityName into (Module.Service, EntityName).
// For a 3-part name like Module.Service.Entity, it returns (Module.Service, Entity).
// For a 2-part name, it returns the name as-is and empty entity name.
func splitContractRef(name ast.QualifiedName) (ast.QualifiedName, string, error) {
	// The qualified name from the parser has Module and Name parts.
	// For "Module.Service.Entity", the parser gives Module="Module", Name="Service.Entity"
	// We need to split Name into service name and entity name.
	parts := strings.SplitN(name.Name, ".", 2)
	if len(parts) != 2 {
		return name, "", fmt.Errorf("expected Module.Service.EntityName, got %s.%s", name.Module, name.Name)
	}

	svcName := ast.QualifiedName{
		Module: name.Module,
		Name:   parts[0],
	}
	return svcName, parts[1], nil
}

// formatEdmType returns a human-readable type string for a property.
func formatEdmType(p *mpr.EdmProperty) string {
	t := p.Type
	if p.MaxLength != "" {
		t += "(" + p.MaxLength + ")"
	}
	if p.Scale != "" {
		t += " Scale=" + p.Scale
	}
	return t
}

// shortenEdmType removes namespace prefix from a type name.
func shortenEdmType(t string) string {
	if strings.HasPrefix(t, "Collection(") {
		inner := t[len("Collection(") : len(t)-1]
		if idx := strings.LastIndex(inner, "."); idx >= 0 {
			inner = inner[idx+1:]
		}
		return "Collection(" + inner + ")"
	}
	if idx := strings.LastIndex(t, "."); idx >= 0 {
		return t[idx+1:]
	}
	return t
}

// edmToMendixType maps an Edm type to a Mendix attribute type string for MDL output.
func edmToMendixType(p *mpr.EdmProperty) string {
	switch p.Type {
	case "Edm.String":
		if p.MaxLength != "" && p.MaxLength != "max" {
			return "String(" + p.MaxLength + ")"
		}
		return "String(200)"
	case "Edm.Int32":
		return "Integer"
	case "Edm.Int64":
		return "Long"
	case "Edm.Decimal":
		return "Decimal"
	case "Edm.Boolean":
		return "Boolean"
	case "Edm.DateTime", "Edm.DateTimeOffset":
		return "DateTime"
	case "Edm.Date":
		return "DateTime"
	case "Edm.Binary":
		return "String(200)"
	case "Edm.Guid":
		return "String(36)"
	case "Edm.Double", "Edm.Single":
		return "Decimal"
	case "Edm.Byte", "Edm.SByte", "Edm.Int16":
		return "Integer"
	default:
		return "String(200)"
	}
}

// ============================================================================
// CREATE EXTERNAL ENTITIES (bulk)
// ============================================================================

// createExternalEntities handles CREATE [OR MODIFY] EXTERNAL ENTITIES FROM Module.Service [INTO Module] [ENTITIES (...)].
// It reads entity types from the cached $metadata and creates external entities in the domain model.
func (e *Executor) createExternalEntities(s *ast.CreateExternalEntitiesStmt) error {
	if e.writer == nil {
		return fmt.Errorf("not connected to a project in write mode")
	}

	doc, svcQN, err := e.parseServiceContract(s.ServiceRef)
	if err != nil {
		return err
	}

	// Build entity set lookup: entity type qualified name → entity set name
	esMap := make(map[string]string)
	for _, es := range doc.EntitySets {
		esMap[es.EntityType] = es.Name
	}

	// Build filter set if entity names specified
	filterSet := make(map[string]bool)
	for _, name := range s.EntityNames {
		filterSet[strings.ToLower(name)] = true
	}

	// Determine target module
	targetModule := s.TargetModule
	if targetModule == "" {
		targetModule = s.ServiceRef.Module
	}

	var created, skipped, failed int

	for _, schema := range doc.Schemas {
		for _, et := range schema.EntityTypes {
			// Apply entity name filter
			if len(filterSet) > 0 && !filterSet[strings.ToLower(et.Name)] {
				continue
			}

			entitySetName := esMap[schema.Namespace+"."+et.Name]
			if entitySetName == "" {
				entitySetName = et.Name + "s" // fallback
			}

			// Build attributes from properties
			var attrs []ast.Attribute
			for _, p := range et.Properties {
				// Skip key properties named ID (Mendix manages its own ID)
				isKey := false
				for _, k := range et.KeyProperties {
					if p.Name == k {
						isKey = true
						break
					}
				}
				if isKey && p.Name == "ID" {
					continue
				}

				attrs = append(attrs, ast.Attribute{
					Name: p.Name,
					Type: edmToAstDataType(p),
				})
			}

			stmt := &ast.CreateExternalEntityStmt{
				Name:           ast.QualifiedName{Module: targetModule, Name: et.Name},
				ServiceRef:     s.ServiceRef,
				EntitySet:      entitySetName,
				RemoteName:     et.Name,
				Countable:      true,
				Attributes:     attrs,
				CreateOrModify: s.CreateOrModify,
			}

			if err := e.execCreateExternalEntity(stmt); err != nil {
				fmt.Fprintf(e.output, "  FAILED: %s.%s — %v\n", targetModule, et.Name, err)
				failed++
			} else {
				created++
			}
		}
	}

	if skipped > 0 || failed > 0 {
		fmt.Fprintf(e.output, "\nImported %d entities from %s (%d failed)\n", created, svcQN, failed)
	} else {
		fmt.Fprintf(e.output, "\nImported %d entities from %s into %s\n", created, svcQN, targetModule)
	}

	return nil
}

// edmToAstDataType converts an Edm property to an AST data type.
func edmToAstDataType(p *mpr.EdmProperty) ast.DataType {
	switch p.Type {
	case "Edm.String":
		length := 200
		if p.MaxLength != "" && p.MaxLength != "max" {
			if n, err := fmt.Sscanf(p.MaxLength, "%d", &length); n == 0 || err != nil {
				length = 200
			}
		}
		return ast.DataType{Kind: ast.TypeString, Length: length}
	case "Edm.Int32":
		return ast.DataType{Kind: ast.TypeInteger}
	case "Edm.Int64":
		return ast.DataType{Kind: ast.TypeLong}
	case "Edm.Decimal":
		return ast.DataType{Kind: ast.TypeDecimal}
	case "Edm.Boolean":
		return ast.DataType{Kind: ast.TypeBoolean}
	case "Edm.DateTime", "Edm.DateTimeOffset", "Edm.Date":
		return ast.DataType{Kind: ast.TypeDateTime}
	case "Edm.Double", "Edm.Single":
		return ast.DataType{Kind: ast.TypeDecimal}
	case "Edm.Byte", "Edm.SByte", "Edm.Int16":
		return ast.DataType{Kind: ast.TypeInteger}
	case "Edm.Guid":
		return ast.DataType{Kind: ast.TypeString, Length: 36}
	case "Edm.Binary":
		return ast.DataType{Kind: ast.TypeString, Length: 200}
	default:
		return ast.DataType{Kind: ast.TypeString, Length: 200}
	}
}

// ============================================================================
// AsyncAPI Contract Commands
// ============================================================================

// showContractChannels handles SHOW CONTRACT CHANNELS FROM Module.Service.
func (e *Executor) showContractChannels(name *ast.QualifiedName) error {
	if name == nil {
		return fmt.Errorf("service name required: SHOW CONTRACT CHANNELS FROM Module.Service")
	}

	doc, svcQN, err := e.parseAsyncAPIContract(*name)
	if err != nil {
		return err
	}

	if len(doc.Channels) == 0 {
		fmt.Fprintf(e.output, "No channels found in contract for %s.\n", svcQN)
		return nil
	}

	type row struct {
		channel   string
		operation string
		opID      string
		message   string
	}

	var rows []row

	for _, ch := range doc.Channels {
		rows = append(rows, row{ch.Name, ch.OperationType, ch.OperationID, ch.MessageRef})
	}

	result := &TableResult{
		Columns: []string{"Channel", "Operation", "OperationID", "Message"},
		Summary: fmt.Sprintf("(%d channels in %s contract)", len(rows), svcQN),
	}
	for _, r := range rows {
		result.Rows = append(result.Rows, []any{r.channel, r.operation, r.opID, r.message})
	}
	return e.writeResult(result)
}

// showContractMessages handles SHOW CONTRACT MESSAGES FROM Module.Service.
func (e *Executor) showContractMessages(name *ast.QualifiedName) error {
	if name == nil {
		return fmt.Errorf("service name required: SHOW CONTRACT MESSAGES FROM Module.Service")
	}

	doc, svcQN, err := e.parseAsyncAPIContract(*name)
	if err != nil {
		return err
	}

	if len(doc.Messages) == 0 {
		fmt.Fprintf(e.output, "No messages found in contract for %s.\n", svcQN)
		return nil
	}

	type row struct {
		name        string
		title       string
		contentType string
		props       int
	}

	var rows []row

	for _, msg := range doc.Messages {
		rows = append(rows, row{msg.Name, msg.Title, msg.ContentType, len(msg.Properties)})
	}

	sort.Slice(rows, func(i, j int) bool {
		return strings.ToLower(rows[i].name) < strings.ToLower(rows[j].name)
	})

	result := &TableResult{
		Columns: []string{"Message", "Title", "ContentType", "Props"},
		Summary: fmt.Sprintf("(%d messages in %s contract)", len(rows), svcQN),
	}
	for _, r := range rows {
		result.Rows = append(result.Rows, []any{r.name, r.title, r.contentType, r.props})
	}
	return e.writeResult(result)
}

// describeContractMessage handles DESCRIBE CONTRACT MESSAGE Module.Service.MessageName.
func (e *Executor) describeContractMessage(name ast.QualifiedName) error {
	svcName, msgName, err := splitContractRef(name)
	if err != nil {
		return err
	}

	doc, svcQN, err := e.parseAsyncAPIContract(svcName)
	if err != nil {
		return err
	}

	msg := doc.FindMessage(msgName)
	if msg == nil {
		return fmt.Errorf("message %q not found in contract for %s", msgName, svcQN)
	}

	fmt.Fprintf(e.output, "%s\n", msg.Name)
	if msg.Title != "" {
		fmt.Fprintf(e.output, "  Title: %s\n", msg.Title)
	}
	if msg.Description != "" {
		fmt.Fprintf(e.output, "  Description: %s\n", msg.Description)
	}
	if msg.ContentType != "" {
		fmt.Fprintf(e.output, "  ContentType: %s\n", msg.ContentType)
	}

	if len(msg.Properties) > 0 {
		fmt.Fprintln(e.output)
		nameWidth := len("Property")
		typeWidth := len("Type")
		for _, p := range msg.Properties {
			if len(p.Name) > nameWidth {
				nameWidth = len(p.Name)
			}
			t := asyncTypeString(p)
			if len(t) > typeWidth {
				typeWidth = len(t)
			}
		}

		fmt.Fprintf(e.output, "  %-*s  %-*s\n", nameWidth, "Property", typeWidth, "Type")
		fmt.Fprintf(e.output, "  %s  %s\n", strings.Repeat("-", nameWidth), strings.Repeat("-", typeWidth))
		for _, p := range msg.Properties {
			fmt.Fprintf(e.output, "  %-*s  %-*s\n", nameWidth, p.Name, typeWidth, asyncTypeString(p))
		}
	}

	return nil
}

// parseAsyncAPIContract finds a business event service by name and parses its cached AsyncAPI document.
func (e *Executor) parseAsyncAPIContract(name ast.QualifiedName) (*mpr.AsyncAPIDocument, string, error) {
	services, err := e.reader.ListBusinessEventServices()
	if err != nil {
		return nil, "", fmt.Errorf("failed to list business event services: %w", err)
	}

	h, err := e.getHierarchy()
	if err != nil {
		return nil, "", fmt.Errorf("failed to build hierarchy: %w", err)
	}

	for _, svc := range services {
		modID := h.FindModuleID(svc.ContainerID)
		modName := h.GetModuleName(modID)

		if !strings.EqualFold(modName, name.Module) || !strings.EqualFold(svc.Name, name.Name) {
			continue
		}

		svcQN := modName + "." + svc.Name

		if svc.Document == "" {
			return nil, svcQN, fmt.Errorf("no cached AsyncAPI contract for %s. This service has no Document field (it may be a publisher, not a consumer)", svcQN)
		}

		doc, err := mpr.ParseAsyncAPI(svc.Document)
		if err != nil {
			return nil, svcQN, fmt.Errorf("failed to parse AsyncAPI contract for %s: %w", svcQN, err)
		}

		return doc, svcQN, nil
	}

	return nil, "", fmt.Errorf("business event service not found: %s.%s", name.Module, name.Name)
}

// asyncTypeString formats an AsyncAPI property type for display.
func asyncTypeString(p *mpr.AsyncAPIProperty) string {
	if p.Format != "" {
		return p.Type + " (" + p.Format + ")"
	}
	return p.Type
}
