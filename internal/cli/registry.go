package cli

import (
	"fmt"

	"github.com/alecerf/devstrap/internal/index"
	"github.com/alecerf/devstrap/internal/tool"
)

// ToolNames returns the tool names from the loaded index.
func ToolNames() []string {
	idx, err := index.Load()
	if err != nil {
		return nil
	}

	return idx.ToolNames()
}

// BuildTools constructs all Tool instances from the index, keyed by name.
// Returns the ordered list of names alongside the map.
func BuildTools(baseDir string, plat tool.Platform) ([]string, map[string]tool.Tool, error) {
	idx, err := index.Load()
	if err != nil {
		return nil, nil, fmt.Errorf("load index: %w", err)
	}

	names := make([]string, 0, len(idx.Definitions))
	tools := make(map[string]tool.Tool, len(idx.Definitions))

	for _, def := range idx.Definitions {
		t, err := index.NewTool(def, baseDir, plat)
		if err != nil {
			fmt.Printf("  %s skipping %s: %v\n", yellowBold("•"), def.Name, err)

			continue
		}

		names = append(names, def.Name)
		tools[def.Name] = t
	}

	return names, tools, nil
}
