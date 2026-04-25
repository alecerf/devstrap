// Package tool implements the "devstrap tool" command and its subcommands.
package tool

import (
	"github.com/alecerf/devstrap/internal/registry"
	"github.com/spf13/cobra"
)

// NewCmd returns the "tool" parent command with its subcommands registered.
func NewCmd(paths *registry.Paths) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tool",
		Short: "Manage development tools",
		Long:  "Commands for checking, upgrading, and listing development tools.",
	}

	cmd.AddCommand(
		newInstallCmd(paths),
		newUpdateCmd(paths),
		newUpgradeCmd(paths),
		newListCmd(paths),
	)

	return cmd
}
