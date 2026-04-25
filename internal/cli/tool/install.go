package tool

import (
	"strings"

	"github.com/alecerf/devstrap/internal/registry"
	"github.com/spf13/cobra"
)

func newInstallCmd(paths *registry.Paths) *cobra.Command {
	var version string

	cmd := &cobra.Command{
		Use:   "install <tools...>",
		Short: "Install one or more development tools",
		Long: "Install downloads and sets up the specified tools.\n\nAvailable tools: " +
			strings.Join(toolNames(), ", "),
		Args:      cobra.MinimumNArgs(1),
		ValidArgs: toolNames(),
		RunE: func(_ *cobra.Command, args []string) error {
			return runAction(*paths, args, version, "installed")
		},
	}

	cmd.Flags().StringVar(&version, "version", "", "install a specific version instead of the latest")

	return cmd
}
