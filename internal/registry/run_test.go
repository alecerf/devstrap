package registry_test

import (
	"context"
	"errors"
	"testing"

	"github.com/alecerf/devstrap/internal/registry"
)

const (
	versionOne = "1.0.0"
	versionTwo = "2.0.0"
	extraData  = "extra-data"
)

var errMock = errors.New("mock error")

// mockTool implements the registry.Tool interface for testing.
type mockTool struct {
	name           string
	fetchLatest    func() (string, string, error)
	fetchVersion   func(string) (string, error)
	currentVersion func() (string, error)
	install        func(string, string) error
}

func (m *mockTool) Name() string { return m.name }

func (m *mockTool) FetchLatest(_ context.Context) (string, string, error) {
	return m.fetchLatest()
}

func (m *mockTool) FetchVersion(_ context.Context, version string) (string, error) {
	return m.fetchVersion(version)
}

func (m *mockTool) CurrentVersion(_ context.Context) (string, error) {
	return m.currentVersion()
}

func (m *mockTool) Install(_ context.Context, _ func(string), version, extra string) error {
	return m.install(version, extra)
}

func noop(string) {}

func assertCheckInfo(t *testing.T, got, want registry.CheckInfo) {
	t.Helper()

	if got.Name != want.Name {
		t.Errorf("Name = %q, want %q", got.Name, want.Name)
	}

	if got.Current != want.Current {
		t.Errorf("Current = %q, want %q", got.Current, want.Current)
	}

	if got.Latest != want.Latest {
		t.Errorf("Latest = %q, want %q", got.Latest, want.Latest)
	}

	if got.UpToDate != want.UpToDate {
		t.Errorf("UpToDate = %v, want %v", got.UpToDate, want.UpToDate)
	}

	if !errors.Is(got.Err, want.Err) {
		t.Errorf("Err = %v, want %v", got.Err, want.Err)
	}
}

func assertResult(t *testing.T, got, want registry.Result) {
	t.Helper()

	if got.Name != want.Name {
		t.Errorf("Name = %q, want %q", got.Name, want.Name)
	}

	if got.Status != want.Status {
		t.Errorf("Status = %q, want %q", got.Status, want.Status)
	}

	if got.Version != want.Version {
		t.Errorf("Version = %q, want %q", got.Version, want.Version)
	}

	if !errors.Is(got.Err, want.Err) {
		t.Errorf("Err = %v, want %v", got.Err, want.Err)
	}
}

func TestIsNewer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		latest  string
		current string
		want    bool
	}{
		{"newer", "1.2.4", "1.2.3", true},
		{"equal", "1.2.3", "1.2.3", false},
		{"older", "1.2.3", "1.2.4", false},
		{"equal strings fallback", "not-semver", "not-semver", false},
		{"different strings fallback", "b", "a", true},
		{"one invalid strings differ", versionOne, "invalid", true},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := registry.ExportIsNewer(testCase.latest, testCase.current)
			if got != testCase.want {
				t.Errorf("isNewer(%q, %q) = %v, want %v",
					testCase.latest, testCase.current, got, testCase.want)
			}
		})
	}
}

type checkTestCase struct {
	name     string
	tool     *mockTool
	wantInfo registry.CheckInfo
}

func TestCheck(t *testing.T) {
	t.Parallel()

	tests := []checkTestCase{
		{
			name: "up to date",
			tool: &mockTool{
				name:           "test-tool",
				fetchLatest:    func() (string, string, error) { return versionOne, "", nil },
				currentVersion: func() (string, error) { return versionOne, nil },
			},
			wantInfo: registry.CheckInfo{
				Name: "test-tool", Current: versionOne, Latest: versionOne, UpToDate: true,
			},
		},
		{
			name: "needs update",
			tool: &mockTool{
				name:           "test-tool",
				fetchLatest:    func() (string, string, error) { return versionTwo, "", nil },
				currentVersion: func() (string, error) { return versionOne, nil },
			},
			wantInfo: registry.CheckInfo{
				Name: "test-tool", Current: versionOne, Latest: versionTwo, UpToDate: false,
			},
		},
		{
			name: "fetch error",
			tool: &mockTool{
				name:        "test-tool",
				fetchLatest: func() (string, string, error) { return "", "", errMock },
			},
			wantInfo: registry.CheckInfo{Name: "test-tool", Err: errMock},
		},
		{
			name: "not installed",
			tool: &mockTool{
				name:           "test-tool",
				fetchLatest:    func() (string, string, error) { return versionOne, "", nil },
				currentVersion: func() (string, error) { return "", errMock },
			},
			wantInfo: registry.CheckInfo{
				Name: "test-tool", Current: "", Latest: versionOne, UpToDate: false,
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			assertCheckInfo(
				t,
				registry.Check(context.Background(), testCase.tool, noop),
				testCase.wantInfo,
			)
		})
	}
}

