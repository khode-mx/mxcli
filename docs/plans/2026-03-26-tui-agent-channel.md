# TUI Agent Communication Channel Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a Unix socket-based agent communication channel to the mxcli TUI, allowing Claude (or other agents) to send MDL commands and receive structured results while humans observe execution in real-time.

**Architecture:** A Unix socket listener runs alongside the bubbletea event loop. Agent commands arrive as JSON over the socket, get converted to `tea.Msg` via `tea.Program.Send()`, execute through existing TUI infrastructure (ExecView's `runMxcli` pattern), and results flow back through a response channel. An optional human-confirmation gate holds responses until the user presses `q`.

**Tech Stack:** Go stdlib (`net`, `encoding/json`), bubbletea `tea.Program.Send()`, Unix domain sockets

---

## Task 1: Protocol Types (`tui/agent_protocol.go`)

**Files:**
- Create: `cmd/mxcli/tui/agent_protocol.go`
- Test: `cmd/mxcli/tui/agent_protocol_test.go`

**Step 1: Write failing test for request parsing**

```go
// agent_protocol_test.go
package tui

import (
	"encoding/json"
	"testing"
)

func TestAgentRequestParsing(t *testing.T) {
	raw := `{"id": 1, "action": "exec", "mdl": "CREATE ENTITY M.E (Name: String(100));"}`
	var req AgentRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if req.ID != 1 {
		t.Errorf("id = %d, want 1", req.ID)
	}
	if req.Action != "exec" {
		t.Errorf("action = %q, want exec", req.Action)
	}
	if req.MDL != `CREATE ENTITY M.E (Name: String(100));` {
		t.Errorf("mdl = %q", req.MDL)
	}
}

func TestAgentResponseSerialization(t *testing.T) {
	resp := AgentResponse{ID: 1, OK: true, Result: "Created entity", Mode: "overlay:exec-result"}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded map[string]interface{}
	json.Unmarshal(data, &decoded)
	if decoded["ok"] != true {
		t.Errorf("ok = %v", decoded["ok"])
	}
}

func TestAgentRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     AgentRequest
		wantErr bool
	}{
		{"valid exec", AgentRequest{ID: 1, Action: "exec", MDL: "SHOW ENTITIES"}, false},
		{"valid check", AgentRequest{ID: 2, Action: "check"}, false},
		{"valid state", AgentRequest{ID: 3, Action: "state"}, false},
		{"valid navigate", AgentRequest{ID: 4, Action: "navigate", Target: "entity:M.E"}, false},
		{"missing id", AgentRequest{Action: "exec", MDL: "SHOW"}, true},
		{"unknown action", AgentRequest{ID: 1, Action: "unknown"}, true},
		{"exec without mdl", AgentRequest{ID: 1, Action: "exec"}, true},
		{"navigate without target", AgentRequest{ID: 1, Action: "navigate"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /mnt/data_sdd/gh/mxcli-wt-03 && go test ./cmd/mxcli/tui/ -run TestAgent -v`
Expected: FAIL — types not defined

**Step 3: Implement protocol types**

```go
// agent_protocol.go
package tui

import "fmt"

// AgentRequest is a JSON command from an external agent (e.g. Claude).
type AgentRequest struct {
	ID     int    `json:"id"`
	Action string `json:"action"`           // "exec", "check", "state", "navigate"
	MDL    string `json:"mdl,omitempty"`     // for "exec"
	Target string `json:"target,omitempty"`  // for "navigate" (e.g. "entity:Module.Entity")
}

// Validate checks that the request has all required fields for its action.
func (r AgentRequest) Validate() error {
	if r.ID == 0 {
		return fmt.Errorf("missing request id")
	}
	switch r.Action {
	case "exec":
		if r.MDL == "" {
			return fmt.Errorf("exec action requires mdl field")
		}
	case "check", "state":
		// no extra fields needed
	case "navigate":
		if r.Target == "" {
			return fmt.Errorf("navigate action requires target field")
		}
	default:
		return fmt.Errorf("unknown action: %q", r.Action)
	}
	return nil
}

// AgentResponse is the JSON response sent back to the agent.
type AgentResponse struct {
	ID     int    `json:"id"`
	OK     bool   `json:"ok"`
	Result string `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
	Mode   string `json:"mode,omitempty"` // e.g. "overlay:exec-result"
}
```

**Step 4: Run test to verify it passes**

Run: `cd /mnt/data_sdd/gh/mxcli-wt-03 && go test ./cmd/mxcli/tui/ -run TestAgent -v`
Expected: PASS

**Step 5: Commit**

```bash
git add cmd/mxcli/tui/agent_protocol.go cmd/mxcli/tui/agent_protocol_test.go
git commit -m "feat(tui): add agent protocol types for external communication"
```

---

## Task 2: Tea Messages for Agent Commands (`tui/agent_msgs.go`)

**Files:**
- Create: `cmd/mxcli/tui/agent_msgs.go`

**Step 1: Define tea.Msg types**

```go
// agent_msgs.go
package tui

