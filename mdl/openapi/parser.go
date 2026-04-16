// SPDX-License-Identifier: Apache-2.0

// Package openapi parses OpenAPI 3.0 JSON specifications and converts them into
// mxcli model types for use with IMPORT REST CLIENT FROM OPENAPI.
package openapi

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/mendixlabs/mxcli/model"
)

// spec is the minimal subset of an OpenAPI 3.0 document that we need.
type spec struct {
	Info    specInfo            `json:"info"`
	Servers []specServer        `json:"servers"`
	Paths   map[string]pathItem `json:"paths"`
	// Components for $ref resolution
	Components specComponents `json:"components"`
}

type specInfo struct {
	Title string `json:"title"`
}

type specServer struct {
	URL string `json:"url"`
}

type pathItem struct {
	Get     *operation `json:"get"`
	Post    *operation `json:"post"`
	Put     *operation `json:"put"`
	Patch   *operation `json:"patch"`
	Delete  *operation `json:"delete"`
	Head    *operation `json:"head"`
	Options *operation `json:"options"`
	// Path-level parameters (inherited by all operations)
	Parameters []parameterOrRef `json:"parameters"`
}

type operation struct {
	OperationID string                   `json:"operationId"`
	Tags        []string                 `json:"tags"`
	Parameters  []parameterOrRef         `json:"parameters"`
	RequestBody *requestBodyOrRef        `json:"requestBody"`
	Responses   map[string]responseOrRef `json:"responses"`
	Deprecated  bool                     `json:"deprecated"`
}

type parameterOrRef struct {
	Ref      string     `json:"$ref"`
	Name     string     `json:"name"`
	In       string     `json:"in"` // "path", "query", "header", "cookie"
	Required bool       `json:"required"`
	Schema   *schemaRef `json:"schema"`
}

type requestBodyOrRef struct {
	Ref string `json:"$ref"`
}

type responseOrRef struct {
	Ref string `json:"$ref"`
}

type schemaRef struct {
	Ref    string `json:"$ref"`
	Type   string `json:"type"`   // "string", "integer", "number", "boolean", "array", "object"
	Format string `json:"format"` // "int32", "int64", "float", "double", etc.
}

type specComponents struct {
	Parameters map[string]parameterOrRef `json:"parameters"`
}

// ParseSpec reads an OpenAPI 3.0 JSON spec and returns a partially-filled
// ConsumedRestService. ContainerID and Name are NOT set — the caller sets them.
// The raw spec bytes are stored in OpenApiContent as-is.
func ParseSpec(specJSON []byte) (*model.ConsumedRestService, error) {
	var s spec
	if err := json.Unmarshal(specJSON, &s); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI spec: %w", err)
	}

	svc := &model.ConsumedRestService{
		OpenApiContent: string(specJSON),
	}

	// Base URL from first server entry
	if len(s.Servers) > 0 {
		svc.BaseUrl = strings.TrimRight(s.Servers[0].URL, "/")
	}

	// Build operations from all paths and HTTP methods
	for path, item := range s.Paths {
		methodOps := []struct {
			method string
			op     *operation
		}{
			{"GET", item.Get},
			{"POST", item.Post},
			{"PUT", item.Put},
			{"PATCH", item.Patch},
			{"DELETE", item.Delete},
			{"HEAD", item.Head},
			{"OPTIONS", item.Options},
		}

		for _, mo := range methodOps {
			if mo.op == nil {
				continue
			}
			op, err := buildOperation(mo.method, path, mo.op, item.Parameters, &s)
			if err != nil {
				return nil, fmt.Errorf("path %s %s: %w", mo.method, path, err)
			}
			svc.Operations = append(svc.Operations, op)
		}
	}

	return svc, nil
}

