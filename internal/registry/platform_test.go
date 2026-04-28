package registry_test

import (
	"errors"
	"testing"

	"github.com/alecerf/devstrap/internal/registry"
)

func assertMapPlatform(t *testing.T, gotOS, gotArch, wantOS, wantArch string) {
	t.Helper()

	if gotOS != wantOS {
		t.Errorf("OS = %q, want %q", gotOS, wantOS)
	}

	if gotArch != wantArch {
		t.Errorf("Arch = %q, want %q", gotArch, wantArch)
	}
}

func mapPlatformDef() registry.Definition {
	return registry.Definition{
		Name: "test-tool",
		Platforms: registry.PlatformSpec{
			OS:   []string{"darwin", "linux"},
			Arch: map[string]string{"amd64": "x64", "arm64": "arm64"},
		},
	}
}

func TestMapPlatform(t *testing.T) {
	t.Parallel()

	def := mapPlatformDef()

	tests := []struct {
		name     string
		plat     registry.Platform
		wantOS   string
		wantArch string
		wantErr  error
	}{
		{
			name: "valid darwin amd64", plat: registry.Platform{OS: "darwin", Arch: "amd64"},
			wantOS: "darwin", wantArch: "x64",
		},
		{
			name:    "unsupported OS",
			plat:    registry.Platform{OS: "windows", Arch: "amd64"},
			wantErr: registry.ErrUnsupportedOS,
		},
		{
			name:    "unsupported arch",
			plat:    registry.Platform{OS: "darwin", Arch: "386"},
			wantErr: registry.ErrUnsupportedArch,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			gotOS, gotArch, err := registry.ExportMapPlatform(testCase.plat, def)

			if testCase.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error wrapping %v, got nil", testCase.wantErr)
				}

				if !errors.Is(err, testCase.wantErr) {
					t.Fatalf("expected error wrapping %v, got %v", testCase.wantErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			assertMapPlatform(t, gotOS, gotArch, testCase.wantOS, testCase.wantArch)
		})
	}
}

func TestRenderTemplate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		tmpl    string
		data    registry.TemplateData
		want    string
		wantErr bool
	}{
		{
			name: "simple OS and Arch",
			tmpl: "{{.OS}}-{{.Arch}}", data: registry.TemplateData{OS: "linux", Arch: "amd64"},
			want: "linux-amd64",
		},
		{
			name: "nested source fields",
			tmpl: "{{.Source.Owner}}/{{.Source.Repo}}",
			data: registry.TemplateData{Source: registry.SourceData{Owner: "foo", Repo: "bar"}},
			want: "foo/bar",
		},
		{
			name: "all fields", tmpl: "{{.Name}}-{{.Version}}-{{.Tag}}",
			data: registry.TemplateData{Name: "tool", Version: "1.0", Tag: "v1.0"},
			want: "tool-1.0-v1.0",
		},
		{
			name: "no template syntax", tmpl: "plain-text",
			data: registry.TemplateData{}, want: "plain-text",
		},
		{
			name: "invalid template", tmpl: "{{.Missing",
			data: registry.TemplateData{}, wantErr: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := registry.ExportRenderTemplate(testCase.tmpl, testCase.data)

			if testCase.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != testCase.want {
				t.Errorf("got %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestRelBinDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		def  registry.Definition
		want string
	}{
		{
			name: "directory mode with go binary",
			def: registry.Definition{
				Install: registry.Install{Mode: "directory"},
				Detect:  registry.Detect{Binary: "go/bin/go"},
			},
			want: "go/bin",
		},
		{
			name: "binary mode",
			def: registry.Definition{
				Install: registry.Install{Mode: "binary"},
				Detect:  registry.Detect{Binary: "some-bin"},
			},
			want: "",
		},
		{
			name: "directory mode but empty binary",
			def: registry.Definition{
				Install: registry.Install{Mode: "directory"},
				Detect:  registry.Detect{Binary: ""},
			},
			want: "",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.def.RelBinDir()
			if got != testCase.want {
				t.Errorf("RelBinDir() = %q, want %q", got, testCase.want)
			}
		})
	}
}
