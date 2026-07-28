package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/awanmh/Nova/internal/security"
)

type RunCmdArgs struct {
	Command string `json:"command"`
}

type RunCmdTool struct {
	rootDir string
}

func NewRunCmdTool(rootDir string) *RunCmdTool {
	return &RunCmdTool{rootDir: rootDir}
}

func (t *RunCmdTool) Name() string {
	return "run_command"
}

func (t *RunCmdTool) Description() string {
	return "Execute a terminal shell command in the workspace root directory."
}

func (t *RunCmdTool) Schema() Schema {
	return Schema{
		Type: "object",
		Properties: map[string]Property{
			"command": {
				Type:        "string",
				Description: "The exact shell command line string to run",
			},
		},
		Required: []string{"command"},
	}
}

func (t *RunCmdTool) RiskLevel() RiskLevel {
	return RiskHighImpact
}

func (t *RunCmdTool) Execute(ctx context.Context, argsJSON string) (string, error) {
	var args RunCmdArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("invalid arguments for run_command: %w", err)
	}

	cmdStr := strings.TrimSpace(args.Command)
	if cmdStr == "" {
		return "", fmt.Errorf("command cannot be empty")
	}

	execCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(execCtx, "cmd.exe", "/c", cmdStr)
	} else {
		cmd = exec.CommandContext(execCtx, "sh", "-c", cmdStr)
	}
	cmd.Dir = t.rootDir

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	output := outBuf.String()
	if errOutput := errBuf.String(); errOutput != "" {
		output += "\nSTDERR:\n" + errOutput
	}

	// Always scrub/redact output before returning
	output = security.Redact(output)

	if err != nil {
		if execCtx.Err() == context.DeadlineExceeded {
			return output, fmt.Errorf("command execution timed out after 30 seconds: %s", output)
		}
		return output, fmt.Errorf("command failed (%v):\n%s", err, output)
	}

	return strings.TrimSpace(output), nil
}
