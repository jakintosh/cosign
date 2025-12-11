package service

import (
	"errors"
	"fmt"
)

// Errors
var (
	ErrNoSignonStore          = errors.New("signon store not configured")
	ErrNoLocationConfigStore  = errors.New("location config store not configured")
	ErrNoLocationOptionStore  = errors.New("location option store not configured")
	ErrNoCampaignStore        = errors.New("campaign store not configured")
	ErrNoKeyStore             = errors.New("key store not configured")
	ErrNoCORSStore            = errors.New("cors store not configured")

	ErrSignonNotFound         = errors.New("signon not found")
	ErrLocationConfigNotFound = errors.New("location config not found")
	ErrLocationOptionNotFound = errors.New("location option not found")
	ErrCampaignNotFound       = errors.New("campaign not found")
	ErrAPIKeyNotFound         = errors.New("api key not found")
	ErrCORSOriginNotFound     = errors.New("cors origin not found")
	ErrInvalidEmail           = errors.New("invalid email address")
	ErrDuplicateEmail         = errors.New("email already signed")
	ErrInvalidLocation        = errors.New("invalid location")
	ErrLocationNotInOptions   = errors.New("location must be from preset options")
	ErrEmptyName              = errors.New("name cannot be empty")
	ErrEmptyEmail             = errors.New("email cannot be empty")
	ErrEmptyLocation          = errors.New("location cannot be empty")
	ErrEmptyCampaignName      = errors.New("campaign name cannot be empty")
	ErrInvalidAPIKeyFormat    = errors.New("invalid api key format")
	ErrUnauthorized           = errors.New("unauthorized")
	ErrOriginNotAllowed       = errors.New("origin not allowed")
)

type DatabaseError struct{ Err error }

func (e DatabaseError) Error() string { return fmt.Sprintf("database error: %v", e.Err) }
func (e DatabaseError) Unwrap() error { return e.Err }
