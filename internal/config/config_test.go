package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/awanmh/Nova/internal/config"
)

func TestConfig_DefaultAndValidate(t *testing.T) {
	cfg := config.DefaultConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected default config to be valid, got: %v", err)
	}

	cfg.Safety.DefaultPolicy = "invalid_policy"
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validation error for invalid policy, got nil")
	}
}

func TestConfig_Migrate(t *testing.T) {
	cfg := &config.Config{
		DefaultModel: "custom-model",
	}
	cfg.Migrate()

	if cfg.DefaultProvider != "ollama" {
		t.Fatalf("expected DefaultProvider to be migrated to 'ollama', got '%s'", cfg.DefaultProvider)
	}
	if cfg.Ollama.Endpoint != "http://localhost:11434" {
		t.Fatalf("expected Ollama endpoint to be migrated, got '%s'", cfg.Ollama.Endpoint)
	}
}

func TestConfig_ApplyEnvironmentOverrides(t *testing.T) {
	cfg := config.DefaultConfig()

	t.Setenv("NOVA_MODEL", "env-model")
	t.Setenv("NOVA_PROVIDER", "openai")
	t.Setenv("OLLAMA_HOST", "http://remote:11434")

	cfg.ApplyEnvironmentOverrides()

	if cfg.DefaultModel != "env-model" {
		t.Fatalf("expected 'env-model', got '%s'", cfg.DefaultModel)
	}
	if cfg.DefaultProvider != "openai" {
		t.Fatalf("expected 'openai', got '%s'", cfg.DefaultProvider)
	}
	if cfg.Ollama.Endpoint != "http://remote:11434" {
		t.Fatalf("expected 'http://remote:11434', got '%s'", cfg.Ollama.Endpoint)
	}
}

func TestConfig_LoadHierarchy(t *testing.T) {
	tempDir := t.TempDir()
	novaDir := filepath.Join(tempDir, ".nova")
	_ = os.MkdirAll(novaDir, 0755)

	projConfigPath := filepath.Join(novaDir, "config.yaml")
	_ = os.WriteFile(projConfigPath, []byte("default_model: project-model\ndefault_provider: ollama\n"), 0644)

	// Load with CLI override
	cfg, err := config.LoadHierarchy(tempDir, "cli-model", "", "")
	if err != nil {
		t.Fatalf("unexpected error loading hierarchy: %v", err)
	}

	// CLI override should win over project-model
	if cfg.DefaultModel != "cli-model" {
		t.Fatalf("expected 'cli-model' from CLI override, got '%s'", cfg.DefaultModel)
	}
}
