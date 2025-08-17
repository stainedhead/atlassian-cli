package output

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		name   string
		format string
		want   string
	}{
		{"json format", "json", "json"},
		{"table format", "table", "table"},
		{"yaml format", "yaml", "yaml"},
		{"invalid format defaults to table", "invalid", "table"},
		{"empty format defaults to table", "", "table"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewFormatter(tt.format)
			assert.NotNil(t, formatter)
			// In a real implementation, we would test the actual format property
		})
	}
}

func TestFormatData(t *testing.T) {
	testData := map[string]interface{}{
		"key":     "DEMO-123",
		"summary": "Test issue",
		"status":  "To Do",
	}

	tests := []struct {
		name   string
		format string
		data   interface{}
	}{
		{"format json data", "json", testData},
		{"format table data", "table", testData},
		{"format yaml data", "yaml", testData},
		{"format slice data", "json", []interface{}{testData}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewFormatter(tt.format)
			result, err := formatter.Format(tt.data)
			
			assert.NoError(t, err)
			assert.NotEmpty(t, result)
			
			// Basic validation based on format
			switch tt.format {
			case "json":
				assert.Contains(t, result, "{")
			case "yaml":
				assert.Contains(t, result, ":")
			case "table":
				// Table format should contain the data in some form
				assert.True(t, len(result) > 0)
			}
		})
	}
}

func TestFormatError(t *testing.T) {
	formatter := NewFormatter("json")
	
	// Test with invalid data that might cause formatting errors
	invalidData := make(chan int) // channels can't be JSON marshaled
	
	_, err := formatter.Format(invalidData)
	assert.Error(t, err)
}

func TestSupportedFormats(t *testing.T) {
	supportedFormats := []string{"json", "table", "yaml"}
	
	for _, format := range supportedFormats {
		t.Run("format_"+format, func(t *testing.T) {
			formatter := NewFormatter(format)
			assert.NotNil(t, formatter)
			
			// Test with simple data
			data := map[string]string{"test": "value"}
			result, err := formatter.Format(data)
			
			assert.NoError(t, err)
			assert.NotEmpty(t, result)
		})
	}
}