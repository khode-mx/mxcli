package tui

import "strings"

// pickerMatch pairs a PickerItem with its fuzzy match score.
type pickerMatch struct {
	item  PickerItem
	score int
}

// fuzzyScore checks if query chars appear in order within target (fzf-style).
// Returns (matched, score). Higher score = better match.
func fuzzyScore(target, query string) (bool, int) {
	tLower := strings.ToLower(target)
	qLower := strings.ToLower(query)
	if len(qLower) == 0 {
		return true, 0
	}
	if len(qLower) > len(tLower) {
		return false, 0
	}
	score := 0
	qi := 0
	prevMatched := false
	for ti := range len(tLower) {
		if qi >= len(qLower) {
			break
		}
		if tLower[ti] == qLower[qi] {
			qi++
			if ti == 0 {
				score += 7
			} else if target[ti-1] == '.' || target[ti-1] == '_' {
				score += 5
			} else if target[ti] >= 'A' && target[ti] <= 'Z' {
				score += 5
			}
			if prevMatched {
				score += 3
			}
			prevMatched = true
		} else {
			prevMatched = false
		}
	}
	if qi < len(qLower) {
		return false, 0
	}
	score += max(0, 50-len(tLower))
	return true, score
}
