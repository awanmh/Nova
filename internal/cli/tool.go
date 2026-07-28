package cli

import (
	"fmt"
	"os"

	"github.com/awanmh/Nova/internal/tools"
	"github.com/spf13/cobra"
)

var toolCmd = &cobra.Command{
	Use:   "tool",
	Short: "Inspect, list, and test built-in NOVA AI execution tools",
}

var toolListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered tools with their descriptions and safety risk classifications",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := os.Getwd()
		reg := tools.NewRegistry()
		if err := tools.RegisterStandardTools(dir, reg); err != nil {
			return fmt.Errorf("error registering tools: %w", err)
		}

		list := reg.List()
		fmt.Println("Registered NOVA Execution Tools:")
		fmt.Println("=====================================================================")
		fmt.Printf("%-18s %-14s %s\n", "TOOL NAME", "RISK LEVEL", "DESCRIPTION")
		fmt.Println("---------------------------------------------------------------------")
		for _, t := range list {
			fmt.Printf("%-18s %-14s %s\n", t.Name(), t.RiskLevel(), t.Description())
		}
		return nil
	},
}

func init() {
	toolCmd.AddCommand(toolListCmd)
	rootCmd.AddCommand(toolCmd)
}
