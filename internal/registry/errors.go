package registry

import "errors"

var (
	errUnsupportedSource      = errors.New("unsupported source type")
	errNoReleaseTag           = errors.New("no release tag found")
	errPathNavigation         = errors.New("cannot navigate JSON path")
	errNoFileMatch            = errors.New("no file matched")
	errVersionNotString       = errors.New("version is not a string")
	errUnsupportedInstallMode = errors.New("unsupported install mode")
	errUnsupportedOS          = errors.New("unsupported OS")
	errUnsupportedArch        = errors.New("unsupported arch")
	errVersionRegexNoMatch    = errors.New("version regex did not match")
)
