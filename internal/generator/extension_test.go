package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TakumiOkayasu/omusubi-codegen/internal/model"
)

// Test that .h files generate .cpp source files (not .c)
func TestGenerateImplementation_HFileGeneratesCpp(t *testing.T) {
	// Create temporary output directory
	tmpDir, err := os.MkdirTemp("", "generator-test-*")
	require.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(tmpDir)

	// Create generator
	cfg := Config{
		TemplateDir: "templates",
		OutputDir:   tmpDir,
	}
	gen := New(cfg)

	// Create ClassInfo for .h file
	classInfo := &model.ClassInfo{
		Name:          "ITestDevice",
		Namespace:     "test",
		SourceFileExt: "h", // Source is .h file
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

	// Check that .h header file was generated
	headerPath := filepath.Join(tmpDir, "test_device.h")
	assert.FileExists(t, headerPath, "Expected header file was not generated")

	// Check that .cpp source file was generated (NOT .c)
	cppPath := filepath.Join(tmpDir, "test_device.cpp")
	assert.FileExists(t, cppPath, "Expected .cpp file was not generated")

	// Check that .c file was NOT generated
	cPath := filepath.Join(tmpDir, "test_device.c")
	assert.NoFileExists(t, cPath, "Unexpected .c file was generated (should be .cpp)")
}

// Test that .hpp files still generate .cpp source files
func TestGenerateImplementation_HppFileGeneratesCpp(t *testing.T) {
	// Create temporary output directory
	tmpDir, err := os.MkdirTemp("", "generator-test-*")
	require.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(tmpDir)

	// Create generator
	cfg := Config{
		TemplateDir: "templates",
		OutputDir:   tmpDir,
	}
	gen := New(cfg)

	// Create ClassInfo for .hpp file
	classInfo := &model.ClassInfo{
		Name:          "ISampleDevice",
		Namespace:     "sample",
		SourceFileExt: "hpp", // Source is .hpp file
		Methods: []model.MethodInfo{
			{
				Name:          "initialize",
				ReturnType:    "bool",
				IsPureVirtual: true,
			},
		},
	}

	// Generate implementation
	derivedClassName := "SampleDevice"
	err = gen.GenerateImplementation(classInfo, derivedClassName)
	require.NoError(t, err, "GenerateImplementation failed")

	// Check that .hpp header file was generated
	headerPath := filepath.Join(tmpDir, "sample_device.hpp")
	assert.FileExists(t, headerPath, "Expected header file was not generated")

	// Check that .cpp source file was generated
	cppPath := filepath.Join(tmpDir, "sample_device.cpp")
	assert.FileExists(t, cppPath, "Expected .cpp file was not generated")
}
