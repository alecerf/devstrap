package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

var (
	greenBold  = color.New(color.FgGreen, color.Bold).SprintFunc()
	redBold    = color.New(color.FgRed, color.Bold).SprintFunc()
	yellowBold = color.New(color.FgYellow, color.Bold).SprintFunc()
	bold       = color.New(color.Bold).SprintFunc()
	dim        = color.New(color.Faint).SprintFunc()
)

// printer formats tool output with aligned names.
type printer struct {
	width int
}

func newPrinter(names []string) printer {
	w := 0

	for _, n := range names {
		if len(n) > w {
			w = len(n)
		}
	}

	return printer{width: w}
}

func (p printer) pad(name string) string {
	return fmt.Sprintf("%-*s", p.width, name)
}

func (p printer) printSuccess(name, msg string) {
	fmt.Printf("  %s %s  %s\n", greenBold("✔"), bold(p.pad(name)), msg)
}

func (p printer) printError(name string, err error) {
	fmt.Printf("  %s %s  %v\n", redBold("✖"), bold(p.pad(name)), err)
}

func (p printer) printInfo(name, msg string) {
	fmt.Printf("  %s %s  %s\n", yellowBold("•"), bold(p.pad(name)), msg)
}

func (p printer) printSummary(total int, label string, count, upToDate, failed int, elapsed time.Duration) {
	parts := []string{bold(strconv.Itoa(total)) + " checked"}

	if count > 0 {
		parts = append(parts, greenBold(strconv.Itoa(count))+" "+label)
	}

	if upToDate > 0 {
		parts = append(parts, greenBold(strconv.Itoa(upToDate))+" up-to-date")
	}

	if failed > 0 {
		parts = append(parts, redBold(strconv.Itoa(failed))+" failed")
	}

	fmt.Printf("\n%s %s\n",
		strings.Join(parts, ", "),
		dim("("+formatDuration(elapsed)+")"),
	)
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}

	m := int(d.Minutes())
	s := int(d.Seconds()) % 60

	return fmt.Sprintf("%dm %ds", m, s)
}
