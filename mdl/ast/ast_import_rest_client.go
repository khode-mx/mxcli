// SPDX-License-Identifier: Apache-2.0

package ast

// ImportRestClientStmt represents:
//
//	IMPORT [OR REPLACE] REST CLIENT Module.Name FROM OPENAPI '/path/to/spec.json'
//
// Reads an OpenAPI 3.0 JSON spec from disk and creates a consumed REST service
// document in the connected Mendix project.
type ImportRestClientStmt struct {
	// Name is the qualified name (Module.ServiceName) to give the imported service.
	Name QualifiedName
	// SpecPath is the file system path to the OpenAPI 3.0 JSON spec.
	SpecPath string
	// Replace indicates that an existing service with the same name should be
	// replaced (IMPORT OR REPLACE). Without this flag, importing over an existing
	// service is an error.
	Replace bool
	// Folder is an optional subfolder within the module (reserved for future use).
	Folder string
	// BaseUrlOverride overrides the BaseUrl parsed from the spec's servers[0].url.
	// Set via: SET BaseUrl: 'https://...'
	BaseUrlOverride string
	// AuthOverride overrides the authentication parsed from the spec (or the default NONE).
	// Set via: SET Authentication: BASIC (Username: '...', Password: '...')
	// nil means no override (use whatever the spec or default provides).
	AuthOverride *RestAuthDef
}

func (s *ImportRestClientStmt) isStatement() {}

// DescribeOpenapiFileStmt represents:
//
//	DESCRIBE OPENAPI FILE '/path/to/spec.json'
//
// A read-only command: parses an OpenAPI spec from disk and outputs a
// CREATE REST CLIENT statement preview without modifying any project.
type DescribeOpenapiFileStmt struct {
	// SpecPath is the file system path to the OpenAPI 3.0 JSON spec.
	SpecPath string
}

func (s *DescribeOpenapiFileStmt) isStatement() {}
