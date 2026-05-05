package registry

import (
	"errors"
	"testing"
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

func TestMapPlatform(t *testing.T) {
	t.Parallel()

	def := Definition{
		Name: "test-tool",
		Platforms: PlatformSpec{
			OS:   map[string]string{"darwin": "darwin", "linux": "linux"},
			Arch: map[string]string{"amd64": "x64", "arm64": "arm64"},
		},
	}

	tests := []struct {
		name     string
		plat     Platform
		wantOS   string
		wantArch string
		wantErr  error
	}{
		{
			name: "valid darwin amd64", plat: Platform{OS: "darwin", Arch: "amd64"},
			wantOS: "darwin", wantArch: "x64",
		},
		{
			name: "valid linux arm64", plat: Platform{OS: "linux", Arch: "arm64"},
			wantOS: "linux", wantArch: "arm64",
		},
		{
			name:    "unsupported OS",
			plat:    Platform{OS: "windows", Arch: "amd64"},
			wantErr: errUnsupportedOS,
		},
		{
			name:    "unsupported arch",
			plat:    Platform{OS: "darwin", Arch: "386"},
			wantErr: errUnsupportedArch,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			gotOS, gotArch, err := mapPlatform(testCase.plat, def)

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

func TestMapPlatformWithOSMapping(t *testing.T) {
	t.Parallel()

	def := Definition{
		Name: "python",
		Platforms: PlatformSpec{
			OS:   map[string]string{"darwin": "apple-darwin", "linux": "unknown-linux-gnu"},
			Arch: map[string]string{"amd64": "x86_64", "arm64": "aarch64"},
		},
	}

	tests := []struct {
		name     string
		plat     Platform
		wantOS   string
		wantArch string
	}{
		{
			name:   "darwin maps to apple-darwin",
			plat:   Platform{OS: "darwin", Arch: "arm64"},
			wantOS: "apple-darwin", wantArch: "aarch64",
		},
		{
			name:   "linux maps to unknown-linux-gnu",
			plat:   Platform{OS: "linux", Arch: "amd64"},
			wantOS: "unknown-linux-gnu", wantArch: "x86_64",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			gotOS, gotArch, err := mapPlatform(testCase.plat, def)
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
		data    TemplateData
		want    string
		wantErr bool
	}{
		{
			name: "simple OS and Arch",
			tmpl: "{{.OS}}-{{.Arch}}", data: TemplateData{OS: "linux", Arch: "amd64"},
			want: "linux-amd64",
		},
		{
			name: "nested source fields",
			tmpl: "{{.Source.Owner}}/{{.Source.Repo}}",
			data: TemplateData{Source: SourceData{Owner: "foo", Repo: "bar"}},
			want: "foo/bar",
		},
		{
			name: "all fields", tmpl: "{{.Name}}-{{.Version}}-{{.Tag}}",
			data: TemplateData{Name: "tool", Version: "1.0", Tag: "v1.0"},
			want: "tool-1.0-v1.0",
		},
		{
			name: "no template syntax", tmpl: "plain-text",
			data: TemplateData{}, want: "plain-text",
		},
		{
			name: "invalid template", tmpl: "{{.Missing",
			data: TemplateData{}, wantErr: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := renderTemplate(testCase.tmpl, testCase.data)

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
		def  Definition
		want string
	}{
		{
			name: "directory mode with go binary",
			def: Definition{
				Install: Install{Mode: "directory"},
				Detect:  Detect{Binary: "go/bin/go"},
			},
			want: "go/bin",
		},
		{
			name: "binary mode",
			def: Definition{
				Install: Install{Mode: "binary"},
				Detect:  Detect{Binary: "some-bin"},
			},
			want: "",
		},
		{
			name: "directory mode but empty binary",
			def: Definition{
				Install: Install{Mode: "directory"},
				Detect:  Detect{Binary: ""},
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
