package tool

import "errors"

var (
	errToolsFailed                  = errors.New("one or more tools failed")
	errUnknownTool                  = errors.New("unknown tool")
	errVersionRequiresSingleTool    = errors.New("--version requires exactly one tool")
	errAllAndToolsMutuallyExclusive = errors.New("--all and tool arguments are mutually exclusive")
	errAllOrToolsRequired           = errors.New("specify tool names or use --all")
	errAllAndVersionConflict        = errors.New("--all and --version are mutually exclusive")
	errDryRunAndVersionConflict     = errors.New("--dry-run and --version are mutually exclusive")
)
