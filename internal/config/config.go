package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the root NOVA configuration.
type Config struct {
	DefaultModel    string       `yaml:"default_model"`
	DefaultProvider string       `yaml:"default_provider"`
	Ollama          OllamaConfig `yaml:"ollama"`
	OpenAI          OpenAIConfig `yaml:"openai"`
	Safety          SafetyConfig `yaml:"safety"`
	UI              UIConfig     `yaml:"ui"`
}

type OllamaConfig struct {
	Endpoint string `yaml:"endpoint"`
}

type OpenAIConfig struct {
	Endpoint string `yaml:"endpoint"`
	APIKey   string `yaml:"api_key"`
}

type SafetyConfig struct {
	DefaultPolicy     string   `yaml:"default_policy"` // "allow", "ask", "deny"
	WorkspaceBoundary bool     `yaml:"workspace_boundary"`
	RedactedPatterns  []string `yaml:"redacted_patterns"`
}

type UIConfig struct {
	Theme        string `yaml:"theme"`
	ShowProgress bool   `yaml:"show_progress"`
}

// DefaultConfig returns the default configuration for NOVA.
func DefaultConfig() *Config {
	return &Config{
		DefaultModel:    "qwen3-coder",
		DefaultProvider: "ollama",
		Ollama: OllamaConfig{
			Endpoint: "http://localhost:11434",
		},
		OpenAI: OpenAIConfig{
			Endpoint: "https://api.openai.com/v1",
		},
		Safety: SafetyConfig{
			DefaultPolicy:     "ask",
			WorkspaceBoundary: true,
			RedactedPatterns: []string{
				"*.env*",
				"*.pem",
				"*.key",
				"credentials.*",
			},
		},
		UI: UIConfig{
			Theme:        "default",
			ShowProgress: true,
		},
	}
}

// Validate checks that the configuration values are well-formed and safe.
func (c *Config) Validate() error {
	if c.DefaultModel == "" {
		return fmt.Errorf("default_model cannot be empty")
	}
	if c.DefaultProvider == "" {
		return fmt.Errorf("default_provider cannot be empty")
	}
	policy := strings.ToLower(c.Safety.DefaultPolicy)
	if policy != "allow" && policy != "ask" && policy != "deny" {
		return fmt.Errorf("invalid safety default_policy '%s': must be allow, ask, or deny", c.Safety.DefaultPolicy)
	}
	return nil
}

// Migrate ensures legacy or missing fields are populated with valid defaults.
func (c *Config) Migrate() {
	def := DefaultConfig()
	if c.DefaultModel == "" {
		c.DefaultModel = def.DefaultModel
	}
	if c.DefaultProvider == "" {
		c.DefaultProvider = def.DefaultProvider
	}
	if c.Ollama.Endpoint == "" {
		c.Ollama.Endpoint = def.Ollama.Endpoint
	}
	if c.OpenAI.Endpoint == "" {
		c.OpenAI.Endpoint = def.OpenAI.Endpoint
	}
	if c.Safety.DefaultPolicy == "" {
		c.Safety.DefaultPolicy = def.Safety.DefaultPolicy
	}
	if len(c.Safety.RedactedPatterns) == 0 {
		c.Safety.RedactedPatterns = def.Safety.RedactedPatterns
	}
	if c.UI.Theme == "" {
		c.UI.Theme = def.UI.Theme
	}
}

// ApplyEnvironmentOverrides overwrites configuration settings with environment variables if present.
func (c *Config) ApplyEnvironmentOverrides() {
	if val := os.Getenv("NOVA_MODEL"); val != "" {
		c.DefaultModel = val
	}
	if val := os.Getenv("NOVA_PROVIDER"); val != "" {
		c.DefaultProvider = val
	}
	if val := os.Getenv("NOVA_POLICY"); val != "" {
		c.Safety.DefaultPolicy = val
	}
	if val := os.Getenv("OPENAI_API_KEY"); val != "" {
		c.OpenAI.APIKey = val
	}
	if val := os.Getenv("OLLAMA_HOST"); val != "" {
		c.Ollama.Endpoint = val
	} else if val := os.Getenv("OLLAMA_ENDPOINT"); val != "" {
		c.Ollama.Endpoint = val
	}
	if val := os.Getenv("OPENAI_ENDPOINT"); val != "" {
		c.OpenAI.Endpoint = val
	}
}

// ApplyCLIOverrides overwrites configuration settings with non-empty CLI flag values.
func (c *Config) ApplyCLIOverrides(model, provider, policy string) {
	if model != "" {
		c.DefaultModel = model
	}
	if provider != "" {
		c.DefaultProvider = provider
	}
	if policy != "" {
		c.Safety.DefaultPolicy = policy
	}
}

// LoadFromPath loads YAML config from disk, or returns default if file does not exist.
func LoadFromPath(path string) (*Config, error) {
	cfg := DefaultConfig()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse yaml config: %w", err)
	}
	cfg.Migrate()
	return cfg, nil
}

// LoadHierarchy loads configuration from Global (~/.nova/config.yaml), Project (.nova/config.yaml),
// Environment variables, and CLI overrides in order of precedence.
func LoadHierarchy(workspaceDir string, cliModel, cliProvider, cliPolicy string) (*Config, error) {
	cfg := DefaultConfig()

	// 1. Global config (~/.nova/config.yaml)
	if home, err := os.UserHomeDir(); err == nil {
		globalPath := filepath.Join(home, ".nova", "config.yaml")
		if globalCfg, err := LoadFromPath(globalPath); err == nil {
			cfg = globalCfg
		}
	}

	// 2. Project config (<workspace>/.nova/config.yaml)
	if workspaceDir != "" {
		projectPath := filepath.Join(workspaceDir, ".nova", "config.yaml")
		if _, err := os.Stat(projectPath); err == nil {
			if projCfg, err := LoadFromPath(projectPath); err == nil {
				// Project config overrides global
				cfg = projCfg
			}
		}
	}

	// 3. Environment variable overrides
	cfg.ApplyEnvironmentOverrides()

	// 4. CLI flag overrides
	cfg.ApplyCLIOverrides(cliModel, cliProvider, cliPolicy)

	// 5. Migrate & Validate
	cfg.Migrate()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// SaveToPath saves YAML config to disk, creating directories if needed.
func (c *Config) SaveToPath(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

