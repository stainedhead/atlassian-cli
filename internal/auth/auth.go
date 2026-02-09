package auth

import (
	"atlassian-cli/internal/types"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// TokenManager defines the interface for storing and retrieving authentication tokens
type TokenManager interface {
	Store(ctx context.Context, creds *types.AuthCredentials) error
	Get(ctx context.Context, serverURL string) (*types.AuthCredentials, error)
	Delete(ctx context.Context, serverURL string) error
	Validate(ctx context.Context, serverURL, email, token string) (*types.UserInfo, error)
}

// MemoryTokenManager implements TokenManager using in-memory storage
// This is primarily for testing and fallback scenarios
type MemoryTokenManager struct {
	credentials map[string]*types.AuthCredentials
	mutex       sync.RWMutex
}

// NewMemoryTokenManager creates a new in-memory token manager
func NewMemoryTokenManager() *MemoryTokenManager {
	return &MemoryTokenManager{
		credentials: make(map[string]*types.AuthCredentials),
	}
}

// Store saves credentials in memory
func (m *MemoryTokenManager) Store(ctx context.Context, creds *types.AuthCredentials) error {
	if err := ValidateCredentials(creds); err != nil {
		return fmt.Errorf("invalid credentials: %w", err)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.credentials[creds.ServerURL] = creds
	return nil
}

// Get retrieves credentials from memory
func (m *MemoryTokenManager) Get(ctx context.Context, serverURL string) (*types.AuthCredentials, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	creds, exists := m.credentials[serverURL]
	if !exists {
		return nil, fmt.Errorf("credentials not found for server: %s", serverURL)
	}

	return creds, nil
}

// Delete removes credentials from memory
func (m *MemoryTokenManager) Delete(ctx context.Context, serverURL string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.credentials, serverURL)
	return nil
}

// Validate validates credentials against the Atlassian API
func (m *MemoryTokenManager) Validate(ctx context.Context, serverURL, email, token string) (*types.UserInfo, error) {
	return ValidateToken(ctx, serverURL, email, token)
}

// ValidateCredentials validates authentication credentials format
func ValidateCredentials(creds *types.AuthCredentials) error {
	if creds == nil {
		return fmt.Errorf("credentials cannot be nil")
	}

	if err := validate.Struct(creds); err != nil {
		return fmt.Errorf("credential validation failed: %w", err)
	}

	return nil
}

// ValidateToken validates an API token against the Atlassian API by calling /rest/api/3/myself
func ValidateToken(ctx context.Context, serverURL, email, token string) (*types.UserInfo, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Build the API endpoint URL
	apiURL := fmt.Sprintf("%s/rest/api/3/myself", serverURL)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set basic auth with email and token
	req.SetBasicAuth(email, token)
	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot reach %s. Check the URL and your network connection: %w", serverURL, err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for authentication failure
	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("authentication failed: invalid email or API token. Generate a new token at https://id.atlassian.com/manage/api-tokens")
	}

	// Check for other HTTP errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse the user info from response
	var userInfo types.UserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	return &userInfo, nil
}
