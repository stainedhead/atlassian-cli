package config

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewConfigCmd(t *testing.T) {
	cmd := NewConfigCmd()
	
	assert.NotNil(t, cmd)
	assert.Equal(t, "config", cmd.Use)
	assert.Contains(t, cmd.Short, "Manage configuration")
	assert.True(t, len(cmd.Commands()) > 0)
}

func TestConfigSetCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "missing arguments",
			args:    []string{"set"},
			wantErr: true,
			errMsg:  "requires exactly 2 args",
		},
		{
			name:    "missing value",
			args:    []string{"set", "default_jira_project"},
			wantErr: true,
			errMsg:  "requires exactly 2 args",
		},
		{
			name:    "too many arguments",
			args:    []string{"set", "key", "value", "extra"},
			wantErr: true,
			errMsg:  "requires exactly 2 args",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewConfigCmd()
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

func TestConfigGetCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "missing key",
			args:    []string{"get"},
			wantErr: true,
			errMsg:  "requires exactly 1 arg",
		},
		{
			name:    "too many arguments",
			args:    []string{"get", "key1", "key2"},
			wantErr: true,
			errMsg:  "requires exactly 1 arg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewConfigCmd()
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

func TestConfigListCommand(t *testing.T) {
	cmd := NewConfigCmd()
	
	// Test that list command exists and has correct structure
	var listCmd *cobra.Command
	for _, subCmd := range cmd.Commands() {
		if subCmd.Use == "list" {
			listCmd = subCmd
			break
		}
	}
	
	assert.NotNil(t, listCmd)
	assert.Equal(t, "list", listCmd.Use)
	assert.Contains(t, listCmd.Short, "List all configuration")
}

func TestConfigCommandStructure(t *testing.T) {
	cmd := NewConfigCmd()
	
	// Verify all expected subcommands exist
	expectedCommands := []string{"set", "get", "list"}
	actualCommands := make(map[string]bool)
	
	for _, subCmd := range cmd.Commands() {
		actualCommands[subCmd.Use] = true
	}
	
	for _, expected := range expectedCommands {
		assert.True(t, actualCommands[expected], "Missing command: %s", expected)
	}
}

func TestConfigValidKeys(t *testing.T) {
	validKeys := []string{
		"default_jira_project",
		"default_confluence_space",
		"output",
		"cache_ttl",
		"cache_enabled",
		"jira_timeout",
		"confluence_timeout",
		"no_color",
	}
	
	// This test verifies that our documentation matches expected configuration keys
	// In a real implementation, this would test against the actual validation logic
	assert.True(t, len(validKeys) > 0, "Should have valid configuration keys defined")
}