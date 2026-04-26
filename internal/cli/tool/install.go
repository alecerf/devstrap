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

func newInstallCmd(paths *registry.Paths) *cobra.Command {
	var (
		version string
		all     bool
		dryRun  bool
	)

	long := "Install downloads and sets up the specified tools.\n\n" +
		"Use --dry-run to check for available upgrades without installing.\n\n" +
		"Available tools: " + strings.Join(toolNames(), ", ")

	cmd := &cobra.Command{
		Use:       "install <tools... | --all>",
		Short:     "Install or upgrade development tools",
		Long:      long,
		ValidArgs: toolNames(),
		RunE: func(_ *cobra.Command, args []string) error {
			if all && version != "" {
				return errAllAndVersionConflict
			}

			if dryRun && version != "" {
				return errDryRunAndVersionConflict
			}

			resolved, err := validateAllFlag(all, args)
			if err != nil {
				return err
			}

			if dryRun {
				return runDryRun(*paths, resolved)
			}

			return runAction(*paths, resolved, version, "installed")
		},
	}

	cmd.Flags().
		StringVar(&version, "version", "", "install a specific version instead of the latest")
	cmd.Flags().BoolVar(&all, "all", false, "install all tools from the index")
	cmd.Flags().
		BoolVar(&dryRun, "dry-run", false, "check for available upgrades without installing")

	return cmd
}

func runDryRun(paths registry.Paths, args []string) error {
	order, all, err := resolveTools(paths, args)
	if err != nil {
		return err
	}

	ctx := context.Background()
	printer := ui.NewPrinter(order)
	start := time.Now()

	var upToDate, updatable, failed int

	for i, name := range order {
		prefix := printer.ProgressPrefix(i, len(order), name)
		spinner := ui.NewSpinner(prefix + "  checking...")

		status := func(msg string) {
			spinner.Update(prefix + "  " + msg)
		}

		info := registry.Check(ctx, all[name], status)

		spinner.Stop()

		switch {
		case info.Err != nil:
			printer.PrintError(name, info.Err)

			failed++
		case info.UpToDate:
			printer.PrintSuccess(name, fmt.Sprintf("up-to-date (%s)", info.Current))

			upToDate++
		case info.Current == "":
			printer.PrintInfo(name, fmt.Sprintf("not installed → %s available", info.Latest))

			updatable++
		default:
			printer.PrintInfo(name, fmt.Sprintf("%s → %s available", info.Current, info.Latest))

			updatable++
		}
	}

	printer.PrintSummary(len(order), "to upgrade", updatable, upToDate, failed, time.Since(start))

	return nil
}
