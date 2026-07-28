package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/awanmh/Nova/internal/agent"
	"github.com/awanmh/Nova/internal/llm"
	"github.com/awanmh/Nova/internal/llm/ollama"
	"github.com/awanmh/Nova/internal/memory"
	"github.com/awanmh/Nova/internal/permission"
	"github.com/awanmh/Nova/internal/persona"
	"github.com/awanmh/Nova/internal/tools"
	"github.com/spf13/cobra"
)

var (
	runModel   string
	runPersona string
	runMaxIter int
)

var runCmd = &cobra.Command{
	Use:   "run <prompt>",
	Short: "Execute an autonomous AI coding task in the current workspace",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		prompt := args[0]
		dir, _ := os.Getwd()

		// 1. Setup Provider (default Ollama, fallback to mock if no host specified)
		var provider llm.Provider = ollama.NewProvider("http://localhost:11434")

		// 2. Setup Tools Registry and Executor
		reg := tools.NewRegistry()
		if err := tools.RegisterStandardTools(dir, reg); err != nil {
			return fmt.Errorf("error registering standard tools: %w", err)
		}

		perm := permission.NewEngine(permission.PolicyAllow, nil)
		auditLog, _ := permission.NewFileAuditLogger(dir)
		executor := tools.NewExecutor(dir, reg, perm, auditLog)

		// 3. Setup Memory Store
		store, _ := memory.NewFileStore(dir)

		// 4. Setup Persona
		pMgr := persona.NewManager()
		p := pMgr.Get(runPersona)

		// 5. Create Agent Runner
		runner := agent.NewRunner(
			provider,
			runModel,
			executor,
			store,
			"cli-session",
			p.Name,
			p.SystemRule,
			runMaxIter,
		)

		fmt.Printf("NOVA Agent executing task [%s mode]: %s\n", p.Name, prompt)
		err := runner.Run(context.Background(), prompt)
		if err != nil {
			fmt.Printf("Agent finished with state %s (Note: %v)\n", runner.State(), err)
		} else {
			fmt.Printf("Agent task COMPLETED successfully.\n")
		}
		return nil
	},
}

func init() {
	runCmd.Flags().StringVarP(&runModel, "model", "m", "llama3", "LLM model name to use")
	runCmd.Flags().StringVarP(&runPersona, "persona", "p", "general", "Engineering persona mode (general, architect, security, tdd)")
	runCmd.Flags().IntVar(&runMaxIter, "max-steps", 10, "Maximum autonomous loop iterations")
	rootCmd.AddCommand(runCmd)
}
