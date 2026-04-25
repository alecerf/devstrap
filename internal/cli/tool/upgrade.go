package tool

import (
	"strings"

	"github.com/spf13/cobra"
)

func newUpgradeCmd(baseDir *string) *cobra.Command {
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
			return runAction(*baseDir, args, version, "upgraded")
		},
	}

	cmd.Flags().StringVar(&version, "version", "", "upgrade to a specific version instead of the latest")

	return cmd
}
