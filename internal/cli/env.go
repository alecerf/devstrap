package cli

import (
	"fmt"

	"github.com/alecerf/devstrap/internal/registry"
	"github.com/alecerf/devstrap/internal/shell"
	"github.com/spf13/cobra"
)

func newEnvCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "env",
		Short: "Print PATH configuration for tools managed by devstrap",
		Long: `Outputs a zsh export PATH statement that includes all directories
managed by devstrap (binary directory and tool-specific bin directories).

Add the following line to your ~/.zshrc:
  eval "$(devstrap env)"`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runEnv()
		},
	}
}

func runEnv() error {
	idx, err := registry.Load()
	if err != nil {
		return fmt.Errorf("load index: %w", err)
	}

	dirs := shell.Paths(paths, idx.Definitions)
	fmt.Println(shell.ZshSnippet(dirs))

	return nil
}
