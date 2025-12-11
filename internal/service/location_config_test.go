package service_test

import (
	"cosign/internal/service"
	"errors"
	"testing"
)

const testCampaignID = "test-campaign-id"

// mockCampaignStore is a test implementation of CampaignStore
type mockCampaignStore struct {
	campaigns map[string]*service.Campaign
	err       error
}

func newMockCampaignStore() *mockCampaignStore {
	return &mockCampaignStore{
		campaigns: make(map[string]*service.Campaign),
	}
}

func (m *mockCampaignStore) Insert(id, name string, allowCustomText bool, createdAt int64) error {
	if m.err != nil {
		return m.err
	}
	m.campaigns[id] = &service.Campaign{
		ID:              id,
		Name:            name,
		AllowCustomText: allowCustomText,
		CreatedAt:       createdAt,
	}
	return nil
}

func (m *mockCampaignStore) GetByID(id string) (*service.Campaign, error) {
	if m.err != nil {
		return nil, m.err
	}
	campaign, ok := m.campaigns[id]
	if !ok {
		return nil, service.ErrCampaignNotFound
	}
	return campaign, nil
}

func (m *mockCampaignStore) List(limit, offset int) ([]*service.Campaign, error) {
	if m.err != nil {
		return nil, m.err
	}
	campaigns := make([]*service.Campaign, 0, len(m.campaigns))
	for _, c := range m.campaigns {
		campaigns = append(campaigns, c)
	}
	return campaigns, nil
}

func (m *mockCampaignStore) Count() (int, error) {
	if m.err != nil {
		return 0, m.err
	}
	return len(m.campaigns), nil
}

func (m *mockCampaignStore) Update(id, name string, allowCustomText bool) error {
	if m.err != nil {
		return m.err
	}
	campaign, ok := m.campaigns[id]
	if !ok {
		return service.ErrCampaignNotFound
	}
	campaign.Name = name
	campaign.AllowCustomText = allowCustomText
	return nil
}

func (m *mockCampaignStore) Delete(id string) error {
	if m.err != nil {
		return m.err
	}
	if _, ok := m.campaigns[id]; !ok {
		return service.ErrCampaignNotFound
	}
	delete(m.campaigns, id)
	return nil
}

// mockLocationOptionStore is a test implementation of LocationOptionStore
type mockLocationOptionStore struct {
	options map[string][]*service.LocationOption // keyed by campaignID
	nextID  int64
	err     error
}

func newMockLocationOptionStore() *mockLocationOptionStore {
	return &mockLocationOptionStore{
		options: make(map[string][]*service.LocationOption),
		nextID:  1,
	}
}

func (m *mockLocationOptionStore) GetOptions(campaignID string) ([]*service.LocationOption, error) {
	if m.err != nil {
		return nil, m.err
	}
	opts, ok := m.options[campaignID]
	if !ok {
		return []*service.LocationOption{}, nil
	}
	return opts, nil
}

func (m *mockLocationOptionStore) AddOption(campaignID, value string, displayOrder int) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	id := m.nextID
	m.nextID++

	opt := &service.LocationOption{
		ID:           id,
		Value:        value,
		DisplayOrder: displayOrder,
	}

	m.options[campaignID] = append(m.options[campaignID], opt)
	return id, nil
}

func (m *mockLocationOptionStore) UpdateOption(campaignID string, id int64, value string, displayOrder int) error {
	if m.err != nil {
		return m.err
	}
	opts, ok := m.options[campaignID]
	if !ok {
		return service.ErrLocationOptionNotFound
	}

	for _, opt := range opts {
		if opt.ID == id {
			opt.Value = value
			opt.DisplayOrder = displayOrder
			return nil
		}
	}
	return service.ErrLocationOptionNotFound
}

