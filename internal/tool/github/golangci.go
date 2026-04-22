package github

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/alecerf/devstrap/internal/tool"
)

var errGolangCIVersionParse = errors.New("could not parse golangci-lint version")

// NewGolangCI returns a Tool that installs or updates golangci-lint.
func NewGolangCI(baseDir string, plat tool.Platform) tool.Tool { //nolint:ireturn // factory function
	return &Tool{
		Owner:    "golangci",
		Repo:     "golangci-lint",
		BinName:  "golangci-lint",
		BinDir:   filepath.Join(baseDir, "bin"),
		Platform: plat,
		ParseVersion: func(output string) (string, error) {
			// "golangci-lint has version 2.11.4 built with ..."
			fields := strings.Fields(output)
			if len(fields) < 4 {
				return "", errGolangCIVersionParse
			}

			return strings.TrimPrefix(fields[3], "v"), nil
		},
	}
}
