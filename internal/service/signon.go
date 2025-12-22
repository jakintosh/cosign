package service

import (
	"regexp"
	"strings"
	"time"
)

// Signon represents a public letter sign-on
type Signon struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Location  string `json:"location"`
	CreatedAt int64  `json:"created_at"`
}

// Signons wraps a list response with pagination metadata
type Signons struct {
	Signons []*Signon `json:"signons"`
	Total   int       `json:"total"`
	Limit   int       `json:"limit"`
	Offset  int       `json:"offset"`
}

// SignonStore interface for data operations
type SignonStore interface {
	Insert(campaignID, name, email, location string, createdAt int64) (int64, error)
	GetByID(campaignID string, id int64) (*Signon, error)
	List(campaignID string, limit, offset int) ([]*Signon, error)
	Count(campaignID string) (int, error)
	Delete(campaignID string, id int64) error
	EmailExists(campaignID, email string) (bool, error)
}

var signonStore SignonStore

// SetSignonStore sets the signon store implementation
func SetSignonStore(s SignonStore) {
	signonStore = s
}

// Email validation regex
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// CreateSignon creates a new sign-on entry with validation
func CreateSignon(
	campaignID string,
	name string,
	email string,
	location string,
	allowDuplicates bool,
) (*Signon, error) {
	if signonStore == nil {
		return nil, ErrNoSignonStore
	}

	// Validate inputs
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(email)
	location = strings.TrimSpace(location)

	if name == "" {
		return nil, ErrEmptyName
	}
	if email == "" {
		return nil, ErrEmptyEmail
	}
	if location == "" {
		return nil, ErrEmptyLocation
	}

	// Validate email format
	if !emailRegex.MatchString(email) {
		return nil, ErrInvalidEmail
	}

	// Check for duplicate email if not allowed
	if !allowDuplicates {
		exists, err := signonStore.EmailExists(campaignID, email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrDuplicateEmail
		}
	}

	// Validate location against configured options
	if err := validateLocation(campaignID, location); err != nil {
		return nil, err
	}

	// Insert
	createdAt := time.Now().Unix()
	id, err := signonStore.Insert(campaignID, name, email, location, createdAt)
	if err != nil {
		return nil, err
	}

	return &Signon{
		ID:        id,
		Name:      name,
		Email:     email,
		Location:  location,
		CreatedAt: createdAt,
	}, nil
}

// GetSignon retrieves a sign-on by ID
func GetSignon(campaignID string, id int64) (*Signon, error) {
	if signonStore == nil {
		return nil, ErrNoSignonStore
	}
	return signonStore.GetByID(campaignID, id)
}

// ListSignons retrieves all sign-ons with pagination and total count
func ListSignons(campaignID string, limit, offset int) (*Signons, error) {
	if signonStore == nil {
		return nil, ErrNoSignonStore
	}

	// Default pagination
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	list, err := signonStore.List(campaignID, limit, offset)
	if err != nil {
		return nil, err
	}

	total, err := signonStore.Count(campaignID)
	if err != nil {
		return nil, err
	}

	return &Signons{
		Signons: list,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	}, nil
}

// DeleteSignon removes a sign-on by ID
func DeleteSignon(campaignID string, id int64) error {
	if signonStore == nil {
		return ErrNoSignonStore
	}
	return signonStore.Delete(campaignID, id)
}

// validateLocation checks if the location is valid based on campaign configuration
func validateLocation(campaignID, location string) error {
	if campaignStore == nil {
		// If no campaign store, allow any location
		return nil
	}

	campaign, err := campaignStore.GetByID(campaignID)
	if err != nil {
		// If campaign doesn't exist, allow any location
		if err == ErrCampaignNotFound {
			return nil
		}
		return err
	}

	// If custom text is allowed, any location is valid
	if campaign.AllowCustomText {
		return nil
	}

	// Otherwise, location must be in the preset options
	options, err := GetCampaignLocations(campaignID)
	if err != nil {
		return err
	}

	// If no options configured, allow any location
	if len(options) == 0 {
		return nil
	}

	// Check if location matches one of the options
	for _, opt := range options {
		if opt.Value == location {
			return nil
		}
	}

	return ErrLocationNotInOptions
}