func (m *mockLocationOptionStore) DeleteOption(campaignID string, id int64) error {
	if m.err != nil {
		return m.err
	}
	opts, ok := m.options[campaignID]
	if !ok {
		return service.ErrLocationOptionNotFound
	}

	for i, opt := range opts {
		if opt.ID == id {
			m.options[campaignID] = append(opts[:i], opts[i+1:]...)
			return nil
		}
	}
	return service.ErrLocationOptionNotFound
}

// Helper to set up test campaign
func setupTestCampaign(campaignStore *mockCampaignStore, allowCustomText bool) {
	campaignStore.Insert(testCampaignID, "Test Campaign", allowCustomText, 1234567890)
}

// Test GetLocationConfig - happy path
func TestGetLocationConfigHappyPath(t *testing.T) {
	campaignStore := newMockCampaignStore()
	setupTestCampaign(campaignStore, true)
	service.SetCampaignStore(campaignStore)

	config, err := service.GetLocationConfig(testCampaignID)

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
	service.SetCampaignStore(nil)

	_, err := service.GetLocationConfig(testCampaignID)

	if err != service.ErrNoCampaignStore {
		t.Errorf("Expected ErrNoCampaignStore, got %v", err)
	}
}

// Test GetLocationConfig - campaign not found
func TestGetLocationConfigCampaignNotFound(t *testing.T) {
	campaignStore := newMockCampaignStore()
	service.SetCampaignStore(campaignStore)

	_, err := service.GetLocationConfig("nonexistent")

	if err != service.ErrCampaignNotFound {
		t.Errorf("Expected ErrCampaignNotFound, got %v", err)
	}
}

// Test SetAllowCustomText - enable
func TestSetAllowCustomTextEnable(t *testing.T) {
	campaignStore := newMockCampaignStore()
	setupTestCampaign(campaignStore, false)
	service.SetCampaignStore(campaignStore)

	err := service.SetAllowCustomText(testCampaignID, true)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	config, _ := service.GetLocationConfig(testCampaignID)
	if !config.AllowCustomText {
		t.Error("Expected AllowCustomText to be updated to true")
	}
}

// Test SetAllowCustomText - disable
func TestSetAllowCustomTextDisable(t *testing.T) {
	campaignStore := newMockCampaignStore()
	setupTestCampaign(campaignStore, true)
	service.SetCampaignStore(campaignStore)

	err := service.SetAllowCustomText(testCampaignID, false)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	config, _ := service.GetLocationConfig(testCampaignID)
	if config.AllowCustomText {
		t.Error("Expected AllowCustomText to be updated to false")
	}
}

// Test SetAllowCustomText - no store
func TestSetAllowCustomTextNoStore(t *testing.T) {
	service.SetCampaignStore(nil)

	err := service.SetAllowCustomText(testCampaignID, true)

	if err != service.ErrNoCampaignStore {
		t.Errorf("Expected ErrNoCampaignStore, got %v", err)
	}
}

// Test SetAllowCustomText - campaign not found
func TestSetAllowCustomTextCampaignNotFound(t *testing.T) {
	campaignStore := newMockCampaignStore()
	service.SetCampaignStore(campaignStore)

	err := service.SetAllowCustomText("nonexistent", true)

	if err != service.ErrCampaignNotFound {
		t.Errorf("Expected ErrCampaignNotFound, got %v", err)
	}
}

// Test GetLocationOptions - happy path
func TestGetLocationOptionsHappyPath(t *testing.T) {
	optionStore := newMockLocationOptionStore()
	optionStore.AddOption(testCampaignID, "New York", 1)
	optionStore.AddOption(testCampaignID, "Boston", 2)
	service.SetLocationOptionStore(optionStore)

	options, err := service.GetLocationOptions(testCampaignID)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(options) != 2 {
		t.Errorf("Expected 2 options, got %d", len(options))
	}
}

// Test GetLocationOptions - empty list
func TestGetLocationOptionsEmpty(t *testing.T) {
	optionStore := newMockLocationOptionStore()
	service.SetLocationOptionStore(optionStore)

	options, err := service.GetLocationOptions(testCampaignID)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(options) != 0 {
		t.Errorf("Expected 0 options, got %d", len(options))
	}
}

