package tool

import "strings"

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

func hasVersionedArg(versions map[string]string) bool {
	for _, v := range versions {
		if v != "" {
			return true
		}
	}

	return false
}
