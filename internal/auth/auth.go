package auth

import (
	"atlassian-cli/internal/types"
	"context"
	"fmt"
	"sync"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// TokenManager defines the interface for storing and retrieving authentication tokens
type TokenManager interface {
	Store(ctx context.Context, creds *types.AuthCredentials) error
	Get(ctx context.Context, serverURL string) (*types.AuthCredentials, error)
	Delete(ctx context.Context, serverURL string) error
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

// ValidateCredentials validates authentication credentials
func ValidateCredentials(creds *types.AuthCredentials) error {
	if creds == nil {
		return fmt.Errorf("credentials cannot be nil")
	}

	if err := validate.Struct(creds); err != nil {
		return fmt.Errorf("credential validation failed: %w", err)
	}

	return nil
}