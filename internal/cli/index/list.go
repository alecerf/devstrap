package index

import (
	"fmt"

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

	printDefinitions(idx.Definitions)

	return nil
}
