package security

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// Redactor handles sanitizing sensitive information from text strings.
type Redactor struct {
	patterns []*regexp.Regexp
}

// NewRedactor creates a new redactor with default secret detection patterns.
func NewRedactor() *Redactor {
	// Common secret patterns: API keys, tokens, private key headers, and env assignment patterns
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(api[_-]?key|secret|token|password|auth)[=:]\s*["']?([A-Za-z0-9_\-\.]{16,})["']?`),
		regexp.MustCompile(`-----BEGIN (RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----[\s\S]*?-----END (RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----`),
		regexp.MustCompile(`(sk-[a-zA-Z0-9]{20,})`),
	}
	return &Redactor{patterns: patterns}
}

// Redact replaces detected secret patterns with "[REDACTED]".
func (r *Redactor) Redact(input string) string {
	output := input
	for _, p := range r.patterns {
		output = p.ReplaceAllStringFunc(output, func(match string) string {
			if strings.HasPrefix(match, "-----BEGIN") {
				return "[REDACTED PRIVATE KEY]"
			}
			parts := strings.SplitN(match, "=", 2)
			if len(parts) == 2 {
				return parts[0] + "=[REDACTED]"
			}
			parts = strings.SplitN(match, ":", 2)
			if len(parts) == 2 {
				return parts[0] + ": [REDACTED]"
			}
			return "[REDACTED]"
		})
	}
	return output
}

var defaultRedactor = NewRedactor()

// Redact is a convenient package-level wrapper for sanitizing secrets from input text.
func Redact(input string) string {
	return defaultRedactor.Redact(input)
}

// ValidateWorkspacePath verifies that targetPath is strictly inside or equal to rootDir.
// Prevents directory traversal attacks via ../ or symlinks.
func ValidateWorkspacePath(rootDir, targetPath string) (string, error) {
	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute root: %w", err)
	}
	absRoot = filepath.Clean(absRoot)

	absTarget, err := filepath.Abs(filepath.Join(absRoot, targetPath))
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute target: %w", err)
	}
	absTarget = filepath.Clean(absTarget)

	// Ensure absTarget starts with absRoot
	rel, err := filepath.Rel(absRoot, absTarget)
	if err != nil || strings.HasPrefix(rel, "..") || rel == ".." {
		return "", fmt.Errorf("security boundary violation: path '%s' is outside workspace '%s'", targetPath, rootDir)
	}

	return absTarget, nil
}

// IsSensitiveFile returns true if the filename matches sensitive patterns like .env, *.pem, *.key.
func IsSensitiveFile(filename string) bool {
	base := strings.ToLower(filepath.Base(filename))
	if strings.HasPrefix(base, ".env") {
		return true
	}
	if strings.HasSuffix(base, ".pem") || strings.HasSuffix(base, ".key") || strings.HasPrefix(base, "credentials.") {
		return true
	}
	return false
}
