// SPDX-License-Identifier: Apache-2.0

package visitor

import (
	"strings"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/mdl/grammar/parser"
)

// ExitImportRestClientStatement handles:
//
//	IMPORT [OR REPLACE] REST CLIENT Module.Name FROM OPENAPI '/path/to/spec.json'
//	  [SET BaseUrl: '...', Authentication: BASIC (Username: '...', Password: '...')]
func (b *Builder) ExitImportRestClientStatement(ctx *parser.ImportRestClientStatementContext) {
	stmt := &ast.ImportRestClientStmt{
		Name:    buildQualifiedName(ctx.QualifiedName()),
		Replace: ctx.REPLACE() != nil,
	}
	if sl := ctx.STRING_LITERAL(); sl != nil {
		stmt.SpecPath = unquoteString(sl.GetText())
	}

	// Parse optional SET properties (BaseUrl, Authentication)
	for _, propCtx := range ctx.AllRestClientProperty() {
		pc, ok := propCtx.(*parser.RestClientPropertyContext)
		if !ok || pc == nil || pc.IdentifierOrKeyword() == nil {
			continue
		}
		key := strings.ToLower(identifierOrKeywordText(pc.IdentifierOrKeyword().(*parser.IdentifierOrKeywordContext)))
		switch key {
		case "baseurl":
			if sl := pc.STRING_LITERAL(); sl != nil {
				stmt.BaseUrlOverride = unquoteString(sl.GetText())
			}
		case "authentication":
			if pc.BASIC() != nil {
				authDef := &ast.RestAuthDef{Scheme: "BASIC"}
				for _, subProp := range pc.AllRestClientProperty() {
					sp, spOk := subProp.(*parser.RestClientPropertyContext)
					if !spOk || sp == nil || sp.IdentifierOrKeyword() == nil {
						continue
					}
					subKey := strings.ToLower(identifierOrKeywordText(sp.IdentifierOrKeyword().(*parser.IdentifierOrKeywordContext)))
					if sl := sp.STRING_LITERAL(); sl != nil {
						switch subKey {
						case "username":
							authDef.Username = unquoteString(sl.GetText())
						case "password":
							authDef.Password = unquoteString(sl.GetText())
						}
					}
				}
				stmt.AuthOverride = authDef
			}
			// NONE or omitted: leave AuthOverride nil (no override)
		}
	}

	b.statements = append(b.statements, stmt)
}
