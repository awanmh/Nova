package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/awanmh/Nova/internal/security"
)

type GitStatusTool struct {
	rootDir string
}

func NewGitStatusTool(rootDir string) *GitStatusTool {
	return &GitStatusTool{rootDir: rootDir}
}

func (t *GitStatusTool) Name() string {
	return "git_status"
}

func (t *GitStatusTool) Description() string {
	return "Inspect git working tree status (modified, staged, and untracked files)."
}

func (t *GitStatusTool) Schema() Schema {
	return Schema{
		Type:       "object",
		Properties: map[string]Property{},
	}
}

func (t *GitStatusTool) RiskLevel() RiskLevel {
	return RiskReadOnly
}

func (t *GitStatusTool) Execute(ctx context.Context, argsJSON string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "status", "--short")
	cmd.Dir = t.rootDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git status error: %w: %s", err, string(out))
	}
	res := strings.TrimSpace(string(out))
	if res == "" {
		return "Working tree clean (no modified files)", nil
	}
	return security.Redact(res), nil
}

type GitDiffTool struct {
	rootDir string
}

func NewGitDiffTool(rootDir string) *GitDiffTool {
	return &GitDiffTool{rootDir: rootDir}
}

func (t *GitDiffTool) Name() string {
	return "git_diff"
}

func (t *GitDiffTool) Description() string {
	return "Inspect git diff of modified files in the workspace."
}

func (t *GitDiffTool) Schema() Schema {
	return Schema{
		Type:       "object",
		Properties: map[string]Property{},
	}
}

func (t *GitDiffTool) RiskLevel() RiskLevel {
	return RiskReadOnly
}

func (t *GitDiffTool) Execute(ctx context.Context, argsJSON string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "diff")
	cmd.Dir = t.rootDir
	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &outBuf
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("git diff error: %w: %s", err, outBuf.String())
	}
	res := strings.TrimSpace(outBuf.String())
	if res == "" {
		return "No unstaged diffs found", nil
	}
	return security.Redact(res), nil
}
