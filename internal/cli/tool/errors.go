package tool

import "errors"

var (
	errToolsFailed                  = errors.New("one or more tools failed")
	errUnknownTool                  = errors.New("unknown tool")
	errAllAndToolsMutuallyExclusive = errors.New("--all and tool arguments are mutually exclusive")
	errAllOrToolsRequired           = errors.New("specify tool names or use --all")
	errAllAndVersionConflict        = errors.New("--all and @version are mutually exclusive")
)
