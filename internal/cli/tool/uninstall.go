package tool

import (
	"context"
	"fmt"
	"os"

	"github.com/alecerf/devstrap/internal/cli/ui"
	"github.com/alecerf/devstrap/internal/registry"
	"github.com/spf13/cobra"
)

func newUninstallCmd(paths *registry.Paths) *cobra.Command {
	cmd := &cobra.Command{
		Use:       "uninstall <tools...>",
		Short:     "Remove installed tools",
		Long:      "Uninstall removes the specified tools from the installation directories.",
		ValidArgs: toolNames(),
		Args:      cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runUninstall(*paths, args)
		},
	}

	return cmd
}

func runUninstall(paths registry.Paths, args []string) error {
	order, all, err := resolveTools(paths, args)
	if err != nil {
		return err
	}

	ctx := context.Background()
	printer := ui.NewPrinter(order)

	var removed, notInstalled, failed int

	for _, name := range order {
		tool := all[name]

		_, verErr := tool.CurrentVersion(ctx)
		if verErr != nil {
			printer.PrintInfo(name, "not installed")

			notInstalled++

			continue
		}

		err = tool.Uninstall()
		if err != nil {
			printer.PrintError(name, err)

			failed++

			continue
		}

		printer.PrintSuccess(name, "removed")

		removed++
	}

	_, _ = fmt.Fprintf(os.Stdout,
		"\n%d checked, %d removed",
		len(order), removed)

	if notInstalled > 0 {
		_, _ = fmt.Fprintf(os.Stdout, ", %d not installed", notInstalled)
	}

	if failed > 0 {
		_, _ = fmt.Fprintf(os.Stdout, ", %d failed", failed)
	}

	_, _ = fmt.Fprintln(os.Stdout)

	if failed > 0 {
		return errToolsFailed
	}

	return nil
}
