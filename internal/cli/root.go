// Package cli implements the devstrap command-line interface.
package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alecerf/devstrap/internal/cli/index"
	"github.com/alecerf/devstrap/internal/cli/tool"
	"github.com/alecerf/devstrap/internal/cli/ui"
	"github.com/alecerf/devstrap/internal/registry"
	"github.com/spf13/cobra"
)

var (
	paths   registry.Paths
	verbose bool
)

func defaultDataDir() string {
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return filepath.Join(dir, "devstrap")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}

	return filepath.Join(home, ".local", "share", "devstrap")
}

func defaultBinDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}

	return filepath.Join(home, ".local", "bin")
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "devstrap",
		Short: "Bootstrap and update development tools",
		Long: `devstrap keeps your development tools up-to-date.

It can install and upgrade tools defined in the devstrap index.
Run "devstrap index update" to fetch the latest tool definitions,
"devstrap tool install --all --dry-run" to see what needs upgrading,
or "devstrap tool install --all" to bring everything to the latest version.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			if verbose {
				ui.Verbosity = ui.Verbose
			}
		},
	}

	root.PersistentFlags().
		StringVar(&paths.DataDir, "data-dir", defaultDataDir(), "directory for tool installations")
	root.PersistentFlags().
		StringVar(&paths.BinDir, "bin-dir", defaultBinDir(), "directory for standalone binaries")
	root.PersistentFlags().
		BoolVarP(&verbose, "verbose", "v", false, "show detailed output")

	root.AddCommand(
		tool.NewCmd(&paths),
		newVersionCmd(),
		newEnvCmd(),
		index.NewCmd(),
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
