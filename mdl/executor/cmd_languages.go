// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"fmt"
)

// showLanguages lists all languages found in the project's translatable strings.
// Requires REFRESH CATALOG FULL to populate the strings table.
func (e *Executor) showLanguages() error {
	if e.catalog == nil {
		return fmt.Errorf("no catalog available — run REFRESH CATALOG FULL first")
	}

	result, err := e.catalog.Query(`
		SELECT Language, COUNT(*) as StringCount
		FROM strings
		WHERE Language != ''
		GROUP BY Language
		ORDER BY StringCount DESC
	`)
	if err != nil {
		return fmt.Errorf("failed to query languages: %w", err)
	}

	if len(result.Rows) == 0 {
		fmt.Fprintln(e.output, "No translatable strings found. Run REFRESH CATALOG FULL to populate the strings table.")
		return nil
	}

	tr := &TableResult{
		Columns: []string{"Language", "Strings"},
		Summary: fmt.Sprintf("(%d languages)", len(result.Rows)),
	}
	for _, row := range result.Rows {
		lang := ""
		count := ""
		if len(row) > 0 {
			lang = fmt.Sprintf("%v", row[0])
		}
		if len(row) > 1 {
			count = fmt.Sprintf("%v", row[1])
		}
		tr.Rows = append(tr.Rows, []any{lang, count})
	}
	return e.writeResult(tr)
}