// Test GetLocationOptions - no store
func TestGetLocationOptionsNoStore(t *testing.T) {
	service.SetLocationOptionStore(nil)

	_, err := service.GetLocationOptions(testCampaignID)

	if err != service.ErrNoLocationOptionStore {
		t.Errorf("Expected ErrNoLocationOptionStore, got %v", err)
	}
}

// Test AddLocationOption - happy path
func TestAddLocationOptionHappyPath(t *testing.T) {
	optionStore := newMockLocationOptionStore()
	service.SetLocationOptionStore(optionStore)

	id, err := service.AddLocationOption(testCampaignID, "New York", 1)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if id == 0 {
		t.Error("Expected non-zero ID")
	}

	options, _ := service.GetLocationOptions(testCampaignID)
	if len(options) != 1 {
		t.Error("Expected option to be added")
	}
}

// Test AddLocationOption - empty value
func TestAddLocationOptionEmptyValue(t *testing.T) {
	optionStore := newMockLocationOptionStore()
	service.SetLocationOptionStore(optionStore)

	_, err := service.AddLocationOption(testCampaignID, "", 1)

	if err != service.ErrEmptyLocation {
		t.Errorf("Expected ErrEmptyLocation, got %v", err)
	}
}

// Test AddLocationOption - whitespace only value
func TestAddLocationOptionWhitespaceOnly(t *testing.T) {
	optionStore := newMockLocationOptionStore()
	service.SetLocationOptionStore(optionStore)

	// Note: value is not trimmed in AddLocationOption
	id, err := service.AddLocationOption(testCampaignID, "   ", 1)

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
	service.SetLocationOptionStore(nil)

	_, err := service.AddLocationOption(testCampaignID, "New York", 1)

	if err != service.ErrNoLocationOptionStore {
		t.Errorf("Expected ErrNoLocationOptionStore, got %v", err)
	}
}

// Test AddLocationOption - multiple options
func TestAddLocationOptionMultiple(t *testing.T) {
	optionStore := newMockLocationOptionStore()
	service.SetLocationOptionStore(optionStore)

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
		id, err := service.AddLocationOption(testCampaignID, tc.value, tc.order)
		if err != nil {
			t.Errorf("Failed to add option %q: %v", tc.value, err)
		}
		if id == 0 {
			t.Error("Expected non-zero ID")
		}
	}

	options, _ := service.GetLocationOptions(testCampaignID)
	if len(options) != 4 {
		t.Errorf("Expected 4 options, got %d", len(options))
	}
}

// Test UpdateLocationOption - happy path
func TestUpdateLocationOptionHappyPath(t *testing.T) {
	optionStore := newMockLocationOptionStore()
	service.SetLocationOptionStore(optionStore)

	id, _ := service.AddLocationOption(testCampaignID, "New York", 1)

	err := service.UpdateLocationOption(testCampaignID, id, "New York City", 1)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	options, _ := service.GetLocationOptions(testCampaignID)
	if len(options) != 1 || options[0].Value != "New York City" {
		t.Error("Expected option to be updated")
	}
}

// Test UpdateLocationOption - empty value
func TestUpdateLocationOptionEmptyValue(t *testing.T) {
	optionStore := newMockLocationOptionStore()
	service.SetLocationOptionStore(optionStore)

	id, _ := service.AddLocationOption(testCampaignID, "New York", 1)

	err := service.UpdateLocationOption(testCampaignID, id, "", 1)

	if err != service.ErrEmptyLocation {
		t.Errorf("Expected ErrEmptyLocation, got %v", err)
	}
}

// Test UpdateLocationOption - nonexistent option
func TestUpdateLocationOptionNonexistent(t *testing.T) {
	optionStore := newMockLocationOptionStore()
	service.SetLocationOptionStore(optionStore)

	err := service.UpdateLocationOption(testCampaignID, 999, "New York", 1)

	if err != service.ErrLocationOptionNotFound {
		t.Errorf("Expected ErrLocationOptionNotFound, got %v", err)
	}
}

