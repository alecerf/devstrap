package registry

import (
	"errors"
	"fmt"
	"testing"
)

func TestNavigatePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		data    any
		path    string
		want    any
		wantErr error
	}{
		{
			name: "simple field",
			data: map[string]any{"version": "1.0"},
			path: ".version",
			want: "1.0",
		},
		{
			name: "array index then field",
			data: []any{map[string]any{"version": "1.2.3"}},
			path: "[0].version",
			want: "1.2.3",
		},
		{
			name: "nested field then array then field",
			data: map[string]any{
				"data": []any{map[string]any{"name": "a"}, map[string]any{"name": "b"}},
			},
			path: ".data[1].name",
			want: "b",
		},
		{
			name: "empty path returns data as-is",
			data: map[string]any{"x": 1},
			path: "",
			want: map[string]any{"x": 1},
		},
		{name: "unclosed bracket", data: []any{"a"}, path: "[0", wantErr: errPathNavigation},
		{name: "invalid index", data: []any{"a"}, path: "[abc].x", wantErr: errPathNavigation},
		{name: "out of bounds", data: []any{"a"}, path: "[5].x", wantErr: errPathNavigation},
		{
			name: "expected array got object", data: map[string]any{"k": "v"},
			path: "[0].x", wantErr: errPathNavigation,
		},
		{
			name: "expected object got array", data: []any{"a"},
			path: ".field", wantErr: errPathNavigation,
		},
		{
			name:    "key not found",
			data:    map[string]any{"a": 1},
			path:    ".missing",
			wantErr: errPathNavigation,
		},
		{
			name:    "unexpected char",
			data:    map[string]any{"a": 1},
			path:    "x",
			wantErr: errPathNavigation,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := navigatePath(testCase.data, testCase.path)
			if testCase.wantErr != nil {
				if !errors.Is(err, testCase.wantErr) {
					t.Fatalf("want error %v, got %v", testCase.wantErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if fmt.Sprint(got) != fmt.Sprint(testCase.want) {
				t.Errorf("got %v, want %v", got, testCase.want)
			}
		})
	}
}

func TestSplitVersionPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		path      string
		wantArray string
		wantField string
		wantErr   error
	}{
		{name: "root array", path: "[0].version", wantArray: "", wantField: ".version"},
		{
			name:      "nested array",
			path:      ".releases[0].version",
			wantArray: ".releases",
			wantField: ".version",
		},
		{
			name:      "multi-level",
			path:      "[0].files[0].name",
			wantArray: "",
			wantField: ".files[0].name",
		},
		{name: "no bracket", path: ".version", wantErr: errPathNavigation},
		{name: "unclosed bracket", path: "[0.version", wantErr: errPathNavigation},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			arrayPath, fieldPath, err := splitVersionPath(testCase.path)
			if testCase.wantErr != nil {
				if !errors.Is(err, testCase.wantErr) {
					t.Fatalf("want error %v, got %v", testCase.wantErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if arrayPath != testCase.wantArray {
				t.Errorf("arrayPath: got %q, want %q", arrayPath, testCase.wantArray)
			}

			if fieldPath != testCase.wantField {
				t.Errorf("fieldPath: got %q, want %q", fieldPath, testCase.wantField)
			}
		})
	}
}

