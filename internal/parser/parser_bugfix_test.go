package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err, "ParseSource failed with valid input")
	require.NotNil(t, fileInfo, "ParseSource returned nil fileInfo")
	require.NotEmpty(t, fileInfo.Classes, "ParseSource did not find any classes")

	classInfo := fileInfo.Classes[0]
	assert.Equal(t, "IDevice", classInfo.Name)
	assert.True(t, classInfo.IsAbstract, "Expected class to be marked as abstract")
}

// TestParseSourceWithEmptyInput tests that ParseSource handles empty input gracefully
func TestParseSourceWithEmptyInput(t *testing.T) {
	p := New()

	source := []byte("")

	fileInfo, err := p.ParseSource(source)
	require.NoError(t, err, "ParseSource failed with empty input")
	require.NotNil(t, fileInfo, "ParseSource returned nil fileInfo")

	// Empty input should result in no classes found, but should not panic
	assert.Empty(t, fileInfo.Classes, "Expected 0 classes for empty input")
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
			assert.NotPanics(t, func() {
				_, _ = p.ParseSource(tc.source)
			}, "ParseSource panicked with input '%s'", tc.name)
		})
	}
}
