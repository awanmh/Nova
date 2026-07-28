package persona

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// LoadFromDirectory loads custom personas from a directory (e.g. .nova/personas)
// supporting .json and .md files.
func LoadFromDirectory(dir string, manager *Manager) error {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil // directory doesn't exist yet, not an error
	}
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		ext := strings.ToLower(filepath.Ext(entry.Name()))

		switch ext {
		case ".json":
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			var p Persona
			if err := json.Unmarshal(data, &p); err == nil && p.Name != "" {
				_ = manager.Register(p)
			}

		case ".md":
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			name := strings.TrimSuffix(entry.Name(), ".md")
			p := parseMarkdownPersona(name, string(data))
			if p.Name != "" {
				_ = manager.Register(p)
			}
		}
	}
	return nil
}

func parseMarkdownPersona(defaultName, content string) Persona {
	lines := strings.Split(content, "\n")
	p := Persona{
		Name:                defaultName,
		Description:         "Custom Engineering Persona (" + defaultName + ")",
		TokenBudgetModifier: 1.0,
	}

	var body []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") && p.Description == "Custom Engineering Persona ("+defaultName+")" {
			p.Description = strings.TrimPrefix(trimmed, "# ")
			continue
		}
		body = append(body, line)
	}
	p.SystemRule = strings.TrimSpace(strings.Join(body, "\n"))
	if p.SystemRule == "" {
		p.SystemRule = "Custom persona instructions for " + defaultName
	}
	return p
}
