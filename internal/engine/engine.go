// Package engine defines the Tool interface and orchestration for installers.
package engine

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"

	"golang.org/x/mod/semver"
)

// Platform holds the detected OS and architecture.
type Platform struct {
	OS   string
	Arch string
}

// DetectPlatform returns the current runtime OS and architecture.
func DetectPlatform() Platform {
	return Platform{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
}

// Result represents the outcome of a single tool operation.
type Result struct {
	Name    string
	Status  string // "up-to-date", "installed"
	Version string
	Err     error
}

// CheckInfo holds version information gathered during a scan.
type CheckInfo struct {
	Name     string
	Current  string // empty if not installed
	Latest   string
	UpToDate bool
	Err      error
}

// Tool is the interface every installer implements.
type Tool interface {
	Name() string
	FetchLatest(ctx context.Context) (version, extra string, err error)
	FetchVersion(ctx context.Context, version string) (extra string, err error)
	CurrentVersion(ctx context.Context) (string, error)
	Install(ctx context.Context, status func(string), version, extra string) error
}

// Check queries the current and latest versions without installing.
func Check(ctx context.Context, t Tool, status func(string)) CheckInfo {
	status("checking for updates...")

	latest, _, err := t.FetchLatest(ctx)
	if err != nil {
		return CheckInfo{Name: t.Name(), Err: err}
	}

	current, err := t.CurrentVersion(ctx)
	if err != nil {
		return CheckInfo{Name: t.Name(), Latest: latest}
	}

	return CheckInfo{
		Name:     t.Name(),
		Current:  current,
		Latest:   latest,
		UpToDate: !isNewer(latest, current),
	}
}

// Run checks for updates and installs if needed.
func Run(ctx context.Context, t Tool, status func(string)) Result {
	status("checking for updates...")

	latest, extra, err := t.FetchLatest(ctx)
	if err != nil {
		return Result{Name: t.Name(), Err: err}
	}

	current, err := t.CurrentVersion(ctx)
	if err == nil && !isNewer(latest, current) {
		return Result{Name: t.Name(), Status: "up-to-date", Version: current}
	}

	status(fmt.Sprintf("installing %s...", latest))

	err = t.Install(ctx, status, latest, extra)
	if err != nil {
		return Result{Name: t.Name(), Err: err}
	}

	return Result{Name: t.Name(), Status: "installed", Version: latest}
}

// RunVersion installs the exact requested version.
// It always performs the installation because version detection may be
// unreliable (e.g. Go's GOTOOLCHAIN auto-forwarding reports a different
// version than the one actually installed on disk).
func RunVersion(ctx context.Context, t Tool, status func(string), version string) Result {
	status(fmt.Sprintf("fetching metadata for %s...", version))

	extra, err := t.FetchVersion(ctx, version)
	if err != nil {
		return Result{Name: t.Name(), Err: err}
	}

	status(fmt.Sprintf("installing %s...", version))

	err = t.Install(ctx, status, version, extra)
	if err != nil {
		return Result{Name: t.Name(), Err: err}
	}

	return Result{Name: t.Name(), Status: "installed", Version: version}
}

// isNewer returns true if latest is a newer semantic version than current.
func isNewer(latest, current string) bool {
	l := "v" + latest

	c := "v" + current
	if !semver.IsValid(l) || !semver.IsValid(c) {
		return latest != current
	}

	return semver.Compare(l, c) > 0
}

// RunVersionCmd runs a binary with the given args and parses the version from output.
func RunVersionCmd(ctx context.Context, bin string, args []string, parse func(string) (string, error)) (string, error) {
	out, err := exec.CommandContext(ctx, bin, args...).Output()
	if err != nil {
		return "", fmt.Errorf("run %s: %w", bin, err)
	}

	return parse(string(out))
}
