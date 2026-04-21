// SPDX-License-Identifier: Apache-2.0

package visitor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
)

func TestConnect_Local(t *testing.T) {
	input := `CONNECT LOCAL '/path/to/project';`
	prog, errs := Build(input)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("Parse error: %v", e)
		}
		return
	}
	stmt, ok := prog.Statements[0].(*ast.ConnectStmt)
	if !ok {
		t.Fatalf("Expected ConnectStmt, got %T", prog.Statements[0])
	}
	if stmt.Path != "/path/to/project" {
		t.Errorf("Expected /path/to/project, got %q", stmt.Path)
	}
}

func TestDisconnect(t *testing.T) {
	input := `DISCONNECT;`
	prog, errs := Build(input)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("Parse error: %v", e)
		}
		return
	}
	_, ok := prog.Statements[0].(*ast.DisconnectStmt)
	if !ok {
		t.Fatalf("Expected DisconnectStmt, got %T", prog.Statements[0])
	}
}
