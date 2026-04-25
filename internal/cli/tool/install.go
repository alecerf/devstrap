package tool

import (
	"strings"

	"github.com/alecerf/devstrap/internal/registry"
	"github.com/spf13/cobra"
)

func newInstallCmd(paths *registry.Paths) *cobra.Command {
	var (
		version string
		all     bool
	)

	cmd := &cobra.Command{
		Use:   "install <tools... | --all>",
		Short: "Install one or more development tools",
		Long: "Install downloads and sets up the specified tools.\n\nAvailable tools: " +
			strings.Join(toolNames(), ", "),
		ValidArgs: toolNames(),
		RunE: func(_ *cobra.Command, args []string) error {
			if all && version != "" {
				return errAllAndVersionConflict
			}

			resolved, err := validateAllFlag(all, args)
			if err != nil {
				return err
			}

			return runAction(*paths, resolved, version, "installed")
		},
	}

	cmd.Flags().StringVar(&version, "version", "", "install a specific version instead of the latest")
	cmd.Flags().BoolVar(&all, "all", false, "install all tools from the index")

	return cmd
}
