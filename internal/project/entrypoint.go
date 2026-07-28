package project

import (
	"path/filepath"
	"sort"
	"strings"
)

// DetectEntryPoints identifies likely application entry points from scanned workspace files.
func DetectEntryPoints(files []FileMeta) []string {
	var entryPoints []string
	seen := make(map[string]bool)

	for _, f := range files {
		pathSlash := filepath.ToSlash(f.Path)

		// Go CLI/Service standard entry points
		if strings.HasPrefix(pathSlash, "cmd/") && strings.HasSuffix(pathSlash, "/main.go") {
			if !seen[pathSlash] {
				entryPoints = append(entryPoints, pathSlash)
				seen[pathSlash] = true
			}
			continue
		}
		if pathSlash == "main.go" || pathSlash == "main.py" || pathSlash == "app.py" || pathSlash == "index.js" || pathSlash == "src/index.ts" || pathSlash == "src/main.rs" || pathSlash == "index.html" {
			if !seen[pathSlash] {
				entryPoints = append(entryPoints, pathSlash)
				seen[pathSlash] = true
			}
		}
	}

	sort.Strings(entryPoints)
	return entryPoints
}
