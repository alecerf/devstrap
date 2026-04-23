package index

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alecerf/devstrap/internal/downloader"
)

// fetchResult holds the output of a version discovery call.
type fetchResult struct {
	Version  string `json:"version"`
	Tag      string `json:"tag"`      // GitHub release tag (may differ from version)
	Filename string `json:"filename"` // matched filename from file_match (if any)
	Checksum string `json:"checksum"` // embedded checksum from file_match (if any)
}

// fetchLatest discovers the latest version of a tool using its source config.
func fetchLatest(ctx context.Context, def Definition, data TemplateData) (fetchResult, error) {
	switch def.Source.Type {
	case "json_api":
		return fetchFromJSONAPI(ctx, def, data)
	case "github_release":
		return fetchFromGitHubRelease(ctx, def)
	default:
		return fetchResult{}, fmt.Errorf("%w: %s", errUnsupportedSource, def.Source.Type)
	}
}

// fetchFromJSONAPI fetches a JSON API endpoint and extracts the version
// (and optionally a file match with checksum).
func fetchFromJSONAPI(ctx context.Context, def Definition, data TemplateData) (fetchResult, error) {
	raw, err := downloader.FetchJSON[json.RawMessage](ctx, def.Source.URL)
	if err != nil {
		return fetchResult{}, fmt.Errorf("fetch %s: %w", def.Source.URL, err)
	}

	var parsed any

	err = json.Unmarshal(raw, &parsed)
	if err != nil {
		return fetchResult{}, fmt.Errorf("parse JSON from %s: %w", def.Source.URL, err)
	}

	versionRaw, err := navigatePath(parsed, def.Source.Version.Path)
	if err != nil {
		return fetchResult{}, fmt.Errorf("extract version: %w", err)
	}

	version, ok := versionRaw.(string)
	if !ok {
		return fetchResult{}, fmt.Errorf("%w at path %q", errVersionNotString, def.Source.Version.Path)
	}

	version = strings.TrimPrefix(version, def.Source.Version.StripPrefix)

	result := fetchResult{Version: version}

	if def.Source.FileMatch != nil {
		fm, err := matchFile(parsed, def.Source.FileMatch, data, version)
		if err != nil {
			return fetchResult{}, fmt.Errorf("file match: %w", err)
		}

		result.Filename = fm.filename
		result.Checksum = fm.checksum
	}

	return result, nil
}

type githubRelease struct {
	TagName string `json:"tag_name"`
}

// fetchFromGitHubRelease queries the GitHub releases API for the latest version.
func fetchFromGitHubRelease(ctx context.Context, def Definition) (fetchResult, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest",
		def.Source.Owner, def.Source.Repo)

	rel, err := downloader.FetchJSON[githubRelease](ctx, url)
	if err != nil {
		return fetchResult{}, fmt.Errorf("fetch github release: %w", err)
	}

	if rel.TagName == "" {
		return fetchResult{}, errNoReleaseTag
	}

	version := strings.TrimPrefix(rel.TagName, "v")

	return fetchResult{
		Version: version,
		Tag:     rel.TagName,
	}, nil
}

type fileMatchResult struct {
	filename string
	checksum string
}

// matchFile searches a JSON array for a file entry whose filename contains
// the rendered pattern, and extracts the checksum field.
func matchFile(root any, fm *FileMatch, data TemplateData, version string) (fileMatchResult, error) {
	data.Version = version

	pattern, err := renderTemplate(fm.FilenameContains, data)
	if err != nil {
		return fileMatchResult{}, fmt.Errorf("render filename pattern: %w", err)
	}

	listRaw, err := navigatePath(root, fm.ListPath)
	if err != nil {
		return fileMatchResult{}, fmt.Errorf("navigate to file list: %w", err)
	}

	list, ok := listRaw.([]any)
	if !ok {
		return fileMatchResult{}, fmt.Errorf("%w: %q is not an array", errPathNavigation, fm.ListPath)
	}

	for _, item := range list {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}

		filename, _ := obj[fm.FilenameField].(string)
		if filename == "" || !strings.Contains(filename, pattern) {
			continue
		}

		checksum, _ := obj[fm.ChecksumField].(string)

		return fileMatchResult{
			filename: filename,
			checksum: checksum,
		}, nil
	}

	return fileMatchResult{}, fmt.Errorf("%w for pattern %q", errNoFileMatch, pattern)
}

// navigatePath traverses a parsed JSON value using a simple path expression.
// Supported syntax:
//   - "[N]" — index into an array
//   - ".field" — access an object key
//
// Example: "[0].version" navigates to the first array element, then its
// "version" field.
func navigatePath(data any, path string) (any, error) {
	current := data
	remaining := path

	for remaining != "" {
		var err error

		switch remaining[0] {
		case '[':
			current, remaining, err = navigateArray(current, remaining, path)
		case '.':
			current, remaining, err = navigateObject(current, remaining, path)
		default:
			return nil, fmt.Errorf("%w: unexpected char %q in %q", errPathNavigation, string(remaining[0]), path)
		}

		if err != nil {
			return nil, err
		}
	}

	return current, nil
}

func navigateArray(current any, remaining, fullPath string) (any, string, error) {
	end := strings.IndexByte(remaining, ']')
	if end == -1 {
		return nil, "", fmt.Errorf("%w: unclosed bracket in %q", errPathNavigation, fullPath)
	}

	idxStr := remaining[1:end]

	var idx int

	_, err := fmt.Sscanf(idxStr, "%d", &idx)
	if err != nil {
		return nil, "", fmt.Errorf("%w: invalid index %q in %q", errPathNavigation, idxStr, fullPath)
	}

	arr, ok := current.([]any)
	if !ok {
		return nil, "", fmt.Errorf("%w: expected array at %q", errPathNavigation, fullPath)
	}

	if idx < 0 || idx >= len(arr) {
		return nil, "", fmt.Errorf("%w: index %d out of range (len %d)", errPathNavigation, idx, len(arr))
	}

	return arr[idx], remaining[end+1:], nil
}

func navigateObject(current any, remaining, fullPath string) (any, string, error) {
	remaining = remaining[1:]

	dot := strings.IndexAny(remaining, ".[")
	if dot == -1 {
		dot = len(remaining)
	}

	key := remaining[:dot]

	obj, ok := current.(map[string]any)
	if !ok {
		return nil, "", fmt.Errorf("%w: expected object for key %q in %q", errPathNavigation, key, fullPath)
	}

	val, exists := obj[key]
	if !exists {
		return nil, "", fmt.Errorf("%w: key %q not found in %q", errPathNavigation, key, fullPath)
	}

	return val, remaining[dot:], nil
}
