package auth

import (
	"atlassian-cli/internal/types"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptedFileTokenManager(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "atlassian-cli-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "credentials.enc")
	manager, err := NewEncryptedFileTokenManager(filePath)
	require.NoError(t, err)

	creds := &types.AuthCredentials{
		ServerURL: "https://test.atlassian.net",
		Email:     "test@example.com",
		Token:     "test-token",
	}

	// Test storing credentials
	err = manager.Store(context.Background(), creds)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(filePath)
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

func TestEncryptedFileTokenManager_MultipleCredentials(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "atlassian-cli-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "credentials.enc")
	manager, err := NewEncryptedFileTokenManager(filePath)
	require.NoError(t, err)

	creds1 := &types.AuthCredentials{
		ServerURL: "https://test1.atlassian.net",
		Email:     "test1@example.com",
		Token:     "test-token-1",
	}

	creds2 := &types.AuthCredentials{
		ServerURL: "https://test2.atlassian.net",
		Email:     "test2@example.com",
		Token:     "test-token-2",
	}

	// Store both credentials
	err = manager.Store(context.Background(), creds1)
	require.NoError(t, err)

	err = manager.Store(context.Background(), creds2)
	require.NoError(t, err)

	// Retrieve both
	retrieved1, err := manager.Get(context.Background(), creds1.ServerURL)
	require.NoError(t, err)
	assert.Equal(t, creds1.Email, retrieved1.Email)

	retrieved2, err := manager.Get(context.Background(), creds2.ServerURL)
	require.NoError(t, err)
	assert.Equal(t, creds2.Email, retrieved2.Email)

	// Delete one
	err = manager.Delete(context.Background(), creds1.ServerURL)
	require.NoError(t, err)

	// First should be gone
	_, err = manager.Get(context.Background(), creds1.ServerURL)
	assert.Error(t, err)

	// Second should still exist
	retrieved2, err = manager.Get(context.Background(), creds2.ServerURL)
	require.NoError(t, err)
	assert.Equal(t, creds2.Email, retrieved2.Email)
}

func TestEncryptedFileTokenManager_EncryptionWorks(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "atlassian-cli-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "credentials.enc")
	manager, err := NewEncryptedFileTokenManager(filePath)
	require.NoError(t, err)

	creds := &types.AuthCredentials{
		ServerURL: "https://test.atlassian.net",
		Email:     "test@example.com",
		Token:     "super-secret-token",
	}

	// Store credentials
	err = manager.Store(context.Background(), creds)
	require.NoError(t, err)

	// Read the raw file
	rawData, err := os.ReadFile(filePath)
	require.NoError(t, err)

	// Verify the file is encrypted (not valid JSON)
	var testJSON map[string]interface{}
	err = json.Unmarshal(rawData, &testJSON)
	assert.Error(t, err, "Encrypted file should not be valid JSON")

	// Verify the plaintext token is not in the file
	assert.NotContains(t, string(rawData), "super-secret-token")
	assert.NotContains(t, string(rawData), "test@example.com")
}

func TestEncryptedFileTokenManager_DefaultPath(t *testing.T) {
	// Test with empty path (should use default)
	manager, err := NewEncryptedFileTokenManager("")
	require.NoError(t, err)

	// Verify it created the path
	home, _ := os.UserHomeDir()
	expectedPath := filepath.Join(home, DefaultCredentialsFile)
	assert.Equal(t, expectedPath, manager.filePath)

	// Clean up if test file was created
	os.RemoveAll(filepath.Dir(expectedPath))
}

func TestEncryptedFileTokenManager_Validate(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "atlassian-cli-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "credentials.enc")
	manager, err := NewEncryptedFileTokenManager(filePath)
	require.NoError(t, err)

	// Validate should call ValidateToken (we can't test the HTTP call here without a server)
	// This just tests that the method exists and can be called
	_, err = manager.Validate(context.Background(), "https://test.atlassian.net", "test@example.com", "token")
	// Expect error since we're not providing a real API endpoint
	assert.Error(t, err)
}

func TestEncryptedFileTokenManager_InvalidCredentials(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "atlassian-cli-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "credentials.enc")
	manager, err := NewEncryptedFileTokenManager(filePath)
	require.NoError(t, err)

	// Try to store invalid credentials
	invalidCreds := &types.AuthCredentials{
		ServerURL: "not-a-url",
		Email:     "invalid-email",
		Token:     "",
	}

	err = manager.Store(context.Background(), invalidCreds)
	assert.Error(t, err)
}

func TestEncryptedFileTokenManager_NonExistentGet(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "atlassian-cli-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "credentials.enc")
	manager, err := NewEncryptedFileTokenManager(filePath)
	require.NoError(t, err)

	// Try to get credentials that don't exist
	_, err = manager.Get(context.Background(), "https://nonexistent.atlassian.net")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "credentials not found")
}

func TestEncryptedFileTokenManager_DeleteNonExistent(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "atlassian-cli-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "credentials.enc")
	manager, err := NewEncryptedFileTokenManager(filePath)
	require.NoError(t, err)

	// Delete non-existent credentials should not error
	err = manager.Delete(context.Background(), "https://nonexistent.atlassian.net")
	assert.NoError(t, err)
}

func TestEncryptedFileTokenManager_Update(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "atlassian-cli-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "credentials.enc")
	manager, err := NewEncryptedFileTokenManager(filePath)
	require.NoError(t, err)

	creds := &types.AuthCredentials{
		ServerURL: "https://test.atlassian.net",
		Email:     "test@example.com",
		Token:     "old-token",
	}

	// Store original credentials
	err = manager.Store(context.Background(), creds)
	require.NoError(t, err)

	// Update with new token
	updatedCreds := &types.AuthCredentials{
		ServerURL: "https://test.atlassian.net",
		Email:     "test@example.com",
		Token:     "new-token",
	}

	err = manager.Store(context.Background(), updatedCreds)
	require.NoError(t, err)

	// Retrieve and verify updated
	retrieved, err := manager.Get(context.Background(), creds.ServerURL)
	require.NoError(t, err)
	assert.Equal(t, "new-token", retrieved.Token)
}
