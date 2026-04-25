package tool

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/alecerf/devstrap/internal/cli/ui"
	"github.com/alecerf/devstrap/internal/registry"
	"github.com/spf13/cobra"
)

func newUpdateCmd(paths *registry.Paths) *cobra.Command {
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
			return runUpdate(*paths, args)
		},
	}
}

func runUpdate(paths registry.Paths, args []string) error {
	order, all, err := resolveTools(paths, args)
	if err != nil {
		return err
	}

	ctx := context.Background()
	p := ui.NewPrinter(order)
	start := time.Now()

	var upToDate, updatable, failed int

	for i, name := range order {
		prefix := p.ProgressPrefix(i, len(order), name)
		sp := ui.NewSpinner(prefix + "  checking...")

		status := func(msg string) {
			sp.Update(prefix + "  " + msg)
		}

		info := registry.Check(ctx, all[name], status)
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
