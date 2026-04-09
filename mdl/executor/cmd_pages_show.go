// SPDX-License-Identifier: Apache-2.0

package executor

import (
	"fmt"
	"sort"
	"strings"
)

// showPages handles SHOW PAGES command.
func (e *Executor) showPages(moduleName string) error {
	// Get hierarchy for module/folder resolution
	h, err := e.getHierarchy()
	if err != nil {
		return fmt.Errorf("failed to build hierarchy: %w", err)
	}

	// Get all pages
	pages, err := e.reader.ListPages()
	if err != nil {
		return fmt.Errorf("failed to list pages: %w", err)
	}

	// Collect rows
	type row struct {
		qualifiedName string
		module        string
		name          string
		folderPath    string
		title         string
		url           string
		params        int
	}
	var rows []row

	for _, p := range pages {
		modID := h.FindModuleID(p.ContainerID)
		modName := h.GetModuleName(modID)
		if moduleName == "" || modName == moduleName {
			qualifiedName := modName + "." + p.Name
			if p.Excluded {
				qualifiedName += " [EXCLUDED]"
			}
			folderPath := h.BuildFolderPath(p.ContainerID)
			title := ""
			if p.Title != nil {
				// Try to get English title first, then any available translation
				title = p.Title.GetTranslation("en_US")
				if title == "" {
					for _, t := range p.Title.Translations {
						title = t
						break
					}
				}
			}
			url := p.URL

			rows = append(rows, row{qualifiedName, modName, p.Name, folderPath, title, url, len(p.Parameters)})
		}
	}

	// Sort by qualified name
	sort.Slice(rows, func(i, j int) bool {
		return strings.ToLower(rows[i].qualifiedName) < strings.ToLower(rows[j].qualifiedName)
	})

	result := &TableResult{
		Columns: []string{"Qualified Name", "Module", "Name", "Folder", "Title", "URL", "Params"},
		Summary: fmt.Sprintf("(%d pages)", len(rows)),
	}
	for _, r := range rows {
		result.Rows = append(result.Rows, []any{r.qualifiedName, r.module, r.name, r.folderPath, r.title, r.url, r.params})
	}
	return e.writeResult(result)
}
