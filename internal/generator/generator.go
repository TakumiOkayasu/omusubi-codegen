package generator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/TakumiOkayasu/omusubi-platform-codegen/internal/model"
)

// Generator handles code generation
type Generator struct {
	templateDir string
	outputDir   string
}

// Config holds generator configuration
type Config struct {
	TemplateDir string
	OutputDir   string
	WithTests   bool
	WithDocs    bool
}

// New creates a new code generator
func New(cfg Config) *Generator {
	return &Generator{
		templateDir: cfg.TemplateDir,
		outputDir:   cfg.OutputDir,
	}
}

// GenerateImplementation generates C++ implementation files from class info
func (g *Generator) GenerateImplementation(classInfo *model.ClassInfo, derivedClassName string) error {
	if err := g.ensureOutputDir(); err != nil {
		return err
	}

	// Generate header file
	if err := g.generateHeader(classInfo, derivedClassName); err != nil {
		return fmt.Errorf("failed to generate header: %w", err)
	}

	// Generate source file
	if err := g.generateSource(classInfo, derivedClassName); err != nil {
		return fmt.Errorf("failed to generate source: %w", err)
	}

	return nil
}

// generateHeader generates the .hpp file
func (g *Generator) generateHeader(classInfo *model.ClassInfo, derivedClassName string) error {
	tmpl, err := g.loadTemplate("class_header.tmpl")
	if err != nil {
		return err
	}

	data := g.prepareTemplateData(classInfo, derivedClassName)

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	filename := strings.ToLower(derivedClassName) + ".hpp"
	outputPath := g.getOutputPath(filename)

	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write header file: %w", err)
	}

	fmt.Printf("Generated: %s\n", outputPath)
	return nil
}

// generateSource generates the .cpp file
func (g *Generator) generateSource(classInfo *model.ClassInfo, derivedClassName string) error {
	tmpl, err := g.loadTemplate("class_source.tmpl")
	if err != nil {
		return err
	}

	data := g.prepareTemplateData(classInfo, derivedClassName)

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	filename := strings.ToLower(derivedClassName) + ".cpp"
	outputPath := g.getOutputPath(filename)

	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write source file: %w", err)
	}

	fmt.Printf("Generated: %s\n", outputPath)
	return nil
}

// prepareTemplateData prepares data for template execution
func (g *Generator) prepareTemplateData(classInfo *model.ClassInfo, derivedClassName string) map[string]interface{} {
	// Filter only pure virtual methods
	pureVirtualMethods := []model.MethodInfo{}
	for _, method := range classInfo.Methods {
		if method.IsPureVirtual {
			pureVirtualMethods = append(pureVirtualMethods, method)
		}
	}

	guardName := strings.ToUpper(derivedClassName) + "_HPP_"

	return map[string]interface{}{
		"ClassName":          derivedClassName,
		"BaseClass":          classInfo.Name,
		"Namespace":          classInfo.Namespace,
		"Methods":            pureVirtualMethods,
		"GuardName":          guardName,
		"HasNamespace":       classInfo.Namespace != "",
		"FormatParameters":   formatParameters,
		"FormatMethodSignature": formatMethodSignature,
	}
}

// loadTemplate loads a template file
func (g *Generator) loadTemplate(name string) (*template.Template, error) {
	tmplPath := filepath.Join(g.templateDir, name)

	funcMap := template.FuncMap{
		"formatParameters":      formatParameters,
		"formatMethodSignature": formatMethodSignature,
		"toLower":              strings.ToLower,
		"toUpper":              strings.ToUpper,
	}

	tmpl, err := template.New(name).Funcs(funcMap).ParseFiles(tmplPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", tmplPath, err)
	}

	return tmpl, nil
}

// formatParameters formats method parameters for C++
func formatParameters(params []model.ParameterInfo) string {
	if len(params) == 0 {
		return ""
	}

	var parts []string
	for _, param := range params {
		part := param.Type
		if param.Name != "" {
			part += " " + param.Name
		}
		if param.DefaultValue != "" {
			part += " = " + param.DefaultValue
		}
		parts = append(parts, part)
	}

	return strings.Join(parts, ", ")
}

// formatMethodSignature formats a complete method signature
func formatMethodSignature(method model.MethodInfo, includeVirtual, includeOverride bool) string {
	var parts []string

	if includeVirtual && method.IsVirtual {
		parts = append(parts, "virtual")
	}

	if method.ReturnType != "" {
		parts = append(parts, method.ReturnType)
	} else {
		parts = append(parts, "void")
	}

	signature := method.Name + "(" + formatParameters(method.Parameters) + ")"
	parts = append(parts, signature)

	if method.IsConst {
		parts = append(parts, "const")
	}

	if includeOverride {
		parts = append(parts, "override")
	}

	return strings.Join(parts, " ")
}

// GenerateTests generates Google Test test files
func (g *Generator) GenerateTests(fileInfo *model.FileInfo) error {
	// TODO: Implement test generation logic
	return fmt.Errorf("not implemented")
}

// ensureOutputDir creates the output directory if it doesn't exist
func (g *Generator) ensureOutputDir() error {
	if err := os.MkdirAll(g.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	return nil
}

// getOutputPath returns the full path for an output file
func (g *Generator) getOutputPath(filename string) string {
	return filepath.Join(g.outputDir, filename)
}

