package template

import (
	"embed"
	"fmt"
	"html/template"
	"io"

	"github.com/TakumiOkayasu/omusubi-platform-codegen/internal/model"
)

//go:embed templates/*
var templatesFS embed.FS

// Manager handles template loading and rendering
type Manager struct {
	templates *template.Template
}

// New creates a new template manager
func New() (*Manager, error) {
	tmpl, err := template.ParseFS(templatesFS, "templates/*.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &Manager{
		templates: tmpl,
	}, nil
}

// RenderClassImplementation renders a C++ class implementation
func (m *Manager) RenderClassImplementation(w io.Writer, class *model.ClassInfo) error {
	// TODO: Implement template rendering
	return fmt.Errorf("not implemented")
}

// RenderClassTest renders a Google Test test file
func (m *Manager) RenderClassTest(w io.Writer, class *model.ClassInfo) error {
	// TODO: Implement template rendering
	return fmt.Errorf("not implemented")
}

// RenderDocumentation renders documentation comments
func (m *Manager) RenderDocumentation(w io.Writer, class *model.ClassInfo) error {
	// TODO: Implement template rendering
	return fmt.Errorf("not implemented")
}
