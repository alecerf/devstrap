// Package updater implements self-update logic for the devstrap binary.
// It queries GitHub Releases for the latest version, downloads and verifies
// the release archive, and replaces the running binary atomically.
package updater

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/alecerf/devstrap/internal/downloader"
)

const (
	githubAPIURL = "https://api.github.com/repos/alecerf/devstrap/releases/latest"
	githubDLBase = "https://github.com/alecerf/devstrap/releases/download"
	checksumFile = "checksums.txt"
	binaryName   = "devstrap"
	execPerm     = 0o750
)

var supportedPlatforms = map[string]map[string]bool{
	"darwin": {"amd64": true, "arm64": true},
	"linux":  {"amd64": true, "arm64": true},
}

type githubRelease struct {
	TagName string `json:"tag_name"`
}

// FetchLatest queries the GitHub API for the latest devstrap release and
// returns the version string without the "v" prefix.
func FetchLatest(ctx context.Context) (string, error) {
	var rel githubRelease

	err := downloader.FetchJSON(ctx, githubAPIURL, &rel)
	if err != nil {
		return "", fmt.Errorf("fetch latest release: %w", err)
	}

	if rel.TagName == "" {
		return "", fmt.Errorf("fetch latest release: %w", errEmptyTag)
	}

	version := rel.TagName
	if len(version) > 0 && version[0] == 'v' {
		version = version[1:]
	}

	return version, nil
}

// Update downloads the devstrap release for the given version, verifies its
// checksum, and replaces the running binary atomically.
func Update(ctx context.Context, version string) error {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	if archs, ok := supportedPlatforms[goos]; !ok || !archs[goarch] {
		return fmt.Errorf("%w: %s/%s", ErrUnsupportedPlatform, goos, goarch)
	}

	tmp, err := os.MkdirTemp("", "devstrap-update-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}

	defer func() { _ = os.RemoveAll(tmp) }()

	newBin, err := downloadAndVerify(ctx, tmp, version, goos, goarch)
	if err != nil {
		return err
	}

	return replaceCurrentBinary(newBin)
}

func downloadAndVerify(ctx context.Context, tmp, version, goos, goarch string) (string, error) {
	archive := fmt.Sprintf("%s_%s_%s_%s.tar.gz", binaryName, version, goos, goarch)
	tag := "v" + version
	archiveURL := fmt.Sprintf("%s/%s/%s", githubDLBase, tag, archive)
	checksumURL := fmt.Sprintf("%s/%s/%s", githubDLBase, tag, checksumFile)

	archivePath := filepath.Join(tmp, archive)

	err := downloader.Download(ctx, archiveURL, archivePath)
	if err != nil {
		return "", fmt.Errorf("download archive: %w", err)
	}

	checksumPath := filepath.Join(tmp, checksumFile)

	err = downloader.Download(ctx, checksumURL, checksumPath)
	if err != nil {
		return "", fmt.Errorf("download checksum: %w", err)
	}

	expectedSHA, err := downloader.ExtractChecksum(checksumPath, archive)
	if err != nil {
		return "", fmt.Errorf("extract checksum: %w", err)
	}

	err = downloader.VerifyChecksum(archivePath, expectedSHA)
	if err != nil {
		return "", fmt.Errorf("verify checksum: %w", err)
	}

	err = downloader.ExtractTarGz(ctx, archivePath, tmp, 0)
	if err != nil {
		return "", fmt.Errorf("extract archive: %w", err)
	}

	newBin := filepath.Join(tmp, binaryName)

	err = os.Chmod(newBin, execPerm)
	if err != nil {
		return "", fmt.Errorf("chmod new binary: %w", err)
	}

	return newBin, nil
}

func replaceCurrentBinary(newBin string) error {
	currentBin, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve current binary: %w", err)
	}

	currentBin, err = filepath.EvalSymlinks(currentBin)
	if err != nil {
		return fmt.Errorf("resolve symlinks: %w", err)
	}

	// Stage the new binary in the same directory so os.Rename is atomic.
	staged, err := os.CreateTemp(filepath.Dir(currentBin), ".devstrap-update-*")
	if err != nil {
		return fmt.Errorf("create staged file: %w", err)
	}

	stagedName := staged.Name()

	defer func() { _ = os.Remove(stagedName) }()

	err = stageFile(newBin, staged)
	if err != nil {
		return fmt.Errorf("stage new binary: %w", err)
	}

	err = os.Rename(stagedName, currentBin)
	if err != nil {
		return fmt.Errorf("replace binary: %w", err)
	}

	return nil
}

func stageFile(src string, dst *os.File) error {
	defer func() { _ = dst.Close() }()

	data, err := os.ReadFile(filepath.Clean(src))
	if err != nil {
		return fmt.Errorf("read %s: %w", src, err)
	}

	_, err = dst.Write(data)
	if err != nil {
		return fmt.Errorf("write staged file: %w", err)
	}

	err = dst.Chmod(execPerm)
	if err != nil {
		return fmt.Errorf("chmod staged file: %w", err)
	}

	return nil
}
