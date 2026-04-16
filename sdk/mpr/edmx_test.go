// SPDX-License-Identifier: Apache-2.0

package mpr

import (
	"testing"
)

const testOData3Metadata = `<?xml version="1.0" encoding="utf-8"?>
<edmx:Edmx Version="1.0" xmlns:edmx="http://schemas.microsoft.com/ado/2007/06/edmx" xmlns:mx="http://www.mendix.com/Protocols/MendixData">
  <edmx:DataServices m:DataServiceVersion="3.0" m:MaxDataServiceVersion="3.0" xmlns:m="http://schemas.microsoft.com/ado/2007/08/dataservices/metadata">
    <Schema Namespace="DefaultNamespace" xmlns="http://schemas.microsoft.com/ado/2009/11/edm">
      <EntityType Name="PurchaseOrder">
        <Documentation>
          <Summary>SAP Purchase Order</Summary>
          <LongDescription>Provides access to Purchase Order information from SAP</LongDescription>
        </Documentation>
        <Key>
          <PropertyRef Name="ID" />
        </Key>
        <Property Name="ID" Type="Edm.Int64" Nullable="false" mx:isAttribute="false" />
        <Property Name="Number" Type="Edm.Int64" />
        <Property Name="Status" Type="Edm.String" />
        <Property Name="SupplierName" Type="Edm.String" MaxLength="200" />
        <Property Name="GrossAmount" Type="Edm.Decimal" />
        <Property Name="DeliveryDate" Type="Edm.DateTimeOffset" />
        <NavigationProperty Name="PurchaseOrderItems" Relationship="DefaultNamespace.PurchaseOrderItem_PurchaseOrder" FromRole="PurchaseOrder" ToRole="PurchaseOrderItems" />
        <NavigationProperty Name="Customer" Relationship="DefaultNamespace.PurchaseOrder_Customer" FromRole="PurchaseOrders" ToRole="Customer" />
      </EntityType>
      <EntityType Name="Customer">
        <Key>
          <PropertyRef Name="ID" />
        </Key>
        <Property Name="ID" Type="Edm.Int64" Nullable="false" />
        <Property Name="Name" Type="Edm.String" MaxLength="200" />
        <Property Name="ContactEmail" Type="Edm.String" MaxLength="200" />
      </EntityType>
      <EntityContainer Name="Entities" m:IsDefaultEntityContainer="true">
        <EntitySet Name="PurchaseOrders" EntityType="DefaultNamespace.PurchaseOrder" />
        <EntitySet Name="Customers" EntityType="DefaultNamespace.Customer" />
      </EntityContainer>
    </Schema>
  </edmx:DataServices>
</edmx:Edmx>`

const testOData4Metadata = `<?xml version="1.0" encoding="utf-8"?>
<edmx:Edmx Version="4.0" xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx">
  <edmx:DataServices>
    <Schema Namespace="DefaultNamespace" xmlns="http://docs.oasis-open.org/odata/ns/edm">
      <EntityType Name="Product">
        <Key>
          <PropertyRef Name="ID" />
        </Key>
        <Property Name="ID" Type="Edm.Int64" Nullable="false" />
        <Property Name="Title" Type="Edm.String" MaxLength="200" />
        <Property Name="Weight" Type="Edm.Decimal" Scale="variable" />
        <Property Name="IsFragile" Type="Edm.Boolean" Nullable="false" />
        <NavigationProperty Name="Parts" Type="Collection(DefaultNamespace.Part)" Partner="Product" />
        <Annotation Term="Org.OData.Core.V1.Description" String="Product Inventory" />
      </EntityType>
      <EntityType Name="Part">
        <Key>
          <PropertyRef Name="ID" />
        </Key>
        <Property Name="ID" Type="Edm.Int64" Nullable="false" />
        <Property Name="Title" Type="Edm.String" MaxLength="200" />
        <NavigationProperty Name="Product" Type="DefaultNamespace.Product" Partner="Parts" />
      </EntityType>
      <Action Name="CreateOrder">
        <Parameter Name="OrderData" Type="DefaultNamespace.OrderInput" />
        <ReturnType Type="DefaultNamespace.OrderResult" />
      </Action>
      <Function Name="GetTopProducts">
        <Parameter Name="Count" Type="Edm.Int32" />
        <ReturnType Type="Collection(DefaultNamespace.Product)" />
      </Function>
      <EntityContainer Name="Entities">
        <EntitySet Name="Products" EntityType="DefaultNamespace.Product" />
        <EntitySet Name="Parts" EntityType="DefaultNamespace.Part" />
      </EntityContainer>
    </Schema>
  </edmx:DataServices>
</edmx:Edmx>`

