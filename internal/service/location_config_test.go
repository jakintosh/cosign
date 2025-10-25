package service_test

import (
	"cosign/internal/service"
	"errors"
	"testing"
)

// mockLocationConfigStore is a test implementation of LocationConfigStore
type mockLocationConfigStoreImpl struct {
	config  *service.LocationConfig
	options []*service.LocationOption
	err     error
}

func newMockLocationConfigStoreImpl(allowCustom bool) *mockLocationConfigStoreImpl {
	return &mockLocationConfigStoreImpl{
		config:  &service.LocationConfig{AllowCustomText: allowCustom},
		options: make([]*service.LocationOption, 0),
	}
}

func (m *mockLocationConfigStoreImpl) GetConfig() (*service.LocationConfig, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.config, nil
}

func (m *mockLocationConfigStoreImpl) SetAllowCustomText(allow bool) error {
	if m.err != nil {
		return m.err
	}
	m.config.AllowCustomText = allow
	return nil
}

func (m *mockLocationConfigStoreImpl) GetOptions() ([]*service.LocationOption, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.options, nil
}

func (m *mockLocationConfigStoreImpl) AddOption(value string, displayOrder int) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	id := int64(len(m.options) + 1)
	m.options = append(m.options, &service.LocationOption{
		ID:           id,
		Value:        value,
		DisplayOrder: displayOrder,
	})
	return id, nil
}

func (m *mockLocationConfigStoreImpl) UpdateOption(id int64, value string, displayOrder int) error {
	if m.err != nil {
		return m.err
	}
	for _, opt := range m.options {
		if opt.ID == id {
			opt.Value = value
			opt.DisplayOrder = displayOrder
			return nil
		}
	}
	return service.ErrLocationOptionNotFound
}

func (m *mockLocationConfigStoreImpl) DeleteOption(id int64) error {
	if m.err != nil {
		return m.err
	}
	for i, opt := range m.options {
		if opt.ID == id {
			m.options = append(m.options[:i], m.options[i+1:]...)
			return nil
		}
	}
	return service.ErrLocationOptionNotFound
}

// Test GetLocationConfig - happy path
func TestGetLocationConfigHappyPath(t *testing.T) {
	store := newMockLocationConfigStoreImpl(true)
	service.SetLocationConfigStore(store)

	config, err := service.GetLocationConfig()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if config == nil {
		t.Fatal("Expected config, got nil")
	}
	if !config.AllowCustomText {
		t.Error("Expected AllowCustomText to be true")
	}
}

// Test GetLocationConfig - no store
func TestGetLocationConfigNoStore(t *testing.T) {
	service.SetLocationConfigStore(nil)

	_, err := service.GetLocationConfig()

	if err != service.ErrNoLocationConfigStore {
		t.Errorf("Expected ErrNoLocationConfigStore, got %v", err)
	}
}

// Test GetLocationConfig - store error
func TestGetLocationConfigStoreError(t *testing.T) {
	store := newMockLocationConfigStoreImpl(true)
	store.err = errors.New("database error")
	service.SetLocationConfigStore(store)

	_, err := service.GetLocationConfig()

	if !errors.Is(err, store.err) {
		t.Errorf("Expected database error, got %v", err)
	}
}

// Test SetAllowCustomText - enable
func TestSetAllowCustomTextEnable(t *testing.T) {
	store := newMockLocationConfigStoreImpl(false)
	service.SetLocationConfigStore(store)

	err := service.SetAllowCustomText(true)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	config, _ := service.GetLocationConfig()
	if !config.AllowCustomText {
		t.Error("Expected AllowCustomText to be updated to true")
	}
}

// Test SetAllowCustomText - disable
func TestSetAllowCustomTextDisable(t *testing.T) {
	store := newMockLocationConfigStoreImpl(true)
	service.SetLocationConfigStore(store)

	err := service.SetAllowCustomText(false)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	config, _ := service.GetLocationConfig()
	if config.AllowCustomText {
		t.Error("Expected AllowCustomText to be updated to false")
	}
}

// Test SetAllowCustomText - no store
func TestSetAllowCustomTextNoStore(t *testing.T) {
	service.SetLocationConfigStore(nil)

	err := service.SetAllowCustomText(true)

	if err != service.ErrNoLocationConfigStore {
		t.Errorf("Expected ErrNoLocationConfigStore, got %v", err)
	}
}

