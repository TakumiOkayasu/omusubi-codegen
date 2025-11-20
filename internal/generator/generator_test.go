package generator

import "testing"

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"MyDevice", "my_device"},
		{"MyTestDevice", "my_test_device"},
		{"CustomIDevice", "custom_idevice"},
		{"HTTPServer", "httpserver"},
		{"DeviceID", "device_id"},
		{"IOBuffer", "iobuffer"},
		{"simple", "simple"},
		{"ALLCAPS", "allcaps"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toSnakeCase(tt.input)
			if result != tt.expected {
				t.Errorf("toSnakeCase(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}
