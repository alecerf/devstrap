package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/alecerf/devstrap/internal/downloader"
)

const regexMinMatches = 2

// Installer implements Tool using a declarative Definition.
type Installer struct {
	def   Definition
	paths Paths
	os    string
	arch  string
}

// NewTool creates an Installer from a Definition, paths, and platform.
func NewTool(def Definition, paths Paths, plat Platform) (*Installer, error) {
	mappedOS, mappedArch, err := mapPlatform(plat, def)
	if err != nil {
		return nil, err
	}

	return &Installer{
		def:   def,
		paths: paths,
		os:    mappedOS,
		arch:  mappedArch,
	}, nil
}

// Name returns the tool's name as defined in the index.
func (i *Installer) Name() string { return i.def.Name }

// FetchLatest queries the source for the latest version and returns the version,
// serialised fetch metadata (extra), and any error.
func (i *Installer) FetchLatest(ctx context.Context) (string, string, error) {
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

// FetchVersion queries the source for a specific version and returns serialised
// fetch metadata (extra) or an error.
func (i *Installer) FetchVersion(ctx context.Context, version string) (string, error) {
	data := i.templateData()

	result, err := fetchVersion(ctx, i.def, data, version)
	if err != nil {
		return "", err
	}

	extra, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("marshal fetch result: %w", err)
	}

	return string(extra), nil
}

// CurrentVersion runs the tool's version command and parses the result.
func (i *Installer) CurrentVersion(ctx context.Context) (string, error) {
	var bin string

	if i.def.Install.Mode == "binary" || i.def.Install.Mode == "direct" {
		bin = filepath.Join(i.paths.BinDir, i.def.Install.BinaryName)
	} else {
		bin = filepath.Join(i.paths.DataDir, i.def.Detect.Binary)
	}

	versionRegex, err := regexp.Compile(i.def.Detect.VersionRegex)
	if err != nil {
		return "", fmt.Errorf("compile version regex: %w", err)
	}

	parse := func(output string) (string, error) {
		matches := versionRegex.FindStringSubmatch(output)
		if len(matches) < regexMinMatches {
			return "", fmt.Errorf(
				"%w: regex %q, output %q",
				errVersionRegexNoMatch,
				i.def.Detect.VersionRegex,
				output,
			)
		}

		return strings.TrimPrefix(matches[1], "v"), nil
	}

	v, err := RunVersionCmd(ctx, bin, i.def.Detect.Args, parse)
	if err != nil {
		return "", fmt.Errorf("detect %s version: %w", i.def.Name, err)
	}

	return v, nil
}

// Install downloads, verifies, and installs the tool at the given version using
// the serialised fetch metadata produced by FetchLatest or FetchVersion.
func (i *Installer) Install(ctx context.Context, status func(string), version, extra string) error {
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

	switch i.def.Install.Mode {
	case "directory":
		status("extracting...")

		return i.installDirectory(ctx, tmp, archiveFile)
	case "binary":
		status("extracting...")

		return i.installBinary(ctx, data, tmp, archiveFile)
	case "direct":
		status("installing...")

		return i.installDirect(archiveFile)
	default:
		return fmt.Errorf("%w: %s", errUnsupportedInstallMode, i.def.Install.Mode)
	}
}

func (i *Installer) templateData() TemplateData {
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

func (i *Installer) verifyChecksum(
	ctx context.Context,
	status func(string),
	data TemplateData,
	result fetchResult,
	tmp, archiveFile string,
) error {
	// Embedded checksum from file_match (e.g. Go releases).
	if result.Checksum != "" {
		status("verifying checksum...")

		err := downloader.VerifyChecksum(archiveFile, result.Checksum)
		if err != nil {
			return fmt.Errorf("verify checksum: %w", err)
		}

		return nil
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

func (i *Installer) installDirectory(ctx context.Context, tmp, archiveFile string) error {
	extractDir := filepath.Join(tmp, "extract")

	err := os.MkdirAll(extractDir, dirPerm)
	if err != nil {
		return fmt.Errorf("create extract dir: %w", err)
	}

	err = downloader.ExtractTarGz(ctx, archiveFile, extractDir, i.def.Install.StripComponents)
	if err != nil {
		return fmt.Errorf("extract archive: %w", err)
	}

	dest := filepath.Join(i.paths.DataDir, i.def.Install.Dest)
	_ = os.RemoveAll(dest)

	err = os.MkdirAll(filepath.Dir(dest), dirPerm)
	if err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}

	err = os.Rename(extractDir, dest)
	if err != nil {
		return fmt.Errorf("rename to %s: %w", dest, err)
	}

	return nil
}

func (i *Installer) installDirect(downloadedFile string) error {
	err := downloader.InstallBinary(downloadedFile, i.paths.BinDir, i.def.Install.BinaryName)
	if err != nil {
		return fmt.Errorf("install binary: %w", err)
	}

	return nil
}

func (i *Installer) installBinary(
	ctx context.Context,
	data TemplateData,
	tmp, archiveFile string,
) error {
	err := downloader.ExtractTarGz(ctx, archiveFile, tmp, 0)
	if err != nil {
		return fmt.Errorf("extract archive: %w", err)
	}

	archivePath, err := renderTemplate(i.def.Install.ArchivePath, data)
	if err != nil {
		return fmt.Errorf("render archive path: %w", err)
	}

	binaryPath := filepath.Join(tmp, archivePath)
	destDir := i.paths.BinDir

	err = downloader.InstallBinary(binaryPath, destDir, i.def.Install.BinaryName)
	if err != nil {
		return fmt.Errorf("install binary: %w", err)
	}

	return nil
}
