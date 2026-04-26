package tool

import (
	"context"
	"fmt"
	"os"

	"github.com/alecerf/devstrap/internal/cli/ui"
	"github.com/alecerf/devstrap/internal/registry"
	"github.com/spf13/cobra"
)

func newListCmd(paths *registry.Paths) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available tools and their installed versions",
		Run: func(_ *cobra.Command, _ []string) {
			runList(*paths)
		},
	}
}

func runList(paths registry.Paths) {
	order, all, err := resolveTools(paths, nil)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stdout, "  %s %v\n", ui.RedBold("✖"), err)

		return
	}

	printer := ui.NewPrinter(order)
	ctx := context.Background()

	for _, name := range order {
		ver, err := all[name].CurrentVersion(ctx)
		if err != nil {
			_, _ = fmt.Fprintf(
				os.Stdout,
				"  %s  %s\n",
				ui.Bold(printer.Pad(name)),
				ui.Dim("not installed"),
			)
		} else {
			_, _ = fmt.Fprintf(
				os.Stdout,
				"  %s  %s\n",
				ui.Bold(printer.Pad(name)),
				ui.GreenBold(ver),
			)
		}
	}
}
