package tools

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "runtime"
    "strconv"
    "strings"
    "syscall"

    "github.com/charmbracelet/crush/internal/permission"
)

type KillParams struct {
    PID       int    `json:"pid"`
    Signal    string `json:"signal,omitempty"`
    KillGroup bool   `json:"kill_group,omitempty"`
}

type killTool struct {
    permissions permission.Service
    workingDir  string
}

const (
    KillToolName    = "kill"
    killDescription = `Terminate a running process by PID. On Unix-like systems you can optionally terminate the entire process group.

Parameters:
- pid: the target process ID (required)
- signal: optional signal name (e.g., SIGTERM, SIGKILL). Defaults to SIGTERM on Unix. Ignored on Windows.
- kill_group: when true on Unix, sends the signal to the whole process group.

Use with care. Prefer SIGTERM before SIGKILL.`
)

func NewKillTool(permissions permission.Service, workingDir string) BaseTool {
    return &killTool{permissions: permissions, workingDir: workingDir}
}

func (k *killTool) Name() string { return KillToolName }

func (k *killTool) Info() ToolInfo {
    return ToolInfo{
        Name:        KillToolName,
        Description: killDescription,
        Parameters: map[string]any{
            "pid": map[string]any{
                "type":        "number",
                "description": "Process ID to terminate",
            },
            "signal": map[string]any{
                "type":        "string",
                "description": "Signal to send (Unix only, defaults to SIGTERM)",
            },
            "kill_group": map[string]any{
                "type":        "boolean",
                "description": "When true, kill the entire process group (Unix only)",
            },
        },
        Required: []string{"pid"},
    }
}

func (k *killTool) Run(ctx context.Context, call ToolCall) (ToolResponse, error) {
    var params KillParams
    if err := json.Unmarshal([]byte(call.Input), &params); err != nil {
        return NewTextErrorResponse("invalid parameters"), nil
    }
    if params.PID <= 0 {
        return NewTextErrorResponse("invalid pid"), nil
    }

    sessionID, messageID := GetContextValues(ctx)
    if sessionID == "" || messageID == "" {
        return ToolResponse{}, fmt.Errorf("session ID and message ID are required for killing a process")
    }

    // Permission gate
    desc := fmt.Sprintf("Terminate process PID=%d", params.PID)
    if params.KillGroup && runtime.GOOS != "windows" {
        desc += " (process group)"
    }
    p := k.permissions.Request(permission.CreatePermissionRequest{
        SessionID:   sessionID,
        ToolCallID:  call.ID,
        Path:        k.workingDir,
        ToolName:    KillToolName,
        Action:      "execute",
        Description: desc,
        Params:      map[string]any{"pid": params.PID, "signal": params.Signal, "kill_group": params.KillGroup},
    })
    if !p {
        return ToolResponse{}, permission.ErrorPermissionDenied
    }

    if runtime.GOOS == "windows" {
        // Best-effort kill on Windows
        proc, err := os.FindProcess(params.PID)
        if err != nil {
            return NewTextErrorResponse(err.Error()), nil
        }
        if err := proc.Kill(); err != nil {
            return NewTextErrorResponse(fmt.Sprintf("failed to kill process: %v", err)), nil
        }
        return NewTextResponse(fmt.Sprintf("Terminated process PID %d", params.PID)), nil
    }

    // Unix-like: map signal
    sig := syscall.SIGTERM
    if params.Signal != "" {
        if s, err := parseSignal(params.Signal); err == nil {
            sig = s
        }
    }

    target := params.PID
    if params.KillGroup {
        // negative pid targets the process group
        target = -params.PID
    }
    if err := syscall.Kill(target, sig); err != nil {
        return NewTextErrorResponse(fmt.Sprintf("failed to send %s to %s%d: %v", sig.String(), groupPrefix(params.KillGroup), params.PID, err)), nil
    }

    scope := "process"
    if params.KillGroup {
        scope = "process group"
    }
    msg := fmt.Sprintf("Sent %s to %s %d", sig.String(), scope, params.PID)
    return NewTextResponse(msg), nil
}

// Shared helpers
func parseSignal(s string) (syscall.Signal, error) {
    us := strings.ToUpper(strings.TrimSpace(s))
    // Allow numeric
    if n, err := strconv.Atoi(us); err == nil {
        return syscall.Signal(n), nil
    }
    switch us {
    case "SIGTERM", "TERM":
        return syscall.SIGTERM, nil
    case "SIGKILL", "KILL":
        return syscall.SIGKILL, nil
    case "SIGINT", "INT":
        return syscall.SIGINT, nil
    case "SIGHUP", "HUP":
        return syscall.SIGHUP, nil
    default:
        return syscall.SIGTERM, nil
    }
}

func groupPrefix(group bool) string {
    if group { return "PGID " }
    return "PID "
}
