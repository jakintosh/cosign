package service

// LocationConfigStore interface for location configuration operations
type LocationConfigStore interface {
	GetConfig() (*LocationConfig, error)
	SetAllowCustomText(allow bool) error
	GetOptions() ([]*LocationOption, error)
	AddOption(value string, displayOrder int) (int64, error)
	UpdateOption(id int64, value string, displayOrder int) error
	DeleteOption(id int64) error
}

var locationConfigStore LocationConfigStore

// SetLocationConfigStore sets the location config store implementation
func SetLocationConfigStore(s LocationConfigStore) {
	locationConfigStore = s
}

// GetLocationConfig retrieves the current location field configuration
func GetLocationConfig() (*LocationConfig, error) {
	if locationConfigStore == nil {
		return nil, ErrNoLocationConfigStore
	}
	return locationConfigStore.GetConfig()
}

// SetAllowCustomText updates whether custom location text is allowed
func SetAllowCustomText(allow bool) error {
	if locationConfigStore == nil {
		return ErrNoLocationConfigStore
	}
	return locationConfigStore.SetAllowCustomText(allow)
}

// GetLocationOptions retrieves all preset location options
func GetLocationOptions() ([]*LocationOption, error) {
	if locationConfigStore == nil {
		return nil, ErrNoLocationConfigStore
	}
	return locationConfigStore.GetOptions()
}

// AddLocationOption adds a new preset location option
func AddLocationOption(value string, displayOrder int) (int64, error) {
	if locationConfigStore == nil {
		return 0, ErrNoLocationConfigStore
	}

	if value == "" {
		return 0, ErrEmptyLocation
	}

	return locationConfigStore.AddOption(value, displayOrder)
}

// UpdateLocationOption updates an existing location option
func UpdateLocationOption(id int64, value string, displayOrder int) error {
	if locationConfigStore == nil {
		return ErrNoLocationConfigStore
	}

	if value == "" {
		return ErrEmptyLocation
	}

	return locationConfigStore.UpdateOption(id, value, displayOrder)
}

// DeleteLocationOption removes a preset location option
func DeleteLocationOption(id int64) error {
	if locationConfigStore == nil {
		return ErrNoLocationConfigStore
	}
	return locationConfigStore.DeleteOption(id)
}

// GetLocationConfigWithOptions returns both config and options together
func GetLocationConfigWithOptions() (*LocationConfig, []*LocationOption, error) {
	if locationConfigStore == nil {
		return nil, nil, ErrNoLocationConfigStore
	}

	config, err := locationConfigStore.GetConfig()
	if err != nil {
		return nil, nil, err
	}

	options, err := locationConfigStore.GetOptions()
	if err != nil {
		return nil, nil, err
	}

	return config, options, nil
}
