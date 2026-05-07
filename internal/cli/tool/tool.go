// Package tool implements the "devstrap tool" subcommands.
package tool

import (
	"github.com/alecerf/devstrap/internal/registry"
	"github.com/spf13/cobra"
)

// NewCmd returns the "tool" parent command with its subcommands.
func NewCmd(paths *registry.Paths) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tool",
		Short: "Manage development tools",
	}

	cmd.AddCommand(
		newInstallCmd(paths),
		newListCmd(paths),
		newUninstallCmd(paths),
	)

	return cmd
}
