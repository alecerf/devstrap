package registry

import (
	"bytes"
	"fmt"
	"text/template"
)

// TemplateData holds the variables available to Go templates in tool definitions.
type TemplateData struct {
	Version  string
	OS       string
	Arch     string
	Name     string
	Filename string
	Tag      string
	Source   SourceData
}

// SourceData holds source-specific template variables.
type SourceData struct {
	Owner string
	Repo  string
}

// renderTemplate executes a Go text/template string with the given data.
func renderTemplate(tmpl string, data TemplateData) (string, error) {
	t, err := template.New("").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("parse template %q: %w", tmpl, err)
	}

	var buf bytes.Buffer

	err = t.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("execute template %q: %w", tmpl, err)
	}

	return buf.String(), nil
}
