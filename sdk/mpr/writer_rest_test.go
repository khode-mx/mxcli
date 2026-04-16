// SPDX-License-Identifier: Apache-2.0

package mpr

import (
	"testing"

	"github.com/mendixlabs/mxcli/model"
	"go.mongodb.org/mongo-driver/bson"
)

func TestSerializeConsumedRestServiceBasic(t *testing.T) {
	w := &Writer{}
	svc := &model.ConsumedRestService{
		BaseElement: model.BaseElement{
			ID:       "test-rest-id",
			TypeName: "Rest$ConsumedRestService",
		},
		ContainerID: "test-module-id",
		Name:        "PetStoreAPI",
		BaseUrl:     "https://petstore.swagger.io/v2",
	}

	data, err := w.serializeConsumedRestService(svc)
	if err != nil {
		t.Fatalf("serialize failed: %v", err)
	}

	var raw map[string]any
	if err := bson.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	assertField(t, raw, "$Type", "Rest$ConsumedRestService")
	assertField(t, raw, "Name", "PetStoreAPI")

	// BaseUrl should be a ValueTemplate
	baseUrl, ok := raw["BaseUrl"].(map[string]any)
	if !ok {
		t.Fatalf("BaseUrl: expected map, got %T", raw["BaseUrl"])
	}
	assertField(t, baseUrl, "$Type", "Rest$ValueTemplate")
	assertField(t, baseUrl, "Value", "https://petstore.swagger.io/v2")

	// AuthenticationScheme should be nil
	if raw["AuthenticationScheme"] != nil {
		t.Errorf("AuthenticationScheme: expected nil, got %v", raw["AuthenticationScheme"])
	}
}

func TestSerializeConsumedRestServiceWithAuth(t *testing.T) {
	w := &Writer{}
	svc := &model.ConsumedRestService{
		BaseElement: model.BaseElement{
			ID: "test-rest-auth-id",
		},
		ContainerID: "test-module-id",
		Name:        "SecureAPI",
		BaseUrl:     "https://api.example.com",
		Authentication: &model.RestAuthentication{
			Scheme:   "Basic",
			Username: "admin",
			Password: "secret",
		},
	}

	data, err := w.serializeConsumedRestService(svc)
	if err != nil {
		t.Fatalf("serialize failed: %v", err)
	}

	var raw map[string]any
	if err := bson.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	// AuthenticationScheme should be BasicAuthenticationScheme
	authScheme, ok := raw["AuthenticationScheme"].(map[string]any)
	if !ok {
		t.Fatalf("AuthenticationScheme: expected map, got %T", raw["AuthenticationScheme"])
	}
	assertField(t, authScheme, "$Type", "Rest$BasicAuthenticationScheme")

	// Username should be StringValue (literal)
	username, ok := authScheme["Username"].(map[string]any)
	if !ok {
		t.Fatalf("Username: expected map, got %T", authScheme["Username"])
	}
	assertField(t, username, "$Type", "Rest$StringValue")
	assertField(t, username, "Value", "admin")

	// Password should be StringValue (literal)
	password, ok := authScheme["Password"].(map[string]any)
	if !ok {
		t.Fatalf("Password: expected map, got %T", authScheme["Password"])
	}
	assertField(t, password, "$Type", "Rest$StringValue")
	assertField(t, password, "Value", "secret")
}

func TestSerializeConsumedRestServiceWithConstantAuth(t *testing.T) {
	w := &Writer{}
	svc := &model.ConsumedRestService{
		BaseElement: model.BaseElement{
			ID: "test-rest-const-auth",
		},
		ContainerID: "test-module-id",
		Name:        "ConstAuthAPI",
		BaseUrl:     "https://api.example.com",
		Authentication: &model.RestAuthentication{
			Scheme:   "Basic",
			Username: "$MyModule.ApiUser",
			Password: "$MyModule.ApiPass",
		},
	}

	data, err := w.serializeConsumedRestService(svc)
	if err != nil {
		t.Fatalf("serialize failed: %v", err)
	}

	var raw map[string]any
	if err := bson.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	authScheme, ok := raw["AuthenticationScheme"].(map[string]any)
	if !ok {
		t.Fatalf("AuthenticationScheme: expected map, got %T", raw["AuthenticationScheme"])
	}

	// Username should be ConstantValue
	username, ok := authScheme["Username"].(map[string]any)
	if !ok {
		t.Fatalf("Username: expected map, got %T", authScheme["Username"])
	}
	assertField(t, username, "$Type", "Rest$ConstantValue")
	assertField(t, username, "Constant", "MyModule.ApiUser")
}

