package issue

import (
	"bytes"
	"testing"

	"atlassian-cli/internal/auth"
	"github.com/stretchr/testify/assert"
)

func TestNewIssueSearchCmd(t *testing.T) {
	tokenManager := auth.NewMemoryTokenManager()
	cmd := NewIssueCmd(tokenManager)
	
	// Find search subcommand
	var searchCmd *cobra.Command
	for _, subCmd := range cmd.Commands() {
		if subCmd.Use == "search" {
			searchCmd = subCmd
			break
		}
	}
	
	assert.NotNil(t, searchCmd, "Issue search command should exist")
	assert.Equal(t, "search", searchCmd.Use)
	assert.Contains(t, searchCmd.Short, "Search issues")
}

func TestIssueSearchFlags(t *testing.T) {
	tokenManager := auth.NewMemoryTokenManager()
	cmd := NewIssueCmd(tokenManager)
	
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
	assert.NotNil(t, flags.Lookup("project"), "Should have --project flag")
	assert.NotNil(t, flags.Lookup("jql"), "Should have --jql flag")
	assert.NotNil(t, flags.Lookup("assignee"), "Should have --assignee flag")
	assert.NotNil(t, flags.Lookup("status"), "Should have --status flag")
	assert.NotNil(t, flags.Lookup("type"), "Should have --type flag")
	assert.NotNil(t, flags.Lookup("limit"), "Should have --limit flag")
}

func TestIssueSearchCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "search with JQL",
			args:    []string{"search", "--jql", "project = DEMO"},
			wantErr: true, // Will fail until authentication is mocked
		},
		{
			name:    "search with filters",
			args:    []string{"search", "--project", "DEMO", "--status", "Open"},
			wantErr: true, // Will fail until authentication is mocked
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenManager := auth.NewMemoryTokenManager()
			cmd := NewIssueCmd(tokenManager)
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