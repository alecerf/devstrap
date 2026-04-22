package cli

import (
	"github.com/alecerf/devstrap/internal/tool"
	"github.com/alecerf/devstrap/internal/tool/github"
	"github.com/alecerf/devstrap/internal/tool/golang"
	"github.com/alecerf/devstrap/internal/tool/node"
)

// entry pairs a tool name with a constructor.
type entry struct {
	name    string
	newTool func(baseDir string, plat tool.Platform) tool.Tool
}

// registry defines the tools and their install order.
var registry = []entry{
	{"go", golang.New},
	{"node", node.New},
	{"golangci-lint", github.NewGolangCI},
}

// ToolNames returns the registered tool names in order.
func ToolNames() []string {
	names := make([]string, len(registry))
	for i, r := range registry {
		names[i] = r.name
	}

	return names
}

// BuildTools constructs all Tool instances keyed by name and returns the
// ordered list of names alongside the map.
func BuildTools(baseDir string, plat tool.Platform) ([]string, map[string]tool.Tool) {
	names := make([]string, len(registry))
	tools := make(map[string]tool.Tool, len(registry))

	for i, r := range registry {
		names[i] = r.name
		tools[r.name] = r.newTool(baseDir, plat)
	}

	return names, tools
}