func TestParseEdmxOData3(t *testing.T) {
	doc, err := ParseEdmx(testOData3Metadata)
	if err != nil {
		t.Fatalf("ParseEdmx failed: %v", err)
	}

	if doc.Version != "1.0" {
		t.Errorf("expected version 1.0, got %s", doc.Version)
	}

	if len(doc.Schemas) != 1 {
		t.Fatalf("expected 1 schema, got %d", len(doc.Schemas))
	}

	schema := doc.Schemas[0]
	if schema.Namespace != "DefaultNamespace" {
		t.Errorf("expected namespace DefaultNamespace, got %s", schema.Namespace)
	}

	if len(schema.EntityTypes) != 2 {
		t.Fatalf("expected 2 entity types, got %d", len(schema.EntityTypes))
	}

	// Check PurchaseOrder
	po := schema.EntityTypes[0]
	if po.Name != "PurchaseOrder" {
		t.Errorf("expected PurchaseOrder, got %s", po.Name)
	}
	if po.Summary != "SAP Purchase Order" {
		t.Errorf("expected summary 'SAP Purchase Order', got '%s'", po.Summary)
	}
	if len(po.KeyProperties) != 1 || po.KeyProperties[0] != "ID" {
		t.Errorf("expected key [ID], got %v", po.KeyProperties)
	}
	if len(po.Properties) != 6 {
		t.Errorf("expected 6 properties, got %d", len(po.Properties))
	}
	if len(po.NavigationProperties) != 2 {
		t.Errorf("expected 2 nav properties, got %d", len(po.NavigationProperties))
	}

	// Check SupplierName property
	var supplierProp *EdmProperty
	for _, p := range po.Properties {
		if p.Name == "SupplierName" {
			supplierProp = p
			break
		}
	}
	if supplierProp == nil {
		t.Fatal("SupplierName property not found")
	}
	if supplierProp.Type != "Edm.String" {
		t.Errorf("expected Edm.String, got %s", supplierProp.Type)
	}
	if supplierProp.MaxLength != "200" {
		t.Errorf("expected MaxLength 200, got %s", supplierProp.MaxLength)
	}

	// Check entity sets
	if len(doc.EntitySets) != 2 {
		t.Fatalf("expected 2 entity sets, got %d", len(doc.EntitySets))
	}
	if doc.EntitySets[0].Name != "PurchaseOrders" {
		t.Errorf("expected PurchaseOrders, got %s", doc.EntitySets[0].Name)
	}

	// Check FindEntityType
	found := doc.FindEntityType("DefaultNamespace.Customer")
	if found == nil {
		t.Error("FindEntityType('DefaultNamespace.Customer') returned nil")
	}
	if found != nil && found.Name != "Customer" {
		t.Errorf("expected Customer, got %s", found.Name)
	}
}

