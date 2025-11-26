package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
		{"M5Stack", "m5stack"},
		{"M5StackConnectableContext", "m5stack_connectable_context"},
		{"M5StackSpan", "m5stack_span"},
		{"ESP32Device", "esp32device"},
		{"STM32Controller", "stm32controller"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toSnakeCase(tt.input)
			assert.Equal(t, tt.expected, result, "toSnakeCase(%q)", tt.input)
		})
	}
}
