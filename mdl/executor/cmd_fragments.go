// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/mendixlabs/mxcli/mdl/ast"
	"github.com/mendixlabs/mxcli/sdk/pages"
)

// execDefineFragment stores a fragment definition in the executor's session state.
func (e *Executor) execDefineFragment(s *ast.DefineFragmentStmt) error {
	if e.fragments == nil {
		e.fragments = make(map[string]*ast.DefineFragmentStmt)
	}
	if _, exists := e.fragments[s.Name]; exists {
		return fmt.Errorf("fragment %q already defined", s.Name)
	}
	e.fragments[s.Name] = s
	fmt.Fprintf(e.output, "Defined fragment %s (%d widgets)\n", s.Name, len(s.Widgets))
	return nil
}

// showFragments lists all defined fragments in the current session.
func (e *Executor) showFragments() error {
	if len(e.fragments) == 0 {
		fmt.Fprintln(e.output, "No fragments defined.")
		return nil
	}

	// Sort by name for consistent output
	names := make([]string, 0, len(e.fragments))
	for name := range e.fragments {
		names = append(names, name)
	}
	sort.Strings(names)

	fmt.Fprintf(e.output, "%-30s %s\n", "Fragment", "Widgets")
	fmt.Fprintf(e.output, "%-30s %s\n", strings.Repeat("-", 30), strings.Repeat("-", 10))
	for _, name := range names {
		frag := e.fragments[name]
		fmt.Fprintf(e.output, "%-30s %d\n", name, len(frag.Widgets))
	}
	return nil
}

// describeFragment outputs a fragment's definition as MDL.
func (e *Executor) describeFragment(name ast.QualifiedName) error {
	if e.fragments == nil {
		return fmt.Errorf("fragment %q not found", name.Name)
	}
	frag, ok := e.fragments[name.Name]
	if !ok {
		return fmt.Errorf("fragment %q not found", name.Name)
	}

	fmt.Fprintf(e.output, "DEFINE FRAGMENT %s AS {\n", frag.Name)
	for _, w := range frag.Widgets {
		outputASTWidgetMDL(e.output, w, 1)
	}
	fmt.Fprintln(e.output, "};")
	return nil
}

// describeFragmentFrom handles DESCRIBE FRAGMENT FROM PAGE/SNIPPET ... WIDGET ... command.
// It finds a named widget in a page or snippet and outputs it as MDL.
func (e *Executor) describeFragmentFrom(s *ast.DescribeFragmentFromStmt) error {
	if e.reader == nil {
		return fmt.Errorf("not connected to a project")
	}

	h, err := e.getHierarchy()
	if err != nil {
		return fmt.Errorf("failed to build hierarchy: %w", err)
	}

	var rawWidgets []rawWidget

	switch s.ContainerType {
	case "PAGE":
		allPages, err := e.reader.ListPages()
		if err != nil {
			return fmt.Errorf("failed to list pages: %w", err)
		}
		var foundPage *pages.Page
		for _, p := range allPages {
			modID := h.FindModuleID(p.ContainerID)
			modName := h.GetModuleName(modID)
			if p.Name == s.ContainerName.Name && (s.ContainerName.Module == "" || modName == s.ContainerName.Module) {
				foundPage = p
				break
			}
		}
		if foundPage == nil {
			return fmt.Errorf("page %s not found", s.ContainerName.String())
		}
		rawWidgets = e.getPageWidgetsFromRaw(foundPage.ID)

	case "SNIPPET":
		allSnippets, err := e.reader.ListSnippets()
		if err != nil {
			return fmt.Errorf("failed to list snippets: %w", err)
		}
		var foundSnippet *pages.Snippet
		for _, sn := range allSnippets {
			modID := h.FindModuleID(sn.ContainerID)
			modName := h.GetModuleName(modID)
			if sn.Name == s.ContainerName.Name && (s.ContainerName.Module == "" || modName == s.ContainerName.Module) {
				foundSnippet = sn
				break
			}
		}
		if foundSnippet == nil {
			return fmt.Errorf("snippet %s not found", s.ContainerName.String())
		}
		rawWidgets = e.getSnippetWidgetsFromRaw(foundSnippet.ID)
	}

	// Find the widget by name
	target := findRawWidgetByName(rawWidgets, s.WidgetName)
	if target == nil {
		return fmt.Errorf("widget %q not found in %s %s", s.WidgetName, strings.ToLower(s.ContainerType), s.ContainerName.String())
	}

	// Output as MDL
	e.outputWidgetMDLV3(*target, 0)
	return nil
}

