package cli

import "github.com/spf13/cobra"

func newToolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tool",
		Short: "Manage development tools",
		Long:  "Commands for checking, upgrading, and listing development tools.",
	}

	cmd.AddCommand(
		newUpdateCmd(),
		newUpgradeCmd(),
		newListCmd(),
	)

	return cmd
}
