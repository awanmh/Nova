package context

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/awanmh/Nova/internal/security"
)

// SearchSnippet scans a single file for case-insensitive occurrences of query and extracts
// snippets with +/- 2 surrounding lines of context.
func SearchSnippet(rootDir, relPath, query string) ([]Item, error) {
	if query == "" || security.IsSensitiveFile(relPath) {
		return nil, nil
	}

	absPath, err := security.ValidateWorkspacePath(rootDir, relPath)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(absPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	var items []Item
	lowerQuery := strings.ToLower(query)

	for i, line := range lines {
		if strings.Contains(strings.ToLower(line), lowerQuery) {
			start := i - 2
			if start < 0 {
				start = 0
			}
			end := i + 2
			if end >= len(lines) {
				end = len(lines) - 1
			}

			snippetLines := lines[start : end+1]
			content := fmt.Sprintf("// %s:%d\n%s", filepath.ToSlash(relPath), i+1, strings.Join(snippetLines, "\n"))
			items = append(items, Item{
				Type:     ItemTypeSearch,
				Path:     filepath.ToSlash(relPath),
				Content:  content,
				Score:    0.7,
				TokenCnt: EstimateTokens(content),
			})
		}
	}

	return items, nil
}
