package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"atlassian-cli/internal/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name           string
		configContent  string
		envVars        map[string]string
		expectedConfig *types.Config
		expectError    bool
	}{
		{
			name: "valid config file",
			configContent: `
api_endpoint: "https://test.atlassian.net"
email: "test@example.com"
token: "test-token"
default_jira_project: "TEST"
default_confluence_space: "DEV"
timeout: "60s"
output: "json"
debug: true
`,
			expectedConfig: &types.Config{
				APIEndpoint:            "https://test.atlassian.net",
				Email:                  "test@example.com",
				Token:                  "test-token",
				DefaultJiraProject:     "TEST",
				DefaultConfluenceSpace: "DEV",
				Timeout:                60 * time.Second,
				Output:                 "json",
				Debug:                  true,
			},
			expectError: false,
		},
		{
			name: "environment variables override",
			configContent: `
api_endpoint: "https://test.atlassian.net"
email: "test@example.com"
token: "test-token"
output: "table"
`,
			envVars: map[string]string{
				"ATLASSIAN_DEFAULT_JIRA_PROJECT":     "ENV_PROJECT",
				"ATLASSIAN_DEFAULT_CONFLUENCE_SPACE": "ENV_SPACE",
				"ATLASSIAN_OUTPUT":                   "yaml",
			},
			expectedConfig: &types.Config{
				APIEndpoint:            "https://test.atlassian.net",
				Email:                  "test@example.com",
				Token:                  "test-token",
				DefaultJiraProject:     "ENV_PROJECT",
				DefaultConfluenceSpace: "ENV_SPACE",
				Output:                 "yaml",
				Timeout:                30 * time.Second, // default
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary config file
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "config.yaml")
			err := os.WriteFile(configFile, []byte(tt.configContent), 0644)
			require.NoError(t, err)

			// Set environment variables
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			// Load config
			config, err := LoadConfig(configFile)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedConfig.APIEndpoint, config.APIEndpoint)
			assert.Equal(t, tt.expectedConfig.Email, config.Email)
			assert.Equal(t, tt.expectedConfig.Token, config.Token)
			assert.Equal(t, tt.expectedConfig.DefaultJiraProject, config.DefaultJiraProject)
			assert.Equal(t, tt.expectedConfig.DefaultConfluenceSpace, config.DefaultConfluenceSpace)
			assert.Equal(t, tt.expectedConfig.Output, config.Output)
			assert.Equal(t, tt.expectedConfig.Debug, config.Debug)

			if tt.expectedConfig.Timeout != 0 {
				assert.Equal(t, tt.expectedConfig.Timeout, config.Timeout)
			}
		})
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	config := &types.Config{
		APIEndpoint:            "https://test.atlassian.net",
		Email:                  "test@example.com",
		Token:                  "test-token",
		DefaultJiraProject:     "PROJ",
		DefaultConfluenceSpace: "SPACE",
		Timeout:                45 * time.Second,
		Output:                 "table",
		Debug:                  false,
	}

	err := SaveConfig(configFile, config)
	require.NoError(t, err)

	// Verify file was created and can be loaded
	loadedConfig, err := LoadConfig(configFile)
	require.NoError(t, err)

	assert.Equal(t, config.APIEndpoint, loadedConfig.APIEndpoint)
	assert.Equal(t, config.Email, loadedConfig.Email)
	assert.Equal(t, config.DefaultJiraProject, loadedConfig.DefaultJiraProject)
	assert.Equal(t, config.DefaultConfluenceSpace, loadedConfig.DefaultConfluenceSpace)
}
