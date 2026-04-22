package cli

import (
	"context"
	"fmt"

	"github.com/alecerf/devstrap/internal/tool"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available tools and their installed versions",
		Run:   runList,
	}
}

func runList(_ *cobra.Command, _ []string) {
	plat := tool.DetectPlatform()
	order, all := BuildTools(baseDir, plat)
	p := newPrinter(order)
	ctx := context.Background()

	for _, name := range order {
		ver, err := all[name].CurrentVersion(ctx)
		if err != nil {
			fmt.Printf("  %s  %s\n", bold(p.pad(name)), dim("not installed"))
		} else {
			fmt.Printf("  %s  %s\n", bold(p.pad(name)), greenBold(ver))
		}
	}
}
