// SPDX-License-Identifier: Apache-2.0

package openapi

import (
	"testing"
)

// minimalSpec builds a minimal valid OpenAPI 3.0 JSON spec with the given paths JSON.
func minimalSpec(servers, paths string) []byte {
	return []byte(`{"openapi":"3.0.1","info":{"title":"Test"},"servers":` + servers + `,"paths":` + paths + `}`)
}

func TestParseSpec_BasicOperation(t *testing.T) {
	json := minimalSpec(
		`[{"url":"https://api.example.com/v1"}]`,
		`{"/pets/{petId}":{"get":{"operationId":"getPet","parameters":[{"name":"petId","in":"path","schema":{"type":"string"}},{"name":"format","in":"query","schema":{"type":"string"}}],"responses":{"200":{}}}}}`,
	)

	svc, err := ParseSpec(json)
	if err != nil {
		t.Fatalf("ParseSpec error: %v", err)
	}

	if svc.BaseUrl != "https://api.example.com/v1" {
		t.Errorf("BaseUrl: got %q, want %q", svc.BaseUrl, "https://api.example.com/v1")
	}
	if len(svc.Operations) != 1 {
		t.Fatalf("Operations count: got %d, want 1", len(svc.Operations))
	}

	op := svc.Operations[0]
	if op.Name != "getPet" {
		t.Errorf("Name: got %q, want %q", op.Name, "getPet")
	}
	if op.HttpMethod != "GET" {
		t.Errorf("HttpMethod: got %q, want %q", op.HttpMethod, "GET")
	}
	if op.Path != "/pets/{petId}" {
		t.Errorf("Path: got %q, want %q", op.Path, "/pets/{petId}")
	}
	if len(op.Parameters) != 1 || op.Parameters[0].Name != "petId" {
		t.Errorf("Parameters: got %v", op.Parameters)
	}
	if len(op.QueryParameters) != 1 || op.QueryParameters[0].Name != "format" {
		t.Errorf("QueryParameters: got %v", op.QueryParameters)
	}
	if op.ResponseType != "JSON" {
		t.Errorf("ResponseType: got %q, want JSON", op.ResponseType)
	}
}

func TestParseSpec_NoServers(t *testing.T) {
	json := minimalSpec(
		`[]`,
		`{"/ping":{"get":{"operationId":"ping","responses":{"200":{}}}}}`,
	)

	svc, err := ParseSpec(json)
	if err != nil {
		t.Fatalf("ParseSpec error: %v", err)
	}
	if svc.BaseUrl != "" {
		t.Errorf("BaseUrl: got %q, want empty string when no servers", svc.BaseUrl)
	}
}

func TestParseSpec_BaseUrlTrimTrailingSlash(t *testing.T) {
	json := minimalSpec(
		`[{"url":"https://api.example.com/v1/"}]`,
		`{"/ping":{"get":{"operationId":"ping","responses":{"200":{}}}}}`,
	)

	svc, err := ParseSpec(json)
	if err != nil {
		t.Fatalf("ParseSpec error: %v", err)
	}
	if svc.BaseUrl != "https://api.example.com/v1" {
		t.Errorf("BaseUrl: trailing slash not stripped, got %q", svc.BaseUrl)
	}
}

func TestParseSpec_DeleteNoBody(t *testing.T) {
	json := minimalSpec(
		`[{"url":"https://api.example.com"}]`,
		`{"/items/{id}":{"delete":{"operationId":"deleteItem","requestBody":{},"responses":{"204":{}}}}}`,
	)

	svc, err := ParseSpec(json)
	if err != nil {
		t.Fatalf("ParseSpec error: %v", err)
	}
	if len(svc.Operations) != 1 {
		t.Fatalf("Operations count: got %d, want 1", len(svc.Operations))
	}
	op := svc.Operations[0]
	if op.BodyType != "" {
		t.Errorf("DELETE BodyType: got %q, want empty (Studio Pro drops body on DELETE)", op.BodyType)
	}
	if op.ResponseType != "NONE" {
		t.Errorf("ResponseType: got %q, want NONE (204 response)", op.ResponseType)
	}
}

func TestParseSpec_DuplicateParams(t *testing.T) {
	// Spec has $type and $key declared twice in query params — matches the
	// Capital API bug. Both occurrences must be preserved (Studio Pro behaviour).
	json := minimalSpec(
		`[{"url":"https://api.example.com"}]`,
		`{"/items":{"get":{"operationId":"getItems","parameters":[{"name":"type","in":"query","schema":{"type":"string"}},{"name":"key","in":"query","schema":{"type":"string"}},{"name":"type","in":"query","schema":{"type":"string"}},{"name":"key","in":"query","schema":{"type":"string"}}],"responses":{"200":{}}}}}`,
	)

	svc, err := ParseSpec(json)
	if err != nil {
		t.Fatalf("ParseSpec error: %v", err)
	}
	if len(svc.Operations) != 1 {
		t.Fatalf("Operations count: got %d, want 1", len(svc.Operations))
	}
	op := svc.Operations[0]
	if len(op.QueryParameters) != 4 {
		t.Errorf("QueryParameters count: got %d, want 4 (both duplicates preserved)", len(op.QueryParameters))
	}
}

func TestParseSpec_PathParamOverridesPathLevel(t *testing.T) {
	// Operation-level parameter with same name+in as path-level should override it.
	json := minimalSpec(
		`[{"url":"https://api.example.com"}]`,
		`{"/items/{id}":{"parameters":[{"name":"id","in":"path","schema":{"type":"string"}}],"get":{"operationId":"getItem","parameters":[{"name":"id","in":"path","schema":{"type":"integer","format":"int32"}}],"responses":{"200":{}}}}}`,
	)

	svc, err := ParseSpec(json)
	if err != nil {
		t.Fatalf("ParseSpec error: %v", err)
	}
	op := svc.Operations[0]
	if len(op.Parameters) != 1 {
		t.Fatalf("Parameters count: got %d, want 1 (operation overrides path-level)", len(op.Parameters))
	}
	if op.Parameters[0].DataType != "Integer" {
		t.Errorf("DataType: got %q, want Integer (operation-level override)", op.Parameters[0].DataType)
	}
}

func TestParseSpec_TypeMapping(t *testing.T) {
	json := minimalSpec(
		`[{"url":"https://api.example.com"}]`,
		`{"/test":{"get":{"operationId":"test","parameters":[`+
			`{"name":"a","in":"query","schema":{"type":"integer","format":"int32"}},`+
			`{"name":"b","in":"query","schema":{"type":"integer","format":"int64"}},`+
			`{"name":"c","in":"query","schema":{"type":"number"}},`+
			`{"name":"d","in":"query","schema":{"type":"boolean"}},`+
			`{"name":"e","in":"query","schema":{"type":"string"}}`+
			`],"responses":{"200":{}}}}}`,
	)

	svc, err := ParseSpec(json)
	if err != nil {
		t.Fatalf("ParseSpec error: %v", err)
	}
	op := svc.Operations[0]
	expected := []string{"Integer", "Long", "Decimal", "Boolean", "String"}
	if len(op.QueryParameters) != len(expected) {
		t.Fatalf("QueryParameters count: got %d, want %d", len(op.QueryParameters), len(expected))
	}
	for i, want := range expected {
		if op.QueryParameters[i].DataType != want {
			t.Errorf("param[%d] DataType: got %q, want %q", i, op.QueryParameters[i].DataType, want)
		}
	}
}
