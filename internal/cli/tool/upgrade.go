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

func newUpgradeCmd(baseDir *string) *cobra.Command {
	var version string

	long := "Upgrade installs or upgrades the specified tools" +
		" (or all tools if none given).\n\nAvailable tools: " +
		strings.Join(toolNames(), ", ")

	cmd := &cobra.Command{
		Use:       "upgrade [tools...]",
		Short:     "Upgrade development tools to their latest versions",
		Long:      long,
		ValidArgs: toolNames(),
		RunE: func(_ *cobra.Command, args []string) error {
			return runUpgrade(*baseDir, args, version)
		},
	}

	cmd.Flags().StringVar(&version, "version", "", "upgrade to a specific version instead of the latest")

	return cmd
}

func runUpgrade(baseDir string, args []string, version string) error {
	version = strings.TrimPrefix(version, "v")

	if version != "" && len(args) != 1 {
		return errVersionRequiresSingleTool
	}

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

	c := runTools(ctx, order, all, version, p)

	p.PrintSummary(len(order), "upgraded", c.acted, c.upToDate, c.failed, time.Since(start))

	if c.failed > 0 {
		return errToolsFailed
	}

	return nil
}
