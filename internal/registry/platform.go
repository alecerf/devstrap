package registry

import (
	"fmt"
	"runtime"
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
	mappedOS, osFound := def.Platforms.OS[plat.OS]
	if !osFound {
		return "", "", fmt.Errorf("%w %q for tool %q", errUnsupportedOS, plat.OS, def.Name)
	}

	mappedArch, archFound := def.Platforms.Arch[plat.Arch]
	if !archFound {
		return "", "", fmt.Errorf("%w %q for tool %q", errUnsupportedArch, plat.Arch, def.Name)
	}

	return mappedOS, mappedArch, nil
}
