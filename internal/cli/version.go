package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	// Version is the semantic version of NOVA.
	Version = "0.1.0-dev"
	// BuildDate is the timestamp when binary was compiled.
	BuildDate = "2026-07-28"
	// GitCommit is the commit hash of the build.
	GitCommit = "dev"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the build version and metadata of NOVA",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("NOVA v%s\n", Version)
		fmt.Printf("commit: %s\n", GitCommit)
		fmt.Printf("build: %s\n", BuildDate)
		fmt.Printf("platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}