// Test UpdateLocationOption - no store
func TestUpdateLocationOptionNoStore(t *testing.T) {
	service.SetLocationOptionStore(nil)

	err := service.UpdateLocationOption(testCampaignID, 1, "New York", 1)

	if err != service.ErrNoLocationOptionStore {
		t.Errorf("Expected ErrNoLocationOptionStore, got %v", err)
	}
}

// Test UpdateLocationOption - change display order
func TestUpdateLocationOptionChangeOrder(t *testing.T) {
	optionStore := newMockLocationOptionStore()
	service.SetLocationOptionStore(optionStore)

	id, _ := service.AddLocationOption(testCampaignID, "New York", 1)

	err := service.UpdateLocationOption(testCampaignID, id, "New York", 5)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	options, _ := service.GetLocationOptions(testCampaignID)
	if options[0].DisplayOrder != 5 {
		t.Error("Expected display order to be updated")
	}
}

// Test DeleteLocationOption - happy path
func TestDeleteLocationOptionHappyPath(t *testing.T) {
	optionStore := newMockLocationOptionStore()
	service.SetLocationOptionStore(optionStore)

	id, _ := service.AddLocationOption(testCampaignID, "New York", 1)

	err := service.DeleteLocationOption(testCampaignID, id)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	options, _ := service.GetLocationOptions(testCampaignID)
	if len(options) != 0 {
		t.Error("Expected option to be deleted")
	}
}

// Test DeleteLocationOption - nonexistent option
func TestDeleteLocationOptionNonexistent(t *testing.T) {
	optionStore := newMockLocationOptionStore()
	service.SetLocationOptionStore(optionStore)

	err := service.DeleteLocationOption(testCampaignID, 999)

	if err != service.ErrLocationOptionNotFound {
		t.Errorf("Expected ErrLocationOptionNotFound, got %v", err)
	}
}

// Test DeleteLocationOption - no store
func TestDeleteLocationOptionNoStore(t *testing.T) {
	service.SetLocationOptionStore(nil)

	err := service.DeleteLocationOption(testCampaignID, 1)

	if err != service.ErrNoLocationOptionStore {
		t.Errorf("Expected ErrNoLocationOptionStore, got %v", err)
	}
}

