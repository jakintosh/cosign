package service

import (
	"strings"
	"time"

	"github.com/google/uuid" // from go.mod
)

// Campaign represents a sign-on campaign
type Campaign struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	AllowCustomText bool   `json:"allow_custom_text"`
	CreatedAt       int64  `json:"created_at"`
}

// Campaigns wraps a list response with pagination metadata
type Campaigns struct {
	Campaigns []*Campaign `json:"campaigns"`
	Total     int         `json:"total"`
	Limit     int         `json:"limit"`
	Offset    int         `json:"offset"`
}

// CampaignStore interface for campaign data operations
type CampaignStore interface {
	Insert(id, name string, allowCustomText bool, createdAt int64) error
	GetByID(id string) (*Campaign, error)
	List(limit, offset int) ([]*Campaign, error)
	Count() (int, error)
	Update(id, name string, allowCustomText bool) error
	Delete(id string) error
}

var campaignStore CampaignStore

// SetCampaignStore sets the campaign store implementation
func SetCampaignStore(s CampaignStore) {
	campaignStore = s
}

// CreateCampaign creates a new campaign with a generated UUID
func CreateCampaign(name string) (*Campaign, error) {
	if campaignStore == nil {
		return nil, ErrNoCampaignStore
	}

	// Validate name
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrEmptyCampaignName
	}

	// Generate UUID
	id := uuid.New().String()

	// Create campaign with default settings
	createdAt := time.Now().Unix()
	err := campaignStore.Insert(id, name, true, createdAt)
	if err != nil {
		return nil, err
	}

	return &Campaign{
		ID:              id,
		Name:            name,
		AllowCustomText: true,
		CreatedAt:       createdAt,
	}, nil
}

// GetCampaign retrieves a campaign by ID
func GetCampaign(id string) (*Campaign, error) {
	if campaignStore == nil {
		return nil, ErrNoCampaignStore
	}
	return campaignStore.GetByID(id)
}

// ListCampaigns retrieves all campaigns with pagination and total count
func ListCampaigns(limit, offset int) (*Campaigns, error) {
	if campaignStore == nil {
		return nil, ErrNoCampaignStore
	}

	// Default pagination
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	campaigns, err := campaignStore.List(limit, offset)
	if err != nil {
		return nil, err
	}

	total, err := campaignStore.Count()
	if err != nil {
		return nil, err
	}

	return &Campaigns{
		Campaigns: campaigns,
		Total:     total,
		Limit:     limit,
		Offset:    offset,
	}, nil
}

// UpdateCampaign updates a campaign's name and config
func UpdateCampaign(id, name string, allowCustomText bool) error {
	if campaignStore == nil {
		return ErrNoCampaignStore
	}

	// Validate name
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrEmptyCampaignName
	}

	return campaignStore.Update(id, name, allowCustomText)
}

// DeleteCampaign removes a campaign by ID
func DeleteCampaign(id string) error {
	if campaignStore == nil {
		return ErrNoCampaignStore
	}
	return campaignStore.Delete(id)
}
