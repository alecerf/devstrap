package registry_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/alecerf/devstrap/internal/registry"
)

type navigatePathCase struct {
	name    string
	data    any
	path    string
	want    any
	wantErr error
}

func navigatePathCases() []navigatePathCase {
	return []navigatePathCase{
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
		{name: "nested field then array then field", data: map[string]any{
			"data": []any{map[string]any{"name": "a"}, map[string]any{"name": "b"}},
		}, path: ".data[1].name", want: "b"},
		{
			name: "empty path returns data as-is",
			data: map[string]any{"x": 1},
			path: "",
			want: map[string]any{"x": 1},
		},
		{
			name: "unclosed bracket", data: []any{"a"},
			path: "[0", wantErr: registry.ErrPathNavigation,
		},
		{
			name: "invalid index", data: []any{"a"},
			path: "[abc].x", wantErr: registry.ErrPathNavigation,
		},
		{
			name: "out of bounds", data: []any{"a"},
			path: "[5].x", wantErr: registry.ErrPathNavigation,
		},
		{
			name: "expected array got object", data: map[string]any{"k": "v"},
			path: "[0].x", wantErr: registry.ErrPathNavigation,
		},
		{
			name: "expected object got array", data: []any{"a"},
			path: ".field", wantErr: registry.ErrPathNavigation,
		},
		{
			name: "key not found", data: map[string]any{"a": 1},
			path: ".missing", wantErr: registry.ErrPathNavigation,
		},
		{
			name: "unexpected char at start", data: map[string]any{"a": 1},
			path: "x", wantErr: registry.ErrPathNavigation,
		},
	}
}

func TestNavigatePath(t *testing.T) {
	t.Parallel()

	for _, testCase := range navigatePathCases() {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := registry.ExportNavigatePath(testCase.data, testCase.path)
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

type splitVersionPathCase struct {
	name      string
	path      string
	wantArray string
	wantField string
	wantErr   error
}

func splitVersionPathCases() []splitVersionPathCase {
	return []splitVersionPathCase{
		{name: "root array", path: "[0].version", wantArray: "", wantField: ".version"},
		{
			name:      "nested array",
			path:      ".releases[0].version",
			wantArray: ".releases",
			wantField: ".version",
		},
		{
			name:      "multi-level field path",
			path:      "[0].files[0].name",
			wantArray: "",
			wantField: ".files[0].name",
		},
		{name: "no bracket", path: ".version", wantErr: registry.ErrPathNavigation},
		{name: "unclosed bracket", path: "[0.version", wantErr: registry.ErrPathNavigation},
	}
}

func TestSplitVersionPath(t *testing.T) {
	t.Parallel()

	for _, testCase := range splitVersionPathCases() {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			arrayPath, fieldPath, err := registry.ExportSplitVersionPath(testCase.path)
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

func newFileMatch() *registry.FileMatch {
	return &registry.FileMatch{
		FilenameField:    "name",
		FilenameContains: "{{.OS}}-{{.Arch}}",
		ChecksumField:    "sha256",
	}
}

type searchFileListCase struct {
	name    string
	list    []any
	matcher *registry.FileMatch
	data    registry.TemplateData
	version string
	want    registry.ExportFileMatchResult
	wantErr error
}

func searchFileListCases() []searchFileListCase {
	templateData := registry.TemplateData{OS: "linux", Arch: "amd64"}

	return []searchFileListCase{
		{
			name:    "match found",
			list:    []any{map[string]any{"name": "tool-linux-amd64.tar.gz", "sha256": "abc123"}},
			matcher: newFileMatch(),
			data:    templateData,
			version: "1.0.0",
			want: registry.ExportFileMatchResult{
				Filename: "tool-linux-amd64.tar.gz",
				Checksum: "abc123",
			},
		},
		{
			name:    "no match",
			list:    []any{map[string]any{"name": "tool-windows-amd64.zip", "sha256": "def456"}},
			matcher: newFileMatch(), data: templateData, version: "1.0.0",
			wantErr: registry.ErrNoFileMatch,
		},
		{
			name: "skip non-objects",
			list: []any{"a string", 42, map[string]any{
				"name": "tool-linux-amd64.tar.gz", "sha256": "matched",
			}},
			matcher: newFileMatch(),
			data:    templateData,
			version: "1.0.0",
			want: registry.ExportFileMatchResult{
				Filename: "tool-linux-amd64.tar.gz",
				Checksum: "matched",
			},
		},
		{
			name: "empty list", list: []any{},
			matcher: newFileMatch(), data: templateData, version: "1.0.0",
			wantErr: registry.ErrNoFileMatch,
		},
	}
}

func TestSearchFileList(t *testing.T) {
	t.Parallel()

	for _, testCase := range searchFileListCases() {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := registry.ExportSearchFileList(
				testCase.list,
				testCase.matcher,
				testCase.data,
				testCase.version,
			)
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

type findVersionEntryCase struct {
	name    string
	root    any
	vCfg    registry.VersionExtract
	version string
	wantErr error
}

func findVersionEntryCases() []findVersionEntryCase {
	return []findVersionEntryCase{
		{
			name:    "root array match",
			root:    []any{map[string]any{"version": "1.0"}, map[string]any{"version": "2.0"}},
			vCfg:    registry.VersionExtract{Path: "[0].version"},
			version: "2.0",
		},
		{
			name: "nested array with strip prefix",
			root: map[string]any{"releases": []any{
				map[string]any{"version": "v1.0"}, map[string]any{"version": "v2.0"},
			}},
			vCfg:    registry.VersionExtract{Path: ".releases[0].version", StripPrefix: "v"},
			version: "2.0",
		},
		{
			name: "version not found", root: []any{map[string]any{"version": "1.0"}},
			vCfg: registry.VersionExtract{Path: "[0].version"}, version: "9.9.9",
			wantErr: registry.ErrVersionNotFound,
		},
		{
			name: "root is not array when expected", root: map[string]any{"key": "val"},
			vCfg: registry.VersionExtract{Path: "[0].version"}, version: "1.0",
			wantErr: registry.ErrPathNavigation,
		},
	}
}

func TestFindVersionEntry(t *testing.T) {
	t.Parallel()

	for _, testCase := range findVersionEntryCases() {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := registry.ExportFindVersionEntry(
				testCase.root,
				testCase.vCfg,
				testCase.version,
			)
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