func TestParseEdmxOData4(t *testing.T) {
	doc, err := ParseEdmx(testOData4Metadata)
	if err != nil {
		t.Fatalf("ParseEdmx failed: %v", err)
	}

	if doc.Version != "4.0" {
		t.Errorf("expected version 4.0, got %s", doc.Version)
	}

	schema := doc.Schemas[0]

	// Check Product entity
	product := schema.EntityTypes[0]
	if product.Name != "Product" {
		t.Errorf("expected Product, got %s", product.Name)
	}
	if product.Summary != "Product Inventory" {
		t.Errorf("expected summary 'Product Inventory', got '%s'", product.Summary)
	}

	// Check navigation property with Collection type
	if len(product.NavigationProperties) != 1 {
		t.Fatalf("expected 1 nav property, got %d", len(product.NavigationProperties))
	}
	nav := product.NavigationProperties[0]
	if nav.Name != "Parts" {
		t.Errorf("expected Parts, got %s", nav.Name)
	}
	if nav.TargetType != "Part" {
		t.Errorf("expected target type Part, got %s", nav.TargetType)
	}
	if !nav.IsMany {
		t.Error("expected IsMany=true for Collection type")
	}

	// Check Part navigation property (single)
	part := schema.EntityTypes[1]
	partNav := part.NavigationProperties[0]
	if partNav.TargetType != "Product" {
		t.Errorf("expected target type Product, got %s", partNav.TargetType)
	}
	if partNav.IsMany {
		t.Error("expected IsMany=false for single type")
	}

	// Check actions
	if len(doc.Actions) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(doc.Actions))
	}
	createOrder := doc.Actions[0]
	if createOrder.Name != "CreateOrder" {
		t.Errorf("expected CreateOrder, got %s", createOrder.Name)
	}
	if len(createOrder.Parameters) != 1 {
		t.Errorf("expected 1 parameter, got %d", len(createOrder.Parameters))
	}
	if createOrder.ReturnType != "DefaultNamespace.OrderResult" {
		t.Errorf("expected return type DefaultNamespace.OrderResult, got %s", createOrder.ReturnType)
	}

	// Check function
	getTop := doc.Actions[1]
	if getTop.Name != "GetTopProducts" {
		t.Errorf("expected GetTopProducts, got %s", getTop.Name)
	}
	if getTop.ReturnType != "Collection(DefaultNamespace.Product)" {
		t.Errorf("expected return type Collection(DefaultNamespace.Product), got %s", getTop.ReturnType)
	}

	// Check entity sets
	if len(doc.EntitySets) != 2 {
		t.Fatalf("expected 2 entity sets, got %d", len(doc.EntitySets))
	}
}

func TestParseEdmxEmpty(t *testing.T) {
	_, err := ParseEdmx("")
	if err == nil {
		t.Error("expected error for empty metadata")
	}
}

const testCapabilitiesMetadata = `<?xml version="1.0" encoding="utf-8"?>
<edmx:Edmx Version="4.0" xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx">
  <edmx:DataServices>
    <Schema Namespace="DefaultNamespace" xmlns="http://docs.oasis-open.org/odata/ns/edm">
      <EntityType Name="Order">
        <Key><PropertyRef Name="OrderId" /></Key>
        <Property Name="OrderId" Type="Edm.Int64" Nullable="false">
          <Annotation Term="Org.OData.Core.V1.Computed" Bool="true" />
        </Property>
        <Property Name="OrderNumber" Type="Edm.String" MaxLength="32">
          <Annotation Term="Org.OData.Core.V1.Immutable" Bool="true" />
        </Property>
        <Property Name="CustomerName" Type="Edm.String" MaxLength="200" />
      </EntityType>
      <EntityType Name="OrderLine">
        <Key><PropertyRef Name="LineId" /></Key>
        <Property Name="LineId" Type="Edm.Int64" Nullable="false" />
      </EntityType>
      <EntityContainer Name="Container">
        <EntitySet Name="Orders" EntityType="DefaultNamespace.Order">
          <Annotation Term="Org.OData.Capabilities.V1.InsertRestrictions">
            <Record>
              <PropertyValue Property="Insertable" Bool="true" />
              <PropertyValue Property="NonInsertableProperties">
                <Collection>
                  <PropertyPath>OrderId</PropertyPath>
                </Collection>
              </PropertyValue>
              <PropertyValue Property="NonInsertableNavigationProperties">
                <Collection>
                  <NavigationPropertyPath>Lines</NavigationPropertyPath>
                </Collection>
              </PropertyValue>
            </Record>
          </Annotation>
          <Annotation Term="Org.OData.Capabilities.V1.UpdateRestrictions">
            <Record>
              <PropertyValue Property="Updatable" Bool="true" />
              <PropertyValue Property="NonUpdatableProperties">
                <Collection>
                  <PropertyPath>OrderId</PropertyPath>
                  <PropertyPath>OrderNumber</PropertyPath>
                </Collection>
              </PropertyValue>
            </Record>
          </Annotation>
          <Annotation Term="Org.OData.Capabilities.V1.DeleteRestrictions">
            <Record><PropertyValue Property="Deletable" Bool="true" /></Record>
          </Annotation>
        </EntitySet>
        <EntitySet Name="OrderLines" EntityType="DefaultNamespace.OrderLine" />
      </EntityContainer>
    </Schema>
  </edmx:DataServices>
</edmx:Edmx>`

