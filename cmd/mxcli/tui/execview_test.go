package tui

import "testing"

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
