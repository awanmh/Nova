package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/awanmh/Nova/internal/persona"
	"github.com/spf13/cobra"
)

var personaCmd = &cobra.Command{
	Use:   "persona",
	Short: "Inspect and list available NOVA engineering personas and behavioral modes",
}

var personaListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all built-in and custom engineering personas",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := persona.NewManager()
		dir, _ := os.Getwd()
		_ = persona.LoadFromDirectory(filepath.Join(dir, ".nova", "personas"), mgr)

		list := mgr.List()
		fmt.Println("Available NOVA Engineering Personas:")
		fmt.Println("=================================================================================")
		fmt.Printf("%-14s %-38s %s\n", "NAME", "DESCRIPTION", "BUDGET MOD")
		fmt.Println("---------------------------------------------------------------------------------")
		for _, p := range list {
			mod := p.TokenBudgetModifier
			if mod == 0 {
				mod = 1.0
			}
			fmt.Printf("%-14s %-38s %.1fx\n", p.Name, p.Description, mod)
		}
		return nil
	},
}

func init() {
	personaCmd.AddCommand(personaListCmd)
	rootCmd.AddCommand(personaCmd)
}