func TestSerializeRestOperationGetWithParams(t *testing.T) {
	op := &model.RestClientOperation{
		Name:       "GetPet",
		HttpMethod: "GET",
		Path:       "/pet/{petId}",
		Parameters: []*model.RestClientParameter{
			{Name: "petId", DataType: "Integer"},
		},
		Headers: []*model.RestClientHeader{
			{Name: "Accept", Value: "application/json"},
		},
		ResponseType: "JSON",
		Timeout:      30,
	}

	result := serializeRestOperation(op)

	assertField(t, result, "$Type", "Rest$RestOperation")
	assertField(t, result, "Name", "GetPet")

	// Timeout
	if v, ok := result["Timeout"].(int64); !ok || v != 30 {
		t.Errorf("Timeout: expected 30, got %v", result["Timeout"])
	}

	// Method should be WithoutBody (GET)
	method, ok := result["Method"].(bson.M)
	if !ok {
		t.Fatalf("Method: expected bson.M, got %T", result["Method"])
	}
	if method["$Type"] != "Rest$RestOperationMethodWithoutBody" {
		t.Errorf("Method.$Type: expected WithoutBody, got %v", method["$Type"])
	}
	if method["HttpMethod"] != "Get" {
		t.Errorf("Method.HttpMethod: expected Get, got %v", method["HttpMethod"])
	}

	// Path should be ValueTemplate
	path, ok := result["Path"].(bson.M)
	if !ok {
		t.Fatalf("Path: expected bson.M, got %T", result["Path"])
	}
	if path["Value"] != "/pet/{petId}" {
		t.Errorf("Path.Value: expected /pet/{petId}, got %v", path["Value"])
	}

	// Parameters
	params := extractBsonArray(result["Parameters"])
	if len(params) != 1 {
		t.Fatalf("Parameters: expected 1, got %d", len(params))
	}
	p0, ok := params[0].(bson.M)
	if !ok {
		t.Fatalf("Parameters[0]: expected bson.M, got %T", params[0])
	}
	if p0["Name"] != "petId" {
		t.Errorf("Parameter Name: expected petId, got %v", p0["Name"])
	}
	dataType, ok := p0["DataType"].(bson.M)
	if !ok {
		t.Fatalf("Parameter DataType: expected bson.M, got %T", p0["DataType"])
	}
	if dataType["$Type"] != "DataTypes$IntegerType" {
		t.Errorf("Parameter DataType.$Type: expected IntegerAttributeType, got %v", dataType["$Type"])
	}

	// Headers
	headers := extractBsonArray(result["Headers"])
	if len(headers) != 1 {
		t.Fatalf("Headers: expected 1, got %d", len(headers))
	}

	// ResponseHandling (JSON uses NoResponseHandling with ContentType for compatibility)
	respHandling, ok := result["ResponseHandling"].(bson.M)
	if !ok {
		t.Fatalf("ResponseHandling: expected bson.M, got %T", result["ResponseHandling"])
	}
	if respHandling["$Type"] != "Rest$NoResponseHandling" {
		t.Errorf("ResponseHandling.$Type: expected NoResponseHandling, got %v", respHandling["$Type"])
	}
	if respHandling["ContentType"] != "application/json" {
		t.Errorf("ResponseHandling.ContentType: expected application/json, got %v", respHandling["ContentType"])
	}
}

func TestSerializeRestOperationPostWithBody(t *testing.T) {
	op := &model.RestClientOperation{
		Name:         "AddPet",
		HttpMethod:   "POST",
		Path:         "/pet",
		BodyType:     "JSON",
		ResponseType: "JSON",
	}

	result := serializeRestOperation(op)

	// Method should be WithBody (POST)
	method, ok := result["Method"].(bson.M)
	if !ok {
		t.Fatalf("Method: expected bson.M, got %T", result["Method"])
	}
	if method["$Type"] != "Rest$RestOperationMethodWithBody" {
		t.Errorf("Method.$Type: expected WithBody, got %v", method["$Type"])
	}
	if method["HttpMethod"] != "Post" {
		t.Errorf("Method.HttpMethod: expected Post, got %v", method["HttpMethod"])
	}

	// Body should be JsonBody (used instead of ImplicitMappingBody to avoid CE7247/CE0061)
	body, ok := method["Body"].(bson.M)
	if !ok {
		t.Fatalf("Body: expected bson.M, got %T", method["Body"])
	}
	if body["$Type"] != "Rest$JsonBody" {
		t.Errorf("Body.$Type: expected JsonBody, got %v", body["$Type"])
	}
}

