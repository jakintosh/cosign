package service

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

// LocationOptionStore interface for location option operations (config is now in Campaign)
type LocationOptionStore interface {
	GetOptions(campaignID string) ([]*LocationOption, error)
	AddOption(campaignID, value string, displayOrder int) (int64, error)
	UpdateOption(campaignID string, id int64, value string, displayOrder int) error
	DeleteOption(campaignID string, id int64) error
}

var locationOptionStore LocationOptionStore

// SetLocationOptionStore sets the location option store implementation
func SetLocationOptionStore(s LocationOptionStore) {
	locationOptionStore = s
}

// GetLocationConfig retrieves the current location field configuration for a campaign
func GetLocationConfig(campaignID string) (*LocationConfig, error) {
	if campaignStore == nil {
		return nil, ErrNoCampaignStore
	}

	campaign, err := campaignStore.GetByID(campaignID)
	if err != nil {
		return nil, err
	}

	return &LocationConfig{
		AllowCustomText: campaign.AllowCustomText,
	}, nil
}

// SetAllowCustomText updates whether custom location text is allowed for a campaign
func SetAllowCustomText(campaignID string, allow bool) error {
	if campaignStore == nil {
		return ErrNoCampaignStore
	}

	campaign, err := campaignStore.GetByID(campaignID)
	if err != nil {
		return err
	}

	return campaignStore.Update(campaignID, campaign.Name, allow)
}

// GetLocationOptions retrieves all preset location options for a campaign
func GetLocationOptions(campaignID string) ([]*LocationOption, error) {
	if locationOptionStore == nil {
		return nil, ErrNoLocationOptionStore
	}
	return locationOptionStore.GetOptions(campaignID)
}

// AddLocationOption adds a new preset location option to a campaign
func AddLocationOption(campaignID, value string, displayOrder int) (int64, error) {
	if locationOptionStore == nil {
		return 0, ErrNoLocationOptionStore
	}

	if value == "" {
		return 0, ErrEmptyLocation
	}

	return locationOptionStore.AddOption(campaignID, value, displayOrder)
}

// UpdateLocationOption updates an existing location option
func UpdateLocationOption(campaignID string, id int64, value string, displayOrder int) error {
	if locationOptionStore == nil {
		return ErrNoLocationOptionStore
	}

	if value == "" {
		return ErrEmptyLocation
	}

	return locationOptionStore.UpdateOption(campaignID, id, value, displayOrder)
}

// DeleteLocationOption removes a preset location option
func DeleteLocationOption(campaignID string, id int64) error {
	if locationOptionStore == nil {
		return ErrNoLocationOptionStore
	}
	return locationOptionStore.DeleteOption(campaignID, id)
}

// GetLocationConfigWithOptions retrieves both config and options for a campaign
func GetLocationConfigWithOptions(campaignID string) (*LocationConfig, []*LocationOption, error) {
	config, err := GetLocationConfig(campaignID)
	if err != nil {
		return nil, nil, err
	}

	options, err := GetLocationOptions(campaignID)
	if err != nil {
		return nil, nil, err
	}

	return config, options, nil
}
