package cli

import "testing"

func TestZshSnippet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		dirs []string
		want string
	}{
		{name: "empty list", dirs: nil, want: ""},
		{
			name: "single dir",
			dirs: []string{"/usr/local/bin"},
			want: `export PATH="/usr/local/bin:$PATH"`,
		},
		{
			name: "multiple dirs",
			dirs: []string{"/a", "/b", "/c"},
			want: `export PATH="/a:/b:/c:$PATH"`,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got := zshSnippet(testCase.dirs); got != testCase.want {
				t.Errorf("zshSnippet(%v) = %q, want %q", testCase.dirs, got, testCase.want)
			}
		})
	}
}
