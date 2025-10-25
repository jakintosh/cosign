package service

import "errors"

// Signon represents a public letter sign-on
type Signon struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Location  string `json:"location"`
	CreatedAt int64  `json:"created_at"`
}

// LocationConfig holds the location field configuration
type LocationConfig struct {
	AllowCustomText bool `json:"allow_custom_text"`
}

// LocationOption represents a preset location option
type LocationOption struct {
	ID           int64  `json:"id"`
	Value        string `json:"value"`
	DisplayOrder int    `json:"display_order"`
}

// APIKey represents an API key for authentication
type APIKey struct {
	ID        string `json:"id"`
	Hash      []byte `json:"-"`
	Salt      []byte `json:"-"`
	CreatedAt int64  `json:"created_at"`
}

// Errors
var (
	ErrNoSignonStore             = errors.New("signon store not configured")
	ErrNoLocationConfigStore     = errors.New("location config store not configured")
	ErrNoKeyStore                = errors.New("key store not configured")
	ErrNoCORSStore               = errors.New("cors store not configured")
	ErrSignonNotFound            = errors.New("signon not found")
	ErrLocationConfigNotFound    = errors.New("location config not found")
	ErrLocationOptionNotFound    = errors.New("location option not found")
	ErrAPIKeyNotFound            = errors.New("api key not found")
	ErrCORSOriginNotFound        = errors.New("cors origin not found")
	ErrInvalidEmail              = errors.New("invalid email address")
	ErrDuplicateEmail            = errors.New("email already signed")
	ErrInvalidLocation           = errors.New("invalid location")
	ErrLocationNotInOptions      = errors.New("location must be from preset options")
	ErrEmptyName                 = errors.New("name cannot be empty")
	ErrEmptyEmail                = errors.New("email cannot be empty")
	ErrEmptyLocation             = errors.New("location cannot be empty")
	ErrInvalidAPIKeyFormat       = errors.New("invalid api key format")
	ErrUnauthorized              = errors.New("unauthorized")
	ErrOriginNotAllowed          = errors.New("origin not allowed")
)
