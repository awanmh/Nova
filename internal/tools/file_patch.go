package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/awanmh/Nova/internal/security"
)

type ApplyPatchArgs struct {
	Path        string `json:"path"`
	Target      string `json:"target_content"`
	Replacement string `json:"replacement_content"`
}

type ApplyPatchTool struct {
	rootDir string
}

func NewApplyPatchTool(rootDir string) *ApplyPatchTool {
	return &ApplyPatchTool{rootDir: rootDir}
}

func (t *ApplyPatchTool) Name() string {
	return "apply_patch"
}

func (t *ApplyPatchTool) Description() string {
	return "Replace an exact target code substring in a workspace file with replacement content."
}

func (t *ApplyPatchTool) Schema() Schema {
	return Schema{
		Type: "object",
		Properties: map[string]Property{
			"path": {
				Type:        "string",
				Description: "Relative path to the target file inside the workspace",
			},
			"target_content": {
				Type:        "string",
				Description: "The exact character sequence in the existing file to replace",
			},
			"replacement_content": {
				Type:        "string",
				Description: "The new content that replaces target_content",
			},
		},
		Required: []string{"path", "target_content", "replacement_content"},
	}
}

func (t *ApplyPatchTool) RiskLevel() RiskLevel {
	return RiskLowModify
}

func (t *ApplyPatchTool) Execute(ctx context.Context, argsJSON string) (string, error) {
	var args ApplyPatchArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("invalid arguments for apply_patch: %w", err)
	}

	if security.IsSensitiveFile(args.Path) {
		return "", errors.New("access denied: cannot patch sensitive or secret file")
	}

	absPath, err := security.ValidateWorkspacePath(t.rootDir, args.Path)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file '%s': %w", args.Path, err)
	}

	content := string(data)
	count := strings.Count(content, args.Target)
	if count == 0 {
		return "", fmt.Errorf("target content not found in file '%s'", args.Path)
	}
	if count > 1 {
		return "", fmt.Errorf("target content matches %d times in file '%s'; must be unique", count, args.Path)
	}

	patched := strings.Replace(content, args.Target, args.Replacement, 1)
	if err := os.WriteFile(absPath, []byte(patched), 0644); err != nil {
		return "", fmt.Errorf("failed to write patched file '%s': %w", args.Path, err)
	}

	return fmt.Sprintf("Successfully patched file '%s'", args.Path), nil
}
