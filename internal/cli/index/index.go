// Package index implements the "devstrap index" subcommands.
package index

import (
	"fmt"
	"os"

	"github.com/alecerf/devstrap/internal/cli/ui"
	"github.com/alecerf/devstrap/internal/registry"
	"github.com/spf13/cobra"
)

// NewCmd returns the "index" parent command with its subcommands.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "index",
		Short: "Manage the tool index",
	}

	cmd.AddCommand(
		newUpdateCmd(),
		newListCmd(),
		newSearchCmd(),
	)

	return cmd
}

func printDefinitions(defs []registry.Definition) {
	names := make([]string, len(defs))
	for i, def := range defs {
		names[i] = def.Name
	}

	printer := ui.NewPrinter(names)

	for _, def := range defs {
		desc := def.Description
		if desc == "" {
			desc = ui.Dim("no description")
		}

		_, _ = fmt.Fprintf(os.Stdout, "  %s  %s\n", ui.Bold(printer.Pad(def.Name)), desc)
	}
}
