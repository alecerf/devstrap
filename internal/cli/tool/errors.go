package tool

import "errors"

var (
	errToolsFailed               = errors.New("one or more tools failed")
	errUnknownTool               = errors.New("unknown tool")
	errVersionRequiresSingleTool = errors.New("--version requires exactly one tool")
)
