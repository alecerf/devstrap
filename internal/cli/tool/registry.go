package tool

import (
	"fmt"
	"os"
	"strings"

	"github.com/alecerf/devstrap/internal/cli/ui"
	"github.com/alecerf/devstrap/internal/registry"
)

func toolNames() []string {
	idx, err := registry.Load()
	if err != nil {
		return nil
	}

	return idx.ToolNames()
}

// validateAllFlag checks the --all flag against the provided tool arguments.
// It returns the effective args slice: nil when --all is set (meaning all tools),
// or the original args otherwise.
func validateAllFlag(all bool, args []string) ([]string, error) {
	if all && len(args) > 0 {
		return nil, errAllAndToolsMutuallyExclusive
	}

	if !all && len(args) == 0 {
		return nil, errAllOrToolsRequired
	}

	if all {
		return nil, nil
	}

	return args, nil
}

// resolveTools loads all tool definitions, validates the given args (if any),
// and returns the ordered name list filtered to args.
func resolveTools(paths registry.Paths, args []string) ([]string, map[string]registry.Tool, error) {
	plat := registry.DetectPlatform()

	idx, err := registry.Load()
	if err != nil {
		return nil, nil, fmt.Errorf("load index: %w", err)
	}

	names := make([]string, 0, len(idx.Definitions))
	tools := make(map[string]registry.Tool, len(idx.Definitions))

	for _, def := range idx.Definitions {
		tool, toolErr := registry.NewTool(def, paths, plat)
		if toolErr != nil {
			_, _ = fmt.Fprintf(
				os.Stdout,
				"  %s skipping %s: %v\n",
				ui.YellowBold("•"),
				def.Name,
				toolErr,
			)

			continue
		}

		names = append(names, def.Name)
		tools[def.Name] = tool
	}

	if len(args) > 0 {
		for _, name := range args {
			if _, ok := tools[name]; !ok {
				return nil, nil, fmt.Errorf(
					"%w: %q (valid: %s)",
					errUnknownTool,
					name,
					strings.Join(names, ", "),
				)
			}
		}

		names = args
	}

	return names, tools, nil
}
