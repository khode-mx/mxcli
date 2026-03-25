package tui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// fakeKeyMsg creates a tea.KeyMsg for the given key string.
func fakeKeyMsg(key string) tea.KeyMsg {
	switch key {
	case "ctrl+e":
		return tea.KeyMsg{Type: tea.KeyCtrlE}
	case "ctrl+o":
		return tea.KeyMsg{Type: tea.KeyCtrlO}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	}
}

func TestExecView_Mode(t *testing.T) {
	ev := NewExecView("mxcli", "/tmp/test.mpr", 80, 24)
	if ev.Mode() != ModeExec {
		t.Errorf("expected ModeExec, got %v", ev.Mode())
	}
}

func TestExecView_StatusInfo(t *testing.T) {
	ev := NewExecView("mxcli", "/tmp/test.mpr", 80, 24)
	info := ev.StatusInfo()
	if info.Mode != "Exec" {
		t.Errorf("expected mode 'Exec', got %q", info.Mode)
	}
}

func TestExecView_StatusInfo_Picking(t *testing.T) {
	ev := NewExecView("mxcli", "/tmp/test.mpr", 80, 24)
	ev.picking = true
	info := ev.StatusInfo()
	if len(info.Breadcrumb) != 2 || info.Breadcrumb[1] != "Open File" {
		t.Errorf("expected breadcrumb [Execute MDL, Open File], got %v", info.Breadcrumb)
	}
}

func TestRefreshMDLCandidates(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files and directories
	os.WriteFile(filepath.Join(tmpDir, "script1.mdl"), []byte("SHOW ENTITIES"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "script2.mdl"), []byte("SHOW MODULES"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("not mdl"), 0644)
	os.WriteFile(filepath.Join(tmpDir, ".hidden.mdl"), []byte("hidden"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)

	ev := NewExecView("mxcli", "/tmp/test.mpr", 80, 24)
	ev.pathInput.SetValue(tmpDir + string(os.PathSeparator))
	ev.refreshMDLCandidates()

	// Should include 2 .mdl files + 1 directory, exclude .txt and hidden files
	if len(ev.pathCandidates) != 3 {
		names := make([]string, len(ev.pathCandidates))
		for i, c := range ev.pathCandidates {
			names[i] = c.name
		}
		t.Fatalf("expected 3 candidates (2 mdl + 1 dir), got %d: %v", len(ev.pathCandidates), names)
	}

	mdlCount := 0
	dirCount := 0
	for _, c := range ev.pathCandidates {
		if c.isMDL {
			mdlCount++
		}
		if c.isDir {
			dirCount++
		}
	}
	if mdlCount != 2 {
		t.Errorf("expected 2 MDL files, got %d", mdlCount)
	}
	if dirCount != 1 {
		t.Errorf("expected 1 directory, got %d", dirCount)
	}
}

func TestRefreshMDLCandidates_WithPrefix(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "alpha.mdl"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmpDir, "beta.mdl"), []byte(""), 0644)

	ev := NewExecView("mxcli", "/tmp/test.mpr", 80, 24)
	ev.pathInput.SetValue(filepath.Join(tmpDir, "al"))
	ev.refreshMDLCandidates()

	if len(ev.pathCandidates) != 1 {
		t.Fatalf("expected 1 candidate matching 'al' prefix, got %d", len(ev.pathCandidates))
	}
	if ev.pathCandidates[0].name != "alpha.mdl" {
		t.Errorf("expected alpha.mdl, got %s", ev.pathCandidates[0].name)
	}
}

func TestRefreshMDLCandidates_EmptyPath(t *testing.T) {
	ev := NewExecView("mxcli", "/tmp/test.mpr", 80, 24)
	ev.pathInput.SetValue("")
	ev.refreshMDLCandidates()

	if ev.pathCandidates != nil {
		t.Errorf("expected nil candidates for empty path, got %d", len(ev.pathCandidates))
	}
}

func TestPickerCursorDown_Wraps(t *testing.T) {
	ev := NewExecView("mxcli", "/tmp/test.mpr", 80, 24)
	ev.pathCandidates = []mdlCandidate{
		{name: "a.mdl", isMDL: true},
		{name: "b.mdl", isMDL: true},
		{name: "c.mdl", isMDL: true},
	}
	ev.pathCursor = 2

	ev.pickerCursorDown()
	if ev.pathCursor != 0 {
		t.Errorf("expected cursor to wrap to 0, got %d", ev.pathCursor)
	}
}

func TestPickerCursorUp_Wraps(t *testing.T) {
	ev := NewExecView("mxcli", "/tmp/test.mpr", 80, 24)
	ev.pathCandidates = []mdlCandidate{
		{name: "a.mdl", isMDL: true},
		{name: "b.mdl", isMDL: true},
		{name: "c.mdl", isMDL: true},
	}
	ev.pathCursor = 0

	ev.pickerCursorUp()
	if ev.pathCursor != 2 {
		t.Errorf("expected cursor to wrap to 2, got %d", ev.pathCursor)
	}
}

func TestPickerCursorDown_ScrollsViewport(t *testing.T) {
	ev := NewExecView("mxcli", "/tmp/test.mpr", 80, 24)
	// Create more candidates than execPickerMaxVisible
	candidates := make([]mdlCandidate, 15)
	for i := range candidates {
		candidates[i] = mdlCandidate{name: "f.mdl", isMDL: true}
	}
	ev.pathCandidates = candidates
	ev.pathCursor = execPickerMaxVisible - 1
	ev.pathScroll = 0

	ev.pickerCursorDown()
	if ev.pathScroll != 1 {
		t.Errorf("expected scroll to advance to 1, got %d", ev.pathScroll)
	}
}

func TestPickerCursorUp_ScrollsViewport(t *testing.T) {
	ev := NewExecView("mxcli", "/tmp/test.mpr", 80, 24)
	candidates := make([]mdlCandidate, 15)
	for i := range candidates {
		candidates[i] = mdlCandidate{name: "f.mdl", isMDL: true}
	}
	ev.pathCandidates = candidates
	ev.pathCursor = 5
	ev.pathScroll = 5

	ev.pickerCursorUp()
	if ev.pathCursor != 4 {
		t.Errorf("expected cursor at 4, got %d", ev.pathCursor)
	}
	if ev.pathScroll != 4 {
		t.Errorf("expected scroll at 4, got %d", ev.pathScroll)
	}
}

func TestPickerCursorDown_EmptyCandidates(t *testing.T) {
	ev := NewExecView("mxcli", "/tmp/test.mpr", 80, 24)
	ev.pathCandidates = nil
	ev.pickerCursorDown() // should not panic
}

func TestUpdateEditor_CtrlE_EmptyTextarea(t *testing.T) {
	ev := NewExecView("mxcli", "/tmp/test.mpr", 80, 24)
	// textarea starts empty
	updated, _ := ev.updateEditor(fakeKeyMsg("ctrl+e"))
	updatedEV := updated.(ExecView)
	if updatedEV.flash != "Nothing to execute" {
		t.Errorf("expected flash 'Nothing to execute', got %q", updatedEV.flash)
	}
}

func TestUpdateEditor_Esc_ClearsFlash(t *testing.T) {
	ev := NewExecView("mxcli", "/tmp/test.mpr", 80, 24)
	ev.flash = "some message"
	updated, _ := ev.updateEditor(fakeKeyMsg("esc"))
	updatedEV := updated.(ExecView)
	if updatedEV.flash != "" {
		t.Errorf("expected flash cleared, got %q", updatedEV.flash)
	}
}
