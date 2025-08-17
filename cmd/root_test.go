package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedOutput string
		expectError    bool
	}{
		{
			name:           "help command",
			args:           []string{"--help"},
			expectedOutput: "Atlassian CLI is a command-line tool",
			expectError:    false,
		},
		{
			name:           "version command",
			args:           []string{"--version"},
			expectedOutput: "atlassian-cli version",
			expectError:    false,
		},
		{
			name:        "invalid command",
			args:        []string{"invalid"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new root command for each test to avoid state pollution
			cmd := newRootCmd()
			
			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expectedOutput != "" {
					assert.Contains(t, output.String(), tt.expectedOutput)
				}
			}
		})
	}
}

func TestRootCommandFlags(t *testing.T) {
	cmd := newRootCmd()
	
	// Test that global flags are properly registered
	assert.NotNil(t, cmd.PersistentFlags().Lookup("config"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("output"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("verbose"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("jira-project"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("confluence-space"))
}