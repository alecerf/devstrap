package tool

import (
	"strings"

	"github.com/alecerf/devstrap/internal/registry"
	"github.com/spf13/cobra"
)

func newUpgradeCmd(paths *registry.Paths) *cobra.Command {
	var version string

	long := "Upgrade installs or upgrades the specified tools" +
		" (or all tools if none given).\n\nAvailable tools: " +
		strings.Join(toolNames(), ", ")

	cmd := &cobra.Command{
		Use:       "upgrade [tools...]",
		Short:     "Upgrade development tools to their latest versions",
		Long:      long,
		ValidArgs: toolNames(),
		RunE: func(_ *cobra.Command, args []string) error {
			return runAction(*paths, args, version, "upgraded")
		},
	}

	cmd.Flags().StringVar(&version, "version", "", "upgrade to a specific version instead of the latest")

	return cmd
}
