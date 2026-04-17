// SPDX-License-Identifier: Apache-2.0

package visitor

import (
	"strings"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/mdl/grammar/parser"
)

// ExitCreateModelStatement bridges the createModelStatement parse tree to
// an ast.CreateModelStmt. Property values are dispatched by alternative:
// identifierOrKeyword (Provider), qualifiedName (Key), or STRING_LITERAL
// (Portal-populated metadata).
func (b *Builder) ExitCreateModelStatement(ctx *parser.CreateModelStatementContext) {
	stmt := &ast.CreateModelStmt{
		Name: buildQualifiedName(ctx.QualifiedName()),
	}
	stmt.Documentation = findDocCommentText(ctx)

	for _, p := range ctx.AllModelProperty() {
		propCtx := p.(*parser.ModelPropertyContext)
		// The first identifierOrKeyword is the property name; the second
		// (if present) or the literal is the value.
		idents := propCtx.AllIdentifierOrKeyword()
		if len(idents) == 0 {
			continue
		}
		key := strings.ToLower(idents[0].GetText())
		switch key {
		case "provider":
			if len(idents) > 1 {
				stmt.Provider = idents[1].GetText()
			}
		case "key":
			if qn := propCtx.QualifiedName(); qn != nil {
				k := buildQualifiedName(qn)
				stmt.Key = &k
			}
		case "displayname":
			if lit := propCtx.STRING_LITERAL(); lit != nil {
				stmt.DisplayName = unquoteString(lit.GetText())
			}
		case "keyname":
			if lit := propCtx.STRING_LITERAL(); lit != nil {
				stmt.KeyName = unquoteString(lit.GetText())
			}
		case "keyid":
			if lit := propCtx.STRING_LITERAL(); lit != nil {
				stmt.KeyID = unquoteString(lit.GetText())
			}
		case "environment":
			if lit := propCtx.STRING_LITERAL(); lit != nil {
				stmt.Environment = unquoteString(lit.GetText())
			}
		case "resourcename":
			if lit := propCtx.STRING_LITERAL(); lit != nil {
				stmt.ResourceName = unquoteString(lit.GetText())
			}
		case "deeplinkurl":
			if lit := propCtx.STRING_LITERAL(); lit != nil {
				stmt.DeepLinkURL = unquoteString(lit.GetText())
			}
		}
	}

	b.statements = append(b.statements, stmt)
}
