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
		all    bool
		dryRun bool
	)

	long := "Install downloads and sets up the specified tools.\n\n" +
		"Pin a specific version with tool@version (e.g. go@1.22.0).\n" +
		"Use --dry-run to check for available upgrades without installing."

	if names := toolNames(); len(names) > 0 {
		long += "\n\nAvailable tools: " + strings.Join(names, ", ")
	} else {
		long += "\n\nRun 'devstrap index update' to see available tools."
	}

	cmd := &cobra.Command{
		Use:       "install <tools[@version]... | --all>",
		Short:     "Install or upgrade development tools",
		Long:      long,
		ValidArgs: toolNames(),
		RunE: func(_ *cobra.Command, args []string) error {
			names, versions := parseArgs(args)

			if all && hasVersionedArg(versions) {
				return errAllAndVersionConflict
			}

			resolved, err := validateAllFlag(all, names)
			if err != nil {
				return err
			}

			if dryRun {
				return runDryRun(*paths, resolved)
			}

			return runInstall(*paths, resolved, versions)
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "install all tools from the index")
	cmd.Flags().
		BoolVar(&dryRun, "dry-run", false, "check for available upgrades without installing")

	return cmd
}

type toolResult struct {
	acted    int
	upToDate int
	failed   int
}

// iterateTools resolves tools and runs process on each with a spinner.
func iterateTools(
	paths registry.Paths,
	args []string,
	process func(ctx context.Context, name string, tool registry.Tool, status func(string)) (string, string, error),
) (toolResult, error) {
	order, all, err := resolveTools(paths, args)
	if err != nil {
		return toolResult{}, err
	}

	ctx := context.Background()
	printer := ui.NewPrinter(order)
	start := time.Now()

	var counts toolResult

	for i, name := range order {
		prefix := printer.ProgressPrefix(i, len(order), name)
		spinner := ui.NewSpinner(prefix + "  checking...")

		status := func(msg string) {
			spinner.Update(prefix + "  " + msg)
		}

		outcome, msg, fnErr := process(ctx, name, all[name], status)

		spinner.Stop()

		switch outcome {
		case "error":
			printer.PrintError(name, fnErr)

			counts.failed++
		case "up-to-date":
			printer.PrintSuccess(name, msg)

			counts.upToDate++
		case "info":
			printer.PrintInfo(name, msg)

			counts.acted++
		default:
			printer.PrintSuccess(name, msg)

			counts.acted++
		}
	}

	printer.PrintSummary(
		len(order), "installed", counts.acted, counts.upToDate, counts.failed, time.Since(start),
	)

	return counts, nil
}

func runInstall(paths registry.Paths, args []string, versions map[string]string) error {
	counts, err := iterateTools(
		paths,
		args,
		func(ctx context.Context, name string, tool registry.Tool, status func(string)) (string, string, error) {
			var res registry.Result

			if v := versions[name]; v != "" {
				res = registry.RunVersion(ctx, tool, status, v)
			} else {
				res = registry.Run(ctx, tool, status)
			}

			if res.Err != nil {
				return "error", "", res.Err
			}

			if res.Status == "up-to-date" {
				return "up-to-date", fmt.Sprintf("up-to-date (%s)", res.Version), nil
			}

			return "success", "installed " + res.Version, nil
		},
	)
	if err != nil {
		return err
	}

	if counts.failed > 0 {
		return errToolsFailed
	}

	return nil
}

func runDryRun(paths registry.Paths, args []string) error {
	counts, err := iterateTools(
		paths,
		args,
		func(ctx context.Context, _ string, tool registry.Tool, status func(string)) (string, string, error) {
			info := registry.Check(ctx, tool, status)

			if info.Err != nil {
				return "error", "", info.Err
			}

			if info.UpToDate {
				return "up-to-date", fmt.Sprintf("up-to-date (%s)", info.Current), nil
			}

			if info.Current == "" {
				return "info", fmt.Sprintf("not installed → %s available", info.Latest), nil
			}

			return "info", fmt.Sprintf("%s → %s available", info.Current, info.Latest), nil
		},
	)
	if err != nil {
		return err
	}

	if counts.failed > 0 {
		return errToolsFailed
	}

	return nil
}