// findRawWidgetByName recursively searches the widget tree for a widget with the given name.
func findRawWidgetByName(widgets []rawWidget, name string) *rawWidget {
	for i := range widgets {
		if widgets[i].Name == name {
			return &widgets[i]
		}
		// Search children
		if found := findRawWidgetByName(widgets[i].Children, name); found != nil {
			return found
		}
		// Search rows (for LayoutGrid)
		for _, row := range widgets[i].Rows {
			for _, col := range row.Columns {
				if found := findRawWidgetByName(col.Widgets, name); found != nil {
					return found
				}
			}
		}
		// Search filter/controlbar widgets
		if found := findRawWidgetByName(widgets[i].FilterWidgets, name); found != nil {
			return found
		}
		if found := findRawWidgetByName(widgets[i].ControlBar, name); found != nil {
			return found
		}
	}
	return nil
}

// outputASTWidgetMDL outputs an AST WidgetV3 as MDL text.
func outputASTWidgetMDL(w io.Writer, widget *ast.WidgetV3, indent int) {
	prefix := strings.Repeat("  ", indent)

	// Widget type and name
	fmt.Fprintf(w, "%s%s %s", prefix, widget.Type, widget.Name)

	// Properties (excluding internal ones like Prefix)
	props := formatASTWidgetProps(widget)
	if props != "" {
		fmt.Fprintf(w, " (%s)", props)
	}

	// Children
	if len(widget.Children) > 0 {
		fmt.Fprintln(w, " {")
		for _, child := range widget.Children {
			outputASTWidgetMDL(w, child, indent+1)
		}
		fmt.Fprintf(w, "%s}\n", prefix)
	} else {
		fmt.Fprintln(w)
	}
}

// formatASTWidgetProps formats widget properties as "Key: Value, Key: Value".
func formatASTWidgetProps(w *ast.WidgetV3) string {
	if len(w.Properties) == 0 {
		return ""
	}

	// Sort keys for deterministic output
	keys := make([]string, 0, len(w.Properties))
	for k := range w.Properties {
		if k == "Prefix" {
			continue // Internal property, not MDL
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	if len(keys) == 0 {
		return ""
	}

	var parts []string
	for _, k := range keys {
		v := w.Properties[k]
		parts = append(parts, fmt.Sprintf("%s: %s", k, formatASTPropertyValue(v)))
	}
	return strings.Join(parts, ", ")
}

// formatASTPropertyValue formats a property value for MDL output.
func formatASTPropertyValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("'%s'", val)
	case int:
		return fmt.Sprintf("%d", val)
	case int64:
		return fmt.Sprintf("%d", val)
	case float64:
		return fmt.Sprintf("%g", val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	case *ast.DataSourceV3:
		return formatDataSourceV3(val)
	case *ast.ActionV3:
		return formatActionV3(val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func formatDataSourceV3(ds *ast.DataSourceV3) string {
	switch ds.Type {
	case "parameter":
		return ds.Reference
	case "database":
		return "DATABASE " + ds.Reference
	case "microflow":
		return "MICROFLOW " + ds.Reference
	case "nanoflow":
		return "NANOFLOW " + ds.Reference
	case "association":
		return "ASSOCIATION " + ds.Reference
	case "selection":
		return "SELECTION " + ds.Reference
	default:
		return ds.Reference
	}
}

func formatActionV3(a *ast.ActionV3) string {
	switch a.Type {
	case "save":
		if a.ClosePage {
			return "SAVE_CHANGES CLOSE_PAGE"
		}
		return "SAVE_CHANGES"
	case "cancel":
		if a.ClosePage {
			return "CANCEL_CHANGES CLOSE_PAGE"
		}
		return "CANCEL_CHANGES"
	case "close":
		return "CLOSE_PAGE"
	case "delete":
		return "DELETE_OBJECT"
	case "showPage":
		return "SHOW_PAGE " + a.Target
	case "microflow":
		return "MICROFLOW " + a.Target
	case "nanoflow":
		return "NANOFLOW " + a.Target
	case "signOut":
		return "SIGN_OUT"
	case "completeTask":
		return "COMPLETE_TASK '" + strings.ReplaceAll(a.OutcomeValue, "'", "''") + "'"
	default:
		return a.Type
	}
}
