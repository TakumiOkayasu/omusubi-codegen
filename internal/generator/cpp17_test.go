package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TakumiOkayasu/omusubi-codegen/internal/model"
)

// Test that generated code includes C++17 features
func TestGenerateImplementation_Cpp17Features(t *testing.T) {
	// Create temporary output directory
	tmpDir, err := os.MkdirTemp("", "generator-cpp17-test-*")
	require.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(tmpDir)

	// Create generator
	cfg := Config{
		TemplateDir: "templates",
		OutputDir:   tmpDir,
	}
	gen := New(cfg)

	// Create ClassInfo
	classInfo := &model.ClassInfo{
		Name:          "IDevice",
		Namespace:     "omusubi",
		SourceFileExt: "hpp",
		Methods: []model.MethodInfo{
			{
				Name:          "initialize",
				ReturnType:    "bool",
				IsPureVirtual: true,
			},
		},
	}

	// Generate implementation
	derivedClassName := "TestDevice"
	err = gen.GenerateImplementation(classInfo, derivedClassName)
	require.NoError(t, err, "GenerateImplementation failed")

	// Read generated header file
	headerPath := filepath.Join(tmpDir, "test_device.hpp")
	content, err := os.ReadFile(headerPath)
	require.NoError(t, err, "Failed to read generated header")

	headerContent := string(content)

	// Check for C++17 features
	tests := []struct {
		name    string
		pattern string
	}{
		{"noexcept destructor", "virtual ~TestDevice() noexcept = default;"},
		{"deleted move constructor", "TestDevice(TestDevice&&) = delete;"},
		{"deleted move assignment", "TestDevice& operator=(TestDevice&&) = delete;"},
		{"C++17 note in comment", "@note C++17 compliant implementation"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Contains(t, headerContent, tt.pattern,
				"Generated header does not contain expected C++17 feature")
		})
	}
}
