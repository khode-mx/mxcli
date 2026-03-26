package tui

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestAgentExecEndToEnd(t *testing.T) {
	sockPath := filepath.Join(t.TempDir(), "agent.sock")

	var messages []tea.Msg
	sender := func(msg tea.Msg) { messages = append(messages, msg) }

	listener, err := NewAgentListener(sockPath, sender, true)
	if err != nil {
		t.Fatalf("NewAgentListener: %v", err)
	}
	defer listener.Close()

	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	// Send exec request
	req := AgentRequest{ID: 42, Action: "exec", MDL: "SHOW ENTITIES"}
	data, _ := json.Marshal(req)
	data = append(data, '\n')
	if _, err := conn.Write(data); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Wait for message dispatch
	time.Sleep(200 * time.Millisecond)
	if len(messages) == 0 {
		t.Fatal("no messages received")
	}

	execMsg, ok := messages[0].(AgentExecMsg)
	if !ok {
		t.Fatalf("expected AgentExecMsg, got %T", messages[0])
	}
	if execMsg.MDL != "SHOW ENTITIES" {
		t.Errorf("mdl = %q, want SHOW ENTITIES", execMsg.MDL)
	}
	if execMsg.RequestID != 42 {
		t.Errorf("requestID = %d, want 42", execMsg.RequestID)
	}

	// Simulate TUI response
	execMsg.ResponseCh <- AgentResponse{ID: 42, OK: true, Result: "done"}

	// Read response from socket
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	var resp AgentResponse
	if err := json.Unmarshal(buf[:n], &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.ID != 42 || !resp.OK {
		t.Errorf("resp = %+v", resp)
	}
}

func TestAgentInvalidRequests(t *testing.T) {
	sockPath := filepath.Join(t.TempDir(), "agent.sock")
	listener, err := NewAgentListener(sockPath, func(tea.Msg) {}, false)
	if err != nil {
		t.Fatalf("NewAgentListener: %v", err)
	}
	defer listener.Close()

	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	// Invalid JSON
	conn.Write([]byte("not json\n"))
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var resp AgentResponse
	json.Unmarshal(buf[:n], &resp)
	if resp.OK {
		t.Error("expected ok=false for invalid json")
	}

	// Missing required fields
	req := AgentRequest{Action: "exec"} // missing ID and MDL
	data, _ := json.Marshal(req)
	data = append(data, '\n')
	conn.Write(data)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err = conn.Read(buf)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	json.Unmarshal(buf[:n], &resp)
	if resp.OK {
		t.Error("expected ok=false for missing fields")
	}
}

func TestAgentSocketCleanup(t *testing.T) {
	sockPath := filepath.Join(t.TempDir(), "agent.sock")
	listener, err := NewAgentListener(sockPath, func(tea.Msg) {}, false)
	if err != nil {
		t.Fatalf("NewAgentListener: %v", err)
	}
	if _, err := os.Stat(sockPath); err != nil {
		t.Errorf("socket should exist: %v", err)
	}
	listener.Close()
	if _, err := os.Stat(sockPath); !os.IsNotExist(err) {
		t.Error("socket should be removed after close")
	}
}

func TestAgentMultipleActions(t *testing.T) {
	sockPath := filepath.Join(t.TempDir(), "agent.sock")

	var messages []tea.Msg
	sender := func(msg tea.Msg) { messages = append(messages, msg) }

	listener, err := NewAgentListener(sockPath, sender, true)
	if err != nil {
		t.Fatalf("NewAgentListener: %v", err)
	}
	defer listener.Close()

	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	// Test state action
	req := AgentRequest{ID: 1, Action: "state"}
	data, _ := json.Marshal(req)
	data = append(data, '\n')
	conn.Write(data)
	time.Sleep(100 * time.Millisecond)
	if len(messages) == 0 {
		t.Fatal("no state message received")
	}
	stateMsg, ok := messages[0].(AgentStateMsg)
	if !ok {
		t.Fatalf("expected AgentStateMsg, got %T", messages[0])
	}
	stateMsg.ResponseCh <- AgentResponse{ID: 1, OK: true, Result: `{"mode":"Browse"}`}

	// Read state response
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 4096)
	n, _ := conn.Read(buf)
	var resp AgentResponse
	json.Unmarshal(buf[:n], &resp)
	if !resp.OK || resp.ID != 1 {
		t.Errorf("state resp = %+v", resp)
	}

	// Test navigate action
	messages = nil
	req = AgentRequest{ID: 2, Action: "navigate", Target: "entity:MyModule.Customer"}
	data, _ = json.Marshal(req)
	data = append(data, '\n')
	conn.Write(data)
	time.Sleep(100 * time.Millisecond)
	if len(messages) == 0 {
		t.Fatal("no navigate message received")
	}
	navMsg, ok := messages[0].(AgentNavigateMsg)
	if !ok {
		t.Fatalf("expected AgentNavigateMsg, got %T", messages[0])
	}
	if navMsg.Target != "entity:MyModule.Customer" {
		t.Errorf("target = %q", navMsg.Target)
	}
	navMsg.ResponseCh <- AgentResponse{ID: 2, OK: true}

	// Read navigate response
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, _ = conn.Read(buf)
	json.Unmarshal(buf[:n], &resp)
	if !resp.OK || resp.ID != 2 {
		t.Errorf("navigate resp = %+v", resp)
	}
}
