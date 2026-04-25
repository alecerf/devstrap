// Package index implements the "devstrap index" command and its subcommands.
package index

import "github.com/spf13/cobra"

// NewCmd returns the "index" parent command with its subcommands registered.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "index",
		Short: "Manage the tool index",
		Long:  "Commands for managing the remote tool index that defines available tools.",
	}

	cmd.AddCommand(
		newUpdateCmd(),
		newListCmd(),
		newSearchCmd(),
	)

	return cmd
}
