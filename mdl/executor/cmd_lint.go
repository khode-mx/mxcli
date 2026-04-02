// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/mdl/linter"
	"github.com/mendixlabs/mxcli/mdl/linter/rules"
)

// execLint executes a LINT statement.
func (e *Executor) execLint(s *ast.LintStmt) error {
	if e.reader == nil {
		return fmt.Errorf("not connected to a project")
	}

	// Handle SHOW LINT RULES
	if s.ShowRules {
		return e.showLintRules()
	}

	// Ensure catalog is built
	if e.catalog == nil {
		fmt.Fprintln(e.output, "Building catalog for linting...")
		if err := e.buildCatalog(false); err != nil {
			return fmt.Errorf("failed to build catalog: %w", err)
		}
	}

	// Create lint context
	ctx := linter.NewLintContext(e.catalog)
	ctx.SetReader(e.reader)

	// Load configuration
	projectDir := filepath.Dir(e.mprPath)
	configPath := linter.FindConfigFile(projectDir)
	config, err := linter.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(e.output, "Warning: failed to load lint config: %v\n", err)
		config = linter.DefaultConfig()
	}

	// Set excluded modules from config
	if len(config.ExcludeModules) > 0 {
		ctx.SetExcludedModules(config.ExcludeModules)
	}

	// Create linter and register built-in rules
	lint := linter.New(ctx)
	lint.AddRule(rules.NewNamingConventionRule())
	lint.AddRule(rules.NewEmptyMicroflowRule())
	lint.AddRule(rules.NewDomainModelSizeRule())
	lint.AddRule(rules.NewValidationFeedbackRule())
	lint.AddRule(rules.NewImageSourceRule())
	lint.AddRule(rules.NewMissingTranslationsRule())

	// Load custom Starlark rules
	rulesDir := filepath.Join(projectDir, ".claude", "lint-rules")
	starlarkRules, err := linter.LoadStarlarkRulesFromDir(rulesDir)
	if err != nil {
		fmt.Fprintf(e.output, "Warning: failed to load custom rules: %v\n", err)
	}
	for _, rule := range starlarkRules {
		lint.AddRule(rule)
	}

	// Apply configuration
	config.ApplyConfig(lint)

	// Handle module filtering
	if s.Target != nil && s.ModuleOnly {
		// Only lint specific module - set all others as excluded
		ctx.SetExcludedModules(nil) // Clear any existing exclusions
		// This is a simplified approach - ideally we'd filter in the linter
		fmt.Fprintf(e.output, "Linting module: %s\n", s.Target.Module)
	}

	// Run linting
	violations, err := lint.Run(context.Background())
	if err != nil {
		return fmt.Errorf("linting failed: %w", err)
	}

	// Filter violations if targeting specific module
	if s.Target != nil && s.ModuleOnly {
		filtered := make([]linter.Violation, 0)
		for _, v := range violations {
			if v.Location.Module == s.Target.Module {
				filtered = append(filtered, v)
			}
		}
		violations = filtered
	}

	// Output results
	var format linter.OutputFormat
	switch s.Format {
	case ast.LintFormatJSON:
		format = linter.OutputFormatJSON
	case ast.LintFormatSARIF:
		format = linter.OutputFormatSARIF
	default:
		format = linter.OutputFormatText
	}

	formatter := linter.GetFormatter(format, false)
	return formatter.Format(violations, e.output)
}

// showLintRules displays available lint rules.
func (e *Executor) showLintRules() error {
	fmt.Fprintln(e.output, "Built-in rules:")
	fmt.Fprintln(e.output)

	// Create a temporary linter with built-in rules
	lint := linter.New(nil)
	lint.AddRule(rules.NewNamingConventionRule())
	lint.AddRule(rules.NewEmptyMicroflowRule())
	lint.AddRule(rules.NewDomainModelSizeRule())
	lint.AddRule(rules.NewValidationFeedbackRule())
	lint.AddRule(rules.NewImageSourceRule())
	lint.AddRule(rules.NewMissingTranslationsRule())

	for _, rule := range lint.Rules() {
		fmt.Fprintf(e.output, "  %s (%s)\n", rule.ID(), rule.Name())
		fmt.Fprintf(e.output, "    %s\n", rule.Description())
		fmt.Fprintf(e.output, "    Category: %s, Default Severity: %s\n", rule.Category(), rule.DefaultSeverity())
		fmt.Fprintln(e.output)
	}

	// Show custom Starlark rules if connected
	if e.mprPath != "" {
		projectDir := filepath.Dir(e.mprPath)
		rulesDir := filepath.Join(projectDir, ".claude", "lint-rules")
		starlarkRules, err := linter.LoadStarlarkRulesFromDir(rulesDir)
		if err == nil && len(starlarkRules) > 0 {
			fmt.Fprintln(e.output, "Custom rules (from .claude/lint-rules/):")
			fmt.Fprintln(e.output)
			for _, rule := range starlarkRules {
				fmt.Fprintf(e.output, "  %s (%s)\n", rule.ID(), rule.Name())
				fmt.Fprintf(e.output, "    %s\n", rule.Description())
				fmt.Fprintf(e.output, "    Category: %s, Default Severity: %s\n", rule.Category(), rule.DefaultSeverity())
				fmt.Fprintln(e.output)
			}
		}
	}

	return nil
}
