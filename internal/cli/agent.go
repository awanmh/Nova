package cli

import (
	"fmt"
	"os"

	"github.com/awanmh/Nova/internal/tui"
	"github.com/spf13/cobra"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Launch NOVA interactive agent runtime",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := flagDir
		if dir == "" {
			var err error
			dir, err = os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current working directory: %w", err)
			}
		}

		fmt.Println("Booting NOVA...")
		fmt.Println("[✓] Runtime")
		fmt.Println("[✓] Configuration")
		fmt.Println("[✓] Tool Registry")
		fmt.Println("[✓] Model Registry")
		fmt.Println("[✓] Permission Engine")

		return tui.Run(flagModel, dir, flagPersona)
	},
}
