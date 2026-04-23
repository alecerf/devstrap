package index

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/alecerf/devstrap/internal/downloader"
	"github.com/alecerf/devstrap/internal/tool"
)

// installer implements tool.Tool using a declarative Definition.
type installer struct {
	def     Definition
	baseDir string
	os      string
	arch    string
}

// NewTool creates a tool.Tool from a Definition, baseDir, and platform.
//
//nolint:ireturn // factory function
func NewTool(def Definition, baseDir string, plat tool.Platform) (tool.Tool, error) {
	mappedOS, mappedArch, err := mapPlatform(plat, def)
	if err != nil {
		return nil, err
	}

	return &installer{
		def:     def,
		baseDir: baseDir,
		os:      mappedOS,
		arch:    mappedArch,
	}, nil
}

func (i *installer) Name() string { return i.def.Name }

func (i *installer) templateData() TemplateData {
	return TemplateData{
		OS:   i.os,
		Arch: i.arch,
		Name: i.def.Name,
		Source: SourceData{
			Owner: i.def.Source.Owner,
			Repo:  i.def.Source.Repo,
		},
	}
}

func (i *installer) FetchLatest(ctx context.Context) (string, string, error) {
	data := i.templateData()

	result, err := fetchLatest(ctx, i.def, data)
	if err != nil {
		return "", "", err
	}

	// Encode the full fetchResult as extra so Install() can use it.
	extra, err := json.Marshal(result)
	if err != nil {
		return "", "", fmt.Errorf("marshal fetch result: %w", err)
	}

	return result.Version, string(extra), nil
}

func (i *installer) CurrentVersion(ctx context.Context) (string, error) {
	bin := filepath.Join(i.baseDir, i.def.Detect.Binary)

	re, err := regexp.Compile(i.def.Detect.VersionRegex)
	if err != nil {
		return "", fmt.Errorf("compile version regex: %w", err)
	}

	parse := func(output string) (string, error) {
		m := re.FindStringSubmatch(output)
		if len(m) < 2 {
			return "", fmt.Errorf("%w: regex %q, output %q", errVersionRegexNoMatch, i.def.Detect.VersionRegex, output)
		}

		return strings.TrimPrefix(m[1], "v"), nil
	}

	v, err := tool.RunVersionCmd(ctx, bin, i.def.Detect.Args, parse)
	if err != nil {
		return "", fmt.Errorf("detect %s version: %w", i.def.Name, err)
	}

	return v, nil
}

func (i *installer) Install(ctx context.Context, status func(string), version, extra string) error {
	var result fetchResult

	err := json.Unmarshal([]byte(extra), &result)
	if err != nil {
		return fmt.Errorf("unmarshal fetch result: %w", err)
	}

	data := i.templateData()
	data.Version = version
	data.Tag = result.Tag
	data.Filename = result.Filename

	tmp, err := os.MkdirTemp("", "devstrap-"+i.def.Name+"-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}

	defer func() { _ = os.RemoveAll(tmp) }()

	archiveURL, err := renderTemplate(i.def.Download.URL, data)
	if err != nil {
		return fmt.Errorf("render download URL: %w", err)
	}

	archiveFile := filepath.Join(tmp, filepath.Base(archiveURL))

	err = downloader.Download(ctx, archiveURL, archiveFile)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}

	err = i.verifyChecksum(ctx, status, data, result, tmp, archiveFile)
	if err != nil {
		return err
	}

	status("extracting...")

	switch i.def.Install.Mode {
	case "directory":
		return i.installDirectory(ctx, tmp, archiveFile)
	case "binary":
		return i.installBinary(ctx, data, tmp, archiveFile)
	default:
		return fmt.Errorf("%w: %s", errUnsupportedInstallMode, i.def.Install.Mode)
	}
}

func (i *installer) verifyChecksum(
	ctx context.Context,
	status func(string),
	data TemplateData,
	result fetchResult,
	tmp, archiveFile string,
) error {
	// Embedded checksum from file_match (e.g. Go releases).
	if result.Checksum != "" {
		status("verifying checksum...")

		return fmt.Errorf("verify checksum: %w", downloader.VerifyChecksum(archiveFile, result.Checksum))
	}

	// Separate checksum file.
	if i.def.Download.Checksum != nil {
		checksumURL, err := renderTemplate(i.def.Download.Checksum.URL, data)
		if err != nil {
			return fmt.Errorf("render checksum URL: %w", err)
		}

		checksumFile := filepath.Join(tmp, filepath.Base(checksumURL))

		err = downloader.Download(ctx, checksumURL, checksumFile)
		if err != nil {
			return fmt.Errorf("download checksum: %w", err)
		}

		status("verifying checksum...")

		expectedSHA, err := downloader.ExtractChecksum(checksumFile, filepath.Base(archiveFile))
		if err != nil {
			return fmt.Errorf("extract checksum: %w", err)
		}

		err = downloader.VerifyChecksum(archiveFile, expectedSHA)
		if err != nil {
			return fmt.Errorf("verify checksum: %w", err)
		}

		return nil
	}

	return nil
}

func (i *installer) installDirectory(ctx context.Context, tmp, archiveFile string) error {
	extractDir := filepath.Join(tmp, "extract")

	err := os.MkdirAll(extractDir, 0o750)
	if err != nil {
		return fmt.Errorf("create extract dir: %w", err)
	}

	err = downloader.ExtractTarGz(ctx, archiveFile, extractDir, i.def.Install.StripComponents)
	if err != nil {
		return fmt.Errorf("extract archive: %w", err)
	}

	dest := filepath.Join(i.baseDir, i.def.Install.Dest)
	_ = os.RemoveAll(dest)

	err = os.Rename(extractDir, dest)
	if err != nil {
		return fmt.Errorf("rename to %s: %w", dest, err)
	}

	return nil
}

func (i *installer) installBinary(ctx context.Context, data TemplateData, tmp, archiveFile string) error {
	err := downloader.ExtractTarGz(ctx, archiveFile, tmp, 0)
	if err != nil {
		return fmt.Errorf("extract archive: %w", err)
	}

	archivePath, err := renderTemplate(i.def.Install.ArchivePath, data)
	if err != nil {
		return fmt.Errorf("render archive path: %w", err)
	}

	binaryPath := filepath.Join(tmp, archivePath)
	destDir := filepath.Join(i.baseDir, i.def.Install.Dest)

	err = downloader.InstallBinary(binaryPath, destDir, i.def.Install.BinaryName)
	if err != nil {
		return fmt.Errorf("install binary: %w", err)
	}

	return nil
}