type runTestCase struct {
	name       string
	tool       *mockTool
	wantResult registry.Result
}

func TestRun(t *testing.T) {
	t.Parallel()

	tests := []runTestCase{
		{
			name: "up to date",
			tool: &mockTool{
				name:           "test-tool",
				fetchLatest:    func() (string, string, error) { return versionOne, "", nil },
				currentVersion: func() (string, error) { return versionOne, nil },
			},
			wantResult: registry.Result{
				Name:    "test-tool",
				Status:  "up-to-date",
				Version: versionOne,
			},
		},
		{
			name: "installs new version",
			tool: &mockTool{
				name:           "test-tool",
				fetchLatest:    func() (string, string, error) { return versionTwo, extraData, nil },
				currentVersion: func() (string, error) { return versionOne, nil },
				install:        func(string, string) error { return nil },
			},
			wantResult: registry.Result{
				Name:    "test-tool",
				Status:  "installed",
				Version: versionTwo,
			},
		},
		{
			name: "installs when not installed",
			tool: &mockTool{
				name:           "test-tool",
				fetchLatest:    func() (string, string, error) { return versionOne, extraData, nil },
				currentVersion: func() (string, error) { return "", errMock },
				install:        func(string, string) error { return nil },
			},
			wantResult: registry.Result{
				Name:    "test-tool",
				Status:  "installed",
				Version: versionOne,
			},
		},
		{
			name: "fetch error",
			tool: &mockTool{
				name:        "test-tool",
				fetchLatest: func() (string, string, error) { return "", "", errMock },
			},
			wantResult: registry.Result{Name: "test-tool", Err: errMock},
		},
		{
			name: "install error",
			tool: &mockTool{
				name:           "test-tool",
				fetchLatest:    func() (string, string, error) { return versionTwo, extraData, nil },
				currentVersion: func() (string, error) { return versionOne, nil },
				install:        func(string, string) error { return errMock },
			},
			wantResult: registry.Result{Name: "test-tool", Err: errMock},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			assertResult(
				t,
				registry.Run(context.Background(), testCase.tool, noop),
				testCase.wantResult,
			)
		})
	}
}

func TestRunVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		tool       *mockTool
		version    string
		wantResult registry.Result
	}{
		{
			name: "installs specified version",
			tool: &mockTool{
				name:         "test-tool",
				fetchVersion: func(string) (string, error) { return extraData, nil },
				install:      func(string, string) error { return nil },
			},
			version:    "1.5.0",
			wantResult: registry.Result{Name: "test-tool", Status: "installed", Version: "1.5.0"},
		},
		{
			name: "fetch error",
			tool: &mockTool{
				name:         "test-tool",
				fetchVersion: func(string) (string, error) { return "", errMock },
			},
			version:    "1.5.0",
			wantResult: registry.Result{Name: "test-tool", Err: errMock},
		},
		{
			name: "install error",
			tool: &mockTool{
				name:         "test-tool",
				fetchVersion: func(string) (string, error) { return extraData, nil },
				install:      func(string, string) error { return errMock },
			},
			version:    "1.5.0",
			wantResult: registry.Result{Name: "test-tool", Err: errMock},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			assertResult(
				t,
				registry.RunVersion(context.Background(), testCase.tool, noop, testCase.version),
				testCase.wantResult,
			)
		})
	}
}
