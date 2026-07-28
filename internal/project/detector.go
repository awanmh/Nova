package project

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DetectLanguages extracts unique programming languages from the scanned files.
func DetectLanguages(files []FileMeta) []string {
	langSet := make(map[string]bool)
	for _, f := range files {
		if f.Language != "" && f.Language != "Markdown" && f.Language != "Config" {
			langSet[f.Language] = true
		}
	}
	var langs []string
	for l := range langSet {
		langs = append(langs, l)
	}
	sort.Strings(langs)
	return langs
}

// DetectFrameworks inspects package manifests (go.mod, package.json) to identify libraries/frameworks.
func DetectFrameworks(rootDir string, files []FileMeta) []string {
	var frameworks []string
	fwSet := make(map[string]bool)

	for _, f := range files {
		base := strings.ToLower(filepath.Base(f.Path))

		if base == "go.mod" {
			data, err := os.ReadFile(filepath.Join(rootDir, f.Path))
			if err == nil {
				content := string(data)
				if strings.Contains(content, "github.com/spf13/cobra") {
					fwSet["Cobra (CLI)"] = true
				}
				if strings.Contains(content, "github.com/charmbracelet/bubbletea") {
					fwSet["Bubble Tea (TUI)"] = true
				}
				if strings.Contains(content, "github.com/gin-gonic/gin") {
					fwSet["Gin (HTTP)"] = true
				}
			}
		}

		if base == "package.json" {
			data, err := os.ReadFile(filepath.Join(rootDir, f.Path))
			if err == nil {
				content := string(data)
				if strings.Contains(content, "react") {
					fwSet["React"] = true
				}
				if strings.Contains(content, "next") {
					fwSet["Next.js"] = true
				}
				if strings.Contains(content, "express") {
					fwSet["Express"] = true
				}
			}
		}
	}

	for fw := range fwSet {
		frameworks = append(frameworks, fw)
	}
	sort.Strings(frameworks)
	return frameworks
}
