// Package cli implements the devstrap command-line interface using cobra.
package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var baseDir string

func newRootCmd() *cobra.Command {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}

	defaultBase := filepath.Join(home, "Workspace")

	root := &cobra.Command{
		Use:   "devstrap",
		Short: "Bootstrap and update development tools",
		Long: `devstrap keeps your development tools up-to-date.

It can install and upgrade Go, Node.js, golangci-lint, and more.
Run "devstrap update" to see what needs upgrading,
or "devstrap upgrade" to bring everything to the latest version.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringVar(&baseDir, "base", defaultBase, "base directory for installations")

	root.AddCommand(
		newUpdateCmd(),
		newUpgradeCmd(),
		newListCmd(),
		newVersionCmd(),
	)

	return root
}

// Execute runs the root command.
func Execute() {
	err := newRootCmd().Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
