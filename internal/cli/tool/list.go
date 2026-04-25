package tool

import (
	"context"
	"fmt"

	"github.com/alecerf/devstrap/internal/cli/ui"
	"github.com/alecerf/devstrap/internal/engine"
	"github.com/spf13/cobra"
)

func newListCmd(baseDir *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available tools and their installed versions",
		Run: func(_ *cobra.Command, _ []string) {
			runList(*baseDir)
		},
	}
}

func runList(baseDir string) {
	plat := engine.DetectPlatform()

	order, all, err := buildTools(baseDir, plat)
	if err != nil {
		fmt.Printf("  %s %v\n", ui.RedBold("✖"), err)

		return
	}

	p := ui.NewPrinter(order)
	ctx := context.Background()

	for _, name := range order {
		ver, err := all[name].CurrentVersion(ctx)
		if err != nil {
			fmt.Printf("  %s  %s\n", ui.Bold(p.Pad(name)), ui.Dim("not installed"))
		} else {
			fmt.Printf("  %s  %s\n", ui.Bold(p.Pad(name)), ui.GreenBold(ver))
		}
	}
}
