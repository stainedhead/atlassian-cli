package page

import (
	"bytes"
	"testing"

	"atlassian-cli/internal/auth"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewPageSearchCmd(t *testing.T) {
	tokenManager := auth.NewMemoryTokenManager()
	cmd := NewPageCmd(tokenManager)

	// Find search subcommand
	var searchCmd *cobra.Command
	for _, subCmd := range cmd.Commands() {
		if subCmd.Use == "search" {
			searchCmd = subCmd
			break
		}
	}

	assert.NotNil(t, searchCmd, "Page search command should exist")
	assert.Equal(t, "search", searchCmd.Use)
	assert.Contains(t, searchCmd.Short, "Search pages")
}

func TestPageSearchFlags(t *testing.T) {
	tokenManager := auth.NewMemoryTokenManager()
	cmd := NewPageCmd(tokenManager)

	// Find search subcommand
	var searchCmd *cobra.Command
	for _, subCmd := range cmd.Commands() {
		if subCmd.Use == "search" {
			searchCmd = subCmd
			break
		}
	}

	assert.NotNil(t, searchCmd)

	// Check required flags exist
	flags := searchCmd.Flags()
	assert.NotNil(t, flags.Lookup("space"), "Should have --space flag")
	assert.NotNil(t, flags.Lookup("cql"), "Should have --cql flag")
	assert.NotNil(t, flags.Lookup("text"), "Should have --text flag")
	assert.NotNil(t, flags.Lookup("title"), "Should have --title flag")
	assert.NotNil(t, flags.Lookup("type"), "Should have --type flag")
	assert.NotNil(t, flags.Lookup("limit"), "Should have --limit flag")
}

func TestPageSearchCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "search with CQL",
			args:    []string{"search", "--cql", "space = DEV"},
			wantErr: true, // Will fail until authentication is mocked
		},
		{
			name:    "search with text",
			args:    []string{"search", "--space", "DEV", "--text", "documentation"},
			wantErr: true, // Will fail until authentication is mocked
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenManager := auth.NewMemoryTokenManager()
			cmd := NewPageCmd(tokenManager)
			cmd.SetArgs(tt.args)

			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			err := cmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
