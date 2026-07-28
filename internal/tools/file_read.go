package tools

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/awanmh/Nova/internal/security"
)

type ReadFileArgs struct {
	Path      string `json:"path"`
	StartLine int    `json:"start_line,omitempty"`
	EndLine   int    `json:"end_line,omitempty"`
}

type ReadFileTool struct {
	rootDir string
}

func NewReadFileTool(rootDir string) *ReadFileTool {
	return &ReadFileTool{rootDir: rootDir}
}

func (t *ReadFileTool) Name() string {
	return "read_file"
}

func (t *ReadFileTool) Description() string {
	return "Read the contents of a file within the workspace with optional start and end line numbers (1-indexed)."
}

func (t *ReadFileTool) Schema() Schema {
	return Schema{
		Type: "object",
		Properties: map[string]Property{
			"path": {
				Type:        "string",
				Description: "Relative path to the file inside the workspace",
			},
			"start_line": {
				Type:        "integer",
				Description: "Optional starting line number (1-indexed, inclusive)",
			},
			"end_line": {
				Type:        "integer",
				Description: "Optional ending line number (1-indexed, inclusive)",
			},
		},
		Required: []string{"path"},
	}
}

func (t *ReadFileTool) RiskLevel() RiskLevel {
	return RiskReadOnly
}

func (t *ReadFileTool) Execute(ctx context.Context, argsJSON string) (string, error) {
	var args ReadFileArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("invalid arguments for read_file: %w", err)
	}

	if security.IsSensitiveFile(args.Path) {
		return "", errors.New("access denied: cannot read sensitive or secret file")
	}

	absPath, err := security.ValidateWorkspacePath(t.rootDir, args.Path)
	if err != nil {
		return "", err
	}

	file, err := os.Open(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to open file '%s': %w", args.Path, err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		if args.StartLine > 0 && lineNum < args.StartLine {
			continue
		}
		if args.EndLine > 0 && lineNum > args.EndLine {
			break
		}
		lines = append(lines, fmt.Sprintf("%6d | %s", lineNum, scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return strings.Join(lines, "\n"), nil
}
