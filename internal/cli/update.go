package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/alecerf/devstrap/internal/cli/ui"
	"github.com/alecerf/devstrap/internal/updater"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

func newUpdateCmd(currentVersion string) *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update devstrap to the latest version",
		Long: `Checks GitHub for the latest devstrap release and installs it if a newer
version is available. The running binary is replaced atomically.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runUpdate(currentVersion)
		},
	}
}

func runUpdate(currentVersion string) error {
	ctx := context.Background()

	spinner := ui.NewSpinner("  checking for updates...")

	latest, err := updater.FetchLatest(ctx)

	spinner.Stop()

	if err != nil {
		return fmt.Errorf("fetch latest version: %w", err)
	}

	current := strings.TrimPrefix(currentVersion, "v")

	if !isNewerVersion(latest, current) {
		_, _ = fmt.Fprintf(os.Stdout, "  %s devstrap  already up-to-date (%s)\n",
			ui.GreenBold("✔"), current)

		return nil
	}

	start := time.Now()

	spinner = ui.NewSpinner(fmt.Sprintf("  updating %s → %s...", current, latest))

	status := func(msg string) {
		spinner.Update("  " + msg)
	}

	status(fmt.Sprintf("downloading %s...", latest))

	err = updater.Update(ctx, latest)

	spinner.Stop()

	if err != nil {
		_, _ = fmt.Fprintf(os.Stdout, "  %s devstrap  %v\n", ui.RedBold("✖"), err)

		return fmt.Errorf("update devstrap: %w", err)
	}

	_, _ = fmt.Fprintf(os.Stdout, "  %s devstrap  updated to %s %s\n",
		ui.GreenBold("✔"),
		latest,
		ui.Dim("("+ui.FormatDuration(time.Since(start))+")"),
	)

	return nil
}

// isNewerVersion returns true when latest is strictly newer than current.
func isNewerVersion(latest, current string) bool {
	latestV := "v" + latest
	currentV := "v" + current

	if !semver.IsValid(latestV) || !semver.IsValid(currentV) {
		return latest != current
	}

	return semver.Compare(latestV, currentV) > 0
}
