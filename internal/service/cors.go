package service

import (
	"strings"
	"time"
)

// CORSStore interface for CORS origin operations
type CORSStore interface {
	Add(origin string, createdAt int64) error
	List() ([]string, error)
	Delete(origin string) error
	IsAllowed(origin string) (bool, error)
}

var corsStore CORSStore

// SetCORSStore sets the CORS store implementation
func SetCORSStore(s CORSStore) {
	corsStore = s
}

// AddCORSOrigin adds an allowed origin to the whitelist
func AddCORSOrigin(origin string) error {
	if corsStore == nil {
		return ErrNoCORSStore
	}

	origin = strings.TrimSpace(origin)
	if origin == "" {
		return ErrOriginNotAllowed
	}

	createdAt := time.Now().Unix()
	return corsStore.Add(origin, createdAt)
}

// ListCORSOrigins returns all allowed origins
func ListCORSOrigins() ([]string, error) {
	if corsStore == nil {
		return nil, ErrNoCORSStore
	}
	return corsStore.List()
}

// DeleteCORSOrigin removes an origin from the whitelist
func DeleteCORSOrigin(origin string) error {
	if corsStore == nil {
		return ErrNoCORSStore
	}
	return corsStore.Delete(origin)
}

// IsOriginAllowed checks if an origin is in the whitelist
func IsOriginAllowed(origin string) (bool, error) {
	if corsStore == nil {
		return false, ErrNoCORSStore
	}

	// Empty origin header means non-browser request
	if origin == "" {
		return true, nil
	}

	return corsStore.IsAllowed(origin)
}