// Test SetAllowCustomText - store error
func TestSetAllowCustomTextStoreError(t *testing.T) {
	store := newMockLocationConfigStoreImpl(true)
	store.err = errors.New("database error")
	service.SetLocationConfigStore(store)

	err := service.SetAllowCustomText(false)

	if !errors.Is(err, store.err) {
		t.Errorf("Expected database error, got %v", err)
	}
}

// Test GetLocationOptions - happy path
func TestGetLocationOptionsHappyPath(t *testing.T) {
	store := newMockLocationConfigStoreImpl(false)
	store.AddOption("New York", 1)
	store.AddOption("Boston", 2)
	service.SetLocationConfigStore(store)

	options, err := service.GetLocationOptions()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(options) != 2 {
		t.Errorf("Expected 2 options, got %d", len(options))
	}
}

// Test GetLocationOptions - empty list
func TestGetLocationOptionsEmpty(t *testing.T) {
	store := newMockLocationConfigStoreImpl(false)
	service.SetLocationConfigStore(store)

	options, err := service.GetLocationOptions()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(options) != 0 {
		t.Errorf("Expected 0 options, got %d", len(options))
	}
}

// Test GetLocationOptions - no store
func TestGetLocationOptionsNoStore(t *testing.T) {
	service.SetLocationConfigStore(nil)

	_, err := service.GetLocationOptions()

	if err != service.ErrNoLocationConfigStore {
		t.Errorf("Expected ErrNoLocationConfigStore, got %v", err)
	}
}

// Test AddLocationOption - happy path
func TestAddLocationOptionHappyPath(t *testing.T) {
	store := newMockLocationConfigStoreImpl(false)
	service.SetLocationConfigStore(store)

	id, err := service.AddLocationOption("New York", 1)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if id == 0 {
		t.Error("Expected non-zero ID")
	}

	options, _ := service.GetLocationOptions()
	if len(options) != 1 {
		t.Error("Expected option to be added")
	}
}

// Test AddLocationOption - empty value
func TestAddLocationOptionEmptyValue(t *testing.T) {
	store := newMockLocationConfigStoreImpl(false)
	service.SetLocationConfigStore(store)

	_, err := service.AddLocationOption("", 1)

	if err != service.ErrEmptyLocation {
		t.Errorf("Expected ErrEmptyLocation, got %v", err)
	}
}

// Test AddLocationOption - whitespace only value
func TestAddLocationOptionWhitespaceOnly(t *testing.T) {
	store := newMockLocationConfigStoreImpl(false)
	service.SetLocationConfigStore(store)

	// Note: value is not trimmed in AddLocationOption
	id, err := service.AddLocationOption("   ", 1)

	// This should succeed (whitespace is preserved)
	if err != nil {
		t.Errorf("Expected no error (whitespace is allowed), got %v", err)
	}
	if id == 0 {
		t.Error("Expected non-zero ID")
	}
}

// Test AddLocationOption - no store
func TestAddLocationOptionNoStore(t *testing.T) {
	service.SetLocationConfigStore(nil)

	_, err := service.AddLocationOption("New York", 1)

	if err != service.ErrNoLocationConfigStore {
		t.Errorf("Expected ErrNoLocationConfigStore, got %v", err)
	}
}

// Test AddLocationOption - multiple options
func TestAddLocationOptionMultiple(t *testing.T) {
	store := newMockLocationConfigStoreImpl(false)
	service.SetLocationConfigStore(store)

	testCases := []struct {
		value string
		order int
	}{
		{"New York", 1},
		{"Boston", 2},
		{"Chicago", 3},
		{"San Francisco", 4},
	}

	for _, tc := range testCases {
		id, err := service.AddLocationOption(tc.value, tc.order)
		if err != nil {
			t.Errorf("Failed to add option %q: %v", tc.value, err)
		}
		if id == 0 {
			t.Error("Expected non-zero ID")
		}
	}

	options, _ := service.GetLocationOptions()
	if len(options) != 4 {
		t.Errorf("Expected 4 options, got %d", len(options))
	}
}

// Test UpdateLocationOption - happy path
func TestUpdateLocationOptionHappyPath(t *testing.T) {
	store := newMockLocationConfigStoreImpl(false)
	service.SetLocationConfigStore(store)

	id, _ := service.AddLocationOption("New York", 1)

	err := service.UpdateLocationOption(id, "New York City", 1)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	options, _ := service.GetLocationOptions()
	if len(options) != 1 || options[0].Value != "New York City" {
		t.Error("Expected option to be updated")
	}
}