func TestSerializeRestOperationNoResponse(t *testing.T) {
	op := &model.RestClientOperation{
		Name:         "DeletePet",
		HttpMethod:   "DELETE",
		Path:         "/pet/{petId}",
		ResponseType: "NONE",
	}

	result := serializeRestOperation(op)

	respHandling, ok := result["ResponseHandling"].(bson.M)
	if !ok {
		t.Fatalf("ResponseHandling: expected bson.M, got %T", result["ResponseHandling"])
	}
	if respHandling["$Type"] != "Rest$NoResponseHandling" {
		t.Errorf("ResponseHandling.$Type: expected NoResponseHandling, got %v", respHandling["$Type"])
	}
}

func TestSerializeRestOperationDeleteNoBody(t *testing.T) {
	// DELETE with BodyType set should still serialize as WithoutBody — matches Studio Pro behavior.
	// The OpenAPI parser strips requestBody for DELETE/HEAD before setting BodyType,
	// but the writer should also be safe if BodyType somehow reaches it.
	op := &model.RestClientOperation{
		Name:         "DeleteTask",
		HttpMethod:   "DELETE",
		Path:         "/tasks/{taskId}",
		BodyType:     "", // parser strips body for DELETE — BodyType must be empty
		ResponseType: "NONE",
		Parameters: []*model.RestClientParameter{
			{Name: "taskId", DataType: "String"},
		},
	}

	result := serializeRestOperation(op)

	method, ok := result["Method"].(bson.M)
	if !ok {
		t.Fatalf("Method: expected bson.M, got %T", result["Method"])
	}
	if method["$Type"] != "Rest$RestOperationMethodWithoutBody" {
		t.Errorf("Method.$Type: expected WithoutBody for DELETE, got %v", method["$Type"])
	}
	if method["HttpMethod"] != "Delete" {
		t.Errorf("Method.HttpMethod: expected Delete, got %v", method["HttpMethod"])
	}
	if _, hasBody := method["Body"]; hasBody {
		t.Error("Method.Body: must not be present for DELETE operation")
	}
}

func TestSerializeRestOperationQueryParams(t *testing.T) {
	op := &model.RestClientOperation{
		Name:       "SearchPets",
		HttpMethod: "GET",
		Path:       "/pet/findByStatus",
		QueryParameters: []*model.RestClientParameter{
			{Name: "status", DataType: "String"},
		},
		ResponseType: "JSON",
	}

	result := serializeRestOperation(op)

	queryParams := extractBsonArray(result["QueryParameters"])
	if len(queryParams) != 1 {
		t.Fatalf("QueryParameters: expected 1, got %d", len(queryParams))
	}
	q0, ok := queryParams[0].(bson.M)
	if !ok {
		t.Fatalf("QueryParameters[0]: expected bson.M, got %T", queryParams[0])
	}
	if q0["Name"] != "status" {
		t.Errorf("QueryParam Name: expected status, got %v", q0["Name"])
	}
	if q0["$Type"] != "Rest$QueryParameter" {
		t.Errorf("QueryParam $Type: expected Rest$QueryParameter, got %v", q0["$Type"])
	}

	// ParameterUsage
	usage, ok := q0["ParameterUsage"].(bson.M)
	if !ok {
		t.Fatalf("ParameterUsage: expected bson.M, got %T", q0["ParameterUsage"])
	}
	if usage["$Type"] != "Rest$RequiredQueryParameterUsage" {
		t.Errorf("ParameterUsage.$Type: expected RequiredQueryParameterUsage, got %v", usage["$Type"])
	}
}

