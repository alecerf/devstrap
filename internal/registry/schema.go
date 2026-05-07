package registry

import (
	"context"
	"path"
)

// Tool is the interface every installer implements.
type Tool interface {
	Name() string
	FetchLatest(ctx context.Context) (version, extra string, err error)
	FetchVersion(ctx context.Context, version string) (extra string, err error)
	CurrentVersion(ctx context.Context) (string, error)
	Install(ctx context.Context, status func(string), version, extra string) error
	Uninstall() error
}

// Paths holds the resolved installation directories.
type Paths struct {
	DataDir string
	BinDir  string
}

// Platform holds the detected OS and architecture.
type Platform struct {
	OS   string
	Arch string
}

// Result represents the outcome of a single tool operation.
type Result struct {
	Name    string
	Status  string // "up-to-date", "installed"
	Version string
	Err     error
}

// CheckInfo holds version information gathered during a dry-run scan.
type CheckInfo struct {
	Name     string
	Current  string
	Latest   string
	UpToDate bool
	Err      error
}

// Definition is the top-level structure of a tool definition JSON file.
type Definition struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Homepage    string       `json:"homepage,omitempty"`
	Source      Source       `json:"source"`
	Download    Download     `json:"download"`
	Install     Install      `json:"install"`
	Detect      Detect       `json:"detect"`
	Platforms   PlatformSpec `json:"platforms"`
}

// Source describes how to discover the latest version of a tool.
type Source struct {
	Type      string         `json:"type"` // "json_api" or "github_release"
	URL       string         `json:"url,omitempty"`
	Version   VersionExtract `json:"version"`
	FileMatch *FileMatch     `json:"file_match,omitempty"`
	Owner     string         `json:"owner,omitempty"`
	Repo      string         `json:"repo,omitempty"`
}

// VersionExtract describes how to pull a version string from the API response.
type VersionExtract struct {
	Path        string `json:"path,omitempty"`
	StripPrefix string `json:"strip_prefix,omitempty"`
	AssetRegex  string `json:"asset_regex,omitempty"`
	Pick        string `json:"pick,omitempty"` // "highest" or empty (first match)
}

// FileMatch locates a file entry and its checksum in a JSON API response.
type FileMatch struct {
	ListPath         string `json:"list_path"`
	FilenameField    string `json:"filename_field"`
	FilenameContains string `json:"filename_contains"`
	ChecksumField    string `json:"checksum_field"`
}

// Download describes how to download and verify the tool archive.
type Download struct {
	URL      string    `json:"url"`
	Checksum *Checksum `json:"checksum,omitempty"`
}

// Checksum describes a separate checksum file download.
type Checksum struct {
	URL string `json:"url"`
}

// Install describes how to install the downloaded artifact.
type Install struct {
	Mode            string `json:"mode"` // "directory", "binary", or "direct"
	Dest            string `json:"dest"`
	StripComponents int    `json:"strip_components,omitempty"`
	BinaryName      string `json:"binary_name,omitempty"`
	ArchivePath     string `json:"archive_path,omitempty"`
}

// Detect describes how to detect the currently installed version.
type Detect struct {
	Binary       string   `json:"binary"`
	Args         []string `json:"args"`
	VersionRegex string   `json:"version_regex"`
}

// PlatformSpec maps runtime GOOS/GOARCH values to tool-specific names.
type PlatformSpec struct {
	OS   map[string]string `json:"os"`
	Arch map[string]string `json:"arch"`
}

// RelBinDir returns the bin directory relative to DataDir for directory-mode tools.
func (d *Definition) RelBinDir() string {
	if d.Install.Mode != "directory" || d.Detect.Binary == "" {
		return ""
	}

	return path.Dir(d.Detect.Binary)
}
