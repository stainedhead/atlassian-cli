package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Formatter handles output formatting for different formats
type Formatter struct {
	format string
}

// NewFormatter creates a new formatter with the specified format
func NewFormatter(format string) *Formatter {
	if format == "" {
		format = "table"
	}

	// Validate format
	switch format {
	case "json", "table", "yaml":
		// Valid formats
	default:
		format = "table" // Default fallback
	}

	return &Formatter{
		format: format,
	}
}

// Format formats the data according to the configured format
func (f *Formatter) Format(data interface{}) (string, error) {
	switch f.format {
	case "json":
		return f.formatJSON(data)
	case "yaml":
		return f.formatYAML(data)
	case "table":
		return f.formatTable(data)
	default:
		return f.formatTable(data)
	}
}

// formatJSON formats data as JSON
func (f *Formatter) formatJSON(data interface{}) (string, error) {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(bytes), nil
}

// formatYAML formats data as YAML
func (f *Formatter) formatYAML(data interface{}) (string, error) {
	bytes, err := yaml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal YAML: %w", err)
	}
	return string(bytes), nil
}

// formatTable formats data as a simple table
func (f *Formatter) formatTable(data interface{}) (string, error) {
	switch v := data.(type) {
	case map[string]interface{}:
		return f.formatMapAsTable(v), nil
	case []interface{}:
		if len(v) > 0 {
			return f.formatSliceAsTable(v), nil
		}
		return "No data\n", nil
	case []map[string]interface{}:
		if len(v) > 0 {
			return f.formatMapSliceAsTable(v), nil
		}
		return "No data\n", nil
	default:
		// Fallback to JSON for complex types
		return f.formatJSON(data)
	}
}

// formatMapAsTable formats a single map as a key-value table
func (f *Formatter) formatMapAsTable(data map[string]interface{}) string {
	var buf strings.Builder

	// Find max key length for alignment
	maxKeyLen := 0
	for key := range data {
		if len(key) > maxKeyLen {
			maxKeyLen = len(key)
		}
	}

	// Format as key-value pairs
	for key, value := range data {
		buf.WriteString(fmt.Sprintf("%-*s: %v\n", maxKeyLen, key, value))
	}

	return buf.String()
}

// formatSliceAsTable formats a slice of interfaces as a table
func (f *Formatter) formatSliceAsTable(data []interface{}) string {
	if len(data) == 0 {
		return "No data\n"
	}

	// Check if all items are maps
	allMaps := true
	for _, item := range data {
		if _, ok := item.(map[string]interface{}); !ok {
			allMaps = false
			break
		}
	}

	if allMaps {
		// Convert to []map[string]interface{} and format
		mapSlice := make([]map[string]interface{}, len(data))
		for i, item := range data {
			mapSlice[i] = item.(map[string]interface{})
		}
		return f.formatMapSliceAsTable(mapSlice)
	}

	// Format as simple list
	var buf strings.Builder
	for i, item := range data {
		buf.WriteString(fmt.Sprintf("%d: %v\n", i+1, item))
	}
	return buf.String()
}

// formatMapSliceAsTable formats a slice of maps as a table
func (f *Formatter) formatMapSliceAsTable(data []map[string]interface{}) string {
	if len(data) == 0 {
		return "No data\n"
	}

	// Collect all unique keys
	keySet := make(map[string]bool)
	for _, item := range data {
		for key := range item {
			keySet[key] = true
		}
	}

	// Convert to sorted slice
	var headers []string
	for key := range keySet {
		headers = append(headers, key)
	}

	var buf strings.Builder

	// Write header
	for i, header := range headers {
		if i > 0 {
			buf.WriteString("\t")
		}
		buf.WriteString(strings.ToUpper(header))
	}
	buf.WriteString("\n")

	// Write separator
	for i, header := range headers {
		if i > 0 {
			buf.WriteString("\t")
		}
		buf.WriteString(strings.Repeat("-", len(header)))
	}
	buf.WriteString("\n")

	// Write rows
	for _, item := range data {
		for i, header := range headers {
			if i > 0 {
				buf.WriteString("\t")
			}
			if value, exists := item[header]; exists {
				buf.WriteString(fmt.Sprintf("%v", value))
			} else {
				buf.WriteString("")
			}
		}
		buf.WriteString("\n")
	}

	return buf.String()
}

// PrintToStdout prints formatted data to stdout
func (f *Formatter) PrintToStdout(data interface{}) error {
	output, err := f.Format(data)
	if err != nil {
		return err
	}

	fmt.Fprint(os.Stdout, output)
	return nil
}