import tea "github.com/charmbracelet/bubbletea"

// AgentExecMsg requests MDL execution from an external agent.
// The ResponseCh receives the result after TUI displays it and (optionally) the user confirms.
type AgentExecMsg struct {
	RequestID  int
	MDL        string
	ResponseCh chan<- AgentResponse
}

// AgentCheckMsg requests a syntax/reference check.
type AgentCheckMsg struct {
	RequestID  int
	ResponseCh chan<- AgentResponse
}

// AgentStateMsg requests current TUI state (active view, project path, etc.).
type AgentStateMsg struct {
	RequestID  int
	ResponseCh chan<- AgentResponse
}

// AgentNavigateMsg requests navigation to a specific element.
type AgentNavigateMsg struct {
	RequestID  int
	Target     string // e.g. "entity:Module.Entity"
	ResponseCh chan<- AgentResponse
}

// agentExecDoneMsg carries exec result back to App for agent response.
type agentExecDoneMsg struct {
	RequestID  int
	Output     string
	Success    bool
	ResponseCh chan<- AgentResponse
}

// agentConfirmedMsg is sent when user presses q to confirm agent result.
type agentConfirmedMsg struct {
	RequestID  int
	ResponseCh chan<- AgentResponse
	Output     string
	Success    bool
}

// Ensure messages satisfy tea.Msg (they do implicitly).
var (
	_ tea.Msg = AgentExecMsg{}
	_ tea.Msg = AgentCheckMsg{}
	_ tea.Msg = AgentStateMsg{}
	_ tea.Msg = AgentNavigateMsg{}
	_ tea.Msg = agentExecDoneMsg{}
	_ tea.Msg = agentConfirmedMsg{}
)
```

**Step 2: Commit**

```bash
git add cmd/mxcli/tui/agent_msgs.go
git commit -m "feat(tui): add bubbletea message types for agent channel"
```

---

## Task 3: Socket Listener (`tui/agent_listener.go`)

**Files:**
- Create: `cmd/mxcli/tui/agent_listener.go`
- Test: `cmd/mxcli/tui/agent_listener_test.go`

**Step 1: Write failing test**

```go
// agent_listener_test.go
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

func TestAgentListenerAcceptsConnection(t *testing.T) {
	sockPath := filepath.Join(t.TempDir(), "test.sock")

	// Collect messages sent to the program
	var received []tea.Msg
	sender := func(msg tea.Msg) { received = append(received, msg) }

	listener, err := NewAgentListener(sockPath, sender, false)
	if err != nil {
		t.Fatalf("NewAgentListener: %v", err)
	}
	defer listener.Close()

	// Connect and send a request
	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	req := AgentRequest{ID: 1, Action: "state"}
	data, _ := json.Marshal(req)
	data = append(data, '\n')
	conn.Write(data)

	// Wait for message to arrive
	time.Sleep(100 * time.Millisecond)
	if len(received) == 0 {
		t.Fatal("expected at least one message")
	}
	if _, ok := received[0].(AgentStateMsg); !ok {
		t.Errorf("expected AgentStateMsg, got %T", received[0])
	}
}

