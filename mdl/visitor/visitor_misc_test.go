// SPDX-License-Identifier: Apache-2.0

package visitor

import (
	"testing"

	"github.com/mendixlabs/mxcli/mdl/ast"
)

func TestSearch(t *testing.T) {
	input := `SEARCH 'customer';`
	prog, errs := Build(input)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("Parse error: %v", e)
		}
		return
	}
	stmt, ok := prog.Statements[0].(*ast.SearchStmt)
	if !ok {
		t.Fatalf("Expected SearchStmt, got %T", prog.Statements[0])
	}
	if stmt.Query != "customer" {
		t.Errorf("Got Query %q", stmt.Query)
	}
}

func TestExecuteScript(t *testing.T) {
	input := `EXECUTE SCRIPT 'myscript.mdl';`
	prog, errs := Build(input)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("Parse error: %v", e)
		}
		return
	}
	stmt, ok := prog.Statements[0].(*ast.ExecuteScriptStmt)
	if !ok {
		t.Fatalf("Expected ExecuteScriptStmt, got %T", prog.Statements[0])
	}
	if stmt.Path != "myscript.mdl" {
		t.Errorf("Got Path %q", stmt.Path)
	}
}

func TestHelp(t *testing.T) {
	input := `help;`
	prog, errs := Build(input)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("Parse error: %v", e)
		}
		return
	}
	_, ok := prog.Statements[0].(*ast.HelpStmt)
	if !ok {
		t.Fatalf("Expected HelpStmt, got %T", prog.Statements[0])
	}
}

func TestUpdate(t *testing.T) {
	input := `UPDATE;`
	prog, errs := Build(input)
	if len(errs) > 0 {
		for _, e := range errs {
			t.Errorf("Parse error: %v", e)
		}
		return
	}
	_, ok := prog.Statements[0].(*ast.UpdateStmt)
	if !ok {
		t.Fatalf("Expected UpdateStmt, got %T", prog.Statements[0])
	}
}
