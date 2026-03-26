---
name: tui-agent-channel
description: Use when Claude needs to send MDL commands to a running mxcli TUI, execute operations with human oversight, or automate TUI interactions via the agent socket. Also use when setting up Claude-TUI integration for supervised Mendix project modifications.
---

# TUI Agent Channel

Send MDL commands to a running mxcli TUI over a Unix socket. The human sees every operation in real-time and confirms before Claude proceeds.

## Setup

### 1. Start TUI with agent socket

```bash
# Human confirmation mode (default) — user presses q to confirm each result
mxcli tui -p app.mpr --agent-socket /tmp/mxcli-agent.sock

# Auto-proceed mode — results display but Claude continues immediately
mxcli tui -p app.mpr --agent-socket /tmp/mxcli-agent.sock --agent-auto-proceed
```

### 2. Send commands via socat

```bash
# For async commands (exec, check), use printf with \n and socat -t timeout
printf '{"id":1,"action":"exec","mdl":"SHOW ENTITIES"}\n' | socat -t 30 - UNIX-CONNECT:/tmp/mxcli-agent.sock

# For instant commands (state, navigate), echo works fine
echo '{"id":2,"action":"state"}' | socat - UNIX-CONNECT:/tmp/mxcli-agent.sock
```

## Protocol

JSON-line protocol over Unix socket. One request per line, one JSON response per line.

### Request

```json
{"id": 1, "action": "exec", "mdl": "CREATE ENTITY Mod.E (Name: String(100));"}
```

| Field | Required | Description |
|-------|----------|-------------|
| `id` | always | Unique request ID (nonzero integer) |
| `action` | always | `exec`, `check`, `state`, `navigate` |
| `mdl` | exec | MDL statement(s) to execute |
| `target` | navigate | Element reference, e.g. `entity:Module.Entity` |

### Response

```json
{"id": 1, "ok": true, "result": "Created entity: Mod.E", "mode": "overlay:exec-result"}
```

| Field | Description |
|-------|-------------|
| `id` | Echoed request ID |
| `ok` | `true` if operation succeeded |
| `result` | Output text (same as TUI overlay content) |
| `error` | Error message (when `ok` is `false`) |
| `mode` | TUI state after operation |

## Actions

| Action | Purpose | Example |
|--------|---------|---------|
| `exec` | Execute MDL | `{"id":1,"action":"exec","mdl":"CREATE ENTITY M.E (X: String(100));"}` |
| `check` | Run `mx check` | `{"id":2,"action":"check"}` |
| `state` | Query TUI state | `{"id":3,"action":"state"}` |
| `navigate` | Navigate to element | `{"id":4,"action":"navigate","target":"entity:M.E"}` |

## Human Confirmation Flow

Without `--agent-auto-proceed`:

1. Claude sends command via socket
2. TUI executes and shows result in overlay (human sees it)
3. **Socket blocks** until human presses `q` or `Esc` to dismiss overlay
4. Response is sent back to Claude

This ensures the human reviews every operation before Claude continues.

## Typical Claude Workflow

```bash
# 1. Check current state
echo '{"id":1,"action":"state"}' | socat - UNIX-CONNECT:/tmp/mxcli-agent.sock

# 2. Execute MDL
echo '{"id":2,"action":"exec","mdl":"CREATE ENTITY MyModule.Customer (Name: String(200), Email: String(200));"}' | socat - UNIX-CONNECT:/tmp/mxcli-agent.sock

# 3. Verify with check
echo '{"id":3,"action":"check"}' | socat - UNIX-CONNECT:/tmp/mxcli-agent.sock

# 4. Navigate to see result
echo '{"id":4,"action":"navigate","target":"entity:MyModule.Customer"}' | socat - UNIX-CONNECT:/tmp/mxcli-agent.sock
```

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| Missing newline after JSON | Append `\n` — protocol is line-delimited |
| `id: 0` | ID must be nonzero |
| `exec` without `mdl` field | Always include MDL text for exec action |
| Socket not found | Ensure TUI is running with `--agent-socket` |
| Response never arrives | In confirmation mode, human must press `q` in TUI |
| `echo` pipe closes before async response | Use `printf '...\n' \| socat -t 30` for exec/check (keeps connection open) |
| Overlay stays open after auto-proceed | By design — human can still review; press `q`/`Esc` to dismiss |

## Architecture Note

bubbletea copies the model at `tea.NewProgram()` time. Fields set on the original variable after that are **not visible in `Update()`**. The `agentAutoProceed` flag is set via `SetAgentAutoProceed()` before `NewProgram`. The `agentListener` (set after) is only used for lifecycle management in `cmd_tui.go`, never accessed from `Update()`.

## Key Files

- `cmd/mxcli/tui/agent_protocol.go` — Request/Response types
- `cmd/mxcli/tui/agent_msgs.go` — tea.Msg types
- `cmd/mxcli/tui/agent_listener.go` — Unix socket server
- `cmd/mxcli/tui/app.go` — Update handler (search `AgentExecMsg`)
- `cmd/mxcli/cmd_tui.go` — CLI flags
