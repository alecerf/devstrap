// Package ui provides terminal UI helpers for the devstrap CLI.
package ui

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

// Color helper functions for terminal output.
var (
	GreenBold  = color.New(color.FgGreen, color.Bold).SprintFunc()
	RedBold    = color.New(color.FgRed, color.Bold).SprintFunc()
	YellowBold = color.New(color.FgYellow, color.Bold).SprintFunc()
	Bold       = color.New(color.Bold).SprintFunc()
	Dim        = color.New(color.Faint).SprintFunc()
)

const secondsPerMinute = 60

// Printer formats tool output with aligned names.
type Printer struct {
	width int
}

// NewPrinter creates a Printer calibrated to the given name list.
func NewPrinter(names []string) Printer {
	maxWidth := 0

	for _, n := range names {
		if len(n) > maxWidth {
			maxWidth = len(n)
		}
	}

	return Printer{width: maxWidth}
}

// Pad right-pads a name to the column width.
func (p Printer) Pad(name string) string {
	return fmt.Sprintf("%-*s", p.width, name)
}

// ProgressPrefix formats a "[n/m] name" progress prefix.
func (p Printer) ProgressPrefix(i, total int, name string) string {
	return fmt.Sprintf("%s %s",
		Dim(fmt.Sprintf("[%d/%d]", i+1, total)),
		Bold(p.Pad(name)),
	)
}

// PrintSuccess prints a green checkmark line.
func (p Printer) PrintSuccess(name, msg string) {
	_, _ = fmt.Fprintf(os.Stdout, "  %s %s  %s\n", GreenBold("✔"), Bold(p.Pad(name)), msg)
}

// PrintError prints a red cross line.
func (p Printer) PrintError(name string, err error) {
	_, _ = fmt.Fprintf(os.Stdout, "  %s %s  %v\n", RedBold("✖"), Bold(p.Pad(name)), err)
}

// PrintInfo prints a yellow bullet line.
func (p Printer) PrintInfo(name, msg string) {
	_, _ = fmt.Fprintf(os.Stdout, "  %s %s  %s\n", YellowBold("•"), Bold(p.Pad(name)), msg)
}

// PrintSummary prints a summary line with counts and elapsed time.
func (p Printer) PrintSummary(
	total int,
	label string,
	count, upToDate, failed int,
	elapsed time.Duration,
) {
	parts := []string{Bold(strconv.Itoa(total)) + " checked"}

	if count > 0 {
		parts = append(parts, GreenBold(strconv.Itoa(count))+" "+label)
	}

	if upToDate > 0 {
		parts = append(parts, GreenBold(strconv.Itoa(upToDate))+" up-to-date")
	}

	if failed > 0 {
		parts = append(parts, RedBold(strconv.Itoa(failed))+" failed")
	}

	_, _ = fmt.Fprintf(os.Stdout, "\n%s %s\n",
		strings.Join(parts, ", "),
		Dim("("+FormatDuration(elapsed)+")"),
	)
}

// FormatDuration formats a duration for human display.
func FormatDuration(dur time.Duration) string {
	if dur < time.Minute {
		return fmt.Sprintf("%.1fs", dur.Seconds())
	}

	m := int(dur.Minutes())
	s := int(dur.Seconds()) % secondsPerMinute

	return fmt.Sprintf("%dm %ds", m, s)
}
