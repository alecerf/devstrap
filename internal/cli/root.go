// Package cli implements the devstrap command-line interface using cobra.
package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alecerf/devstrap/internal/cli/index"
	"github.com/alecerf/devstrap/internal/cli/tool"
	"github.com/alecerf/devstrap/internal/registry"
	"github.com/spf13/cobra"
)

var paths registry.Paths

// defaultDataDir returns the default data directory for tool installations,
// respecting $XDG_DATA_HOME (fallback ~/.local/share/devstrap).
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

// defaultBinDir returns the default directory for standalone binaries (~/.local/bin).
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
	}

	root.PersistentFlags().
		StringVar(&paths.DataDir, "data-dir", defaultDataDir(), "directory for tool installations")
	root.PersistentFlags().
		StringVar(&paths.BinDir, "bin-dir", defaultBinDir(), "directory for standalone binaries")

	root.AddCommand(
		tool.NewCmd(&paths),
		newVersionCmd(),
		newEnvCmd(),
		newUpdateCmd(Version),
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
