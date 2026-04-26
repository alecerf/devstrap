package registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alecerf/devstrap/internal/downloader"
)

const (
	// DefaultRemoteURL is the base URL for the default devstrap index repository.
	DefaultRemoteURL = "https://raw.githubusercontent.com/alecerf/devstrap-index/trunk"

	indexFilename = "index.json"
	toolsFilename = "tools.json"
	dirPerm       = 0o750
	filePerm      = 0o600
)

// ErrNoIndex is returned when the local index cache does not exist.
var ErrNoIndex = errors.New("no index found — run 'devstrap index update' first")

// Index holds all loaded tool definitions.
type Index struct {
	Definitions []Definition
}

// CacheDir returns the default index cache directory.
// It respects $XDG_CACHE_HOME, falling back to ~/.cache/devstrap.
func CacheDir() (string, error) {
	if dir := os.Getenv("XDG_CACHE_HOME"); dir != "" {
		return filepath.Join(dir, "devstrap"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}

	return filepath.Join(home, ".cache", "devstrap"), nil
}

// Load reads the index from the local cache directory.
// Returns ErrNoIndex if the cache does not exist.
func Load() (*Index, error) {
	cacheDir, err := CacheDir()
	if err != nil {
		return nil, err
	}

	return LoadFrom(filepath.Join(cacheDir, indexFilename))
}

// LoadFrom reads the index from the given JSON file.
func LoadFrom(path string) (*Index, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNoIndex
		}

		return nil, fmt.Errorf("read index: %w", err)
	}

	var defs []Definition

	err = json.Unmarshal(data, &defs)
	if err != nil {
		return nil, fmt.Errorf("parse index: %w", err)
	}

	return &Index{Definitions: defs}, nil
}

// Update fetches the index from the remote repository and writes it to the
// local cache directory.
func Update(ctx context.Context, remoteURL string) error {
	cacheDir, err := CacheDir()
	if err != nil {
		return err
	}

	return UpdateTo(ctx, remoteURL, cacheDir)
}

// UpdateTo fetches the tool listing and individual definitions from the remote
// URL, assembles them into a single index file, and writes it to dir.
func UpdateTo(ctx context.Context, remoteURL, dir string) error {
	err := os.MkdirAll(dir, dirPerm)
	if err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}

	var toolNames []string

	err = downloader.FetchJSON(ctx, remoteURL+"/"+toolsFilename, &toolNames)
	if err != nil {
		return fmt.Errorf("fetch tool listing: %w", err)
	}

	defs := make([]Definition, 0, len(toolNames))

	for _, toolPath := range toolNames {
		var def Definition

		err = downloader.FetchJSON(ctx, remoteURL+"/"+toolPath, &def)
		if err != nil {
			return fmt.Errorf("fetch tool %s: %w", toolPath, err)
		}

		defs = append(defs, def)
	}

	data, err := json.MarshalIndent(defs, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal index: %w", err)
	}

	localPath := filepath.Join(dir, indexFilename)

	err = os.WriteFile(localPath, data, filePerm)
	if err != nil {
		return fmt.Errorf("write index: %w", err)
	}

	return nil
}

// ToolNames returns the names of all tools in the index.
func (idx *Index) ToolNames() []string {
	names := make([]string, len(idx.Definitions))
	for i, def := range idx.Definitions {
		names[i] = def.Name
	}

	return names
}

// FindTool returns the definition for the named tool, or nil if not found.
func (idx *Index) FindTool(name string) *Definition {
	for i := range idx.Definitions {
		if idx.Definitions[i].Name == name {
			return &idx.Definitions[i]
		}
	}

	return nil
}