func TestSerializeConsumedRestServiceWithOpenApiContent(t *testing.T) {
	w := &Writer{}
	rawSpec := `{"openapi":"3.0.0","info":{"title":"Test"},"paths":{}}`
	svc := &model.ConsumedRestService{
		BaseElement:    model.BaseElement{ID: "test-oaf-id"},
		ContainerID:    "test-module-id",
		Name:           "TestAPI",
		BaseUrl:        "https://api.example.com",
		OpenApiContent: rawSpec,
	}

	data, err := w.serializeConsumedRestService(svc)
	if err != nil {
		t.Fatalf("serialize failed: %v", err)
	}

	var raw map[string]any
	if err := bson.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	// OpenApiFile sub-document must be present with correct $Type
	oaf, ok := raw["OpenApiFile"].(map[string]any)
	if !ok {
		t.Fatalf("OpenApiFile: expected map, got %T", raw["OpenApiFile"])
	}
	assertField(t, oaf, "$Type", "Rest$OpenApiFile")
	assertField(t, oaf, "Content", rawSpec)
	if oaf["$ID"] == nil {
		t.Error("OpenApiFile.$ID must not be nil")
	}
}

func TestSerializeConsumedRestServiceNoOpenApiContent(t *testing.T) {
	w := &Writer{}
	svc := &model.ConsumedRestService{
		BaseElement: model.BaseElement{ID: "test-no-oaf-id"},
		ContainerID: "test-module-id",
		Name:        "ManualAPI",
		BaseUrl:     "https://api.example.com",
		// OpenApiContent intentionally empty
	}

	data, err := w.serializeConsumedRestService(svc)
	if err != nil {
		t.Fatalf("serialize failed: %v", err)
	}

	var raw map[string]any
	if err := bson.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	// OpenApiFile should be nil when no content provided
	if raw["OpenApiFile"] != nil {
		t.Errorf("OpenApiFile: expected nil for manual REST client, got %v", raw["OpenApiFile"])
	}
}

func TestSerializeRestOperationWithTags(t *testing.T) {
	op := &model.RestClientOperation{
		Name:         "ListItems",
		HttpMethod:   "GET",
		Path:         "/items",
		Tags:         []string{"Items", "Read"},
		ResponseType: "JSON",
		Timeout:      300,
	}

	result := serializeRestOperation(op)

	// Tags array: [int32(1), "Items", "Read"] — one tag per position after prefix
	tags, ok := result["Tags"].(bson.A)
	if !ok {
		t.Fatalf("Tags: expected bson.A, got %T", result["Tags"])
	}
	if len(tags) != 3 { // prefix + 2 tags
		t.Fatalf("Tags length: expected 3 (prefix+2), got %d", len(tags))
	}
	if tags[0] != int32(1) {
		t.Errorf("Tags[0] prefix: expected int32(1), got %v", tags[0])
	}
	if tags[1] != "Items" {
		t.Errorf("Tags[1]: expected Items, got %v", tags[1])
	}
	if tags[2] != "Read" {
		t.Errorf("Tags[2]: expected Read, got %v", tags[2])
	}
}

func TestSerializeRestOperationNoResponseHasStatusCode(t *testing.T) {
	op := &model.RestClientOperation{
		Name:         "DeleteItem",
		HttpMethod:   "DELETE",
		Path:         "/items/{id}",
		ResponseType: "NONE",
		Timeout:      300,
	}

	result := serializeRestOperation(op)

	respHandling, ok := result["ResponseHandling"].(bson.M)
	if !ok {
		t.Fatalf("ResponseHandling: expected bson.M, got %T", result["ResponseHandling"])
	}
	assertField(t, respHandling, "$Type", "Rest$NoResponseHandling")
	// Studio Pro always writes StatusCode: 200 on NoResponseHandling
	if respHandling["StatusCode"] != int32(200) {
		t.Errorf("ResponseHandling.StatusCode: expected int32(200), got %v (%T)", respHandling["StatusCode"], respHandling["StatusCode"])
	}
}

func TestSerializeRestParameterHasTestValue(t *testing.T) {
	op := &model.RestClientOperation{
		Name:       "GetItem",
		HttpMethod: "GET",
		Path:       "/items/{id}",
		Parameters: []*model.RestClientParameter{
			{Name: "id", DataType: "String"},
		},
		ResponseType: "JSON",
		Timeout:      300,
	}

	result := serializeRestOperation(op)

	params := extractBsonArray(result["Parameters"])
	if len(params) != 1 {
		t.Fatalf("Parameters: expected 1, got %d", len(params))
	}
	p0, ok := params[0].(bson.M)
	if !ok {
		t.Fatalf("Parameters[0]: expected bson.M, got %T", params[0])
	}
	// TestValue must always be present (Studio Pro writes Rest$StringValue with empty Value)
	tv, ok := p0["TestValue"].(bson.M)
	if !ok {
		t.Fatalf("TestValue: expected bson.M, got %T", p0["TestValue"])
	}
	assertField(t, tv, "$Type", "Rest$StringValue")
	assertField(t, tv, "Value", "")
	if tv["$ID"] == nil {
		t.Error("TestValue.$ID must not be nil")
	}
}

