package tool

import (
	"strings"

	"github.com/alecerf/devstrap/internal/registry"
	"github.com/spf13/cobra"
)

func newUpgradeCmd(paths *registry.Paths) *cobra.Command {
	var (
		version string
		all     bool
	)

	long := "Upgrade installs or upgrades the specified tools.\n\nAvailable tools: " +
		strings.Join(toolNames(), ", ")

	cmd := &cobra.Command{
		Use:       "upgrade <tools... | --all>",
		Short:     "Upgrade development tools to their latest versions",
		Long:      long,
		ValidArgs: toolNames(),
		RunE: func(_ *cobra.Command, args []string) error {
			if all && version != "" {
				return errAllAndVersionConflict
			}

			resolved, err := validateAllFlag(all, args)
			if err != nil {
				return err
			}

			return runAction(*paths, resolved, version, "upgraded")
		},
	}

	cmd.Flags().StringVar(&version, "version", "", "upgrade to a specific version instead of the latest")
	cmd.Flags().BoolVar(&all, "all", false, "upgrade all tools from the index")

	return cmd
}
