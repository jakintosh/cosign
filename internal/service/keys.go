package service

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"
	"time"
)

// APIKey represents an API key for authentication
type APIKey struct {
	ID        string `json:"id"`
	Hash      []byte `json:"-"`
	Salt      []byte `json:"-"`
	CreatedAt int64  `json:"created_at"`
}

// KeyStore interface for API key operations
type KeyStore interface {
	Insert(id string, hash, salt []byte, createdAt int64) error
	GetByID(id string) (*APIKey, error)
	List() ([]*APIKey, error)
	Delete(id string) error
}

var keyStore KeyStore

// SetKeyStore sets the key store implementation
func SetKeyStore(s KeyStore) {
	keyStore = s
}

// CreateAPIKey generates a new API key with format {id}.{secret}
func CreateAPIKey(id string) (string, error) {
	if keyStore == nil {
		return "", ErrNoKeyStore
	}

	if id == "" {
		id = "key-" + generateRandomString(8)
	}

	// Generate random secret (32 bytes)
	secret := generateRandomString(32)

	// Generate random salt
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Hash the secret with salt
	hash := hashSecret(secret, salt)

	// Store in database
	createdAt := time.Now().Unix()
	if err := keyStore.Insert(id, hash, salt, createdAt); err != nil {
		return "", err
	}

	// Return full key in format {id}.{secret}
	return fmt.Sprintf("%s.%s", id, secret), nil
}

// VerifyAPIKey validates an API key against stored hash
func VerifyAPIKey(fullKey string) (bool, error) {
	if keyStore == nil {
		return false, ErrNoKeyStore
	}

	// Parse key format: {id}.{secret}
	parts := strings.SplitN(fullKey, ".", 2)
	if len(parts) != 2 {
		return false, ErrInvalidAPIKeyFormat
	}

	id := parts[0]
	secret := parts[1]

	// Get stored key
	key, err := keyStore.GetByID(id)
	if err != nil {
		if err == ErrAPIKeyNotFound {
			return false, nil
		}
		return false, err
	}

	// Hash provided secret with stored salt
	providedHash := hashSecret(secret, key.Salt)

	// Constant-time comparison
	if subtle.ConstantTimeCompare(providedHash, key.Hash) == 1 {
		return true, nil
	}

	return false, nil
}

// ListAPIKeys returns all API key IDs (not secrets)
func ListAPIKeys() ([]*APIKey, error) {
	if keyStore == nil {
		return nil, ErrNoKeyStore
	}
	return keyStore.List()
}

// DeleteAPIKey removes an API key by ID
func DeleteAPIKey(id string) error {
	if keyStore == nil {
		return ErrNoKeyStore
	}
	return keyStore.Delete(id)
}

// hashSecret creates a SHA256 hash of secret + salt
func hashSecret(secret string, salt []byte) []byte {
	h := sha256.New()
	h.Write([]byte(secret))
	h.Write(salt)
	return h.Sum(nil)
}

// generateRandomString creates a random base64url string
func generateRandomString(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		panic(err) // Should never happen
	}
	return base64.RawURLEncoding.EncodeToString(bytes)
}
