package parser

import (
	"testing"
)

// TestParseSourceWithValidInput tests that ParseSource properly handles valid C++ code
func TestParseSourceWithValidInput(t *testing.T) {
	p := New()

	source := []byte(`
namespace omusubi {
class IDevice {
public:
    virtual ~IDevice() = default;
    virtual void initialize() = 0;
    virtual int read(uint8_t* buffer, size_t size) = 0;
};
}
`)

	fileInfo, err := p.ParseSource(source)
	if err != nil {
		t.Fatalf("ParseSource failed with valid input: %v", err)
	}

	if fileInfo == nil {
		t.Fatal("ParseSource returned nil fileInfo")
	}

	if len(fileInfo.Classes) == 0 {
		t.Fatal("ParseSource did not find any classes")
	}

	classInfo := fileInfo.Classes[0]
	if classInfo.Name != "IDevice" {
		t.Errorf("Expected class name 'IDevice', got '%s'", classInfo.Name)
	}

	if !classInfo.IsAbstract {
		t.Error("Expected class to be marked as abstract")
	}
}

// TestParseSourceWithEmptyInput tests that ParseSource handles empty input gracefully
func TestParseSourceWithEmptyInput(t *testing.T) {
	p := New()

	source := []byte("")

	fileInfo, err := p.ParseSource(source)
	if err != nil {
		t.Fatalf("ParseSource failed with empty input: %v", err)
	}

	if fileInfo == nil {
		t.Fatal("ParseSource returned nil fileInfo")
	}

	// Empty input should result in no classes found, but should not panic
	if len(fileInfo.Classes) != 0 {
		t.Errorf("Expected 0 classes for empty input, got %d", len(fileInfo.Classes))
	}
}

// TestParseSourceDoesNotPanic tests that ParseSource does not panic with various inputs
func TestParseSourceDoesNotPanic(t *testing.T) {
	p := New()

	testCases := []struct {
		name   string
		source []byte
	}{
		{"empty", []byte("")},
		{"whitespace only", []byte("   \n\t\n   ")},
		{"comment only", []byte("// This is a comment\n/* Another comment */")},
		{"invalid syntax", []byte("class {{{")},
		{"partial class", []byte("class Foo {")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("ParseSource panicked with input '%s': %v", tc.name, r)
				}
			}()

			_, err := p.ParseSource(tc.source)
			// We don't care about errors here, just that it doesn't panic
			_ = err
		})
	}
}
