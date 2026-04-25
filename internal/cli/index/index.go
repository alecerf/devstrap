// Package index implements the "devstrap index" command and its subcommands.
package index

import (
	"fmt"

	"github.com/alecerf/devstrap/internal/cli/ui"
	"github.com/alecerf/devstrap/internal/registry"
	"github.com/spf13/cobra"
)

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

// printDefinitions formats and prints a list of tool definitions.
func printDefinitions(defs []registry.Definition) {
	names := make([]string, len(defs))
	for i, def := range defs {
		names[i] = def.Name
	}

	p := ui.NewPrinter(names)

	for _, def := range defs {
		desc := def.Description
		if desc == "" {
			desc = ui.Dim("no description")
		}

		fmt.Printf("  %s  %s\n", ui.Bold(p.Pad(def.Name)), desc)
	}
}