// buildOperation converts a single OpenAPI operation into a RestClientOperation.
func buildOperation(method, path string, op *operation, pathParams []parameterOrRef, s *spec) (*model.RestClientOperation, error) {
	methodUpper := strings.ToUpper(method)
	restOp := &model.RestClientOperation{
		Name:         sanitizeOperationID(op.OperationID, method, path),
		HttpMethod:   methodUpper,
		Path:         path,
		Tags:         op.Tags,
		ResponseType: "JSON", // default; overridden below if no 200 response
		Timeout:      300,    // Mendix default, matching Studio Pro output
	}

	// Merge path-level parameters with operation-level parameters.
	// Operation parameters override path-level parameters with the same name+in.
	allParams := mergeParams(pathParams, op.Parameters, s)

	for _, p := range allParams {
		// Resolve $ref if needed
		resolved := resolveParam(p, s)

		typeName := schemaToMendixType(resolved.Schema)
		param := &model.RestClientParameter{
			Name:      resolved.Name,
			DataType:  typeName,
			TestValue: "", // Studio Pro writes empty string for OpenAPI imports
		}
		switch resolved.In {
		case "path":
			restOp.Parameters = append(restOp.Parameters, param)
		case "query":
			restOp.QueryParameters = append(restOp.QueryParameters, param)
			// header and cookie params are omitted — Mendix manages these differently
		}
	}

	// Request body: if present, mark operation as having a JSON body.
	// DELETE and HEAD cannot have a body in Mendix — match Studio Pro's silent-drop behavior.
	if op.RequestBody != nil && methodUpper != "DELETE" && methodUpper != "HEAD" {
		restOp.BodyType = "JSON"
	}

	// Response type: check for a 2xx response
	hasSuccessResponse := false
	for code := range op.Responses {
		if code == "200" || code == "201" || code == "2XX" || code == "default" {
			hasSuccessResponse = true
			break
		}
	}
	if !hasSuccessResponse {
		restOp.ResponseType = "NONE"
	}

	return restOp, nil
}

// mergeParams merges path-level and operation-level parameters.
// Operation parameters with the same (name, in) pair override path-level ones.
// Within-list duplicates are preserved as-is, matching Studio Pro behaviour.
func mergeParams(pathParams, opParams []parameterOrRef, s *spec) []parameterOrRef {
	type key struct{ name, in string }
	// Index only path-level params so operation-level params can override them.
	// Within-list duplicates (e.g. duplicate query params in the spec) are kept
	// as-is rather than silently deduplicated.
	pathIndex := make(map[key]int)
	merged := make([]parameterOrRef, 0, len(pathParams)+len(opParams))

	for _, p := range pathParams {
		r := resolveParam(p, s)
		k := key{r.Name, r.In}
		if _, exists := pathIndex[k]; !exists {
			pathIndex[k] = len(merged)
			merged = append(merged, p)
		}
	}
	for _, p := range opParams {
		r := resolveParam(p, s)
		k := key{r.Name, r.In}
		if idx, exists := pathIndex[k]; exists {
			merged[idx] = p // operation overrides path-level param
		} else {
			merged = append(merged, p)
		}
	}
	return merged
}

// resolveParam resolves a $ref parameter to its concrete definition.
func resolveParam(p parameterOrRef, s *spec) parameterOrRef {
	if p.Ref == "" {
		return p
	}
	// Only support #/components/parameters/<name>
	const prefix = "#/components/parameters/"
	if strings.HasPrefix(p.Ref, prefix) {
		name := strings.TrimPrefix(p.Ref, prefix)
		if resolved, ok := s.Components.Parameters[name]; ok {
			return resolved
		}
	}
	return p
}

// schemaToMendixType converts an OpenAPI schema type to a Mendix data type name.
func schemaToMendixType(schema *schemaRef) string {
	if schema == nil {
		return "String"
	}
	switch schema.Type {
	case "integer":
		if schema.Format == "int64" {
			return "Long"
		}
		return "Integer"
	case "number":
		return "Decimal"
	case "boolean":
		return "Boolean"
	default:
		return "String"
	}
}

// nonAlphanumRe matches characters that are not valid in a Mendix identifier.
var nonAlphanumRe = regexp.MustCompile(`[^a-zA-Z0-9_]`)

// sanitizeOperationID converts an OpenAPI operationId (or generates one from
// the method+path) into a valid Mendix operation name (alphanumeric + underscore,
// starting with a letter).
func sanitizeOperationID(operationID, method, path string) string {
	name := operationID
	if name == "" {
		// Generate: METHOD_path_segments (e.g. GET__api_v1_pets → GET_api_v1_pets)
		slug := nonAlphanumRe.ReplaceAllString(path, "_")
		slug = strings.Trim(slug, "_")
		name = method + "_" + slug
	}

	// Replace any remaining invalid characters
	name = nonAlphanumRe.ReplaceAllString(name, "_")

	// Collapse consecutive underscores
	for strings.Contains(name, "__") {
		name = strings.ReplaceAll(name, "__", "_")
	}
	name = strings.Trim(name, "_")

	// Must start with a letter
	if len(name) > 0 && (name[0] >= '0' && name[0] <= '9') {
		name = "op_" + name
	}

	if name == "" {
		name = strings.ToLower(method) + "_operation"
	}
	return name
}