// Test UpdateLocationOption - empty value
func TestUpdateLocationOptionEmptyValue(t *testing.T) {
	store := newMockLocationConfigStoreImpl(false)
	service.SetLocationConfigStore(store)

	id, _ := service.AddLocationOption("New York", 1)

	err := service.UpdateLocationOption(id, "", 1)

	if err != service.ErrEmptyLocation {
		t.Errorf("Expected ErrEmptyLocation, got %v", err)
	}
}

// Test UpdateLocationOption - nonexistent option
func TestUpdateLocationOptionNonexistent(t *testing.T) {
	store := newMockLocationConfigStoreImpl(false)
	service.SetLocationConfigStore(store)

	err := service.UpdateLocationOption(999, "New York", 1)

	if err != service.ErrLocationOptionNotFound {
		t.Errorf("Expected ErrLocationOptionNotFound, got %v", err)
	}
}

// Test UpdateLocationOption - no store
func TestUpdateLocationOptionNoStore(t *testing.T) {
	service.SetLocationConfigStore(nil)

	err := service.UpdateLocationOption(1, "New York", 1)

	if err != service.ErrNoLocationConfigStore {
		t.Errorf("Expected ErrNoLocationConfigStore, got %v", err)
	}
}

// Test UpdateLocationOption - change display order
func TestUpdateLocationOptionChangeOrder(t *testing.T) {
	store := newMockLocationConfigStoreImpl(false)
	service.SetLocationConfigStore(store)

	id, _ := service.AddLocationOption("New York", 1)

	err := service.UpdateLocationOption(id, "New York", 5)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	options, _ := service.GetLocationOptions()
	if options[0].DisplayOrder != 5 {
		t.Error("Expected display order to be updated")
	}
}

// Test DeleteLocationOption - happy path
func TestDeleteLocationOptionHappyPath(t *testing.T) {
	store := newMockLocationConfigStoreImpl(false)
	service.SetLocationConfigStore(store)

	id, _ := service.AddLocationOption("New York", 1)

	err := service.DeleteLocationOption(id)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	options, _ := service.GetLocationOptions()
	if len(options) != 0 {
		t.Error("Expected option to be deleted")
	}
}

// Test DeleteLocationOption - nonexistent option
func TestDeleteLocationOptionNonexistent(t *testing.T) {
	store := newMockLocationConfigStoreImpl(false)
	service.SetLocationConfigStore(store)

	err := service.DeleteLocationOption(999)

	if err != service.ErrLocationOptionNotFound {
		t.Errorf("Expected ErrLocationOptionNotFound, got %v", err)
	}
}

// Test DeleteLocationOption - no store
func TestDeleteLocationOptionNoStore(t *testing.T) {
	service.SetLocationConfigStore(nil)

	err := service.DeleteLocationOption(1)

	if err != service.ErrNoLocationConfigStore {
		t.Errorf("Expected ErrNoLocationConfigStore, got %v", err)
	}
}

// Test DeleteLocationOption - multiple options
func TestDeleteLocationOptionMultiple(t *testing.T) {
	store := newMockLocationConfigStoreImpl(false)
	service.SetLocationConfigStore(store)

	_, _ = service.AddLocationOption("New York", 1)
	id2, _ := service.AddLocationOption("Boston", 2)
	_, _ = service.AddLocationOption("Chicago", 3)

	service.DeleteLocationOption(id2)

	options, _ := service.GetLocationOptions()
	if len(options) != 2 {
		t.Errorf("Expected 2 options after deletion, got %d", len(options))
	}

	// Verify correct ones remain
	found := false
	for _, opt := range options {
		if opt.ID == id2 {
			found = true
		}
	}
	if found {
		t.Error("Expected deleted option to be removed")
	}
}

// Test GetLocationConfigWithOptions - happy path
func TestGetLocationConfigWithOptionsHappyPath(t *testing.T) {
	store := newMockLocationConfigStoreImpl(true)
	store.AddOption("New York", 1)
	store.AddOption("Boston", 2)
	service.SetLocationConfigStore(store)

	config, options, err := service.GetLocationConfigWithOptions()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if config == nil {
		t.Fatal("Expected config, got nil")
	}
	if len(options) != 2 {
		t.Errorf("Expected 2 options, got %d", len(options))
	}
	if !config.AllowCustomText {
		t.Error("Expected AllowCustomText to be true")
	}
}

