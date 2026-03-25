// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"github.com/mendixlabs/mxcli/sdk/pages"
	"go.mongodb.org/mongo-driver/bson"
)

// =============================================================================
// Pluggable Widget Engine — Core Types and Operation Registry
// =============================================================================

// WidgetDefinition describes how to construct a pluggable widget from MDL syntax.
// Loaded from embedded JSON definition files (*.def.json).
type WidgetDefinition struct {
	WidgetID         string                `json:"widgetId"`
	MDLName          string                `json:"mdlName"`
	TemplateFile     string                `json:"templateFile"`
	DefaultEditable  string                `json:"defaultEditable"`
	DefaultSelection string                `json:"defaultSelection,omitempty"`
	PropertyMappings []PropertyMapping     `json:"propertyMappings,omitempty"`
	ChildSlots       []ChildSlotMapping    `json:"childSlots,omitempty"`
	Modes            map[string]WidgetMode `json:"modes,omitempty"`
}

// WidgetMode defines a conditional configuration variant for a widget.
// For example, ComboBox has "enumeration" and "association" modes.
type WidgetMode struct {
	Condition        string             `json:"condition,omitempty"`
	Description      string             `json:"description,omitempty"`
	PropertyMappings []PropertyMapping  `json:"propertyMappings"`
	ChildSlots       []ChildSlotMapping `json:"childSlots,omitempty"`
}

// PropertyMapping maps an MDL source (attribute, association, literal, etc.)
// to a pluggable widget property key via a named operation.
type PropertyMapping struct {
	PropertyKey string `json:"propertyKey"`
	Source      string `json:"source,omitempty"`
	Value       string `json:"value,omitempty"`
	Operation   string `json:"operation"`
	Default     string `json:"default,omitempty"`
}

// ChildSlotMapping maps an MDL child container (e.g., TEMPLATE, FILTER) to a
// widget property that holds child widgets.
type ChildSlotMapping struct {
	PropertyKey  string `json:"propertyKey"`
	MDLContainer string `json:"mdlContainer"`
	Operation    string `json:"operation"`
}

// BuildContext carries resolved values from MDL parsing for use by operations.
type BuildContext struct {
	AttributePath string
	AssocPath     string
	EntityName    string
	PrimitiveVal  string
	DataSource    pages.DataSource
	ChildWidgets  []bson.D
}

// OperationFunc updates a template object's property identified by propertyKey.
// It receives the current object BSON, the property type ID map, the target key,
// and the build context containing resolved values.
type OperationFunc func(obj bson.D, propTypeIDs map[string]pages.PropertyTypeIDEntry, propertyKey string, ctx *BuildContext) bson.D

// OperationRegistry maps operation names to their implementations.
type OperationRegistry struct {
	operations map[string]OperationFunc
}

// NewOperationRegistry creates a registry pre-loaded with the 5 built-in operations.
func NewOperationRegistry() *OperationRegistry {
	reg := &OperationRegistry{
		operations: make(map[string]OperationFunc),
	}
	reg.Register("attribute", opAttribute)
	reg.Register("association", opAssociation)
	reg.Register("primitive", opPrimitive)
	reg.Register("datasource", opDatasource)
	reg.Register("widgets", opWidgets)
	return reg
}

// Register adds or replaces an operation by name.
func (r *OperationRegistry) Register(name string, fn OperationFunc) {
	r.operations[name] = fn
}

// Lookup returns the operation function for the given name, or nil if not found.
func (r *OperationRegistry) Lookup(name string) OperationFunc {
	return r.operations[name]
}

// =============================================================================
// Built-in Operations
// =============================================================================

// opAttribute sets an attribute reference on a widget property.
func opAttribute(obj bson.D, propTypeIDs map[string]pages.PropertyTypeIDEntry, propertyKey string, ctx *BuildContext) bson.D {
	if ctx.AttributePath == "" {
		return obj
	}
	return updateWidgetPropertyValue(obj, propTypeIDs, propertyKey, func(val bson.D) bson.D {
		return setAttributeRef(val, ctx.AttributePath)
	})
}

// opAssociation sets an association reference (AttributeRef + EntityRef) on a widget property.
func opAssociation(obj bson.D, propTypeIDs map[string]pages.PropertyTypeIDEntry, propertyKey string, ctx *BuildContext) bson.D {
	if ctx.AssocPath == "" {
		return obj
	}
	return updateWidgetPropertyValue(obj, propTypeIDs, propertyKey, func(val bson.D) bson.D {
		return setAssociationRef(val, ctx.AssocPath, ctx.EntityName)
	})
}

// opPrimitive sets a primitive string value on a widget property.
func opPrimitive(obj bson.D, propTypeIDs map[string]pages.PropertyTypeIDEntry, propertyKey string, ctx *BuildContext) bson.D {
	if ctx.PrimitiveVal == "" {
		return obj
	}
	return updateWidgetPropertyValue(obj, propTypeIDs, propertyKey, func(val bson.D) bson.D {
		return setPrimitiveValue(val, ctx.PrimitiveVal)
	})
}

// opDatasource sets a data source on a widget property.
func opDatasource(obj bson.D, propTypeIDs map[string]pages.PropertyTypeIDEntry, propertyKey string, ctx *BuildContext) bson.D {
	if ctx.DataSource == nil {
		return obj
	}
	return updateWidgetPropertyValue(obj, propTypeIDs, propertyKey, func(val bson.D) bson.D {
		return setDataSource(val, ctx.DataSource)
	})
}

// opWidgets replaces the Widgets array in a widget property value with child widgets.
func opWidgets(obj bson.D, propTypeIDs map[string]pages.PropertyTypeIDEntry, propertyKey string, ctx *BuildContext) bson.D {
	if len(ctx.ChildWidgets) == 0 {
		return obj
	}
	return updateWidgetPropertyValue(obj, propTypeIDs, propertyKey, func(val bson.D) bson.D {
		return setChildWidgets(val, ctx.ChildWidgets)
	})
}

// setChildWidgets replaces the Widgets field in a WidgetValue with the given child widgets.
func setChildWidgets(val bson.D, childWidgets []bson.D) bson.D {
	widgetsArr := bson.A{int32(2)} // version marker
	for _, w := range childWidgets {
		widgetsArr = append(widgetsArr, w)
	}

	result := make(bson.D, 0, len(val))
	for _, elem := range val {
		if elem.Key == "Widgets" {
			result = append(result, bson.E{Key: "Widgets", Value: widgetsArr})
		} else {
			result = append(result, elem)
		}
	}
	return result
}
