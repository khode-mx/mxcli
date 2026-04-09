// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mendixlabs/mxcli/mdl/ast"
)

// showPublishedRestServices handles SHOW PUBLISHED REST SERVICES [IN module] command.
func (e *Executor) showPublishedRestServices(moduleName string) error {
	services, err := e.reader.ListPublishedRestServices()
	if err != nil {
		return fmt.Errorf("failed to list published REST services: %w", err)
	}

	h, err := e.getHierarchy()
	if err != nil {
		return fmt.Errorf("failed to build hierarchy: %w", err)
	}

	type row struct {
		module        string
		qualifiedName string
		path          string
		version       string
		resources     int
		operations    int
	}
	var rows []row

	for _, svc := range services {
		modID := h.FindModuleID(svc.ContainerID)
		modName := h.GetModuleName(modID)
		if moduleName != "" && !strings.EqualFold(modName, moduleName) {
			continue
		}

		qn := modName + "." + svc.Name
		opCount := 0
		for _, res := range svc.Resources {
			opCount += len(res.Operations)
		}

		path := svc.Path
		if len(path) > 50 {
			path = path[:47] + "..."
		}

		rows = append(rows, row{modName, qn, path, svc.Version, len(svc.Resources), opCount})
	}

	if len(rows) == 0 {
		fmt.Fprintln(e.output, "No published REST services found.")
		return nil
	}

	sort.Slice(rows, func(i, j int) bool {
		return strings.ToLower(rows[i].qualifiedName) < strings.ToLower(rows[j].qualifiedName)
	})

	result := &TableResult{
		Columns: []string{"Module", "QualifiedName", "Path", "Version", "Resources", "Operations"},
		Summary: fmt.Sprintf("(%d published REST services)", len(rows)),
	}
	for _, r := range rows {
		result.Rows = append(result.Rows, []any{r.module, r.qualifiedName, r.path, r.version, r.resources, r.operations})
	}
	return e.writeResult(result)
}

// describePublishedRestService handles DESCRIBE PUBLISHED REST SERVICE command.
func (e *Executor) describePublishedRestService(name ast.QualifiedName) error {
	services, err := e.reader.ListPublishedRestServices()
	if err != nil {
		return fmt.Errorf("failed to list published REST services: %w", err)
	}

	h, err := e.getHierarchy()
	if err != nil {
		return fmt.Errorf("failed to build hierarchy: %w", err)
	}

	for _, svc := range services {
		modID := h.FindModuleID(svc.ContainerID)
		modName := h.GetModuleName(modID)
		qualifiedName := modName + "." + svc.Name

		if !strings.EqualFold(modName, name.Module) || !strings.EqualFold(svc.Name, name.Name) {
			continue
		}

		// Output as re-executable MDL
		fmt.Fprintf(e.output, "CREATE PUBLISHED REST SERVICE %s (\n", qualifiedName)
		fmt.Fprintf(e.output, "  Path: '%s'", svc.Path)
		if svc.Version != "" {
			fmt.Fprintf(e.output, ",\n  Version: '%s'", svc.Version)
		}
		if svc.ServiceName != "" {
			fmt.Fprintf(e.output, ",\n  ServiceName: '%s'", svc.ServiceName)
		}
		fmt.Fprintln(e.output, "\n)")

		if len(svc.Resources) > 0 {
			fmt.Fprintln(e.output, "{")
			for _, res := range svc.Resources {
				fmt.Fprintf(e.output, "  RESOURCE '%s' {\n", res.Name)
				for _, op := range res.Operations {
					deprecated := ""
					if op.Deprecated {
						deprecated = " DEPRECATED"
					}
					mf := ""
					if op.Microflow != "" {
						mf = fmt.Sprintf(" MICROFLOW %s", op.Microflow)
					}
					summary := ""
					if op.Summary != "" {
						summary = fmt.Sprintf(" -- %s", op.Summary)
					}
					fmt.Fprintf(e.output, "    %s %s%s%s;%s\n",
						op.HTTPMethod, op.Path, mf, deprecated, summary)
				}
				fmt.Fprintln(e.output, "  }")
			}
			fmt.Fprintln(e.output, "};")
		} else {
			fmt.Fprintln(e.output, ";")
		}
		fmt.Fprintln(e.output, "/")

		return nil
	}

	return fmt.Errorf("published REST service not found: %s", name)
}
