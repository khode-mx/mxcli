// SPDX-License-Identifier: Apache-2.0

package ast

// CreateModelStmt represents:
//
//	CREATE MODEL Module.Name (
//	  Provider: MxCloudGenAI,
//	  Key: Module.SomeConstant
//	  -- optional Portal-populated fields:
//	  [, DisplayName: '...']
//	  [, KeyName: '...']
//	  [, KeyId: '...']
//	  [, Environment: '...']
//	  [, ResourceName: '...']
//	  [, DeepLinkURL: '...']
//	);
type CreateModelStmt struct {
	Name             QualifiedName
	Documentation    string
	Provider         string         // "MxCloudGenAI" by default
	Key              *QualifiedName // qualified name of the String constant holding the Portal key
	DisplayName      string         // optional Portal-populated metadata
	KeyName          string         // optional Portal-populated metadata
	KeyID            string         // optional Portal-populated metadata
	Environment      string         // optional Portal-populated metadata
	ResourceName     string         // optional Portal-populated metadata
	DeepLinkURL      string         // optional Portal-populated metadata
}

func (s *CreateModelStmt) isStatement() {}

// DropModelStmt represents: DROP MODEL Module.Name
type DropModelStmt struct {
	Name QualifiedName
}

func (s *DropModelStmt) isStatement() {}
