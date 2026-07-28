package cli

import (
	"context"
	"fmt"

	"github.com/awanmh/Nova/internal/llm"
	"github.com/awanmh/Nova/internal/llm/ollama"
	"github.com/spf13/cobra"
)

var askModel string

var askCmd = &cobra.Command{
	Use:   "ask <question>",
	Short: "Ask NOVA a quick one-shot question without executing tools",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		question := args[0]
		provider := ollama.NewProvider("http://localhost:11434")

		fmt.Printf("NOVA Question [%s]: %s\n", askModel, question)
		resp, err := provider.Chat(context.Background(), &llm.ChatRequest{
			Model: askModel,
			Messages: []llm.Message{
				{
					Role:    "system",
					Content: "You are NOVA, a helpful software engineering assistant. Answer the question concisely.",
				},
				{
					Role:    "user",
					Content: question,
				},
			},
			Temperature: 0.2,
		})

		if err != nil {
			return fmt.Errorf("LLM chat error: %w", err)
		}

		fmt.Println("\nNOVA Response:")
		fmt.Println("-------------------------------------------------------------------------")
		fmt.Println(resp.Message.Content)
		return nil
	},
}

func init() {
	askCmd.Flags().StringVarP(&askModel, "model", "m", "llama3", "LLM model name to use")
	rootCmd.AddCommand(askCmd)
}
