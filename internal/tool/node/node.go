// Package node implements a Tool that installs or updates Node.js.
package node

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

var errNoVersions = errors.New("no versions found")

// New returns a Tool that installs or updates Node.js.
func New(baseDir string, plat tool.Platform) tool.Tool { //nolint:ireturn // factory function
	return &installer{
		baseDir:  baseDir,
		platform: plat,
		binPath:  filepath.Join(baseDir, "node", "bin", "node"),
	}
}

type installer struct {
	baseDir  string
	platform tool.Platform
	binPath  string
}

func (n *installer) Name() string { return "node" }

func (n *installer) CurrentVersion(ctx context.Context) (string, error) {
	v, err := tool.RunVersionCmd(ctx, n.binPath, []string{"-v"}, parseVersion)
	if err != nil {
		return "", fmt.Errorf("get node version: %w", err)
	}

	return v, nil
}

type version struct {
	Version string `json:"version"`
}

func (n *installer) FetchLatest(ctx context.Context) (string, string, error) {
	versions, err := downloader.FetchJSON[[]version](ctx, "https://nodejs.org/dist/index.json")
	if err != nil {
		return "", "", fmt.Errorf("fetch node versions: %w", err)
	}

	if len(versions) == 0 {
		return "", "", errNoVersions
	}

	return strings.TrimPrefix(versions[0].Version, "v"), "", nil
}

func (n *installer) Install(ctx context.Context, status func(string), latest, _ string) error {
	tmp, err := os.MkdirTemp("", "devstrap-node-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}

	defer func() { _ = os.RemoveAll(tmp) }()

	filename := fmt.Sprintf("node-v%s-%s.tar.gz", latest, fileSuffix(n.platform))
	archive := filepath.Join(tmp, filename)
	baseURL := "https://nodejs.org/dist/v" + latest

	err = downloader.DownloadAndVerify(ctx, status, baseURL, filename, "SHASUMS256.txt", tmp, archive)
	if err != nil {
		return fmt.Errorf("download and verify: %w", err)
	}

	status("extracting...")

	installDir := filepath.Join(tmp, "node-install")

	err = os.MkdirAll(installDir, 0o750)
	if err != nil {
		return fmt.Errorf("create %s: %w", installDir, err)
	}

	err = downloader.ExtractTarGz(ctx, archive, installDir, 1)
	if err != nil {
		return fmt.Errorf("extract archive: %w", err)
	}

	dest := filepath.Join(n.baseDir, "node")
	_ = os.RemoveAll(dest)

	err = os.Rename(installDir, dest)
	if err != nil {
		return fmt.Errorf("rename to %s: %w", dest, err)
	}

	return nil
}

func parseVersion(output string) (string, error) {
	return strings.TrimSpace(strings.TrimPrefix(output, "v")), nil
}

func fileSuffix(plat tool.Platform) string {
	arch := plat.Arch
	if arch == "amd64" {
		arch = "x64"
	}

	return fmt.Sprintf("%s-%s", plat.OS, arch)
}
