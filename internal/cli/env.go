package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecerf/devstrap/internal/registry"
	"github.com/spf13/cobra"
)

func newEnvCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "env",
		Short: "Print PATH configuration for tools managed by devstrap",
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

	dirs := shellPaths(paths, idx.Definitions)
	_, _ = fmt.Fprintln(os.Stdout, zshSnippet(dirs))

	return nil
}

// shellPaths computes directories to add to PATH.
func shellPaths(paths registry.Paths, defs []registry.Definition) []string {
	seen := make(map[string]struct{})

	var dirs []string

	add := func(dir string) {
		if _, ok := seen[dir]; ok {
			return
		}

		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			return
		}

		seen[dir] = struct{}{}
		dirs = append(dirs, dir)
	}

	add(paths.BinDir)

	for i := range defs {
		rel := defs[i].RelBinDir()
		if rel != "" {
			add(filepath.Join(paths.DataDir, rel))
		}
	}

	return dirs
}

func zshSnippet(dirs []string) string {
	if len(dirs) == 0 {
		return ""
	}

	return `export PATH="` + strings.Join(dirs, ":") + `:$PATH"`
}
