package issue

import (
	"bytes"
	"testing"

	"atlassian-cli/internal/auth"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestNewIssueCmd(t *testing.T) {
	tokenManager := auth.NewMemoryTokenManager()
	cmd := NewIssueCmd(tokenManager)
	
	assert.NotNil(t, cmd)
	assert.Equal(t, "issue", cmd.Use)
	assert.Contains(t, cmd.Short, "Manage JIRA issues")
	assert.True(t, len(cmd.Commands()) > 0)
}

func TestIssueCreateCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		errMsg   string
	}{
		{
			name:    "missing required flags",
			args:    []string{"create"},
			wantErr: true,
			errMsg:  "required flag",
		},
		{
			name:    "missing summary",
			args:    []string{"create", "--type", "Story"},
			wantErr: true,
			errMsg:  "required flag",
		},
		{
			name:    "missing type",
			args:    []string{"create", "--summary", "Test issue"},
			wantErr: true,
			errMsg:  "required flag",
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

func TestIssueGetCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "missing issue key",
			args:    []string{"get"},
			wantErr: true,
			errMsg:  "requires exactly 1 arg",
		},
		{
			name:    "too many arguments",
			args:    []string{"get", "DEMO-123", "DEMO-124"},
			wantErr: true,
			errMsg:  "requires exactly 1 arg",
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

func TestIssueListCommand(t *testing.T) {
	tokenManager := auth.NewMemoryTokenManager()
	cmd := NewIssueCmd(tokenManager)
	
	// Test that list command exists
	listCmd := findCommand(cmd, "list")
	assert.NotNil(t, listCmd)
	assert.Equal(t, "list", listCmd.Use)
}

func TestIssueUpdateCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "missing issue key",
			args:    []string{"update"},
			wantErr: true,
			errMsg:  "requires exactly 1 arg",
		},
		{
			name:    "too many arguments",
			args:    []string{"update", "DEMO-123", "DEMO-124"},
			wantErr: true,
			errMsg:  "requires exactly 1 arg",
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

func TestIssueCommandFlags(t *testing.T) {
	tokenManager := auth.NewMemoryTokenManager()
	cmd := NewIssueCmd(tokenManager)
	
	// Test global flags are inherited
	assert.NotNil(t, cmd.PersistentFlags().Lookup("jira-project"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("output"))
	
	// Test create command flags
	createCmd := findCommand(cmd, "create")
	if createCmd != nil {
		flags := createCmd.Flags()
		assert.NotNil(t, flags.Lookup("type"))
		assert.NotNil(t, flags.Lookup("summary"))
		assert.NotNil(t, flags.Lookup("description"))
		assert.NotNil(t, flags.Lookup("assignee"))
		assert.NotNil(t, flags.Lookup("priority"))
	}
}

// Helper function to find a subcommand
func findCommand(parent *cobra.Command, name string) *cobra.Command {
	for _, cmd := range parent.Commands() {
		if cmd.Use == name {
			return cmd
		}
	}
	return nil
}