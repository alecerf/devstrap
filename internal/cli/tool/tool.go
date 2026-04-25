// Package tool implements the "devstrap tool" command and its subcommands.
package tool

import "github.com/spf13/cobra"

// NewCmd returns the "tool" parent command with its subcommands registered.
func NewCmd(baseDir *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tool",
		Short: "Manage development tools",
		Long:  "Commands for checking, upgrading, and listing development tools.",
	}

	cmd.AddCommand(
		newInstallCmd(baseDir),
		newUpdateCmd(baseDir),
		newUpgradeCmd(baseDir),
		newListCmd(baseDir),
	)

	return cmd
}
