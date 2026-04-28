package registry

// ExportNavigatePath exposes navigatePath for external testing.
var ExportNavigatePath = navigatePath

// ExportSplitVersionPath exposes splitVersionPath for external testing.
var ExportSplitVersionPath = splitVersionPath

// ExportFindVersionEntry exposes findVersionEntry for external testing.
var ExportFindVersionEntry = findVersionEntry

// ExportIsNewer exposes isNewer for external testing.
var ExportIsNewer = isNewer

// ExportMapPlatform exposes mapPlatform for external testing.
var ExportMapPlatform = mapPlatform

// ExportRenderTemplate exposes renderTemplate for external testing.
var ExportRenderTemplate = renderTemplate

// Exported sentinel errors for external testing.
var (
	ErrPathNavigation  = errPathNavigation
	ErrNoFileMatch     = errNoFileMatch
	ErrVersionNotFound = errVersionNotFound
	ErrUnsupportedOS   = errUnsupportedOS
	ErrUnsupportedArch = errUnsupportedArch
)

// ExportFileMatchResult holds exported fields from fileMatchResult.
type ExportFileMatchResult struct {
	Filename string
	Checksum string
}

// ExportSearchFileList wraps searchFileList with exported types.
func ExportSearchFileList(
	list []any, fileMatcher *FileMatch, data TemplateData, version string,
) (ExportFileMatchResult, error) {
	res, err := searchFileList(list, fileMatcher, data, version)

	return ExportFileMatchResult{Filename: res.filename, Checksum: res.checksum}, err
}
