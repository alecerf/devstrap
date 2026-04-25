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
	p := ui.NewPrinter(order)
	start := time.Now()

	c := runTools(ctx, order, all, version, p)

	p.PrintSummary(len(order), label, c.acted, c.upToDate, c.failed, time.Since(start))

	if c.failed > 0 {
		return errToolsFailed
	}

	return nil
}

func runTools(
	ctx context.Context,
	names []string,
	tools map[string]registry.Tool,
	version string,
	p ui.Printer,
) runCounts {
	var c runCounts

	for i, name := range names {
		prefix := p.ProgressPrefix(i, len(names), name)
		sp := ui.NewSpinner(prefix + "  checking...")

		status := func(msg string) {
			sp.Update(prefix + "  " + msg)
		}

		var res registry.Result
		if version != "" {
			res = registry.RunVersion(ctx, tools[name], status, version)
		} else {
			res = registry.Run(ctx, tools[name], status)
		}

		sp.Stop()

		switch {
		case res.Err != nil:
			p.PrintError(name, res.Err)
			c.failed++
		case res.Status == "up-to-date":
			p.PrintSuccess(name, fmt.Sprintf("up-to-date (%s)", res.Version))
			c.upToDate++
		default:
			p.PrintSuccess(name, "installed "+res.Version)
			c.acted++
		}
	}

	return c
}
