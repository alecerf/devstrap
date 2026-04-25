package index

import (
	"fmt"
	"strings"

	"github.com/alecerf/devstrap/internal/cli/ui"
	"github.com/alecerf/devstrap/internal/registry"
	"github.com/spf13/cobra"
)

func newSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Search tools by name or description",
		Args:  cobra.ExactArgs(1),
		RunE:  runSearch,
	}
}

func runSearch(_ *cobra.Command, args []string) error {
	idx, err := registry.Load()
	if err != nil {
		return fmt.Errorf("load index: %w", err)
	}

	query := strings.ToLower(args[0])
	var matches []registry.Definition

	for _, def := range idx.Definitions {
		if strings.Contains(strings.ToLower(def.Name), query) ||
			strings.Contains(strings.ToLower(def.Description), query) {
			matches = append(matches, def)
		}
	}

	if len(matches) == 0 {
		fmt.Printf("  No tools matching %q\n", args[0])

		return nil
	}

	names := make([]string, len(matches))
	for i, def := range matches {
		names[i] = def.Name
	}

	p := ui.NewPrinter(names)

	for _, def := range matches {
		desc := def.Description
		if desc == "" {
			desc = ui.Dim("no description")
		}

		fmt.Printf("  %s  %s\n", ui.Bold(p.Pad(def.Name)), desc)
	}

	return nil
}
