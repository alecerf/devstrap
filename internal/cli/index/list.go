package index

import (
	"fmt"

	"github.com/alecerf/devstrap/internal/cli/ui"
	"github.com/alecerf/devstrap/internal/registry"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all tools available in the index",
		RunE:  runList,
	}
}

func runList(_ *cobra.Command, _ []string) error {
	idx, err := registry.Load()
	if err != nil {
		return fmt.Errorf("load index: %w", err)
	}

	if len(idx.Definitions) == 0 {
		fmt.Println("  No tools in the index.")

		return nil
	}

	names := make([]string, len(idx.Definitions))
	for i, def := range idx.Definitions {
		names[i] = def.Name
	}

	p := ui.NewPrinter(names)

	for _, def := range idx.Definitions {
		desc := def.Description
		if desc == "" {
			desc = ui.Dim("no description")
		}

		fmt.Printf("  %s  %s\n", ui.Bold(p.Pad(def.Name)), desc)
	}

	return nil
}