// Test DeleteLocationOption - multiple options
func TestDeleteLocationOptionMultiple(t *testing.T) {
	optionStore := newMockLocationOptionStore()
	service.SetLocationOptionStore(optionStore)

	_, _ = service.AddLocationOption(testCampaignID, "New York", 1)
	id2, _ := service.AddLocationOption(testCampaignID, "Boston", 2)
	_, _ = service.AddLocationOption(testCampaignID, "Chicago", 3)

	service.DeleteLocationOption(testCampaignID, id2)

	options, _ := service.GetLocationOptions(testCampaignID)
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
	campaignStore := newMockCampaignStore()
	setupTestCampaign(campaignStore, true)
	service.SetCampaignStore(campaignStore)

	optionStore := newMockLocationOptionStore()
	optionStore.AddOption(testCampaignID, "New York", 1)
	optionStore.AddOption(testCampaignID, "Boston", 2)
	service.SetLocationOptionStore(optionStore)

	config, options, err := service.GetLocationConfigWithOptions(testCampaignID)

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
	campaignStore := newMockCampaignStore()
	setupTestCampaign(campaignStore, false)
	service.SetCampaignStore(campaignStore)

	optionStore := newMockLocationOptionStore()
	service.SetLocationOptionStore(optionStore)

	config, options, err := service.GetLocationConfigWithOptions(testCampaignID)

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

// Test GetLocationConfigWithOptions - no campaign store
func TestGetLocationConfigWithOptionsNoCampaignStore(t *testing.T) {
	service.SetCampaignStore(nil)
	service.SetLocationOptionStore(newMockLocationOptionStore())

	_, _, err := service.GetLocationConfigWithOptions(testCampaignID)

	if err != service.ErrNoCampaignStore {
		t.Errorf("Expected ErrNoCampaignStore, got %v", err)
	}
}

// Test GetLocationConfigWithOptions - no option store
func TestGetLocationConfigWithOptionsNoOptionStore(t *testing.T) {
	campaignStore := newMockCampaignStore()
	setupTestCampaign(campaignStore, true)
	service.SetCampaignStore(campaignStore)
	service.SetLocationOptionStore(nil)

	_, _, err := service.GetLocationConfigWithOptions(testCampaignID)

	if err != service.ErrNoLocationOptionStore {
		t.Errorf("Expected ErrNoLocationOptionStore, got %v", err)
	}
}

// Test GetLocationConfigWithOptions - campaign not found
func TestGetLocationConfigWithOptionsCampaignNotFound(t *testing.T) {
	campaignStore := newMockCampaignStore()
	service.SetCampaignStore(campaignStore)
	service.SetLocationOptionStore(newMockLocationOptionStore())

	_, _, err := service.GetLocationConfigWithOptions("nonexistent")

	if err != service.ErrCampaignNotFound {
		t.Errorf("Expected ErrCampaignNotFound, got %v", err)
	}
}

// Test AddLocationOption - store error
func TestAddLocationOptionStoreError(t *testing.T) {
	optionStore := newMockLocationOptionStore()
	optionStore.err = errors.New("database error")
	service.SetLocationOptionStore(optionStore)

	_, err := service.AddLocationOption(testCampaignID, "New York", 1)

	if !errors.Is(err, optionStore.err) {
		t.Errorf("Expected database error, got %v", err)
	}
}

// Test UpdateLocationOption - store error
func TestUpdateLocationOptionStoreError(t *testing.T) {
	optionStore := newMockLocationOptionStore()
	service.SetLocationOptionStore(optionStore)

	id, _ := service.AddLocationOption(testCampaignID, "New York", 1)

	// Now set error
	optionStore.err = errors.New("database error")

	err := service.UpdateLocationOption(testCampaignID, id, "Boston", 2)

	if !errors.Is(err, optionStore.err) {
		t.Errorf("Expected database error, got %v", err)
	}
}

// Test DeleteLocationOption - store error
func TestDeleteLocationOptionStoreError(t *testing.T) {
	optionStore := newMockLocationOptionStore()
	service.SetLocationOptionStore(optionStore)

	id, _ := service.AddLocationOption(testCampaignID, "New York", 1)

	// Now set error
	optionStore.err = errors.New("database error")

	err := service.DeleteLocationOption(testCampaignID, id)

	if !errors.Is(err, optionStore.err) {
		t.Errorf("Expected database error, got %v", err)
	}
}

// Test location option workflow
func TestLocationOptionWorkflow(t *testing.T) {
	optionStore := newMockLocationOptionStore()
	service.SetLocationOptionStore(optionStore)

	// Add options
	id1, _ := service.AddLocationOption(testCampaignID, "New York", 1)
	id2, _ := service.AddLocationOption(testCampaignID, "Boston", 2)
	_, _ = service.AddLocationOption(testCampaignID, "Chicago", 3)

	options, _ := service.GetLocationOptions(testCampaignID)
	if len(options) != 3 {
		t.Errorf("Expected 3 options, got %d", len(options))
	}

	// Update one
	service.UpdateLocationOption(testCampaignID, id2, "Boston Updated", 10)

	options, _ = service.GetLocationOptions(testCampaignID)
	for _, opt := range options {
		if opt.ID == id2 && opt.Value != "Boston Updated" {
			t.Error("Expected option to be updated")
		}
	}

	// Delete one
	service.DeleteLocationOption(testCampaignID, id1)

	options, _ = service.GetLocationOptions(testCampaignID)
	if len(options) != 2 {
		t.Errorf("Expected 2 options after delete, got %d", len(options))
	}
}
