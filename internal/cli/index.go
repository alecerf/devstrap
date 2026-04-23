package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/alecerf/devstrap/internal/index"
	"github.com/spf13/cobra"
)

func newIndexCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "index",
		Short: "Manage the tool index",
		Long:  "Commands for managing the remote tool index that defines available tools.",
	}

	cmd.AddCommand(
		newIndexUpdateCmd(),
		newIndexListCmd(),
		newIndexSearchCmd(),
	)

	return cmd
}

func newIndexUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Fetch the latest tool index from the remote repository",
		RunE:  runIndexUpdate,
	}
}

func runIndexUpdate(_ *cobra.Command, _ []string) error {
	ctx := context.Background()
	sp := newSpinner("  fetching index...")

	err := index.Update(ctx, index.DefaultRemoteURL)
	sp.stop()

	if err != nil {
		fmt.Printf("  %s %s\n", redBold("✖"), "failed to update index")

		return fmt.Errorf("update index: %w", err)
	}

	idx, err := index.Load()
	if err != nil {
		return fmt.Errorf("load updated index: %w", err)
	}

	fmt.Printf("  %s index updated (%d tools available)\n",
		greenBold("✔"), len(idx.Definitions))

	return nil
}

func newIndexListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all tools available in the index",
		RunE:  runIndexList,
	}
}

func runIndexList(_ *cobra.Command, _ []string) error {
	idx, err := index.Load()
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

	p := newPrinter(names)

	for _, def := range idx.Definitions {
		desc := def.Description
		if desc == "" {
			desc = dim("no description")
		}

		fmt.Printf("  %s  %s\n", bold(p.pad(def.Name)), desc)
	}

	return nil
}

func newIndexSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Search tools by name or description",
		Args:  cobra.ExactArgs(1),
		RunE:  runIndexSearch,
	}
}

func runIndexSearch(_ *cobra.Command, args []string) error {
	idx, err := index.Load()
	if err != nil {
		return fmt.Errorf("load index: %w", err)
	}

	query := strings.ToLower(args[0])
	var matches []index.Definition

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

	p := newPrinter(names)

	for _, def := range matches {
		desc := def.Description
		if desc == "" {
			desc = dim("no description")
		}

		fmt.Printf("  %s  %s\n", bold(p.pad(def.Name)), desc)
	}

	return nil
}
