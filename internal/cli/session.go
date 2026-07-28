package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/awanmh/Nova/internal/memory"
	"github.com/spf13/cobra"
)

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Manage and inspect NOVA conversation sessions and memory history",
}

var sessionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent conversation sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := os.Getwd()
		store, err := memory.NewFileStore(dir)
		if err != nil {
			return err
		}
		list, err := store.ListSessions(context.Background())
		if err != nil {
			return err
		}
		if len(list) == 0 {
			fmt.Println("No conversation sessions found.")
			return nil
		}
		fmt.Println("NOVA Conversation Sessions:")
		fmt.Println("==========================================================================")
		fmt.Printf("%-20s %-26s %-12s %s\n", "ID", "TITLE", "PERSONA", "UPDATED AT")
		fmt.Println("--------------------------------------------------------------------------")
		for _, s := range list {
			fmt.Printf("%-20s %-26s %-12s %s\n", s.ID, s.Title, s.Persona, s.UpdatedAt.Format("2006-01-02 15:04"))
		}
		return nil
	},
}

var sessionExportCmd = &cobra.Command{
	Use:   "export <session-id>",
	Short: "Export a session transcript to markdown format",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := os.Getwd()
		store, err := memory.NewFileStore(dir)
		if err != nil {
			return err
		}
		md, err := memory.ExportMarkdown(context.Background(), store, args[0])
		if err != nil {
			return fmt.Errorf("export error: %w", err)
		}
		fmt.Println(md)
		return nil
	},
}

func init() {
	sessionCmd.AddCommand(sessionListCmd, sessionExportCmd)
	rootCmd.AddCommand(sessionCmd)
}
