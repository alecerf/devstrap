package registry

import (
	"fmt"
	"runtime"
	"slices"
)

// DetectPlatform returns the current runtime OS and architecture.
func DetectPlatform() Platform {
	return Platform{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
}

// mapPlatform maps the runtime OS and architecture to tool-specific values
// using the definition's platform configuration.
func mapPlatform(plat Platform, def Definition) (string, string, error) {
	if !slices.Contains(def.Platforms.OS, plat.OS) {
		return "", "", fmt.Errorf("%w %q for tool %q", errUnsupportedOS, plat.OS, def.Name)
	}

	mapped, ok := def.Platforms.Arch[plat.Arch]
	if !ok {
		return "", "", fmt.Errorf("%w %q for tool %q", errUnsupportedArch, plat.Arch, def.Name)
	}

	return plat.OS, mapped, nil
}
