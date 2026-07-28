package tools

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/awanmh/Nova/internal/security"
)

type SearchFilesArgs struct {
	Query string `json:"query"`
}

type SearchFilesTool struct {
	rootDir string
}

func NewSearchFilesTool(rootDir string) *SearchFilesTool {
	return &SearchFilesTool{rootDir: rootDir}
}

func (t *SearchFilesTool) Name() string {
	return "search_files"
}

func (t *SearchFilesTool) Description() string {
	return "Search for a keyword or string pattern across source files in the workspace."
}

func (t *SearchFilesTool) Schema() Schema {
	return Schema{
		Type: "object",
		Properties: map[string]Property{
			"query": {
				Type:        "string",
				Description: "Case-insensitive string keyword or pattern to search for",
			},
		},
		Required: []string{"query"},
	}
}

func (t *SearchFilesTool) RiskLevel() RiskLevel {
	return RiskReadOnly
}

func (t *SearchFilesTool) Execute(ctx context.Context, argsJSON string) (string, error) {
	var args SearchFilesArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("invalid arguments for search_files: %w", err)
	}
	if strings.TrimSpace(args.Query) == "" {
		return "", fmt.Errorf("search query cannot be empty")
	}

	lowerQuery := strings.ToLower(args.Query)
	var matches []string

	err := filepath.WalkDir(t.rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "node_modules", "vendor", "bin", "dist", ".nova":
				return filepath.SkipDir
			}
			return nil
		}

		rel, err := filepath.Rel(t.rootDir, path)
		if err != nil {
			return nil
		}
		relSlash := filepath.ToSlash(rel)
		if security.IsSensitiveFile(relSlash) {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			if strings.Contains(strings.ToLower(line), lowerQuery) {
				matches = append(matches, fmt.Sprintf("%s:%d: %s", relSlash, lineNum, strings.TrimSpace(line)))
				if len(matches) >= 50 {
					return filepath.SkipAll
				}
			}
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	if len(matches) == 0 {
		return fmt.Sprintf("No matches found for query '%s'", args.Query), nil
	}

	return strings.Join(matches, "\n"), nil
}
