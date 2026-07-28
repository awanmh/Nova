package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/awanmh/Nova/internal/config"
	"github.com/awanmh/Nova/internal/security"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage or inspect NOVA configuration (~/.nova/config.yaml)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigShow()
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current active configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigShow()
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration parameter (e.g. openai.api_key, default_model, default_provider)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := strings.ToLower(strings.TrimSpace(args[0]))
		val := strings.TrimSpace(args[1])

		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to locate user home directory: %w", err)
		}
		configPath := filepath.Join(home, ".nova", "config.yaml")
		cfg, err := config.LoadFromPath(configPath)
		if err != nil {
			return err
		}

		switch key {
		case "default_model", "model":
			cfg.DefaultModel = val
		case "default_provider", "provider":
			cfg.DefaultProvider = val
		case "openai.api_key", "api_key", "apikey":
			cfg.OpenAI.APIKey = val
			if cfg.DefaultProvider == "" || cfg.DefaultProvider == "ollama" {
				cfg.DefaultProvider = "openai"
			}
		case "openai.endpoint":
			cfg.OpenAI.Endpoint = val
		case "ollama.endpoint", "ollama_host":
			cfg.Ollama.Endpoint = val
		case "safety.default_policy", "policy":
			cfg.Safety.DefaultPolicy = val
		default:
			return fmt.Errorf("unknown config key '%s'. Supported keys: default_model, default_provider, openai.api_key, openai.endpoint, ollama.endpoint, safety.default_policy", key)
		}

		if err := cfg.SaveToPath(configPath); err != nil {
			return fmt.Errorf("failed to save config to %s: %w", configPath, err)
		}

		displayVal := val
		if strings.Contains(key, "api_key") || strings.Contains(key, "apikey") || strings.Contains(key, "key") {
			displayVal = security.Redact(val)
		}

		fmt.Printf("✅ NOVA configuration updated successfully (%s):\n", configPath)
		fmt.Printf("   %s = %s\n", key, displayVal)
		return nil
	},
}

func runConfigShow() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to locate user home directory: %w", err)
	}
	configPath := filepath.Join(home, ".nova", "config.yaml")

	cfg, err := config.LoadFromPath(configPath)
	if err != nil {
		return err
	}

	fmt.Printf("Current NOVA Configuration (%s):\n", configPath)
	fmt.Println("=====================================================================")
	data, _ := yaml.Marshal(cfg)
	// Scrub secrets in display
	fmt.Print(security.Redact(string(data)))
	return nil
}

func init() {
	configCmd.AddCommand(configShowCmd, configSetCmd)
	rootCmd.AddCommand(configCmd)
}
