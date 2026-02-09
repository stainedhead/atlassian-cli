package config

import (
	"testing"

	"atlassian-cli/internal/types"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

			// Create a mock cobra command with flag
			cmd := &cobra.Command{Use: "test"}
			cmd.Flags().String("project", tt.flagValue, "test flag")
			if tt.flagValue != "" {
				cmd.Flags().Set("project", tt.flagValue)
			}

			// Set up viper with config values
			if tt.config != nil && tt.config.DefaultJiraProject != "" {
				viper.Set("default_jira_project", tt.config.DefaultJiraProject)
				defer viper.Set("default_jira_project", "")
			}

			result, err := ResolveProject(cmd)

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

			// Create a mock cobra command with flag
			cmd := &cobra.Command{Use: "test"}
			cmd.Flags().String("space", tt.flagValue, "test flag")
			if tt.flagValue != "" {
				cmd.Flags().Set("space", tt.flagValue)
			}

			// Set up viper with config values
			if tt.config != nil && tt.config.DefaultConfluenceSpace != "" {
				viper.Set("default_confluence_space", tt.config.DefaultConfluenceSpace)
				defer viper.Set("default_confluence_space", "")
			}

			result, err := ResolveSpace(cmd)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
