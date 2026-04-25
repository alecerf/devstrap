package tool

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/alecerf/devstrap/internal/cli/ui"
	"github.com/alecerf/devstrap/internal/engine"
	"github.com/spf13/cobra"
)

func newUpgradeCmd(baseDir *string) *cobra.Command {
	long := "Upgrade installs or upgrades the specified tools" +
		" (or all tools if none given).\n\nAvailable tools: " +
		strings.Join(toolNames(), ", ")

	return &cobra.Command{
		Use:       "upgrade [tools...]",
		Short:     "Upgrade development tools to their latest versions",
		Long:      long,
		ValidArgs: toolNames(),
		RunE: func(_ *cobra.Command, args []string) error {
			return runUpgrade(*baseDir, args)
		},
	}
}

func runUpgrade(baseDir string, args []string) error {
	plat := engine.DetectPlatform()

	order, all, err := buildTools(baseDir, plat)
	if err != nil {
		return err
	}

	if len(args) > 0 {
		for _, name := range args {
			if _, ok := all[name]; !ok {
				return fmt.Errorf("%w: %q (valid: %s)", errUnknownTool, name, strings.Join(order, ", "))
			}
		}

		order = args
	}

	ctx := context.Background()
	p := ui.NewPrinter(order)
	start := time.Now()

	var upgraded, upToDate, failed int

	for i, name := range order {
		prefix := fmt.Sprintf("%s %s",
			ui.Dim(fmt.Sprintf("[%d/%d]", i+1, len(order))),
			ui.Bold(p.Pad(name)),
		)
		sp := ui.NewSpinner(prefix + "  checking...")

		status := func(msg string) {
			sp.Update(prefix + "  " + msg)
		}

		res := engine.Run(ctx, all[name], status)
		sp.Stop()

		switch {
		case res.Err != nil:
			p.PrintError(name, res.Err)
			failed++
		case res.Status == "up-to-date":
			p.PrintSuccess(name, fmt.Sprintf("up-to-date (%s)", res.Version))
			upToDate++
		default:
			p.PrintSuccess(name, "installed "+res.Version)
			upgraded++
		}
	}

	p.PrintSummary(len(order), "upgraded", upgraded, upToDate, failed, time.Since(start))

	if failed > 0 {
		return errToolsFailed
	}

	return nil
}
