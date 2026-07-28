package cli

import (
	"fmt"
	"os"

	"github.com/awanmh/Nova/internal/tui"
	"github.com/spf13/cobra"
)

var (
	flagModel   string
	flagDir     string
	flagPersona string
	flagResume  string
	flagMode    string
)

var rootCmd = &cobra.Command{
	Use:   "nova",
	Short: "NOVA is a local-first AI software engineering agent",
	Long: `NOVA — Local Agent Runtime
A CLI & TUI autonomous software engineering agent that understands,
modifies, executes, and verifies code safely on your local machine.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := flagDir
		if dir == "" {
			dir, _ = os.Getwd()
		}
		return tui.Run(flagModel, dir, flagPersona)
	},
}

// Execute is the primary CLI entry point for NOVA.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&flagModel, "model", "m", "qwen3-coder", "Target AI model name")
	rootCmd.PersistentFlags().StringVarP(&flagDir, "dir", "d", "", "Target workspace directory")
	rootCmd.PersistentFlags().StringVarP(&flagPersona, "persona", "p", "general", "Agent engineering persona")
	rootCmd.PersistentFlags().StringVar(&flagResume, "resume", "", "Resume a previous session ID")
	rootCmd.PersistentFlags().StringVar(&flagMode, "mode", "agent", "Execution mode (agent, chat, plan, review, debug)")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(agentCmd)
}
