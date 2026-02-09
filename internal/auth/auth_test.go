package auth

import (
	"atlassian-cli/internal/types"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestValidateToken_Success(t *testing.T) {
	// Create test server that returns valid user info
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "/rest/api/3/myself", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		// Check basic auth
		username, password, ok := r.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "test@example.com", username)
		assert.Equal(t, "test-token", password)

		// Return valid user info
		userInfo := types.UserInfo{
			AccountID:   "5b10a2844c20165700ede21g",
			DisplayName: "John Doe",
			Email:       "test@example.com",
			Active:      true,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(userInfo)
	}))
	defer server.Close()

	// Test validation
	userInfo, err := ValidateToken(context.Background(), server.URL, "test@example.com", "test-token")
	require.NoError(t, err)
	assert.Equal(t, "John Doe", userInfo.DisplayName)
	assert.Equal(t, "test@example.com", userInfo.Email)
	assert.True(t, userInfo.Active)
}

func TestValidateToken_Unauthorized(t *testing.T) {
	// Create test server that returns 401
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"errorMessages":["Invalid credentials"]}`))
	}))
	defer server.Close()

	// Test validation
	_, err := ValidateToken(context.Background(), server.URL, "test@example.com", "bad-token")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "authentication failed")
	assert.Contains(t, err.Error(), "invalid email or API token")
	assert.Contains(t, err.Error(), "https://id.atlassian.com/manage/api-tokens")
}

func TestValidateToken_ServerError(t *testing.T) {
	// Create test server that returns 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"errorMessages":["Internal server error"]}`))
	}))
	defer server.Close()

	// Test validation
	_, err := ValidateToken(context.Background(), server.URL, "test@example.com", "test-token")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "API request failed with status 500")
}

func TestValidateToken_NetworkError(t *testing.T) {
	// Use invalid URL to trigger network error
	_, err := ValidateToken(context.Background(), "http://invalid-nonexistent-domain-12345.local", "test@example.com", "test-token")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot reach")
}

func TestMemoryTokenManager_Validate(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userInfo := types.UserInfo{
			AccountID:   "123",
			DisplayName: "Test User",
			Email:       "test@example.com",
			Active:      true,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(userInfo)
	}))
	defer server.Close()

	manager := NewMemoryTokenManager()

	// Test validate method
	userInfo, err := manager.Validate(context.Background(), server.URL, "test@example.com", "test-token")
	require.NoError(t, err)
	assert.Equal(t, "Test User", userInfo.DisplayName)
}