func TestSearchFileList(t *testing.T) {
	t.Parallel()

	matcher := &FileMatch{
		FilenameField:    "name",
		FilenameContains: "{{.OS}}-{{.Arch}}",
		ChecksumField:    "sha256",
	}
	data := TemplateData{OS: "linux", Arch: "amd64"}

	tests := []struct {
		name    string
		list    []any
		want    fileMatchResult
		wantErr error
	}{
		{
			name: "match found",
			list: []any{map[string]any{"name": "tool-linux-amd64.tar.gz", "sha256": "abc123"}},
			want: fileMatchResult{filename: "tool-linux-amd64.tar.gz", checksum: "abc123"},
		},
		{
			name:    "no match",
			list:    []any{map[string]any{"name": "tool-windows-amd64.zip", "sha256": "def456"}},
			wantErr: errNoFileMatch,
		},
		{
			name: "skip non-objects",
			list: []any{"a string", 42, map[string]any{
				"name": "tool-linux-amd64.tar.gz", "sha256": "matched",
			}},
			want: fileMatchResult{filename: "tool-linux-amd64.tar.gz", checksum: "matched"},
		},
		{
			name:    "empty list",
			list:    []any{},
			wantErr: errNoFileMatch,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := searchFileList(testCase.list, matcher, data, "1.0.0")
			if testCase.wantErr != nil {
				if !errors.Is(err, testCase.wantErr) {
					t.Fatalf("want error %v, got %v", testCase.wantErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != testCase.want {
				t.Errorf("got %+v, want %+v", got, testCase.want)
			}
		})
	}
}

func TestFindVersionEntry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		root    any
		vCfg    VersionExtract
		version string
		wantErr error
	}{
		{
			name:    "root array match",
			root:    []any{map[string]any{"version": "1.0"}, map[string]any{"version": "2.0"}},
			vCfg:    VersionExtract{Path: "[0].version"},
			version: "2.0",
		},
		{
			name: "nested array with strip prefix",
			root: map[string]any{"releases": []any{
				map[string]any{"version": "v1.0"}, map[string]any{"version": "v2.0"},
			}},
			vCfg:    VersionExtract{Path: ".releases[0].version", StripPrefix: "v"},
			version: "2.0",
		},
		{
			name:    "version not found",
			root:    []any{map[string]any{"version": "1.0"}},
			vCfg:    VersionExtract{Path: "[0].version"},
			version: "9.9.9",
			wantErr: errVersionNotFound,
		},
		{
			name:    "root is not array",
			root:    map[string]any{"key": "val"},
			vCfg:    VersionExtract{Path: "[0].version"},
			version: "1.0",
			wantErr: errPathNavigation,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := findVersionEntry(testCase.root, testCase.vCfg, testCase.version)
			if testCase.wantErr != nil {
				if !errors.Is(err, testCase.wantErr) {
					t.Fatalf("want error %v, got %v", testCase.wantErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got == nil {
				t.Fatal("expected non-nil entry")
			}
		})
	}
}

func TestPickVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		versions []string
		strategy string
		want     string
	}{
		{
			name: "highest picks max", versions: []string{"3.10.5", "3.13.3", "3.12.9", "3.11.12"},
			strategy: "highest", want: "3.13.3",
		},
		{name: "single version", versions: []string{"3.13.3"}, strategy: "highest", want: "3.13.3"},
		{
			name:     "default picks first",
			versions: []string{"3.10.5", "3.13.3"},
			strategy: "",
			want:     "3.10.5",
		},
		{name: "empty list", versions: []string{}, strategy: "highest", want: ""},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got := pickVersion(testCase.versions, testCase.strategy); got != testCase.want {
				t.Errorf("pickVersion() = %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestExtractVersionFromAssets(t *testing.T) {
	t.Parallel()

	def := Definition{
		Source: Source{
			Version: VersionExtract{
				AssetRegex: `cpython-([0-9]+\.[0-9]+\.[0-9]+)\+.*-{{.Arch}}-{{.OS}}-install_only\.tar\.gz`,
				Pick:       "highest",
			},
		},
	}
	data := TemplateData{OS: "apple-darwin", Arch: "aarch64"}
	rel := githubRelease{
		TagName: "20260504",
		Assets: []githubAsset{
			{Name: "cpython-3.10.20+20260504-aarch64-apple-darwin-install_only.tar.gz"},
			{Name: "cpython-3.10.20+20260504-aarch64-apple-darwin-debug-full.tar.zst"},
			{Name: "cpython-3.12.9+20260504-aarch64-apple-darwin-install_only.tar.gz"},
			{Name: "cpython-3.13.3+20260504-aarch64-apple-darwin-install_only.tar.gz"},
			{Name: "cpython-3.13.3+20260504-x86_64-apple-darwin-install_only.tar.gz"},
		},
	}

	result, err := extractVersionFromAssets(rel, def, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Version != "3.13.3" {
		t.Errorf("version = %q, want %q", result.Version, "3.13.3")
	}

	if result.Tag != "20260504" {
		t.Errorf("tag = %q, want %q", result.Tag, "20260504")
	}
}

func TestExtractVersionFromAssetsNoMatch(t *testing.T) {
	t.Parallel()

	def := Definition{
		Source: Source{
			Version: VersionExtract{
				AssetRegex: `cpython-([0-9]+\.[0-9]+\.[0-9]+)\+.*-{{.Arch}}-{{.OS}}-install_only\.tar\.gz`,
				Pick:       "highest",
			},
		},
	}
	rel := githubRelease{
		TagName: "20260504",
		Assets: []githubAsset{
			{Name: "cpython-3.13.3+20260504-aarch64-apple-darwin-install_only.tar.gz"},
		},
	}

	_, err := extractVersionFromAssets(rel, def, TemplateData{OS: "windows", Arch: "x86_64"})
	if err == nil {
		t.Fatal("expected error for no matching assets, got nil")
	}
}
