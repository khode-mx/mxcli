// SPDX-License-Identifier: Apache-2.0

package visitor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
)

func TestCreateExportMapping(t *testing.T) {
	input := `CREATE EXPORT MAPPING MyModule.PetExport WITH JSON STRUCTURE MyModule.PetSchema {
		MyModule.Pet {
			name = Name,
			id = PetId
		}
	};`
	prog, errs := Build(input)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("Parse error: %v", e)
		}
		return
	}
	stmt, ok := prog.Statements[0].(*ast.CreateExportMappingStmt)
	if !ok {
		t.Fatalf("Expected CreateExportMappingStmt, got %T", prog.Statements[0])
	}
	if stmt.Name.Name != "PetExport" {
		t.Errorf("Got name %s", stmt.Name.Name)
	}
	if stmt.SchemaKind != "JSON_STRUCTURE" {
		t.Errorf("Got SchemaKind %q", stmt.SchemaKind)
	}
	if stmt.RootElement == nil {
		t.Fatal("Expected non-nil RootElement")
	}
}
