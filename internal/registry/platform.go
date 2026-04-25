package registry

import (
	"fmt"
	"slices"

	"github.com/alecerf/devstrap/internal/engine"
)

// mapPlatform maps the runtime OS and architecture to tool-specific values
// using the definition's platform configuration.
func mapPlatform(plat engine.Platform, def Definition) (string, string, error) {
	if !slices.Contains(def.Platforms.OS, plat.OS) {
		return "", "", fmt.Errorf("%w %q for tool %q", errUnsupportedOS, plat.OS, def.Name)
	}

	mapped, ok := def.Platforms.Arch[plat.Arch]
	if !ok {
		return "", "", fmt.Errorf("%w %q for tool %q", errUnsupportedArch, plat.Arch, def.Name)
	}

	return plat.OS, mapped, nil
}
