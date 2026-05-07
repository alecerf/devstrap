package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/alecerf/devstrap/internal/downloader"
	"golang.org/x/mod/semver"
)

type fetchResult struct {
	Version  string `json:"version"`
	Tag      string `json:"tag"`
	Filename string `json:"filename"`
	Checksum string `json:"checksum"`
}

func fetchLatest(ctx context.Context, def Definition, data TemplateData) (fetchResult, error) {
	switch def.Source.Type {
	case "json_api":
		return fetchFromJSONAPI(ctx, def, data)
	case "github_release":
		return fetchFromGitHubRelease(ctx, def, data)
	default:
		return fetchResult{}, fmt.Errorf("%w: %s", errUnsupportedSource, def.Source.Type)
	}
}

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
		return fetchVersionFromGitHubRelease(ctx, def, data, version)
	default:
		return fetchResult{}, fmt.Errorf("%w: %s", errUnsupportedSource, def.Source.Type)
	}
}

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
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name string `json:"name"`
}

func fetchFromGitHubRelease(
	ctx context.Context, def Definition, data TemplateData,
) (fetchResult, error) {
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

	// Asset-based version extraction.
	if def.Source.Version.AssetRegex != "" {
		return extractVersionFromAssets(rel, def, data)
	}

	// Simple tag-based version (default behavior).
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

// navigatePath traverses a parsed JSON value using "[N]" and ".field" syntax.
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

func fetchVersionFromGitHubRelease(
	ctx context.Context,
	def Definition,
	data TemplateData,
	version string,
) (fetchResult, error) {
	// Asset-based version lookup: find the requested version in the latest release.
	if def.Source.Version.AssetRegex != "" {
		return fetchVersionFromAssets(ctx, def, data, version)
	}

	// Simple tag-based lookup.
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

func compileAssetRegex(def Definition, data TemplateData) (*regexp.Regexp, error) {
	pattern, err := renderTemplate(def.Source.Version.AssetRegex, data)
	if err != nil {
		return nil, fmt.Errorf("render asset_regex: %w", err)
	}

	assetRegex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("compile asset_regex %q: %w", pattern, err)
	}

	return assetRegex, nil
}

func extractVersionFromAssets(
	rel githubRelease, def Definition, data TemplateData,
) (fetchResult, error) {
	assetRegex, err := compileAssetRegex(def, data)
	if err != nil {
		return fetchResult{}, err
	}

	var versions []string

	for _, asset := range rel.Assets {
		matches := assetRegex.FindStringSubmatch(asset.Name)
		if len(matches) >= regexMinMatches {
			versions = append(versions, matches[1])
		}
	}

	if len(versions) == 0 {
		return fetchResult{}, fmt.Errorf(
			"%w: no assets matching %q",
			errNoFileMatch,
			assetRegex.String(),
		)
	}

	version := pickVersion(versions, def.Source.Version.Pick)

	return fetchResult{
		Version: version,
		Tag:     rel.TagName,
	}, nil
}

func fetchVersionFromAssets(
	ctx context.Context, def Definition, data TemplateData, version string,
) (fetchResult, error) {
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

	assetRegex, err := compileAssetRegex(def, data)
	if err != nil {
		return fetchResult{}, err
	}

	for _, asset := range rel.Assets {
		matches := assetRegex.FindStringSubmatch(asset.Name)
		if len(matches) >= regexMinMatches && matches[1] == version {
			return fetchResult{
				Version: version,
				Tag:     rel.TagName,
			}, nil
		}
	}

	return fetchResult{}, fmt.Errorf("%w: version %q not found in release %s",
		errVersionNotFound, version, rel.TagName)
}

func pickVersion(versions []string, strategy string) string {
	if len(versions) == 0 {
		return ""
	}

	if strategy == "highest" {
		best := versions[0]

		for _, v := range versions[1:] {
			if semver.Compare("v"+v, "v"+best) > 0 {
				best = v
			}
		}

		return best
	}

	// Default: return the first match.
	return versions[0]
}

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
