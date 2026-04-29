package updater

import "errors"

// ErrUnsupportedPlatform is returned when the current OS/arch combination has
// no pre-built release asset.
var ErrUnsupportedPlatform = errors.New("unsupported platform")

// errEmptyTag is returned when the GitHub API response has no tag name.
var errEmptyTag = errors.New("empty tag in GitHub release response")
