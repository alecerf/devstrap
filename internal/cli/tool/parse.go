package tool

import "strings"

// parseArgs splits positional arguments in the form "tool@version" into
// tool names and a versions map. If no "@" is present, the tool gets an
// empty version string (meaning latest). The special version "latest" is
// treated as empty. A leading "v" prefix in the version is silently stripped.
func parseArgs(args []string) ([]string, map[string]string) {
	versions := make(map[string]string, len(args))
	names := make([]string, 0, len(args))

	for _, arg := range args {
		name, ver := splitAtVersion(arg)
		names = append(names, name)
		versions[name] = ver
	}

	return names, versions
}

// splitAtVersion splits "tool@version" into (tool, version).
// Returns (arg, "") if no "@" is found.
func splitAtVersion(arg string) (string, string) {
	atIdx := strings.LastIndex(arg, "@")
	if atIdx < 1 {
		return arg, ""
	}

	name := arg[:atIdx]
	ver := arg[atIdx+1:]

	if strings.EqualFold(ver, "latest") || ver == "" {
		return name, ""
	}

	ver = strings.TrimPrefix(ver, "v")

	return name, ver
}

// hasVersionedArg reports whether any entry in versions is non-empty.
func hasVersionedArg(versions map[string]string) bool {
	for _, v := range versions {
		if v != "" {
			return true
		}
	}

	return false
}
