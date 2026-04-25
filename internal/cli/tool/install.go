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
	var version string

	cmd := &cobra.Command{
		Use:   "install <tools...>",
		Short: "Install one or more development tools",
		Long: "Install downloads and sets up the specified tools.\n\nAvailable tools: " +
			strings.Join(toolNames(), ", "),
		Args:      cobra.MinimumNArgs(1),
		ValidArgs: toolNames(),
		RunE: func(_ *cobra.Command, args []string) error {
			return runInstall(*baseDir, args, version)
		},
	}

	cmd.Flags().StringVar(&version, "version", "", "install a specific version instead of the latest")

	return cmd
}

func runInstall(baseDir string, args []string, version string) error {
	version = strings.TrimPrefix(version, "v")

	if version != "" && len(args) != 1 {
		return errVersionRequiresSingleTool
	}

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

	c := runTools(ctx, args, all, version, p)

	p.PrintSummary(len(args), "installed", c.acted, c.upToDate, c.failed, time.Since(start))

	if c.failed > 0 {
		return errToolsFailed
	}

	return nil
}
