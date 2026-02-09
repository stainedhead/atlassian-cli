package auth

import (
	"atlassian-cli/internal/types"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// DefaultCredentialsFile is the default path for encrypted credentials
	DefaultCredentialsFile = ".atlassian-cli/credentials.enc"
	// PBKDF2Iterations is the number of iterations for key derivation
	PBKDF2Iterations = 100000
	// KeySize is the size of the encryption key in bytes (32 bytes = 256 bits)
	KeySize = 32
)

// EncryptedFileTokenManager implements TokenManager using AES-256-GCM encrypted file
type EncryptedFileTokenManager struct {
	filePath string
	key      []byte
	mutex    sync.RWMutex
}

// NewEncryptedFileTokenManager creates a new encrypted file-based token manager
func NewEncryptedFileTokenManager(filePath string) (*EncryptedFileTokenManager, error) {
	// If empty path, use default in home directory
	if filePath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		filePath = filepath.Join(home, DefaultCredentialsFile)
	}

	// Derive encryption key from machine-specific identifier
	key, err := deriveKey()
	if err != nil {
		return nil, fmt.Errorf("failed to derive encryption key: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create credentials directory: %w", err)
	}

	return &EncryptedFileTokenManager{
		filePath: filePath,
		key:      key,
	}, nil
}

// Store saves credentials in an encrypted file
func (e *EncryptedFileTokenManager) Store(ctx context.Context, creds *types.AuthCredentials) error {
	if err := ValidateCredentials(creds); err != nil {
		return fmt.Errorf("invalid credentials: %w", err)
	}

	e.mutex.Lock()
	defer e.mutex.Unlock()

	// Load existing credentials
	allCreds, err := e.loadAll()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load existing credentials: %w", err)
	}

	// Update or add credentials
	if allCreds == nil {
		allCreds = make(map[string]*types.AuthCredentials)
	}
	allCreds[creds.ServerURL] = creds

	// Save all credentials
	return e.saveAll(allCreds)
}

// Get retrieves credentials from the encrypted file
func (e *EncryptedFileTokenManager) Get(ctx context.Context, serverURL string) (*types.AuthCredentials, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	allCreds, err := e.loadAll()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("credentials not found for server: %s", serverURL)
		}
		return nil, fmt.Errorf("failed to load credentials: %w", err)
	}

	creds, exists := allCreds[serverURL]
	if !exists {
		return nil, fmt.Errorf("credentials not found for server: %s", serverURL)
	}

	return creds, nil
}

// Delete removes credentials from the encrypted file
func (e *EncryptedFileTokenManager) Delete(ctx context.Context, serverURL string) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	allCreds, err := e.loadAll()
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, nothing to delete
			return nil
		}
		return fmt.Errorf("failed to load credentials: %w", err)
	}

	delete(allCreds, serverURL)

	// If no credentials left, delete the file
	if len(allCreds) == 0 {
		if err := os.Remove(e.filePath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete credentials file: %w", err)
		}
		return nil
	}

	return e.saveAll(allCreds)
}

// Validate validates credentials against the Atlassian API
func (e *EncryptedFileTokenManager) Validate(ctx context.Context, serverURL, email, token string) (*types.UserInfo, error) {
	return ValidateToken(ctx, serverURL, email, token)
}

// loadAll loads all credentials from the encrypted file
func (e *EncryptedFileTokenManager) loadAll() (map[string]*types.AuthCredentials, error) {
	// Read encrypted data
	encryptedData, err := os.ReadFile(e.filePath)
	if err != nil {
		return nil, err
	}

	// Decrypt data
	plaintext, err := e.decrypt(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credentials: %w", err)
	}

	// Unmarshal credentials
	var allCreds map[string]*types.AuthCredentials
	if err := json.Unmarshal(plaintext, &allCreds); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credentials: %w", err)
	}

	return allCreds, nil
}

// saveAll saves all credentials to the encrypted file
func (e *EncryptedFileTokenManager) saveAll(allCreds map[string]*types.AuthCredentials) error {
	// Marshal credentials
	plaintext, err := json.Marshal(allCreds)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	// Encrypt data
	encryptedData, err := e.encrypt(plaintext)
	if err != nil {
		return fmt.Errorf("failed to encrypt credentials: %w", err)
	}

	// Write to temp file first (atomic write)
	tempFile := e.filePath + ".tmp"
	if err := os.WriteFile(tempFile, encryptedData, 0600); err != nil {
		return fmt.Errorf("failed to write temporary credentials file: %w", err)
	}

	// Rename temp file to target file (atomic operation)
	if err := os.Rename(tempFile, e.filePath); err != nil {
		os.Remove(tempFile) // Clean up temp file on error
		return fmt.Errorf("failed to rename credentials file: %w", err)
	}

	return nil
}

// encrypt encrypts data using AES-256-GCM
func (e *EncryptedFileTokenManager) encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Create nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt and prepend nonce
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decrypt decrypts data using AES-256-GCM
func (e *EncryptedFileTokenManager) decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}

	// Extract nonce and ciphertext
	nonce := ciphertext[:gcm.NonceSize()]
	ciphertext = ciphertext[gcm.NonceSize():]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// deriveKey derives an encryption key from machine-specific identifiers
func deriveKey() ([]byte, error) {
	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "default-host"
	}

	// Get user ID (UID)
	uid := fmt.Sprintf("%d", os.Getuid())

	// Combine identifiers
	salt := []byte(fmt.Sprintf("%s:%s", hostname, uid))

	// Derive key using PBKDF2
	// Using a fixed password combined with machine-specific salt
	// This is less secure than OS keychain but better than plaintext
	password := []byte("atlassian-cli-encryption-key")
	key := pbkdf2.Key(password, salt, PBKDF2Iterations, KeySize, sha256.New)

	return key, nil
}
