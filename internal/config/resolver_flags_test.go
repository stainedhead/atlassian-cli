package config

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestResolveProjectWithCommandFlag(t *testing.T) {
	tests := []struct {
		name        string
		flagValue   string
		envValue    string
		configValue string
		want        string
		wantErr     bool
	}{
		{
			name:      "command flag takes precedence",
			flagValue: "FLAG-PROJ",
			envValue:  "ENV-PROJ",
			want:      "FLAG-PROJ",
			wantErr:   false,
		},
		{
			name:     "env var when no flag",
			envValue: "ENV-PROJ",
			want:     "ENV-PROJ",
			wantErr:  false,
		},
		{
			name:        "config when no flag or env",
			configValue: "CONFIG-PROJ",
			want:        "CONFIG-PROJ",
			wantErr:     false,
		},
		{
			name:    "error when no value found",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup command with flag
			cmd := &cobra.Command{}
			cmd.Flags().String("project", "", "project key")
			if tt.flagValue != "" {
				cmd.Flags().Set("project", tt.flagValue)
			}

			// Setup environment
			if tt.envValue != "" {
				os.Setenv("ATLASSIAN_DEFAULT_JIRA_PROJECT", tt.envValue)
				defer os.Unsetenv("ATLASSIAN_DEFAULT_JIRA_PROJECT")
			}

			// Setup config
			if tt.configValue != "" {
				viper.Set("default_jira_project", tt.configValue)
				defer viper.Set("default_jira_project", "")
			}

			got, err := ResolveProject(cmd)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestResolveSpaceWithCommandFlag(t *testing.T) {
	tests := []struct {
		name        string
		flagValue   string
		envValue    string
		configValue string
		want        string
		wantErr     bool
	}{
		{
			name:      "command flag takes precedence",
			flagValue: "FLAG-SPACE",
			envValue:  "ENV-SPACE",
			want:      "FLAG-SPACE",
			wantErr:   false,
		},
		{
			name:     "env var when no flag",
			envValue: "ENV-SPACE",
			want:     "ENV-SPACE",
			wantErr:  false,
		},
		{
			name:    "error when no value found",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup command with flag
			cmd := &cobra.Command{}
			cmd.Flags().String("space", "", "space key")
			if tt.flagValue != "" {
				cmd.Flags().Set("space", tt.flagValue)
			}

			// Setup environment
			if tt.envValue != "" {
				os.Setenv("ATLASSIAN_DEFAULT_CONFLUENCE_SPACE", tt.envValue)
				defer os.Unsetenv("ATLASSIAN_DEFAULT_CONFLUENCE_SPACE")
			}

			got, err := ResolveSpace(cmd)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