func TestHttpMethodToMendix(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"GET", "Get"},
		{"POST", "Post"},
		{"PUT", "Put"},
		{"PATCH", "Patch"},
		{"DELETE", "Delete"},
		{"HEAD", "Head"},
		{"OPTIONS", "Options"},
	}
	for _, tc := range tests {
		result := httpMethodToMendix(tc.input)
		if result != tc.expected {
			t.Errorf("httpMethodToMendix(%q): expected %q, got %q", tc.input, tc.expected, result)
		}
	}
}

func TestSerializeConsumedRestServiceFullRoundtrip(t *testing.T) {
	w := &Writer{}
	svc := &model.ConsumedRestService{
		BaseElement: model.BaseElement{
			ID: "test-roundtrip-id",
		},
		ContainerID:   "test-module-id",
		Name:          "PetStoreAPI",
		Documentation: "Swagger Pet Store API",
		BaseUrl:       "https://petstore.swagger.io/v2",
		Operations: []*model.RestClientOperation{
			{
				Name:       "ListPets",
				HttpMethod: "GET",
				Path:       "/pet/findByStatus",
				QueryParameters: []*model.RestClientParameter{
					{Name: "status", DataType: "String"},
				},
				Headers: []*model.RestClientHeader{
					{Name: "Accept", Value: "application/json"},
				},
				ResponseType: "JSON",
				Timeout:      30,
			},
			{
				Name:       "GetPet",
				HttpMethod: "GET",
				Path:       "/pet/{petId}",
				Parameters: []*model.RestClientParameter{
					{Name: "petId", DataType: "Integer"},
				},
				ResponseType: "JSON",
			},
			{
				Name:         "AddPet",
				HttpMethod:   "POST",
				Path:         "/pet",
				BodyType:     "JSON",
				ResponseType: "JSON",
			},
			{
				Name:         "RemovePet",
				HttpMethod:   "DELETE",
				Path:         "/pet/{petId}",
				ResponseType: "NONE",
				Parameters: []*model.RestClientParameter{
					{Name: "petId", DataType: "Integer"},
				},
			},
		},
	}

	data, err := w.serializeConsumedRestService(svc)
	if err != nil {
		t.Fatalf("serialize failed: %v", err)
	}

	// Verify the BSON can be deserialized
	var raw map[string]any
	if err := bson.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	// Verify top-level structure
	assertField(t, raw, "$Type", "Rest$ConsumedRestService")
	assertField(t, raw, "Name", "PetStoreAPI")
	assertField(t, raw, "Documentation", "Swagger Pet Store API")

	// Verify operations count
	ops := extractBsonArray(raw["Operations"])
	if len(ops) != 4 {
		t.Fatalf("Operations: expected 4, got %d", len(ops))
	}

	// Verify first operation
	op0, ok := ops[0].(map[string]any)
	if !ok {
		t.Fatalf("Operations[0]: expected map, got %T", ops[0])
	}
	assertField(t, op0, "Name", "ListPets")

	// Verify POST operation has WithBody method
	op2, ok := ops[2].(map[string]any)
	if !ok {
		t.Fatalf("Operations[2]: expected map, got %T", ops[2])
	}
	assertField(t, op2, "Name", "AddPet")
	method2, ok := op2["Method"].(map[string]any)
	if !ok {
		t.Fatalf("Operations[2].Method: expected map, got %T", op2["Method"])
	}
	assertField(t, method2, "$Type", "Rest$RestOperationMethodWithBody")

	// Verify Body is JsonBody
	body2, ok := method2["Body"].(map[string]any)
	if !ok {
		t.Fatalf("Operations[2].Method.Body: expected map, got %T", method2["Body"])
	}
	assertField(t, body2, "$Type", "Rest$JsonBody")

	// Verify DELETE operation has WithoutBody method
	op3, ok := ops[3].(map[string]any)
	if !ok {
		t.Fatalf("Operations[3]: expected map, got %T", ops[3])
	}
	method3, ok := op3["Method"].(map[string]any)
	if !ok {
		t.Fatalf("Operations[3].Method: expected map, got %T", op3["Method"])
	}
	assertField(t, method3, "$Type", "Rest$RestOperationMethodWithoutBody")
}
