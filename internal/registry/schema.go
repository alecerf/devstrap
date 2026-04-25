// Package registry implements a declarative tool index for devstrap.
// Tool definitions are JSON files fetched from a remote repository and cached
// locally. Each definition describes the full installation pipeline: version
// discovery, download, checksum verification, and installation.
package registry

import "context"

// Tool is the interface every installer implements.
type Tool interface {
	Name() string
	FetchLatest(ctx context.Context) (version, extra string, err error)
	FetchVersion(ctx context.Context, version string) (extra string, err error)
	CurrentVersion(ctx context.Context) (string, error)
	Install(ctx context.Context, status func(string), version, extra string) error
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

// CheckInfo holds version information gathered during a scan.
type CheckInfo struct {
	Name     string
	Current  string // empty if not installed
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
	// Type is the source strategy: "json_api" or "github_release".
	Type string `json:"type"`

	// URL is the API endpoint for "json_api" sources.
	URL string `json:"url,omitempty"`

	// Version configures how to extract the version string from the API response.
	Version VersionExtract `json:"version"`

	// FileMatch optionally locates a specific file entry (and its checksum)
	// within the API response. Used when the checksum is embedded in the
	// version API (e.g. Go releases).
	FileMatch *FileMatch `json:"file_match,omitempty"`

	// Owner is the GitHub repository owner (for "github_release" sources).
	Owner string `json:"owner,omitempty"`

	// Repo is the GitHub repository name (for "github_release" sources).
	Repo string `json:"repo,omitempty"`
}

// VersionExtract describes how to pull a version string from a JSON response.
type VersionExtract struct {
	// Path is a simple expression like "[0].version" to navigate JSON.
	Path string `json:"path"`

	// StripPrefix is removed from the extracted value (e.g. "go" or "v").
	StripPrefix string `json:"strip_prefix,omitempty"`
}

// FileMatch describes how to locate a file entry and its checksum in a JSON
// array embedded in the version API response.
type FileMatch struct {
	// ListPath navigates to the array of file objects (e.g. "[0].files").
	ListPath string `json:"list_path"`

	// FilenameField is the JSON key for the filename (e.g. "filename").
	FilenameField string `json:"filename_field"`

	// FilenameContains is a Go template that must appear in the filename.
	// Example: "{{.OS}}-{{.Arch}}.tar.gz"
	FilenameContains string `json:"filename_contains"`

	// ChecksumField is the JSON key for the SHA-256 checksum (e.g. "sha256").
	ChecksumField string `json:"checksum_field"`
}

// Download describes how to download the tool archive and verify its checksum.
type Download struct {
	// URL is a Go template for the archive download URL.
	URL string `json:"url"`

	// Checksum optionally describes how to obtain the SHA-256 checksum.
	// If nil, the checksum comes from Source.FileMatch (embedded strategy).
	Checksum *Checksum `json:"checksum,omitempty"`
}

// Checksum describes a separate checksum file download.
type Checksum struct {
	// URL is a Go template for the checksum file download URL.
	URL string `json:"url"`
}

// Install describes how to install the downloaded artifact.
type Install struct {
	// Mode is "directory" (extract archive to a folder) or "binary"
	// (extract a single binary).
	Mode string `json:"mode"`

	// Dest is the installation directory relative to baseDir (e.g. "go", "bin").
	Dest string `json:"dest"`

	// StripComponents removes leading path components during extraction
	// (like tar --strip-components). Only used with mode "directory".
	StripComponents int `json:"strip_components,omitempty"`

	// BinaryName is the binary filename. Only used with mode "binary".
	BinaryName string `json:"binary_name,omitempty"`

	// ArchivePath is a Go template for the path to the binary inside the
	// extracted archive. Only used with mode "binary".
	ArchivePath string `json:"archive_path,omitempty"`
}

// Detect describes how to detect the currently installed version.
type Detect struct {
	// Binary is the path to the binary relative to baseDir (e.g. "go/bin/go").
	Binary string `json:"binary"`

	// Args are passed to the binary to get version output.
	Args []string `json:"args"`

	// VersionRegex is a regex with one capture group for the version string.
	VersionRegex string `json:"version_regex"`
}

// PlatformSpec describes supported OS/arch combinations and optional name mappings.
type PlatformSpec struct {
	// OS lists supported operating systems (e.g. ["darwin", "linux"]).
	OS []string `json:"os"`

	// Arch maps runtime GOARCH values to tool-specific arch names.
	// Example: {"amd64": "x64", "arm64": "arm64"}
	Arch map[string]string `json:"arch"`
}