func TestParseEdmxCapabilityAnnotations(t *testing.T) {
	doc, err := ParseEdmx(testCapabilitiesMetadata)
	if err != nil {
		t.Fatalf("ParseEdmx failed: %v", err)
	}

	// Find the Orders entity set.
	var orders *EdmEntitySet
	for _, es := range doc.EntitySets {
		if es.Name == "Orders" {
			orders = es
		}
	}
	if orders == nil {
		t.Fatal("Orders entity set not found")
	}

	if orders.Insertable == nil || !*orders.Insertable {
		t.Errorf("Orders.Insertable = %v, want true", orders.Insertable)
	}
	if orders.Updatable == nil || !*orders.Updatable {
		t.Errorf("Orders.Updatable = %v, want true", orders.Updatable)
	}
	if orders.Deletable == nil || !*orders.Deletable {
		t.Errorf("Orders.Deletable = %v, want true", orders.Deletable)
	}

	wantNonIns := []string{"OrderId"}
	if !stringSliceEqual(orders.NonInsertableProperties, wantNonIns) {
		t.Errorf("NonInsertableProperties = %v, want %v", orders.NonInsertableProperties, wantNonIns)
	}
	wantNonUpd := []string{"OrderId", "OrderNumber"}
	if !stringSliceEqual(orders.NonUpdatableProperties, wantNonUpd) {
		t.Errorf("NonUpdatableProperties = %v, want %v", orders.NonUpdatableProperties, wantNonUpd)
	}
	wantNonInsNav := []string{"Lines"}
	if !stringSliceEqual(orders.NonInsertableNavigationProperties, wantNonInsNav) {
		t.Errorf("NonInsertableNavigationProperties = %v, want %v", orders.NonInsertableNavigationProperties, wantNonInsNav)
	}

	// OrderLines has no annotations → all flags unset.
	var lines *EdmEntitySet
	for _, es := range doc.EntitySets {
		if es.Name == "OrderLines" {
			lines = es
		}
	}
	if lines == nil {
		t.Fatal("OrderLines entity set not found")
	}
	if lines.Insertable != nil || lines.Updatable != nil || lines.Deletable != nil {
		t.Errorf("OrderLines should have nil capability flags, got Insertable=%v Updatable=%v Deletable=%v",
			lines.Insertable, lines.Updatable, lines.Deletable)
	}

	// Per-property Computed/Immutable annotations.
	order := doc.FindEntityType("DefaultNamespace.Order")
	if order == nil {
		t.Fatal("Order entity type not found")
	}
	propByName := map[string]*EdmProperty{}
	for _, p := range order.Properties {
		propByName[p.Name] = p
	}
	if !propByName["OrderId"].Computed {
		t.Errorf("OrderId.Computed = false, want true")
	}
	if !propByName["OrderNumber"].Immutable {
		t.Errorf("OrderNumber.Immutable = false, want true")
	}
	if propByName["CustomerName"].Computed || propByName["CustomerName"].Immutable {
		t.Errorf("CustomerName should have no capability flags")
	}
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestResolveNavType(t *testing.T) {
	tests := []struct {
		input    string
		wantType string
		wantMany bool
	}{
		{"DefaultNamespace.Product", "Product", false},
		{"Collection(DefaultNamespace.Part)", "Part", true},
		{"Edm.String", "String", false},
		{"Product", "Product", false},
	}

	for _, tt := range tests {
		typeName, isMany := resolveNavType(tt.input)
		if typeName != tt.wantType {
			t.Errorf("resolveNavType(%q): got type %q, want %q", tt.input, typeName, tt.wantType)
		}
		if isMany != tt.wantMany {
			t.Errorf("resolveNavType(%q): got isMany=%v, want %v", tt.input, isMany, tt.wantMany)
		}
	}
}
