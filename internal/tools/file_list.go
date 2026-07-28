package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/awanmh/Nova/internal/security"
)

type ListDirArgs struct {
	Path string `json:"path"`
}

type ListDirTool struct {
	rootDir string
}

func NewListDirTool(rootDir string) *ListDirTool {
	return &ListDirTool{rootDir: rootDir}
}

func (t *ListDirTool) Name() string {
	return "list_dir"
}

func (t *ListDirTool) Description() string {
	return "List all files and subdirectories in a directory within the workspace."
}

func (t *ListDirTool) Schema() Schema {
	return Schema{
		Type: "object",
		Properties: map[string]Property{
			"path": {
				Type:        "string",
				Description: "Relative directory path (use '.' for workspace root)",
			},
		},
		Required: []string{"path"},
	}
}

func (t *ListDirTool) RiskLevel() RiskLevel {
	return RiskReadOnly
}

func (t *ListDirTool) Execute(ctx context.Context, argsJSON string) (string, error) {
	var args ListDirArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("invalid arguments for list_dir: %w", err)
	}
	if args.Path == "" {
		args.Path = "."
	}

	absPath, err := security.ValidateWorkspacePath(t.rootDir, args.Path)
	if err != nil {
		return "", err
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to read directory '%s': %w", args.Path, err)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Directory Contents of %s:\n", filepath.ToSlash(args.Path)))
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") && entry.Name() != "." {
			continue // skip dotfiles
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if entry.IsDir() {
			sb.WriteString(fmt.Sprintf("[DIR]  %s/\n", entry.Name()))
		} else {
			sb.WriteString(fmt.Sprintf("[FILE] %-30s (%10d bytes)\n", entry.Name(), info.Size()))
		}
	}

	return strings.TrimSpace(sb.String()), nil
}
