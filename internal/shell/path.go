// Package shell generates PATH configuration for tools managed by devstrap.
package shell

import (
	"os"
	"path/filepath"

	"github.com/alecerf/devstrap/internal/registry"
)

// Paths computes the list of directories that should be added to the user's
// PATH. It always includes binDir, and adds derived bin directories only for
// directory-mode tools whose install directory exists on disk.
func Paths(paths registry.Paths, defs []registry.Definition) []string {
	seen := make(map[string]struct{})

	var dirs []string

	add := func(dir string) {
		if _, ok := seen[dir]; ok {
			return
		}

		if !isDir(dir) {
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

func isDir(path string) bool {
	info, err := os.Stat(path)

	return err == nil && info.IsDir()
}
