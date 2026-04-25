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

func newUpdateCmd(baseDir *string) *cobra.Command {
	long := "Update fetches the latest versions of the specified tools" +
		" (or all if none given)\nand reports whether upgrades are" +
		" available.\n\nAvailable tools: " +
		strings.Join(toolNames(), ", ")

	return &cobra.Command{
		Use:       "update [tools...]",
		Short:     "Fetch latest versions and report available upgrades",
		Long:      long,
		ValidArgs: toolNames(),
		RunE: func(_ *cobra.Command, args []string) error {
			return runUpdate(*baseDir, args)
		},
	}
}

func runUpdate(baseDir string, args []string) error {
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

	var upToDate, updatable, failed int

	for i, name := range order {
		prefix := fmt.Sprintf("%s %s",
			ui.Dim(fmt.Sprintf("[%d/%d]", i+1, len(order))),
			ui.Bold(p.Pad(name)),
		)
		sp := ui.NewSpinner(prefix + "  checking...")

		status := func(msg string) {
			sp.Update(prefix + "  " + msg)
		}

		info := engine.Check(ctx, all[name], status)
		sp.Stop()

		switch {
		case info.Err != nil:
			p.PrintError(name, info.Err)
			failed++
		case info.UpToDate:
			p.PrintSuccess(name, fmt.Sprintf("up-to-date (%s)", info.Current))
			upToDate++
		case info.Current == "":
			p.PrintInfo(name, fmt.Sprintf("not installed → %s available", info.Latest))
			updatable++
		default:
			p.PrintInfo(name, fmt.Sprintf("%s → %s available", info.Current, info.Latest))
			updatable++
		}
	}

	p.PrintSummary(len(order), "to upgrade", updatable, upToDate, failed, time.Since(start))

	return nil
}
