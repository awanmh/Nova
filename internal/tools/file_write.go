package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/awanmh/Nova/internal/security"
)

type WriteFileArgs struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type WriteFileTool struct {
	rootDir string
}

func NewWriteFileTool(rootDir string) *WriteFileTool {
	return &WriteFileTool{rootDir: rootDir}
}

func (t *WriteFileTool) Name() string {
	return "write_file"
}

func (t *WriteFileTool) Description() string {
	return "Create a new file or overwrite an existing file within the workspace."
}

func (t *WriteFileTool) Schema() Schema {
	return Schema{
		Type: "object",
		Properties: map[string]Property{
			"path": {
				Type:        "string",
				Description: "Relative path to the target file inside the workspace",
			},
			"content": {
				Type:        "string",
				Description: "The full content to write to the file",
			},
		},
		Required: []string{"path", "content"},
	}
}

func (t *WriteFileTool) RiskLevel() RiskLevel {
	return RiskLowModify
}

func (t *WriteFileTool) Execute(ctx context.Context, argsJSON string) (string, error) {
	var args WriteFileArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("invalid arguments for write_file: %w", err)
	}

	if security.IsSensitiveFile(args.Path) {
		return "", errors.New("access denied: cannot write to sensitive or secret file")
	}

	absPath, err := security.ValidateWorkspacePath(t.rootDir, args.Path)
	if err != nil {
		return "", err
	}

	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory '%s': %w", dir, err)
	}

	if err := os.WriteFile(absPath, []byte(args.Content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file '%s': %w", args.Path, err)
	}

	return fmt.Sprintf("Successfully wrote %d bytes to %s", len(args.Content), args.Path), nil
}
