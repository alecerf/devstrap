// Package github implements a Tool for binaries distributed via GitHub releases.
package github

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

var errNoReleaseTag = errors.New("no release tag found")

// Tool implements tool.Tool for a binary distributed via GitHub releases.
type Tool struct {
	Owner    string
	Repo     string
	BinName  string
	BinDir   string
	Platform tool.Platform

	// VersionArgs are passed to the installed binary to get its version.
	// Default: ["version"].
	VersionArgs []string

	// ParseVersion extracts a clean version from the command output.
	ParseVersion func(output string) (string, error)

	// Optional overrides. Nil uses sensible defaults.
	ArchiveName  func(version string, plat tool.Platform) string
	ChecksumName func(version string) string
	BinaryPath   func(tmpDir, version string, plat tool.Platform) string
}

// Name returns the tool binary name.
func (g *Tool) Name() string { return g.BinName }

func (g *Tool) archiveName(version string) string {
	if g.ArchiveName != nil {
		return g.ArchiveName(version, g.Platform)
	}

	return fmt.Sprintf("%s-%s-%s.tar.gz", g.BinName, version, g.Platform.FileSuffix())
}

func (g *Tool) checksumName(version string) string {
	if g.ChecksumName != nil {
		return g.ChecksumName(version)
	}

	return fmt.Sprintf("%s-%s-checksums.txt", g.BinName, version)
}

func (g *Tool) binaryPath(tmpDir, version string) string {
	if g.BinaryPath != nil {
		return g.BinaryPath(tmpDir, version, g.Platform)
	}

	return filepath.Join(
		tmpDir,
		fmt.Sprintf("%s-%s-%s", g.BinName, version, g.Platform.FileSuffix()),
		g.BinName,
	)
}

func (g *Tool) versionArgs() []string {
	if len(g.VersionArgs) > 0 {
		return g.VersionArgs
	}

	return []string{"version"}
}

// CurrentVersion returns the installed version by running the binary.
func (g *Tool) CurrentVersion(ctx context.Context) (string, error) {
	bin := filepath.Join(g.BinDir, g.BinName)

	v, err := tool.RunVersionCmd(ctx, bin, g.versionArgs(), g.ParseVersion)
	if err != nil {
		return "", fmt.Errorf("get %s version: %w", g.BinName, err)
	}

	return v, nil
}

type release struct {
	TagName string `json:"tag_name"`
}

// FetchLatest queries the GitHub API for the latest release version and tag.
func (g *Tool) FetchLatest(ctx context.Context) (string, string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", g.Owner, g.Repo)

	rel, err := downloader.FetchJSON[release](ctx, url)
	if err != nil {
		return "", "", fmt.Errorf("fetch github release: %w", err)
	}

	if rel.TagName == "" {
		return "", "", errNoReleaseTag
	}

	return strings.TrimPrefix(rel.TagName, "v"), rel.TagName, nil
}

// Install downloads, verifies, extracts, and installs the binary from a GitHub release.
func (g *Tool) Install(ctx context.Context, status func(string), version, tag string) error {
	tmp, err := os.MkdirTemp("", "devstrap-"+g.BinName+"-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}

	defer func() { _ = os.RemoveAll(tmp) }()

	archiveName := g.archiveName(version)
	checksumName := g.checksumName(version)
	baseURL := fmt.Sprintf("https://github.com/%s/%s/releases/download/%s", g.Owner, g.Repo, tag)
	archive := filepath.Join(tmp, archiveName)

	err = downloader.DownloadAndVerify(ctx, status, baseURL, archiveName, checksumName, tmp, archive)
	if err != nil {
		return fmt.Errorf("download and verify: %w", err)
	}

	status("extracting...")

	err = downloader.ExtractTarGz(ctx, archive, tmp, 0)
	if err != nil {
		return fmt.Errorf("extract archive: %w", err)
	}

	err = downloader.InstallBinary(g.binaryPath(tmp, version), g.BinDir, g.BinName)
	if err != nil {
		return fmt.Errorf("install binary: %w", err)
	}

	return nil
}
