package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/alecerf/devstrap/internal/tool"
	"github.com/spf13/cobra"
)

func newUpgradeCmd() *cobra.Command {
	long := "Upgrade installs or upgrades the specified tools" +
		" (or all tools if none given).\n\nAvailable tools: " +
		strings.Join(ToolNames(), ", ")

	return &cobra.Command{
		Use:       "upgrade [tools...]",
		Short:     "Upgrade development tools to their latest versions",
		Long:      long,
		ValidArgs: ToolNames(),
		RunE:      runUpgrade,
	}
}

func runUpgrade(_ *cobra.Command, args []string) error {
	plat := tool.DetectPlatform()
	order, all := BuildTools(baseDir, plat)

	if len(args) > 0 {
		for _, name := range args {
			if _, ok := all[name]; !ok {
				return fmt.Errorf("%w: %q (valid: %s)", errUnknownTool, name, strings.Join(order, ", "))
			}
		}

		order = args
	}

	ctx := context.Background()
	p := newPrinter(order)
	start := time.Now()

	var upgraded, upToDate, failed int

	for i, name := range order {
		prefix := fmt.Sprintf("%s %s",
			dim(fmt.Sprintf("[%d/%d]", i+1, len(order))),
			bold(p.pad(name)),
		)
		sp := newSpinner(prefix + "  checking...")

		status := func(msg string) {
			sp.update(prefix + "  " + msg)
		}

		res := tool.Run(ctx, all[name], status)
		sp.stop()

		switch {
		case res.Err != nil:
			p.printError(name, res.Err)
			failed++
		case res.Status == "up-to-date":
			p.printSuccess(name, fmt.Sprintf("up-to-date (%s)", res.Version))
			upToDate++
		default:
			p.printSuccess(name, "installed "+res.Version)
			upgraded++
		}
	}

	p.printSummary(len(order), "upgraded", upgraded, upToDate, failed, time.Since(start))

	if failed > 0 {
		return errToolsFailed
	}

	return nil
}
