package tool

import (
	"context"
	"fmt"

	"github.com/alecerf/devstrap/internal/cli/ui"
	"github.com/alecerf/devstrap/internal/engine"
)

type runCounts struct {
	acted    int
	upToDate int
	failed   int
}

func runTools(
	ctx context.Context,
	names []string,
	tools map[string]engine.Tool,
	version string,
	p ui.Printer,
) runCounts {
	var c runCounts

	for i, name := range names {
		prefix := fmt.Sprintf("%s %s",
			ui.Dim(fmt.Sprintf("[%d/%d]", i+1, len(names))),
			ui.Bold(p.Pad(name)),
		)
		sp := ui.NewSpinner(prefix + "  checking...")

		status := func(msg string) {
			sp.Update(prefix + "  " + msg)
		}

		var res engine.Result
		if version != "" {
			res = engine.RunVersion(ctx, tools[name], status, version)
		} else {
			res = engine.Run(ctx, tools[name], status)
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
