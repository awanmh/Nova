package persona_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/awanmh/Nova/internal/persona"
)

func TestBuiltinPersonas(t *testing.T) {
	mgr := persona.NewManager()

	// Verify all 4 canonical engineering personas exist
	names := []string{"general", "architect", "security", "tdd"}
	for _, n := range names {
		p := mgr.Get(n)
		if p.Name != n {
			t.Fatalf("expected persona '%s', got '%s'", n, p.Name)
		}
		if p.SystemRule == "" {
			t.Fatalf("persona '%s' has empty SystemRule", n)
		}
	}

	// Unknown persona should fallback to "general"
	unknown := mgr.Get("non_existent_persona")
	if unknown.Name != "general" {
		t.Fatalf("expected fallback to 'general', got '%s'", unknown.Name)
	}
}

func TestCustomPersonaLoader(t *testing.T) {
	tempDir := t.TempDir()
	mgr := persona.NewManager()

	// Create a custom markdown persona file
	mdContent := `# Database Optimization Specialist
You are an expert in PostgreSQL indexing, query planning, and schema normalisation.
Always check EXPLAIN ANALYZE before optimizing.`

	mdPath := filepath.Join(tempDir, "db_expert.md")
	if err := os.WriteFile(mdPath, []byte(mdContent), 0644); err != nil {
		t.Fatalf("failed to write custom md persona: %v", err)
	}

	// Create a custom json persona file
	jsonContent := `{"name":"ui_designer","description":"UI Designer","system_rule":"Focus on aesthetics and CSS."}`
	jsonPath := filepath.Join(tempDir, "ui_designer.json")
	if err := os.WriteFile(jsonPath, []byte(jsonContent), 0644); err != nil {
		t.Fatalf("failed to write custom json persona: %v", err)
	}

	if err := persona.LoadFromDirectory(tempDir, mgr); err != nil {
		t.Fatalf("LoadFromDirectory failed: %v", err)
	}

	// Verify db_expert was loaded
	dbExpert := mgr.Get("db_expert")
	if dbExpert.Name != "db_expert" {
		t.Fatalf("expected custom persona 'db_expert', got '%s'", dbExpert.Name)
	}
	if dbExpert.Description != "Database Optimization Specialist" {
		t.Fatalf("unexpected description: %s", dbExpert.Description)
	}

	// Verify ui_designer was loaded
	uiDesigner := mgr.Get("ui_designer")
	if uiDesigner.Name != "ui_designer" {
		t.Fatalf("expected custom persona 'ui_designer', got '%s'", uiDesigner.Name)
	}
}
