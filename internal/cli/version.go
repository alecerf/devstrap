package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags.
var Version = "dev"

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the devstrap version",
		Run:   runVersion,
	}
}

func runVersion(_ *cobra.Command, _ []string) {
	_, _ = fmt.Fprintf(os.Stdout, "devstrap %s\n", Version)
}
