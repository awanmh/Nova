package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/awanmh/Nova/internal/permission"
	"github.com/spf13/cobra"
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Inspect tool execution audit trails and safety permission logs",
}

var auditListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent tool executions and permission decisions from .nova/audit.log",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := os.Getwd()
		logger, err := permission.NewFileAuditLogger(dir)
		if err != nil {
			return fmt.Errorf("error opening audit log: %w", err)
		}

		records, err := logger.ReadRecent(context.Background(), 50)
		if err != nil {
			return fmt.Errorf("error reading audit log: %w", err)
		}

		if len(records) == 0 {
			fmt.Println("No audit log records found in .nova/audit.log.")
			return nil
		}

		fmt.Println("Recent Tool Execution Audit Logs:")
		fmt.Println("=====================================================================================")
		fmt.Printf("%-20s %-15s %-12s %-10s %s\n", "TIMESTAMP", "TOOL", "RISK", "STATUS", "DURATION")
		fmt.Println("-------------------------------------------------------------------------------------")
		for _, r := range records {
			ts := r.Timestamp.Format("2006-01-02 15:04:05")
			fmt.Printf("%-20s %-15s %-12s %-10s %dms\n", ts, r.ToolName, r.RiskLevel, r.Status, r.DurationMs)
		}
		return nil
	},
}

func init() {
	auditCmd.AddCommand(auditListCmd)
	rootCmd.AddCommand(auditCmd)
}
