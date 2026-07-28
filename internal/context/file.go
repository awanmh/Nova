package context

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/awanmh/Nova/internal/security"
)

var (
	ErrSensitiveFile = errors.New("cannot load sensitive file into context")
	ErrFileTooLarge  = errors.New("file exceeds maximum context file size limit (500KB)")
)

const maxFileSize = 500 * 1024 // 500 KB

// LoadFile reads a workspace file and wraps it as a Context Item, enforcing security rules and size bounds.
func LoadFile(rootDir, relPath string) (*Item, error) {
	if security.IsSensitiveFile(relPath) {
		return nil, ErrSensitiveFile
	}

	absPath, err := security.ValidateWorkspacePath(rootDir, relPath)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file '%s': %w", relPath, err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("path '%s' is a directory, not a file", relPath)
	}
	if info.Size() > maxFileSize {
		return nil, ErrFileTooLarge
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file '%s': %w", relPath, err)
	}

	content := string(data)
	tokens := EstimateTokens(content)

	return &Item{
		Type:     ItemTypeFile,
		Path:     filepath.ToSlash(relPath),
		Content:  content,
		Score:    1.0,
		TokenCnt: tokens,
	}, nil
}
