package cli

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose NOVA installation, environment, and local dependencies",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("NOVA System Diagnosis:")
		fmt.Println("========================")

		// Check 1: OS and Go runtime
		fmt.Printf("[✓] OS / Architecture      : %s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Printf("[✓] Go runtime             : %s\n", runtime.Version())

		// Check 2: Git availability
		if path, err := exec.LookPath("git"); err == nil {
			fmt.Printf("[✓] Git CLI                : %s\n", path)
		} else {
			fmt.Println("[✗] Git CLI                : not found in PATH")
		}

		// Check 3: ripgrep availability
		if path, err := exec.LookPath("rg"); err == nil {
			fmt.Printf("[✓] ripgrep (rg)           : %s\n", path)
		} else {
			fmt.Println("[!] ripgrep (rg)           : optional search tool not found in PATH")
		}

		// Check 4: Workspace accessibility
		cwd, err := os.Getwd()
		if err == nil {
			fmt.Printf("[✓] Workspace directory    : %s\n", cwd)
		} else {
			fmt.Printf("[✗] Workspace directory    : error (%v)\n", err)
		}

		// Check 5: Ollama reachability
		client := http.Client{Timeout: 1500 * time.Millisecond}
		resp, err := client.Get("http://localhost:11434")
		if err == nil {
			resp.Body.Close()
			fmt.Println("[✓] Ollama server          : http://localhost:11434 (READY)")
		} else {
			fmt.Println("[!] Ollama server          : http://localhost:11434 (not running or unreachable)")
		}

		fmt.Println("\nDiagnosis completed.")
	},
}
