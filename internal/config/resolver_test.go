package config

import (
	"testing"

	"atlassian-cli/internal/types"

	"github.com/stretchr/testify/assert"
)

func TestResolveProject(t *testing.T) {
	tests := []struct {
		name        string
		flagValue   string
		config      *types.Config
		envVar      string
		expected    string
		expectError bool
	}{
		{
			name:      "flag override takes precedence",
			flagValue: "FLAG_PROJECT",
			config: &types.Config{
				DefaultJiraProject: "CONFIG_PROJECT",
			},
			envVar:   "ENV_PROJECT",
			expected: "FLAG_PROJECT",
		},
		{
			name:      "env var when no flag",
			flagValue: "",
			config: &types.Config{
				DefaultJiraProject: "CONFIG_PROJECT",
			},
			envVar:   "ENV_PROJECT",
			expected: "ENV_PROJECT",
		},
		{
			name:      "config when no flag or env",
			flagValue: "",
			config: &types.Config{
				DefaultJiraProject: "CONFIG_PROJECT",
			},
			envVar:   "",
			expected: "CONFIG_PROJECT",
		},
		{
			name:        "error when no value found",
			flagValue:   "",
			config:      &types.Config{},
			envVar:      "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable if provided
			if tt.envVar != "" {
				t.Setenv("ATLASSIAN_DEFAULT_JIRA_PROJECT", tt.envVar)
			}

			result, err := ResolveProject(tt.flagValue, tt.config)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResolveSpace(t *testing.T) {
	tests := []struct {
		name        string
		flagValue   string
		config      *types.Config
		envVar      string
		expected    string
		expectError bool
	}{
		{
			name:      "flag override takes precedence",
			flagValue: "FLAG_SPACE",
			config: &types.Config{
				DefaultConfluenceSpace: "CONFIG_SPACE",
			},
			envVar:   "ENV_SPACE",
			expected: "FLAG_SPACE",
		},
		{
			name:      "env var when no flag",
			flagValue: "",
			config: &types.Config{
				DefaultConfluenceSpace: "CONFIG_SPACE",
			},
			envVar:   "ENV_SPACE",
			expected: "ENV_SPACE",
		},
		{
			name:      "config when no flag or env",
			flagValue: "",
			config: &types.Config{
				DefaultConfluenceSpace: "CONFIG_SPACE",
			},
			envVar:   "",
			expected: "CONFIG_SPACE",
		},
		{
			name:        "error when no value found",
			flagValue:   "",
			config:      &types.Config{},
			envVar:      "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable if provided
			if tt.envVar != "" {
				t.Setenv("ATLASSIAN_DEFAULT_CONFLUENCE_SPACE", tt.envVar)
			}

			result, err := ResolveSpace(tt.flagValue, tt.config)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}