func TestAgentListenerCleansUpSocket(t *testing.T) {
	sockPath := filepath.Join(t.TempDir(), "test.sock")
	listener, err := NewAgentListener(sockPath, func(tea.Msg) {}, false)
	if err != nil {
		t.Fatalf("NewAgentListener: %v", err)
	}
	listener.Close()
	if _, err := os.Stat(sockPath); !os.IsNotExist(err) {
		t.Error("socket file should be removed after Close")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /mnt/data_sdd/gh/mxcli-wt-03 && go test ./cmd/mxcli/tui/ -run TestAgentListener -v -timeout 10s`
Expected: FAIL

**Step 3: Implement listener**

```go
// agent_listener.go
package tui

import (
	"bufio"
	"encoding/json"
	"net"
	"os"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

// AgentListener accepts agent connections on a Unix socket and converts
// JSON requests into tea.Msg values injected into the bubbletea program.
type AgentListener struct {
	socketPath  string
	listener    net.Listener
	sendMsg     func(tea.Msg) // tea.Program.Send or test stub
	autoProceed bool          // skip human confirmation
	mu          sync.Mutex
	closed      bool
	wg          sync.WaitGroup
}

// NewAgentListener creates and starts a Unix socket listener.
// sendMsg is called to inject messages into the bubbletea event loop
// (typically tea.Program.Send).
func NewAgentListener(socketPath string, sendMsg func(tea.Msg), autoProceed bool) (*AgentListener, error) {
	// Remove stale socket if present
	os.Remove(socketPath)

	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, err
	}

	al := &AgentListener{
		socketPath:  socketPath,
		listener:    ln,
		sendMsg:     sendMsg,
		autoProceed: autoProceed,
	}

	al.wg.Add(1)
	go al.acceptLoop()
	return al, nil
}

// Close stops the listener and removes the socket file.
func (al *AgentListener) Close() {
	al.mu.Lock()
	if al.closed {
		al.mu.Unlock()
		return
	}
	al.closed = true
	al.mu.Unlock()

	al.listener.Close()
	al.wg.Wait()
	os.Remove(al.socketPath)
}

// AutoProceed returns whether human confirmation is skipped.
func (al *AgentListener) AutoProceed() bool {
	return al.autoProceed
}

func (al *AgentListener) acceptLoop() {
	defer al.wg.Done()
	for {
		conn, err := al.listener.Accept()
		if err != nil {
			return // listener closed
		}
		al.wg.Add(1)
		go al.handleConnection(conn)
	}
}

func (al *AgentListener) handleConnection(conn net.Conn) {
	defer al.wg.Done()
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB max line
	encoder := json.NewEncoder(conn)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req AgentRequest
		if err := json.Unmarshal(line, &req); err != nil {
			resp := AgentResponse{OK: false, Error: "invalid json: " + err.Error()}
			encoder.Encode(resp)
			continue
		}

		if err := req.Validate(); err != nil {
			resp := AgentResponse{ID: req.ID, OK: false, Error: err.Error()}
			encoder.Encode(resp)
			continue
		}

		// Create a response channel for this request
		responseCh := make(chan AgentResponse, 1)

		// Convert to tea.Msg and inject
		switch req.Action {
		case "exec":
			al.sendMsg(AgentExecMsg{
				RequestID:  req.ID,
				MDL:        req.MDL,
				ResponseCh: responseCh,
			})
		case "check":
			al.sendMsg(AgentCheckMsg{
				RequestID:  req.ID,
				ResponseCh: responseCh,
			})
		case "state":
			al.sendMsg(AgentStateMsg{
				RequestID:  req.ID,
				ResponseCh: responseCh,
			})
		case "navigate":
			al.sendMsg(AgentNavigateMsg{
				RequestID:  req.ID,
				Target:     req.Target,
				ResponseCh: responseCh,
			})
		}

		// Wait for response from TUI
		resp := <-responseCh
		encoder.Encode(resp)
	}
}
```

**Step 4: Run test to verify it passes**

Run: `cd /mnt/data_sdd/gh/mxcli-wt-03 && go test ./cmd/mxcli/tui/ -run TestAgentListener -v -timeout 10s`
Expected: PASS

**Step 5: Commit**

```bash
git add cmd/mxcli/tui/agent_listener.go cmd/mxcli/tui/agent_listener_test.go
git commit -m "feat(tui): add Unix socket listener for agent communication"
```

---

## Task 4: TUI Update Handler Integration (`tui/app.go` modifications)

**Files:**
- Modify: `cmd/mxcli/tui/app.go` (App struct + Update + NewApp)
- Modify: `cmd/mxcli/tui/overlayview.go` (add agent confirmation)

**Step 1: Add agent fields to App struct**

In `app.go`, add to `App` struct:

```go
agentListener   *AgentListener
agentPending    *agentPendingOp // non-nil when waiting for user confirmation
```

Add new type:

```go
// agentPendingOp tracks an in-flight agent operation awaiting user confirmation.
type agentPendingOp struct {
	RequestID  int
	Output     string
	Success    bool
	ResponseCh chan<- AgentResponse
}
```

**Step 2: Add agent message handling to App.Update**

Add cases before the `tea.KeyMsg` case in `App.Update`:

```go
case AgentExecMsg:
	// Execute MDL using same mechanism as ExecView
	mxcliPath := a.mxcliPath
	projectPath := a.activeTabProjectPath()
	requestID := msg.RequestID
	mdlText := msg.MDL
	responseCh := msg.ResponseCh
	return a, func() tea.Msg {
		tmpFile, err := os.CreateTemp("", "mxcli-agent-*.mdl")
		if err != nil {
			return agentExecDoneMsg{
				RequestID: requestID, Output: err.Error(),
				Success: false, ResponseCh: responseCh,
			}
		}
		tmpPath := tmpFile.Name()
		defer os.Remove(tmpPath)
		tmpFile.WriteString(mdlText)
		tmpFile.Close()

		args := []string{"exec"}
		if projectPath != "" {
			args = append(args, "-p", projectPath)
		}
		args = append(args, tmpPath)
		out, execErr := runMxcli(mxcliPath, args...)
		return agentExecDoneMsg{
			RequestID: requestID, Output: out,
			Success: execErr == nil, ResponseCh: responseCh,
		}
	}

case agentExecDoneMsg:
	content := DetectAndHighlight(msg.Output)
	title := "Agent Exec Result"
	if !msg.Success {
		title = "Agent Exec Error"
	}
	ov := NewOverlayView(title, content, a.width, a.height, OverlayViewOpts{})
	a.views.Push(ov)
	if msg.Success {
		if a.watcher != nil {
			a.watcher.Suppress(2 * time.Second)
		}
	}
	// If auto-proceed, respond immediately; otherwise wait for user 'q'
	if a.agentListener != nil && a.agentListener.AutoProceed() {
		msg.ResponseCh <- AgentResponse{
			ID: msg.RequestID, OK: msg.Success,
			Result: msg.Output, Mode: "overlay:exec-result",
		}
		if msg.Success {
			return a, a.Init()
		}
		return a, nil
	}
	// Store pending op for user confirmation
	a.agentPending = &agentPendingOp{
		RequestID: msg.RequestID, Output: msg.Output,
		Success: msg.Success, ResponseCh: msg.ResponseCh,
	}
	return a, func() tea.Msg {
		if msg.Success {
			return nil // tree refresh handled after confirmation
		}
		return nil
	}

case agentConfirmedMsg:
	msg.ResponseCh <- AgentResponse{
		ID: msg.RequestID, OK: msg.Success,
		Result: msg.Output, Mode: "overlay:exec-result",
	}
	a.agentPending = nil
	if msg.Success {
		return a, a.Init()
	}
	return a, nil

case AgentStateMsg:
	mode := a.views.Active().Mode().String()
	projectPath := a.activeTabProjectPath()
	resp := AgentResponse{
		ID: msg.RequestID, OK: true,
		Result: fmt.Sprintf(`{"mode":"%s","project":"%s"}`, mode, projectPath),
		Mode:   "state",
	}
	msg.ResponseCh <- resp
	return a, nil

case AgentCheckMsg:
	mxcliPath := a.mxcliPath
	projectPath := a.activeTabProjectPath()
	requestID := msg.RequestID
	responseCh := msg.ResponseCh
	return a, func() tea.Msg {
		out, err := runMxcli(mxcliPath, "check", "-p", projectPath)
		success := err == nil
		return agentExecDoneMsg{
			RequestID: requestID, Output: out,
			Success: success, ResponseCh: responseCh,
		}
	}

case AgentNavigateMsg:
	target := msg.Target
	// Parse "entity:Module.Entity" format
	if bv, ok := a.views.Base().(BrowserView); ok {
		// Strip prefix like "entity:", "microflow:", etc.
		qname := target
		if idx := strings.Index(target, ":"); idx >= 0 {
			qname = target[idx+1:]
		}
		cmd := bv.navigateToNode(qname)
		a.views.SetBase(bv)
		if tab := a.activeTabPtr(); tab != nil {
			tab.Miller = bv.miller
			tab.UpdateLabel()
			a.syncTabBar()
		}
		msg.ResponseCh <- AgentResponse{
			ID: msg.RequestID, OK: true,
			Result: fmt.Sprintf("navigated to %s", qname), Mode: "browser",
		}
		return a, cmd
	}
	msg.ResponseCh <- AgentResponse{
		ID: msg.RequestID, OK: false, Error: "not in browser mode",
	}
	return a, nil
```

**Step 3: Handle user confirmation in overlay 'q' key**

In the `tea.KeyMsg` handling for overlay view's `q`/`Esc` (which pops the view), add agent confirmation check. In App.Update's `tea.KeyMsg` section, before `a.views.Active().Update(msg)`:

```go
// Agent confirmation: when overlay is popped while agent op is pending
if (msg.String() == "q" || msg.String() == "esc") && a.agentPending != nil && a.views.Active().Mode() == ModeOverlay {
	pending := a.agentPending
	a.agentPending = nil
	a.views.Pop()
	pending.ResponseCh <- AgentResponse{
		ID: pending.RequestID, OK: pending.Success,
		Result: pending.Output, Mode: "overlay:exec-result",
	}
	if pending.Success {
		return a, a.Init()
	}
	return a, nil
}
```

**Step 4: Commit**

```bash
git add cmd/mxcli/tui/app.go
git commit -m "feat(tui): handle agent messages in App.Update with confirmation flow"
```

---

## Task 5: CLI Flag & Startup Wiring (`cmd_tui.go`)

**Files:**
- Modify: `cmd/mxcli/cmd_tui.go`
- Modify: `cmd/mxcli/tui/app.go` (add StartAgentListener method)

**Step 1: Add StartAgentListener to App**

In `app.go`:

```go
// StartAgentListener begins listening on a Unix socket for agent commands.
// Call after tea.NewProgram is created, like StartWatcher.
func (a *App) StartAgentListener(prog *tea.Program, socketPath string, autoProceed bool) error {
	listener, err := NewAgentListener(socketPath, prog.Send, autoProceed)
	if err != nil {
		return err
	}
	a.agentListener = listener
	Trace("app: agent listener started on %s (autoProceed=%v)", socketPath, autoProceed)
	return nil
}

// CloseAgentListener stops the agent listener if running.
func (a *App) CloseAgentListener() {
	if a.agentListener != nil {
		a.agentListener.Close()
	}
}
```

**Step 2: Add CLI flags and wiring**

In `cmd_tui.go`, add to the `Run` function after `m.StartWatcher(p)`:

```go
agentSocket, _ := cmd.Flags().GetString("agent-socket")
agentAutoProceed, _ := cmd.Flags().GetBool("agent-auto-proceed")
if agentSocket != "" {
	if err := m.StartAgentListener(p, agentSocket, agentAutoProceed); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: agent listener failed: %v\n", err)
	}
	defer m.CloseAgentListener()
}
```

In `init()`:

```go
tuiCmd.Flags().String("agent-socket", "", "Unix socket path for agent communication (e.g. /tmp/mxcli-agent.sock)")
tuiCmd.Flags().Bool("agent-auto-proceed", false, "Skip human confirmation for agent operations")
```

**Step 3: Commit**

```bash
git add cmd/mxcli/cmd_tui.go cmd/mxcli/tui/app.go
git commit -m "feat(tui): add --agent-socket and --agent-auto-proceed CLI flags"
```

---

## Task 6: Integration Test (end-to-end)

**Files:**
- Create: `cmd/mxcli/tui/agent_integration_test.go`

**Step 1: Write integration test**

```go
// agent_integration_test.go
package tui

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAgentExecEndToEnd(t *testing.T) {
	sockPath := filepath.Join(t.TempDir(), "agent.sock")

	// Use auto-proceed mode for testing (no human confirmation)
	var messages []interface{}
	sender := func(msg interface{}) { messages = append(messages, msg) }

	listener, err := NewAgentListener(sockPath, sender, true)
	if err != nil {
		t.Fatalf("NewAgentListener: %v", err)
	}
	defer listener.Close()

	// Connect
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

	// The listener dispatches to the sender; verify message arrives
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

func TestAgentInvalidRequest(t *testing.T) {
	sockPath := filepath.Join(t.TempDir(), "agent.sock")
	listener, err := NewAgentListener(sockPath, func(interface{}) {}, false)
	if err != nil {
		t.Fatalf("NewAgentListener: %v", err)
	}
	defer listener.Close()

	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	// Send invalid JSON
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

	// Send valid JSON but missing required fields
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

// Verify socket is cleaned up on listener close
func TestAgentListenerCleanup(t *testing.T) {
	sockPath := filepath.Join(t.TempDir(), "agent.sock")
	listener, err := NewAgentListener(sockPath, func(interface{}) {}, false)
	if err != nil {
		t.Fatalf("NewAgentListener: %v", err)
	}
	// Socket should exist
	if _, err := os.Stat(sockPath); err != nil {
		t.Errorf("socket should exist: %v", err)
	}
	listener.Close()
	// Socket should be removed
	if _, err := os.Stat(sockPath); !os.IsNotExist(err) {
		t.Error("socket should be removed after close")
	}
}
```

**Step 2: Run tests**

Run: `cd /mnt/data_sdd/gh/mxcli-wt-03 && go test ./cmd/mxcli/tui/ -run TestAgent -v -timeout 30s`
Expected: ALL PASS

**Step 3: Commit**

```bash
git add cmd/mxcli/tui/agent_integration_test.go
git commit -m "test(tui): add integration tests for agent communication channel"
```

---

## Task 7: Build Verification & Final Cleanup

**Step 1: Run full build**

Run: `cd /mnt/data_sdd/gh/mxcli-wt-03 && make build`
Expected: Build succeeds

**Step 2: Run all TUI tests**

Run: `cd /mnt/data_sdd/gh/mxcli-wt-03 && go test ./cmd/mxcli/tui/ -v -timeout 60s`
Expected: All tests pass

**Step 3: Manual smoke test**

```bash
# Terminal 1: start TUI with agent socket
./bin/mxcli tui -p /path/to/app.mpr --agent-socket /tmp/mxcli-agent.sock

# Terminal 2: send a command
echo '{"id":1,"action":"state"}' | socat - UNIX-CONNECT:/tmp/mxcli-agent.sock
```

**Step 4: Final commit if any cleanup needed**
