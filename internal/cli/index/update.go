package index

import (
	"context"
	"fmt"
	"os"

	"github.com/alecerf/devstrap/internal/cli/ui"
	"github.com/alecerf/devstrap/internal/registry"
	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Fetch the latest tool index from the remote repository",
		RunE:  runUpdate,
	}
}

func runUpdate(_ *cobra.Command, _ []string) error {
	ctx := context.Background()
	sp := ui.NewSpinner("  fetching index...")

	err := registry.Update(ctx, registry.DefaultRemoteURL)

	sp.Stop()

	if err != nil {
		_, _ = fmt.Fprintf(os.Stdout, "  %s %s\n", ui.RedBold("✖"), "failed to update index")

		return fmt.Errorf("update index: %w", err)
	}

	idx, err := registry.Load()
	if err != nil {
		return fmt.Errorf("load updated index: %w", err)
	}

	_, _ = fmt.Fprintf(os.Stdout, "  %s index updated (%d tools available)\n",
		ui.GreenBold("✔"), len(idx.Definitions))

	return nil
}
