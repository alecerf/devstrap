package registry

import (
	"context"
	"fmt"
	"os/exec"

	"golang.org/x/mod/semver"
)

// Check queries the current and latest versions without installing.
func Check(ctx context.Context, tool Tool, status func(string)) CheckInfo {
	status("checking for updates...")

	latest, _, err := tool.FetchLatest(ctx)
	if err != nil {
		return CheckInfo{Name: tool.Name(), Err: err}
	}

	current, err := tool.CurrentVersion(ctx)
	if err != nil {
		return CheckInfo{Name: tool.Name(), Latest: latest}
	}

	return CheckInfo{
		Name:     tool.Name(),
		Current:  current,
		Latest:   latest,
		UpToDate: !isNewer(latest, current),
	}
}

// Run checks for updates and installs if a newer version is available.
func Run(ctx context.Context, tool Tool, status func(string)) Result {
	status("checking for updates...")

	latest, extra, err := tool.FetchLatest(ctx)
	if err != nil {
		return Result{Name: tool.Name(), Err: err}
	}

	current, err := tool.CurrentVersion(ctx)
	if err == nil && !isNewer(latest, current) {
		return Result{Name: tool.Name(), Status: "up-to-date", Version: current}
	}

	status(fmt.Sprintf("installing %s...", latest))

	err = tool.Install(ctx, status, latest, extra)
	if err != nil {
		return Result{Name: tool.Name(), Err: err}
	}

	return Result{Name: tool.Name(), Status: "installed", Version: latest}
}

// RunVersion installs the exact requested version unconditionally.
func RunVersion(ctx context.Context, tool Tool, status func(string), version string) Result {
	status(fmt.Sprintf("fetching metadata for %s...", version))

	extra, err := tool.FetchVersion(ctx, version)
	if err != nil {
		return Result{Name: tool.Name(), Err: err}
	}

	status(fmt.Sprintf("installing %s...", version))

	err = tool.Install(ctx, status, version, extra)
	if err != nil {
		return Result{Name: tool.Name(), Err: err}
	}

	return Result{Name: tool.Name(), Status: "installed", Version: version}
}

func isNewer(latest, current string) bool {
	latestV := "v" + latest

	c := "v" + current
	if !semver.IsValid(latestV) || !semver.IsValid(c) {
		return latest != current
	}

	return semver.Compare(latestV, c) > 0
}

// RunVersionCmd runs a binary and parses its version output.
func RunVersionCmd(
	ctx context.Context,
	bin string,
	args []string,
	parse func(string) (string, error),
) (string, error) {
	out, err := exec.CommandContext(ctx, bin, args...).Output()
	if err != nil {
		return "", fmt.Errorf("run %s: %w", bin, err)
	}

	return parse(string(out))
}
