package registry

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

// fetchVersion discovers install metadata for a specific version.
func fetchVersion(
	ctx context.Context,
	def Definition,
	data TemplateData,
	version string,
) (fetchResult, error) {
	switch def.Source.Type {
	case "json_api":
		return fetchVersionFromJSONAPI(ctx, def, data, version)
	case "github_release":
		return fetchVersionFromGitHubRelease(ctx, def, version)
	default:
		return fetchResult{}, fmt.Errorf("%w: %s", errUnsupportedSource, def.Source.Type)
	}
}

// fetchFromJSONAPI fetches a JSON API endpoint and extracts the version
// (and optionally a file match with checksum).
func fetchFromJSONAPI(ctx context.Context, def Definition, data TemplateData) (fetchResult, error) {
	var raw json.RawMessage

	err := downloader.FetchJSON(ctx, def.Source.URL, &raw)
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
		return fetchResult{}, fmt.Errorf(
			"%w at path %q",
			errVersionNotString,
			def.Source.Version.Path,
		)
	}

	version = strings.TrimPrefix(version, def.Source.Version.StripPrefix)

	result := fetchResult{Version: version}

	if def.Source.FileMatch != nil {
		fileMatch, err := matchFile(parsed, def.Source.FileMatch, data, version)
		if err != nil {
			return fetchResult{}, fmt.Errorf("file match: %w", err)
		}

		result.Filename = fileMatch.filename
		result.Checksum = fileMatch.checksum
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

	var rel githubRelease

	err := downloader.FetchJSON(ctx, url, &rel)
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
func matchFile(
	root any,
	fileMatcher *FileMatch,
	data TemplateData,
	version string,
) (fileMatchResult, error) {
	listRaw, err := navigatePath(root, fileMatcher.ListPath)
	if err != nil {
		return fileMatchResult{}, fmt.Errorf("navigate to file list: %w", err)
	}

	list, ok := listRaw.([]any)
	if !ok {
		return fileMatchResult{}, fmt.Errorf(
			"%w: %q is not an array",
			errPathNavigation,
			fileMatcher.ListPath,
		)
	}

	return searchFileList(list, fileMatcher, data, version)
}

// matchFileInEntry is like matchFile but operates on a single array entry
// rather than the full API response. It extracts the field portion of
// ListPath (e.g. ".files" from "[0].files") to navigate within the entry.
func matchFileInEntry(
	entry any, fileMatcher *FileMatch, data TemplateData, version string,
) (fileMatchResult, error) {
	_, fieldPath, err := splitVersionPath(fileMatcher.ListPath)
	if err != nil {
		return fileMatchResult{}, fmt.Errorf("parse list path: %w", err)
	}

	listRaw, err := navigatePath(entry, fieldPath)
	if err != nil {
		return fileMatchResult{}, fmt.Errorf("navigate to file list: %w", err)
	}

	list, ok := listRaw.([]any)
	if !ok {
		return fileMatchResult{}, fmt.Errorf("%w: %q is not an array", errPathNavigation, fieldPath)
	}

	return searchFileList(list, fileMatcher, data, version)
}

func searchFileList(
	list []any, fileMatcher *FileMatch, data TemplateData, version string,
) (fileMatchResult, error) {
	data.Version = version

	pattern, err := renderTemplate(fileMatcher.FilenameContains, data)
	if err != nil {
		return fileMatchResult{}, fmt.Errorf("render filename pattern: %w", err)
	}

	for _, item := range list {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}

		filename, _ := obj[fileMatcher.FilenameField].(string)
		if filename == "" || !strings.Contains(filename, pattern) {
			continue
		}

		checksum, _ := obj[fileMatcher.ChecksumField].(string)

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
			return nil, fmt.Errorf(
				"%w: unexpected char %q in %q",
				errPathNavigation,
				string(remaining[0]),
				path,
			)
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
		return nil, "", fmt.Errorf(
			"%w: invalid index %q in %q",
			errPathNavigation,
			idxStr,
			fullPath,
		)
	}

	arr, ok := current.([]any)
	if !ok {
		return nil, "", fmt.Errorf("%w: expected array at %q", errPathNavigation, fullPath)
	}

	if idx < 0 || idx >= len(arr) {
		return nil, "", fmt.Errorf(
			"%w: index %d out of range (len %d)",
			errPathNavigation,
			idx,
			len(arr),
		)
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
		return nil, "", fmt.Errorf(
			"%w: expected object for key %q in %q",
			errPathNavigation,
			key,
			fullPath,
		)
	}

	val, exists := obj[key]
	if !exists {
		return nil, "", fmt.Errorf("%w: key %q not found in %q", errPathNavigation, key, fullPath)
	}

	return val, remaining[dot:], nil
}