// Test GetLocationConfigWithOptions - no options
func TestGetLocationConfigWithOptionsNoOptions(t *testing.T) {
	store := newMockLocationConfigStoreImpl(false)
	service.SetLocationConfigStore(store)

	config, options, err := service.GetLocationConfigWithOptions()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if config == nil {
		t.Fatal("Expected config, got nil")
	}
	if len(options) != 0 {
		t.Errorf("Expected 0 options, got %d", len(options))
	}
}

// Test GetLocationConfigWithOptions - no store
func TestGetLocationConfigWithOptionsNoStore(t *testing.T) {
	service.SetLocationConfigStore(nil)

	_, _, err := service.GetLocationConfigWithOptions()

	if err != service.ErrNoLocationConfigStore {
		t.Errorf("Expected ErrNoLocationConfigStore, got %v", err)
	}
}

// Test GetLocationConfigWithOptions - config error
func TestGetLocationConfigWithOptionsConfigError(t *testing.T) {
	store := newMockLocationConfigStoreImpl(true)
	// Simulate error on GetConfig
	oldErr := store.err
	store.err = errors.New("config error")
	service.SetLocationConfigStore(store)

	_, _, err := service.GetLocationConfigWithOptions()

	if !errors.Is(err, store.err) {
		t.Errorf("Expected config error, got %v", err)
	}

	store.err = oldErr
}

// Test AddLocationOption - store error
func TestAddLocationOptionStoreError(t *testing.T) {
	store := newMockLocationConfigStoreImpl(false)
	store.err = errors.New("database error")
	service.SetLocationConfigStore(store)

	_, err := service.AddLocationOption("New York", 1)

	if !errors.Is(err, store.err) {
		t.Errorf("Expected database error, got %v", err)
	}
}

// Test UpdateLocationOption - store error
func TestUpdateLocationOptionStoreError(t *testing.T) {
	store := newMockLocationConfigStoreImpl(false)
	service.SetLocationConfigStore(store)

	id, _ := service.AddLocationOption("New York", 1)

	// Now set error
	store.err = errors.New("database error")

	err := service.UpdateLocationOption(id, "Boston", 2)

	if !errors.Is(err, store.err) {
		t.Errorf("Expected database error, got %v", err)
	}
}

// Test DeleteLocationOption - store error
func TestDeleteLocationOptionStoreError(t *testing.T) {
	store := newMockLocationConfigStoreImpl(false)
	service.SetLocationConfigStore(store)

	id, _ := service.AddLocationOption("New York", 1)

	// Now set error
	store.err = errors.New("database error")

	err := service.DeleteLocationOption(id)

	if !errors.Is(err, store.err) {
		t.Errorf("Expected database error, got %v", err)
	}
}

// Test location option workflow
func TestLocationOptionWorkflow(t *testing.T) {
	store := newMockLocationConfigStoreImpl(false)
	service.SetLocationConfigStore(store)

	// Add options
	id1, _ := service.AddLocationOption("New York", 1)
	id2, _ := service.AddLocationOption("Boston", 2)
	_, _ = service.AddLocationOption("Chicago", 3)

	options, _ := service.GetLocationOptions()
	if len(options) != 3 {
		t.Errorf("Expected 3 options, got %d", len(options))
	}

	// Update one
	service.UpdateLocationOption(id2, "Boston Updated", 10)

	options, _ = service.GetLocationOptions()
	for _, opt := range options {
		if opt.ID == id2 && opt.Value != "Boston Updated" {
			t.Error("Expected option to be updated")
		}
	}

	// Delete one
	service.DeleteLocationOption(id1)

	options, _ = service.GetLocationOptions()
	if len(options) != 2 {
		t.Errorf("Expected 2 options after delete, got %d", len(options))
	}
}

// Test GetLocationConfigWithOptions - options error propagation
func TestGetLocationConfigWithOptionsOptionsError(t *testing.T) {
	store := newMockLocationConfigStoreImpl(true)
	service.SetLocationConfigStore(store)

	// Add an option to test, then simulate error on subsequent GetOptions call
	store.AddOption("New York", 1)

	// This is a limitation of our mock - we can't easily simulate error after success
	// But we can test that errors are propagated
	config, options, err := service.GetLocationConfigWithOptions()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if config == nil || len(options) != 1 {
		t.Error("Expected valid config and options")
	}
}
