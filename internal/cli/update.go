package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/alecerf/devstrap/internal/tool"
	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	long := "Update fetches the latest versions of the specified tools" +
		" (or all if none given)\nand reports whether upgrades are" +
		" available.\n\nAvailable tools: " +
		strings.Join(ToolNames(), ", ")

	return &cobra.Command{
		Use:       "update [tools...]",
		Short:     "Fetch latest versions and report available upgrades",
		Long:      long,
		ValidArgs: ToolNames(),
		RunE:      runUpdate,
	}
}

func runUpdate(_ *cobra.Command, args []string) error {
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

	var upToDate, updatable, failed int

	for i, name := range order {
		prefix := fmt.Sprintf("%s %s",
			dim(fmt.Sprintf("[%d/%d]", i+1, len(order))),
			bold(p.pad(name)),
		)
		sp := newSpinner(prefix + "  checking...")

		status := func(msg string) {
			sp.update(prefix + "  " + msg)
		}

		info := tool.Check(ctx, all[name], status)
		sp.stop()

		switch {
		case info.Err != nil:
			p.printError(name, info.Err)
			failed++
		case info.UpToDate:
			p.printSuccess(name, fmt.Sprintf("up-to-date (%s)", info.Current))
			upToDate++
		case info.Current == "":
			p.printInfo(name, fmt.Sprintf("not installed → %s available", info.Latest))
			updatable++
		default:
			p.printInfo(name, fmt.Sprintf("%s → %s available", info.Current, info.Latest))
			updatable++
		}
	}

	p.printSummary(len(order), "to upgrade", updatable, upToDate, failed, time.Since(start))

	return nil
}