// fetchVersionFromGitHubRelease queries the GitHub releases API for a specific version.
// It tries the "v"-prefixed tag first, then the plain version string.
func fetchVersionFromGitHubRelease(
	ctx context.Context,
	def Definition,
	version string,
) (fetchResult, error) {
	tags := []string{"v" + version, version}

	for _, tag := range tags {
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s",
			def.Source.Owner, def.Source.Repo, tag)

		var rel githubRelease

		err := downloader.FetchJSON(ctx, url, &rel)
		if err != nil {
			continue
		}

		if rel.TagName != "" {
			return fetchResult{
				Version: version,
				Tag:     rel.TagName,
			}, nil
		}
	}

	return fetchResult{}, fmt.Errorf("%w for version %q", errVersionNotFound, version)
}

// fetchVersionFromJSONAPI fetches the JSON API and searches for a specific version.
func fetchVersionFromJSONAPI(
	ctx context.Context, def Definition, data TemplateData, version string,
) (fetchResult, error) {
	var raw json.RawMessage

	err := downloader.FetchJSON(ctx, def.Source.URL, &raw)
	if err != nil {
		return fetchResult{}, fmt.Errorf("fetch %s: %w", def.Source.URL, err)
	}

	var parsed any

	err = json.Unmarshal(raw, &parsed)
	if err != nil {
		return fetchResult{}, fmt.Errorf("parse JSON from %s: %w", def.Source.URL, err)
	}

	entry, err := findVersionEntry(parsed, def.Source.Version, version)
	if err != nil {
		// Version not in API response (e.g. older Go releases are excluded
		// from the default endpoint). Construct the filename from the naming
		// convention so the download can still proceed.
		if def.Source.FileMatch != nil {
			return constructFetchResult(def, data, version)
		}

		return fetchResult{}, err
	}

	result := fetchResult{Version: version}

	if def.Source.FileMatch != nil {
		fileMatch, err := matchFileInEntry(entry, def.Source.FileMatch, data, version)
		if err != nil {
			return fetchResult{}, fmt.Errorf("file match: %w", err)
		}

		result.Filename = fileMatch.filename
		result.Checksum = fileMatch.checksum
	}

	return result, nil
}

// constructFetchResult builds a fetchResult when the requested version is not
// present in the API response. It synthesises the filename from the tool's
// naming convention: {strip_prefix}{version}.{rendered filename_contains}.
// No embedded checksum is available in this case; the separate checksum file
// is used if the tool definition provides one.
func constructFetchResult(
	def Definition, data TemplateData, version string,
) (fetchResult, error) {
	data.Version = version

	pattern, err := renderTemplate(def.Source.FileMatch.FilenameContains, data)
	if err != nil {
		return fetchResult{}, fmt.Errorf("render filename pattern: %w", err)
	}

	filename := def.Source.Version.StripPrefix + version + "." + pattern

	return fetchResult{Version: version, Filename: filename}, nil
}

// findVersionEntry searches a JSON array for an entry matching the requested version.
// It uses the version path (e.g., "[0].version") to determine the array location
// and the version field within each element.
func findVersionEntry(root any, vCfg VersionExtract, version string) (any, error) {
	arrayPath, fieldPath, err := splitVersionPath(vCfg.Path)
	if err != nil {
		return nil, err
	}

	var arr []any

	if arrayPath == "" {
		a, ok := root.([]any)
		if !ok {
			return nil, fmt.Errorf("%w: root is not an array", errPathNavigation)
		}

		arr = a
	} else {
		raw, err := navigatePath(root, arrayPath)
		if err != nil {
			return nil, fmt.Errorf("navigate to array: %w", err)
		}

		a, ok := raw.([]any)
		if !ok {
			return nil, fmt.Errorf("%w: value at %q is not an array", errPathNavigation, arrayPath)
		}

		arr = a
	}

	for _, entry := range arr {
		vRaw, err := navigatePath(entry, fieldPath)
		if err != nil {
			continue
		}

		versionStr, ok := vRaw.(string)
		if !ok {
			continue
		}

		versionStr = strings.TrimPrefix(versionStr, vCfg.StripPrefix)
		if versionStr == version {
			return entry, nil
		}
	}

	return nil, fmt.Errorf("%w: %q not found in API response", errVersionNotFound, version)
}

// splitVersionPath splits a version path like "[0].version" into the path
// to the containing array and the field path within each element.
//
// Examples:
//
//	"[0].version"            → arrayPath="",          fieldPath=".version"
//	".releases[0].version"   → arrayPath=".releases", fieldPath=".version"
func splitVersionPath(path string) (string, string, error) {
	bracket := strings.Index(path, "[")
	if bracket == -1 {
		return "", "", fmt.Errorf("%w: version path %q has no array index", errPathNavigation, path)
	}

	closeBracket := strings.IndexByte(path[bracket:], ']')
	if closeBracket == -1 {
		return "", "", fmt.Errorf(
			"%w: unclosed bracket in version path %q",
			errPathNavigation,
			path,
		)
	}

	return path[:bracket], path[bracket+closeBracket+1:], nil
}
