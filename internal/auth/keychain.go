package auth

import (
	"atlassian-cli/internal/types"
	"context"
	"encoding/json"
	"fmt"

	"github.com/zalando/go-keyring"
)

const (
	// ServiceName is the service name used for keychain storage
	ServiceName = "atlassian-cli"
)

// KeychainTokenManager implements TokenManager using OS keychain
type KeychainTokenManager struct {
	serviceName string
}

// NewKeychainTokenManager creates a new keychain-based token manager
func NewKeychainTokenManager() *KeychainTokenManager {
	return &KeychainTokenManager{
		serviceName: ServiceName,
	}
}

// Store saves credentials in the OS keychain
func (k *KeychainTokenManager) Store(ctx context.Context, creds *types.AuthCredentials) error {
	if err := ValidateCredentials(creds); err != nil {
		return fmt.Errorf("invalid credentials: %w", err)
	}

	// Marshal credentials to JSON
	data, err := json.Marshal(creds)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	// Store in keychain using serverURL as the account key
	if err := keyring.Set(k.serviceName, creds.ServerURL, string(data)); err != nil {
		return fmt.Errorf("failed to store credentials in keychain: %w", err)
	}

	return nil
}

// Get retrieves credentials from the OS keychain
func (k *KeychainTokenManager) Get(ctx context.Context, serverURL string) (*types.AuthCredentials, error) {
	// Retrieve from keychain
	data, err := keyring.Get(k.serviceName, serverURL)
	if err != nil {
		if err == keyring.ErrNotFound {
			return nil, fmt.Errorf("credentials not found for server: %s", serverURL)
		}
		return nil, fmt.Errorf("failed to retrieve credentials from keychain: %w", err)
	}

	// Unmarshal credentials
	var creds types.AuthCredentials
	if err := json.Unmarshal([]byte(data), &creds); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credentials: %w", err)
	}

	return &creds, nil
}

// Delete removes credentials from the OS keychain
func (k *KeychainTokenManager) Delete(ctx context.Context, serverURL string) error {
	if err := keyring.Delete(k.serviceName, serverURL); err != nil {
		if err == keyring.ErrNotFound {
			// Already deleted, not an error
			return nil
		}
		return fmt.Errorf("failed to delete credentials from keychain: %w", err)
	}

	return nil
}

// Validate validates credentials against the Atlassian API
func (k *KeychainTokenManager) Validate(ctx context.Context, serverURL, email, token string) (*types.UserInfo, error) {
	return ValidateToken(ctx, serverURL, email, token)
}
