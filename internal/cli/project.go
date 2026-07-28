package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/awanmh/Nova/internal/project"
	"github.com/spf13/cobra"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Inspect and analyze local project repository architecture and dependencies",
}

var projectInspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Scan the current workspace and display a comprehensive architectural overview",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := flagDir
		if dir == "" {
			var err error
			dir, err = os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current working directory: %w", err)
			}
		}

		fmt.Printf("Scanning repository at: %s...\n\n", dir)
		scanner := project.NewWorkspaceScanner()
		snap, err := scanner.Scan(context.Background(), dir)
		if err != nil {
			return fmt.Errorf("failed to scan workspace: %w", err)
		}

		fmt.Println("Project Architectural Snapshot:")
		fmt.Println("=====================================================================")
		fmt.Println(snap.Summary)
		fmt.Println("---------------------------------------------------------------------")
		fmt.Printf("Detected Architecture Mode: %s\n", snap.Architecture)
		return nil
	},
}

func init() {
	projectCmd.AddCommand(projectInspectCmd)
	rootCmd.AddCommand(projectCmd)
}
