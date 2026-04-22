// Package golang implements a Tool that installs or updates the Go toolchain.
package golang

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecerf/devstrap/internal/downloader"
	"github.com/alecerf/devstrap/internal/tool"
)

var (
	errNoReleases   = errors.New("no releases found")
	errNoFile       = errors.New("no file found for platform")
	errVersionParse = errors.New("unexpected go version output")
)

// New returns a Tool that installs or updates the Go toolchain.
func New(baseDir string, plat tool.Platform) tool.Tool { //nolint:ireturn // factory function
	return &installer{
		baseDir:  baseDir,
		platform: plat,
		binPath:  filepath.Join(baseDir, "go", "bin", "go"),
	}
}

type installer struct {
	baseDir  string
	platform tool.Platform
	binPath  string
}

func (g *installer) Name() string { return "go" }

func (g *installer) CurrentVersion(ctx context.Context) (string, error) {
	v, err := tool.RunVersionCmd(ctx, g.binPath, []string{"version"}, parseVersion)
	if err != nil {
		return "", fmt.Errorf("get go version: %w", err)
	}

	return v, nil
}

type release struct {
	Version string `json:"version"`
	Files   []file `json:"files"`
}

type file struct {
	Filename string `json:"filename"`
	SHA256   string `json:"sha256"`
}

func (g *installer) FetchLatest(ctx context.Context) (string, string, error) {
	releases, err := downloader.FetchJSON[[]release](ctx, "https://go.dev/dl/?mode=json")
	if err != nil {
		return "", "", fmt.Errorf("fetch go releases: %w", err)
	}

	if len(releases) == 0 {
		return "", "", errNoReleases
	}

	latest := strings.TrimPrefix(releases[0].Version, "go")
	suffix := g.platform.FileSuffix() + ".tar.gz"

	for _, f := range releases[0].Files {
		if strings.HasSuffix(f.Filename, suffix) {
			return latest, f.SHA256, nil
		}
	}

	return "", "", fmt.Errorf("%w: %s", errNoFile, suffix)
}

func (g *installer) Install(ctx context.Context, status func(string), latest, expectedSHA string) error {
	tmp, err := os.MkdirTemp("", "devstrap-go-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}

	defer func() { _ = os.RemoveAll(tmp) }()

	filename := fmt.Sprintf("go%s.%s.tar.gz", latest, g.platform.FileSuffix())
	archive := filepath.Join(tmp, filename)

	err = downloader.Download(ctx, "https://go.dev/dl/"+filename, archive)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}

	status("verifying checksum...")

	err = downloader.VerifyChecksum(archive, expectedSHA)
	if err != nil {
		return fmt.Errorf("verify checksum: %w", err)
	}

	status("extracting...")

	err = downloader.ExtractTarGz(ctx, archive, tmp, 0)
	if err != nil {
		return fmt.Errorf("extract archive: %w", err)
	}

	dest := filepath.Join(g.baseDir, "go")
	_ = os.RemoveAll(dest)

	err = os.Rename(filepath.Join(tmp, "go"), dest)
	if err != nil {
		return fmt.Errorf("rename to %s: %w", dest, err)
	}

	return nil
}

func parseVersion(output string) (string, error) {
	// "go version go1.22.3 darwin/arm64"
	parts := strings.Fields(output)
	if len(parts) < 3 {
		return "", errVersionParse
	}

	return strings.TrimPrefix(parts[2], "go"), nil
}
