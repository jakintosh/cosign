package service_test

import (
	"cosign/internal/service"
	"testing"
)

const testCampaignID = "test-campaign-id"

// mockCampaignStore implements CampaignStore with in-memory data.
type mockCampaignStore struct {
	campaigns map[string]*service.Campaign
	options   map[string][]service.LocationOption
	err       error
}

func newMockCampaignStore() *mockCampaignStore {
	return &mockCampaignStore{
		campaigns: make(map[string]*service.Campaign),
		options:   make(map[string][]service.LocationOption),
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
	delete(m.options, id)
	return nil
}

func (m *mockCampaignStore) GetLocationOptions(campaignID string) ([]*service.LocationOption, error) {
	if m.err != nil {
		return nil, m.err
	}
	opts := m.options[campaignID]
	out := make([]*service.LocationOption, 0, len(opts))
	for i := range opts {
		opt := opts[i]
		out = append(out, &opt)
	}
	return out, nil
}

func (m *mockCampaignStore) ReplaceLocationOptions(campaignID string, options []service.LocationOption) error {
	if m.err != nil {
		return m.err
	}
	if _, ok := m.campaigns[campaignID]; !ok {
		return service.ErrCampaignNotFound
	}
	m.options[campaignID] = append([]service.LocationOption{}, options...)
	return nil
}

func setupTestCampaign(campaignStore *mockCampaignStore, allowCustomText bool) {
	campaignStore.campaigns[testCampaignID] = &service.Campaign{
		ID:              testCampaignID,
		Name:            "Test Campaign",
		AllowCustomText: allowCustomText,
		CreatedAt:       0,
	}
}

func TestGetCampaignLocationsHappyPath(t *testing.T) {
	store := newMockCampaignStore()
	setupTestCampaign(store, true)
	store.options[testCampaignID] = []service.LocationOption{
		{ID: 1, Value: "NYC", DisplayOrder: 1},
	}

	service.SetCampaignStore(store)
	t.Cleanup(func() { service.SetCampaignStore(nil) })

	locs, err := service.GetCampaignLocations(testCampaignID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(locs) != 1 || locs[0].Value != "NYC" {
		t.Errorf("unexpected locations: %+v", locs)
	}
}

func TestSetCampaignLocationsReplacesOptions(t *testing.T) {
	store := newMockCampaignStore()
	setupTestCampaign(store, true)
	store.options[testCampaignID] = []service.LocationOption{
		{ID: 1, Value: "Old", DisplayOrder: 1},
	}

	service.SetCampaignStore(store)
	t.Cleanup(func() { service.SetCampaignStore(nil) })

	newOptions := []service.LocationOption{
		{Value: "New", DisplayOrder: 5},
	}
	if err := service.SetCampaignLocations(testCampaignID, newOptions); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	updated, _ := service.GetCampaignLocations(testCampaignID)
	if len(updated) != 1 || updated[0].Value != "New" || updated[0].DisplayOrder != 5 {
		t.Errorf("unexpected locations after update: %+v", updated)
	}
}

func TestSetCampaignLocationsEmptyLocation(t *testing.T) {
	store := newMockCampaignStore()
	setupTestCampaign(store, true)
	service.SetCampaignStore(store)
	t.Cleanup(func() { service.SetCampaignStore(nil) })

	options := []service.LocationOption{
		{Value: "", DisplayOrder: 1},
	}
	if err := service.SetCampaignLocations(testCampaignID, options); err != service.ErrEmptyLocation {
		t.Fatalf("expected ErrEmptyLocation, got %v", err)
	}
}
