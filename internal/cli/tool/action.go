package tool

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/alecerf/devstrap/internal/cli/ui"
	"github.com/alecerf/devstrap/internal/registry"
)

type runCounts struct {
	acted    int
	upToDate int
	failed   int
}

// runAction is the shared implementation for install and upgrade commands.
func runAction(paths registry.Paths, args []string, version, label string) error {
	version = strings.TrimPrefix(version, "v")

	if version != "" && len(args) != 1 {
		return errVersionRequiresSingleTool
	}

	order, all, err := resolveTools(paths, args)
	if err != nil {
		return err
	}

	ctx := context.Background()
	printer := ui.NewPrinter(order)
	start := time.Now()

	counts := runTools(ctx, order, all, version, printer)

	printer.PrintSummary(
		len(order),
		label,
		counts.acted,
		counts.upToDate,
		counts.failed,
		time.Since(start),
	)

	if counts.failed > 0 {
		return errToolsFailed
	}

	return nil
}

func runTools(
	ctx context.Context,
	names []string,
	tools map[string]registry.Tool,
	version string,
	printer ui.Printer,
) runCounts {
	var counts runCounts

	for i, name := range names {
		prefix := printer.ProgressPrefix(i, len(names), name)
		spinner := ui.NewSpinner(prefix + "  checking...")

		status := func(msg string) {
			spinner.Update(prefix + "  " + msg)
		}

		var res registry.Result
		if version != "" {
			res = registry.RunVersion(ctx, tools[name], status, version)
		} else {
			res = registry.Run(ctx, tools[name], status)
		}

		spinner.Stop()

		switch {
		case res.Err != nil:
			printer.PrintError(name, res.Err)

			counts.failed++
		case res.Status == "up-to-date":
			printer.PrintSuccess(name, fmt.Sprintf("up-to-date (%s)", res.Version))

			counts.upToDate++
		default:
			printer.PrintSuccess(name, "installed "+res.Version)

			counts.acted++
		}
	}

	return counts
}
