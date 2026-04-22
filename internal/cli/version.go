package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags:
//
//	go build -ldflags "-X github.com/alecerf/devstrap/internal/cli.Version=1.0.0"
var Version = "dev"

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the devstrap version",
		Run:   runVersion,
	}
}

func runVersion(_ *cobra.Command, _ []string) {
	fmt.Printf("devstrap %s\n", Version)
}
