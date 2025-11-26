package generator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/TakumiOkayasu/omusubi-codegen/internal/model"
)

// File permission constants
const (
	DirPermission  = 0755 // Directory permission (rwxr-xr-x)
	FilePermission = 0644 // File permission (rw-r--r--)
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

	// Determine file extension based on source file
	headerExt := "hpp"
	if classInfo.SourceFileExt == "h" {
		headerExt = "h"
	}

	// Generate header file
	if err := g.generateHeader(classInfo, derivedClassName, headerExt); err != nil {
		return fmt.Errorf("failed to generate header: %w", err)
	}

	// Generate source file (always .cpp regardless of source extension)
	sourceExt := "cpp"
	if err := g.generateSource(classInfo, derivedClassName, headerExt, sourceExt); err != nil {
		return fmt.Errorf("failed to generate source: %w", err)
	}

	return nil
}

// generateHeader generates the header file (.hpp or .h)
func (g *Generator) generateHeader(classInfo *model.ClassInfo, derivedClassName string, headerExt string) error {
	tmpl, err := g.loadTemplate("class_header.tmpl")
	if err != nil {
		return err
	}

	data := g.prepareTemplateData(classInfo, derivedClassName)

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	filename := toSnakeCase(derivedClassName) + "." + headerExt
	outputPath := g.getOutputPath(filename)

	if err := os.WriteFile(outputPath, buf.Bytes(), FilePermission); err != nil {
		return fmt.Errorf("failed to write header file: %w", err)
	}

	fmt.Printf("Generated: %s\n", outputPath)
	return nil
}

// generateSource generates the source file (.cpp)
func (g *Generator) generateSource(classInfo *model.ClassInfo, derivedClassName string, headerExt string, sourceExt string) error {
	tmpl, err := g.loadTemplate("class_source.tmpl")
	if err != nil {
		return err
	}

	data := g.prepareTemplateData(classInfo, derivedClassName)
	data["HeaderExt"] = headerExt

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	filename := toSnakeCase(derivedClassName) + "." + sourceExt
	outputPath := g.getOutputPath(filename)

	if err := os.WriteFile(outputPath, buf.Bytes(), FilePermission); err != nil {
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

	// Determine base class extension
	baseClassExt := classInfo.SourceFileExt
	if baseClassExt == "" {
		baseClassExt = "hpp"
	}

	return map[string]interface{}{
		"ClassName":             derivedClassName,
		"BaseClass":             classInfo.Name,
		"BaseClassExt":          baseClassExt,
		"Namespace":             classInfo.Namespace,
		"Methods":               pureVirtualMethods,
		"HasNamespace":          classInfo.Namespace != "",
		"FormatParameters":      formatParameters,
		"FormatMethodSignature": formatMethodSignature,
	}
}

// loadTemplate loads a template file from embedded filesystem
func (g *Generator) loadTemplate(name string) (*template.Template, error) {
	funcMap := template.FuncMap{
		"formatParameters":      formatParameters,
		"formatMethodSignature": formatMethodSignature,
		"toLower":               strings.ToLower,
		"toUpper":               strings.ToUpper,
		"toSnakeCase":           toSnakeCase,
	}

	// 埋め込みファイルシステムからテンプレートを読み込む
	tmplPath := "templates/" + name
	tmplContent, err := templatesFS.ReadFile(tmplPath)

	if err != nil {
		return nil, fmt.Errorf("failed to read embedded template %s: %w", name, err)
	}

	tmpl, err := template.New(name).Funcs(funcMap).Parse(string(tmplContent))

	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
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

// ensureOutputDir creates the output directory if it doesn't exist
func (g *Generator) ensureOutputDir() error {
	if err := os.MkdirAll(g.outputDir, DirPermission); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	return nil
}

// getOutputPath returns the full path for an output file
func (g *Generator) getOutputPath(filename string) string {
	return filepath.Join(g.outputDir, filename)
}

// toSnakeCase converts CamelCase or PascalCase to snake_case
// Preserves number+uppercase letter combinations (e.g., M5Stack -> m5stack)
func toSnakeCase(s string) string {
	// Insert underscore before uppercase letters that follow lowercase letters
	// but NOT after numbers (to keep M5Stack as m5stack, not m5_stack)
	re := regexp.MustCompile("([a-z])([A-Z])")
	snake := re.ReplaceAllString(s, "${1}_${2}")

	// Convert to lowercase
	return strings.ToLower(snake)
}
