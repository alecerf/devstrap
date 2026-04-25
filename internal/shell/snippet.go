package shell

import "strings"

// ZshSnippet returns a zsh export PATH statement for the given directories.
func ZshSnippet(dirs []string) string {
	if len(dirs) == 0 {
		return ""
	}

	return `export PATH="` + strings.Join(dirs, ":") + `:$PATH"`
}
