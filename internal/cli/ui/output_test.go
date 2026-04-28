package ui_test

import (
	"testing"
	"time"

	"github.com/alecerf/devstrap/internal/cli/ui"
)

func TestFormatDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		dur  time.Duration
		want string
	}{
		{name: "zero", dur: 0, want: "0.0s"},
		{name: "2.5 seconds", dur: 2500 * time.Millisecond, want: "2.5s"},
		{name: "59.9 seconds", dur: 59900 * time.Millisecond, want: "59.9s"},
		{name: "exactly 60 seconds", dur: 60 * time.Second, want: "1m 0s"},
		{name: "90 seconds", dur: 90 * time.Second, want: "1m 30s"},
		{name: "150 seconds", dur: 150 * time.Second, want: "2m 30s"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got := ui.FormatDuration(testCase.dur); got != testCase.want {
				t.Errorf("FormatDuration(%v) = %q, want %q", testCase.dur, got, testCase.want)
			}
		})
	}
}

func TestPrinterPad(t *testing.T) {
	t.Parallel()

	t.Run("multiple names", func(t *testing.T) {
		t.Parallel()

		printer := ui.NewPrinter([]string{"go", "nodejs", "k9s"})

		if got := printer.Pad("go"); got != "go    " {
			t.Errorf("Pad(\"go\") = %q (len %d), want %q (len 6)", got, len(got), "go    ")
		}

		if got := printer.Pad("nodejs"); got != "nodejs" {
			t.Errorf("Pad(\"nodejs\") = %q (len %d), want %q (len 6)", got, len(got), "nodejs")
		}
	})

	t.Run("empty names", func(t *testing.T) {
		t.Parallel()

		printer := ui.NewPrinter(nil)

		if got := printer.Pad("go"); got != "go" {
			t.Errorf("Pad(\"go\") = %q (len %d), want %q", got, len(got), "go")
		}
	})

	t.Run("single name", func(t *testing.T) {
		t.Parallel()

		printer := ui.NewPrinter([]string{"ab"})

		if got := printer.Pad("x"); got != "x " {
			t.Errorf("Pad(\"x\") = %q (len %d), want %q (len 2)", got, len(got), "x ")
		}
	})
}
