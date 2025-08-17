package auth

import (
	"atlassian-cli/internal/auth"
	"atlassian-cli/internal/types"
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAuthCmd(t *testing.T) {
	tokenManager := auth.NewMemoryTokenManager()
	cmd := NewAuthCmd(tokenManager)

	assert.Equal(t, "auth", cmd.Use)
	assert.True(t, cmd.HasSubCommands())

	// Check that subcommands are registered
	subcommands := cmd.Commands()
	commandNames := make([]string, len(subcommands))
	for i, subcmd := range subcommands {
		commandNames[i] = subcmd.Use
	}

	assert.Contains(t, commandNames, "login")
	assert.Contains(t, commandNames, "logout")
	assert.Contains(t, commandNames, "status")
}

func TestAuthStatusCommand(t *testing.T) {
	tests := []struct {
		name           string
		setupManager   func(manager auth.TokenManager)
		serverURL      string
		expectedOutput string
		expectError    bool
	}{
		{
			name: "authenticated status",
			setupManager: func(manager auth.TokenManager) {
				creds := &types.AuthCredentials{
					ServerURL: "https://test.atlassian.net",
					Email:     "test@example.com",
					Token:     "test-token",
				}
				manager.Store(context.Background(), creds)
			},
			serverURL:      "https://test.atlassian.net",
			expectedOutput: "Authenticated as test@example.com",
			expectError:    false,
		},
		{
			name:           "not authenticated status",
			setupManager:   func(manager auth.TokenManager) {},
			serverURL:      "https://test.atlassian.net",
			expectedOutput: "Not authenticated",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenManager := auth.NewMemoryTokenManager()
			tt.setupManager(tokenManager)

			cmd := newStatusCmd(tokenManager)
			
			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)

			// Set server URL flag
			cmd.Flags().Set("server", tt.serverURL)

			err := cmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Contains(t, output.String(), tt.expectedOutput)
		})
	}
}

func TestAuthLogoutCommand(t *testing.T) {
	tokenManager := auth.NewMemoryTokenManager()

	// Setup - store credentials first
	creds := &types.AuthCredentials{
		ServerURL: "https://test.atlassian.net",
		Email:     "test@example.com",
		Token:     "test-token",
	}
	err := tokenManager.Store(context.Background(), creds)
	require.NoError(t, err)

	// Test logout command
	cmd := newLogoutCmd(tokenManager)
	
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Set server URL flag
	cmd.Flags().Set("server", "https://test.atlassian.net")

	err = cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, output.String(), "Logged out")

	// Verify credentials were deleted
	_, err = tokenManager.Get(context.Background(), "https://test.atlassian.net")
	assert.Error(t, err)
}

func TestValidateAuthFlags(t *testing.T) {
	tests := []struct {
		name        string
		serverURL   string
		email       string
		token       string
		expectError bool
	}{
		{
			name:        "valid flags",
			serverURL:   "https://test.atlassian.net",
			email:       "test@example.com",
			token:       "test-token",
			expectError: false,
		},
		{
			name:        "missing server URL",
			serverURL:   "",
			email:       "test@example.com",
			token:       "test-token",
			expectError: true,
		},
		{
			name:        "invalid email",
			serverURL:   "https://test.atlassian.net",
			email:       "invalid-email",
			token:       "test-token",
			expectError: true,
		},
		{
			name:        "missing token",
			serverURL:   "https://test.atlassian.net",
			email:       "test@example.com",
			token:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAuthFlags(tt.serverURL, tt.email, tt.token)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}