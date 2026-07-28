package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/awanmh/Nova/internal/llm"
	"github.com/awanmh/Nova/internal/llm/llamacpp"
	"github.com/awanmh/Nova/internal/llm/ollama"
	"github.com/awanmh/Nova/internal/llm/openai"
	"github.com/spf13/cobra"
)

var modelCmd = &cobra.Command{
	Use:   "model",
	Short: "Inspect, list, and check available AI models across providers",
}

var modelListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all discovered models across Ollama, OpenAI, and llama.cpp providers",
	RunE: func(cmd *cobra.Command, args []string) error {
		reg := llm.NewRegistry()
		_ = reg.Register(ollama.NewProvider(os.Getenv("OLLAMA_HOST")))
		if key := os.Getenv("OPENAI_API_KEY"); key != "" {
			_ = reg.Register(openai.NewProvider(os.Getenv("OPENAI_ENDPOINT"), key))
		}
		_ = reg.Register(llamacpp.NewProvider(os.Getenv("LLAMACPP_ENDPOINT")))

		ctx := context.Background()
		models, err := reg.DiscoverModels(ctx)
		if err != nil {
			return fmt.Errorf("error discovering models: %w", err)
		}

		if len(models) == 0 {
			fmt.Println("No models discovered. Ensure Ollama (localhost:11434) or llama.cpp (localhost:8080) is running, or OPENAI_API_KEY is set.")
			return nil
		}

		fmt.Println("Discovered AI Models:")
		fmt.Println("=====================================================================")
		fmt.Printf("%-30s %-12s %-8s %-12s\n", "NAME", "PROVIDER", "STATUS", "CAPABILITIES")
		fmt.Println("---------------------------------------------------------------------")
		for _, m := range models {
			caps := "chat,tools"
			if m.Capability.Reasoning {
				caps += ",reasoning"
			}
			fmt.Printf("%-30s %-12s %-8s %-12s\n", m.Name, m.Provider, m.Status, caps)
		}
		return nil
	},
}

func init() {
	modelCmd.AddCommand(modelListCmd)
	rootCmd.AddCommand(modelCmd)
}
