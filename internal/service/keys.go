package service

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"
)

type KeyStore interface {
	CountKeys() (int, error)
	DeleteKey(id string) error
	FetchKey(id string) (salt string, hash string, err error)
	InsertKey(id string, salt string, hash string) error
}

var keyStore KeyStore

func SetKeyStore(s KeyStore) {
	keyStore = s
}

// InitKeys exists to check if the keystore is empty, and if so, to populate it with
// a bootstrap key provided by the apiKey parameter. it exits early if there are any
// existing keys in the store
func InitKeys(
	apiKey string,
) error {
	if keyStore == nil {
		return ErrNoKeyStore
	}

	count, err := keyStore.CountKeys()
	if err != nil {
		return DatabaseError{err}
	}
	if count > 0 {
		return nil
	}

	// generate a salt
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return err
	}

	// extract id and secret from apiKey
	parts := strings.Split(apiKey, ".")
	if len(parts) != 2 {
		return fmt.Errorf("incorrectly formatted apiKey")
	}
	id := parts[0]
	secretHex := parts[1]
	secret, err := hex.DecodeString(secretHex)
	if err != nil {
		return err
	}

	if err := registerKey(id, salt, secret); err != nil {
		return err
	}

	return nil
}

func CreateAPIKey() (
	string,
	error,
) {
	if keyStore == nil {
		return "", ErrNoKeyStore
	}

	idBytes := make([]byte, 8)
	if _, err := rand.Read(idBytes); err != nil {
		return "", err
	}
	saltBytes := make([]byte, 16)
	if _, err := rand.Read(saltBytes); err != nil {
		return "", err
	}
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		return "", err
	}

	id := hex.EncodeToString(idBytes)
	secret := hex.EncodeToString(secretBytes)
	if err := registerKey(id, saltBytes, secretBytes); err != nil {
		return "", err
	}

	token := id + "." + secret
	return token, nil
}

func DeleteAPIKey(
	id string,
) error {
	if keyStore == nil {
		return ErrNoKeyStore
	}
	if err := keyStore.DeleteKey(id); err != nil {
		return DatabaseError{err}
	}
	return nil
}

func VerifyAPIKey(
	token string,
) (
	bool,
	error,
) {
	if keyStore == nil {
		return false, ErrNoKeyStore
	}

	// get parts of key
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return false, nil
	}
	id := parts[0]
	secretHex := parts[1]

	// get salt and hash
	saltHex, hashHex, err := keyStore.FetchKey(id)
	if err != nil {
		// couldn't fetch the key for some reason
		return false, DatabaseError{err}
	}

	// decode the salt
	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		// invalid hex, this should be unreachable
		return false, err
	}

	// decode the secret
	secret, err := hex.DecodeString(secretHex)
	if err != nil {
		// invalid secret, this should be unreachable
		return false, err
	}

	// decode the hash
	hash, err := hex.DecodeString(hashHex)
	if err != nil {
		// invalid hex, this should be unreachable
		return false, err
	}

	// rebuild the hashed secret
	constructedHash := sha256.Sum256(append(salt, secret...))

	// verify hashes are equal in constant time
	if subtle.ConstantTimeCompare(hash, constructedHash[:]) == 1 {
		return true, nil
	}
	return false, nil
}

func registerKey(
	id string,
	saltBytes []byte,
	secretBytes []byte,
) error {
	hashBytes := sha256.Sum256(append(saltBytes, secretBytes...))

	salt := hex.EncodeToString(saltBytes)
	hash := hex.EncodeToString(hashBytes[:])

	if err := keyStore.InsertKey(
		id,
		salt,
		hash,
	); err != nil {
		return DatabaseError{err}
	}
	return nil
}
