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

func newInstallCmd(baseDir *string) *cobra.Command {
	return &cobra.Command{
		Use:   "install <tools...>",
		Short: "Install one or more development tools",
		Long: "Install downloads and sets up the specified tools.\n\nAvailable tools: " +
			strings.Join(toolNames(), ", "),
		Args:      cobra.MinimumNArgs(1),
		ValidArgs: toolNames(),
		RunE: func(_ *cobra.Command, args []string) error {
			return runInstall(*baseDir, args)
		},
	}
}

func runInstall(baseDir string, args []string) error {
	plat := engine.DetectPlatform()

	order, all, err := buildTools(baseDir, plat)
	if err != nil {
		return err
	}

	for _, name := range args {
		if _, ok := all[name]; !ok {
			return fmt.Errorf("%w: %q (valid: %s)", errUnknownTool, name, strings.Join(order, ", "))
		}
	}

	ctx := context.Background()
	p := ui.NewPrinter(args)
	start := time.Now()

	var installed, upToDate, failed int

	for i, name := range args {
		prefix := fmt.Sprintf("%s %s",
			ui.Dim(fmt.Sprintf("[%d/%d]", i+1, len(args))),
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
			installed++
		}
	}

	p.PrintSummary(len(args), "installed", installed, upToDate, failed, time.Since(start))

	if failed > 0 {
		return errToolsFailed
	}

	return nil
}
