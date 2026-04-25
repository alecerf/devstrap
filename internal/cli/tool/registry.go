package tool

import (
	"fmt"

	"github.com/alecerf/devstrap/internal/cli/ui"
	"github.com/alecerf/devstrap/internal/engine"
	"github.com/alecerf/devstrap/internal/registry"
)

func toolNames() []string {
	idx, err := registry.Load()
	if err != nil {
		return nil
	}

	return idx.ToolNames()
}

func buildTools(baseDir string, plat engine.Platform) ([]string, map[string]engine.Tool, error) {
	idx, err := registry.Load()
	if err != nil {
		return nil, nil, fmt.Errorf("load index: %w", err)
	}

	names := make([]string, 0, len(idx.Definitions))
	tools := make(map[string]engine.Tool, len(idx.Definitions))

	for _, def := range idx.Definitions {
		t, err := registry.NewTool(def, baseDir, plat)
		if err != nil {
			fmt.Printf("  %s skipping %s: %v\n", ui.YellowBold("•"), def.Name, err)

			continue
		}

		names = append(names, def.Name)
		tools[def.Name] = t
	}

	return names, tools, nil
}
