package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TakumiOkayasu/omusubi-platform-codegen/internal/model"
)

// Test that .h files generate .cpp source files (not .c)
func TestGenerateImplementation_HFileGeneratesCpp(t *testing.T) {
	// Create temporary output directory
	tmpDir, err := os.MkdirTemp("", "generator-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
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
	if err != nil {
		t.Fatalf("GenerateImplementation failed: %v", err)
	}

	// Check that .h header file was generated
	headerPath := filepath.Join(tmpDir, "test_device.h")
	if _, err := os.Stat(headerPath); os.IsNotExist(err) {
		t.Errorf("Expected header file %s was not generated", headerPath)
	}

	// Check that .cpp source file was generated (NOT .c)
	cppPath := filepath.Join(tmpDir, "test_device.cpp")
	if _, err := os.Stat(cppPath); os.IsNotExist(err) {
		t.Errorf("Expected .cpp file %s was not generated", cppPath)
	}

	// Check that .c file was NOT generated
	cPath := filepath.Join(tmpDir, "test_device.c")
	if _, err := os.Stat(cPath); err == nil {
		t.Errorf("Unexpected .c file %s was generated (should be .cpp)", cPath)
	}
}

// Test that .hpp files still generate .cpp source files
func TestGenerateImplementation_HppFileGeneratesCpp(t *testing.T) {
	// Create temporary output directory
	tmpDir, err := os.MkdirTemp("", "generator-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
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
	if err != nil {
		t.Fatalf("GenerateImplementation failed: %v", err)
	}

	// Check that .hpp header file was generated
	headerPath := filepath.Join(tmpDir, "sample_device.hpp")
	if _, err := os.Stat(headerPath); os.IsNotExist(err) {
		t.Errorf("Expected header file %s was not generated", headerPath)
	}

	// Check that .cpp source file was generated
	cppPath := filepath.Join(tmpDir, "sample_device.cpp")
	if _, err := os.Stat(cppPath); os.IsNotExist(err) {
		t.Errorf("Expected .cpp file %s was not generated", cppPath)
	}
}
