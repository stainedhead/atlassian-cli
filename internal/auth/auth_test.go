package auth

import (
	"atlassian-cli/internal/types"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCredentialsValidation(t *testing.T) {
	tests := []struct {
		name        string
		creds       *types.AuthCredentials
		expectError bool
	}{
		{
			name: "valid credentials",
			creds: &types.AuthCredentials{
				ServerURL: "https://test.atlassian.net",
				Email:     "test@example.com",
				Token:     "test-token",
			},
			expectError: false,
		},
		{
			name: "invalid email",
			creds: &types.AuthCredentials{
				ServerURL: "https://test.atlassian.net",
				Email:     "invalid-email",
				Token:     "test-token",
			},
			expectError: true,
		},
		{
			name: "invalid URL",
			creds: &types.AuthCredentials{
				ServerURL: "not-a-url",
				Email:     "test@example.com",
				Token:     "test-token",
			},
			expectError: true,
		},
		{
			name: "empty token",
			creds: &types.AuthCredentials{
				ServerURL: "https://test.atlassian.net",
				Email:     "test@example.com",
				Token:     "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCredentials(tt.creds)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMemoryTokenManager(t *testing.T) {
	manager := NewMemoryTokenManager()

	creds := &types.AuthCredentials{
		ServerURL: "https://test.atlassian.net",
		Email:     "test@example.com",
		Token:     "test-token",
	}

	// Test storing credentials
	err := manager.Store(context.Background(), creds)
	require.NoError(t, err)

	// Test retrieving credentials
	retrieved, err := manager.Get(context.Background(), creds.ServerURL)
	require.NoError(t, err)
	assert.Equal(t, creds.Email, retrieved.Email)
	assert.Equal(t, creds.Token, retrieved.Token)
	assert.Equal(t, creds.ServerURL, retrieved.ServerURL)

	// Test deleting credentials
	err = manager.Delete(context.Background(), creds.ServerURL)
	require.NoError(t, err)

	// Test retrieving after deletion should fail
	_, err = manager.Get(context.Background(), creds.ServerURL)
	assert.Error(t, err)
}

func TestTokenManagerInterface(t *testing.T) {
	// Test that MemoryTokenManager implements TokenManager interface
	var _ TokenManager = NewMemoryTokenManager()
}