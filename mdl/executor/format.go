// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// OutputFormat controls how command results are rendered.
type OutputFormat string

const (
	// FormatTable renders results as a pipe-delimited markdown table (default).
	FormatTable OutputFormat = "table"
	// FormatJSON renders results as a JSON array of objects.
	FormatJSON OutputFormat = "json"
)

// TableResult holds structured tabular data that can be rendered in multiple formats.
type TableResult struct {
	Columns []string // column headers
	Rows    [][]any  // row data (one slice per row, matching Columns order)
	Summary string   // optional summary line, e.g. "(42 entities)"
}

// writeResult renders a TableResult to e.output in the current format.
func (e *Executor) writeResult(r *TableResult) error {
	if e.format == FormatJSON {
		return e.writeResultJSON(r)
	}
	e.writeResultTable(r)
	return nil
}

// writeResultTable renders a TableResult as a pipe-delimited markdown table.
func (e *Executor) writeResultTable(r *TableResult) {
	if len(r.Columns) == 0 {
		return
	}

	// Calculate column widths from headers and data.
	widths := make([]int, len(r.Columns))
	for i, col := range r.Columns {
		widths[i] = len(col)
	}
	for _, row := range r.Rows {
		for i, val := range row {
			if i >= len(widths) {
				break
			}
			s := formatCellValue(val)
			if len(s) > widths[i] {
				widths[i] = len(s)
			}
		}
	}

	// Print header.
	fmt.Fprint(e.output, "|")
	for i, col := range r.Columns {
		fmt.Fprintf(e.output, " %-*s |", widths[i], col)
	}
	fmt.Fprintln(e.output)

	// Print separator.
	fmt.Fprint(e.output, "|")
	for _, w := range widths {
		fmt.Fprintf(e.output, "-%s-|", strings.Repeat("-", w))
	}
	fmt.Fprintln(e.output)

	// Print rows.
	for _, row := range r.Rows {
		fmt.Fprint(e.output, "|")
		for i := range r.Columns {
			var s string
			if i < len(row) {
				s = formatCellValue(row[i])
			}
			fmt.Fprintf(e.output, " %-*s |", widths[i], s)
		}
		fmt.Fprintln(e.output)
	}

	// Print summary.
	if r.Summary != "" {
		fmt.Fprintf(e.output, "\n%s\n", r.Summary)
	}
}

// writeResultJSON renders a TableResult as a JSON array of objects.
func (e *Executor) writeResultJSON(r *TableResult) error {
	objects := make([]map[string]any, 0, len(r.Rows))
	for _, row := range r.Rows {
		obj := make(map[string]any, len(r.Columns))
		for i, col := range r.Columns {
			if i < len(row) {
				obj[col] = row[i]
			}
		}
		objects = append(objects, obj)
	}

	enc := json.NewEncoder(e.output)
	enc.SetIndent("", "  ")
	return enc.Encode(objects)
}

// writeDescribeJSON wraps a describe handler's output in a JSON envelope.
// In table/text mode it calls fn directly. In JSON mode it captures fn's output
// and wraps it as {"name": ..., "type": ..., "mdl": ...}.
func (e *Executor) writeDescribeJSON(name, objectType string, fn func() error) error {
	if e.format != FormatJSON {
		return fn()
	}

	// Capture the text output from fn.
	var buf bytes.Buffer
	origOutput := e.output
	origGuard := e.guard
	e.output = &buf
	e.guard = nil // disable line guard for capture
	err := fn()
	e.output = origOutput
	e.guard = origGuard
	if err != nil {
		return err
	}

	result := map[string]any{
		"name": name,
		"type": objectType,
		"mdl":  buf.String(),
	}
	enc := json.NewEncoder(e.output)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

// formatCellValue formats a value for table cell display.
func formatCellValue(val any) string {
	if val == nil {
		return ""
	}
	return fmt.Sprintf("%v", val)
}
