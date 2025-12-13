package service_test

import (
	"cosign/internal/service"
	"errors"
	"testing"
)

const testCampaignIDSignon = "test-campaign-id"

// mockSignonStore is a test implementation of SignonStore
type mockSignonStore struct {
	signons map[string]map[int64]*service.Signon // campaignID -> id -> signon
	emails  map[string]map[string]bool           // campaignID -> email -> exists
	nextID  int64
	err     error // error to return on any operation
}

func newMockSignonStore() *mockSignonStore {
	return &mockSignonStore{
		signons: make(map[string]map[int64]*service.Signon),
		emails:  make(map[string]map[string]bool),
		nextID:  1,
	}
}

func (m *mockSignonStore) Insert(campaignID, name, email, location string, createdAt int64) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	if m.signons[campaignID] == nil {
		m.signons[campaignID] = make(map[int64]*service.Signon)
	}
	if m.emails[campaignID] == nil {
		m.emails[campaignID] = make(map[string]bool)
	}

	id := m.nextID
	m.nextID++
	m.signons[campaignID][id] = &service.Signon{
		ID:        id,
		Name:      name,
		Email:     email,
		Location:  location,
		CreatedAt: createdAt,
	}
	m.emails[campaignID][email] = true
	return id, nil
}

func (m *mockSignonStore) GetByID(campaignID string, id int64) (*service.Signon, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.signons[campaignID] == nil {
		return nil, service.ErrSignonNotFound
	}
	signon, exists := m.signons[campaignID][id]
	if !exists {
		return nil, service.ErrSignonNotFound
	}
	return signon, nil
}

func (m *mockSignonStore) List(campaignID string, limit, offset int) ([]*service.Signon, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.signons[campaignID] == nil {
		return []*service.Signon{}, nil
	}

	var signons []*service.Signon
	for _, s := range m.signons[campaignID] {
		signons = append(signons, s)
	}

	// Simple pagination
	if offset >= len(signons) {
		return []*service.Signon{}, nil
	}
	end := min(offset+limit, len(signons))
	return signons[offset:end], nil
}

func (m *mockSignonStore) Count(campaignID string) (int, error) {
	if m.err != nil {
		return 0, m.err
	}
	if m.signons[campaignID] == nil {
		return 0, nil
	}
	return len(m.signons[campaignID]), nil
}

func (m *mockSignonStore) Delete(campaignID string, id int64) error {
	if m.err != nil {
		return m.err
	}
	if m.signons[campaignID] == nil {
		return service.ErrSignonNotFound
	}
	signon, exists := m.signons[campaignID][id]
	if !exists {
		return service.ErrSignonNotFound
	}
	delete(m.signons[campaignID], id)
	if m.emails[campaignID] != nil {
		delete(m.emails[campaignID], signon.Email)
	}
	return nil
}

func (m *mockSignonStore) EmailExists(campaignID, email string) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	if m.emails[campaignID] == nil {
		return false, nil
	}
	return m.emails[campaignID][email], nil
}

// Test CreateSignon - happy path
func TestCreateSignonHappyPath(t *testing.T) {
	store := newMockSignonStore()
	service.SetSignonStore(store)
	campaignStore := newMockCampaignStore()
	setupTestCampaign(campaignStore, true)
	service.SetCampaignStore(campaignStore)
	service.SetLocationOptionStore(nil) // Allow any location

	signon, err := service.CreateSignon(testCampaignIDSignon, "John Doe", "john@example.com", "New York", false)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if signon == nil {
		t.Fatal("Expected signon, got nil")
	}
	if signon.Name != "John Doe" {
		t.Errorf("Expected name 'John Doe', got %q", signon.Name)
	}
	if signon.Email != "john@example.com" {
		t.Errorf("Expected email 'john@example.com', got %q", signon.Email)
	}
	if signon.Location != "New York" {
		t.Errorf("Expected location 'New York', got %q", signon.Location)
	}
}

// Test CreateSignon - with whitespace trimming
func TestCreateSignonTrimsWhitespace(t *testing.T) {
	store := newMockSignonStore()
	service.SetSignonStore(store)
	campaignStore := newMockCampaignStore()
	setupTestCampaign(campaignStore, true)
	service.SetCampaignStore(campaignStore)
	service.SetLocationOptionStore(nil)

	signon, err := service.CreateSignon(testCampaignIDSignon, "  Jane Doe  ", "  jane@example.com  ", "  Boston  ", false)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if signon.Name != "Jane Doe" {
		t.Errorf("Expected trimmed name 'Jane Doe', got %q", signon.Name)
	}
	if signon.Email != "jane@example.com" {
		t.Errorf("Expected trimmed email 'jane@example.com', got %q", signon.Email)
	}
	if signon.Location != "Boston" {
		t.Errorf("Expected trimmed location 'Boston', got %q", signon.Location)
	}
}

// Test CreateSignon - empty name
func TestCreateSignonEmptyName(t *testing.T) {
	store := newMockSignonStore()
	service.SetSignonStore(store)

	_, err := service.CreateSignon(testCampaignIDSignon, "", "john@example.com", "NYC", false)

	if err != service.ErrEmptyName {
		t.Errorf("Expected ErrEmptyName, got %v", err)
	}
}

// Test CreateSignon - empty email
func TestCreateSignonEmptyEmail(t *testing.T) {
	store := newMockSignonStore()
	service.SetSignonStore(store)

	_, err := service.CreateSignon(testCampaignIDSignon, "John Doe", "", "NYC", false)

	if err != service.ErrEmptyEmail {
		t.Errorf("Expected ErrEmptyEmail, got %v", err)
	}
}

// Test CreateSignon - empty location
func TestCreateSignonEmptyLocation(t *testing.T) {
	store := newMockSignonStore()
	service.SetSignonStore(store)

	_, err := service.CreateSignon(testCampaignIDSignon, "John Doe", "john@example.com", "", false)

	if err != service.ErrEmptyLocation {
		t.Errorf("Expected ErrEmptyLocation, got %v", err)
	}
}

// Test CreateSignon - invalid email format
func TestCreateSignonInvalidEmailFormat(t *testing.T) {
	store := newMockSignonStore()
	service.SetSignonStore(store)

	testCases := []string{
		"not-an-email",
		"missing@domain",
		"@nodomain.com",
		"spaces in@email.com",
	}

	for _, email := range testCases {
		_, err := service.CreateSignon(testCampaignIDSignon, "John Doe", email, "NYC", false)
		if err != service.ErrInvalidEmail {
			t.Errorf("Email %q: expected ErrInvalidEmail, got %v", email, err)
		}
	}
}

// Test CreateSignon - duplicate email not allowed
func TestCreateSignonDuplicateEmailNotAllowed(t *testing.T) {
	store := newMockSignonStore()
	service.SetSignonStore(store)
	campaignStore := newMockCampaignStore()
	setupTestCampaign(campaignStore, true)
	service.SetCampaignStore(campaignStore)
	service.SetLocationOptionStore(nil)

	// Create first signon
	_, err := service.CreateSignon(testCampaignIDSignon, "John Doe", "john@example.com", "NYC", false)
	if err != nil {
		t.Fatalf("Failed to create first signon: %v", err)
	}

	// Try to create duplicate
	_, err = service.CreateSignon(testCampaignIDSignon, "Jane Doe", "john@example.com", "Boston", false)
	if err != service.ErrDuplicateEmail {
		t.Errorf("Expected ErrDuplicateEmail, got %v", err)
	}
}

// Test CreateSignon - duplicate email allowed
func TestCreateSignonDuplicateEmailAllowed(t *testing.T) {
	store := newMockSignonStore()
	service.SetSignonStore(store)
	campaignStore := newMockCampaignStore()
	setupTestCampaign(campaignStore, true)
	service.SetCampaignStore(campaignStore)
	service.SetLocationOptionStore(nil)

	// Create first signon
	_, err := service.CreateSignon(testCampaignIDSignon, "John Doe", "john@example.com", "NYC", true)
	if err != nil {
		t.Fatalf("Failed to create first signon: %v", err)
	}

	// Create duplicate (should succeed)
	signon, err := service.CreateSignon(testCampaignIDSignon, "Jane Doe", "john@example.com", "Boston", true)
	if err != nil {
		t.Errorf("Expected no error with allowDuplicates=true, got %v", err)
	}
	if signon == nil {
		t.Fatal("Expected signon, got nil")
	}
}

// Test CreateSignon - location validation with preset options
func TestCreateSignonLocationValidationWithPreset(t *testing.T) {
	store := newMockSignonStore()
	service.SetSignonStore(store)

	campaignStore := newMockCampaignStore()
	setupTestCampaign(campaignStore, false) // Strict mode
	service.SetCampaignStore(campaignStore)

	locationStore := newMockLocationOptionStore()
	locationStore.AddOption(testCampaignIDSignon, "New York", 1)
	locationStore.AddOption(testCampaignIDSignon, "Boston", 2)
	service.SetLocationOptionStore(locationStore)

	// Valid preset location
	signon, err := service.CreateSignon(testCampaignIDSignon, "John Doe", "john@example.com", "New York", false)
	if err != nil {
		t.Errorf("Expected no error for valid preset location, got %v", err)
	}
	if signon == nil {
		t.Fatal("Expected signon, got nil")
	}

	// Invalid custom location
	_, err = service.CreateSignon(testCampaignIDSignon, "Jane Doe", "jane@example.com", "Chicago", false)
	if err != service.ErrLocationNotInOptions {
		t.Errorf("Expected ErrLocationNotInOptions for custom location, got %v", err)
	}
}

// Test CreateSignon - location validation with custom text allowed
func TestCreateSignonLocationValidationCustomAllowed(t *testing.T) {
	store := newMockSignonStore()
	service.SetSignonStore(store)

	campaignStore := newMockCampaignStore()
	setupTestCampaign(campaignStore, true) // Custom text allowed
	service.SetCampaignStore(campaignStore)

	locationStore := newMockLocationOptionStore()
	service.SetLocationOptionStore(locationStore)

	// Custom location should be allowed
	signon, err := service.CreateSignon(testCampaignIDSignon, "John Doe", "john@example.com", "Anywhere", false)
	if err != nil {
		t.Errorf("Expected no error when custom text allowed, got %v", err)
	}
	if signon == nil {
		t.Fatal("Expected signon, got nil")
	}
}

// Test CreateSignon - no signon store
func TestCreateSignonNoStore(t *testing.T) {
	service.SetSignonStore(nil)

	_, err := service.CreateSignon(testCampaignIDSignon, "John Doe", "john@example.com", "NYC", false)

	if err != service.ErrNoSignonStore {
		t.Errorf("Expected ErrNoSignonStore, got %v", err)
	}
}

// Test GetSignon - happy path
func TestGetSignonHappyPath(t *testing.T) {
	store := newMockSignonStore()
	service.SetSignonStore(store)
	campaignStore := newMockCampaignStore()
	setupTestCampaign(campaignStore, true)
	service.SetCampaignStore(campaignStore)
	service.SetLocationOptionStore(nil)

	created, _ := service.CreateSignon(testCampaignIDSignon, "John Doe", "john@example.com", "NYC", false)

	signon, err := service.GetSignon(testCampaignIDSignon, created.ID)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if signon == nil {
		t.Fatal("Expected signon, got nil")
	}
	if signon.ID != created.ID {
		t.Errorf("Expected ID %d, got %d", created.ID, signon.ID)
	}
}

// Test GetSignon - not found
func TestGetSignonNotFound(t *testing.T) {
	store := newMockSignonStore()
	service.SetSignonStore(store)

	_, err := service.GetSignon(testCampaignIDSignon, 999)

	if err != service.ErrSignonNotFound {
		t.Errorf("Expected ErrSignonNotFound, got %v", err)
	}
}

// Test GetSignon - no store
func TestGetSignonNoStore(t *testing.T) {
	service.SetSignonStore(nil)

	_, err := service.GetSignon(testCampaignIDSignon, 1)

	if err != service.ErrNoSignonStore {
		t.Errorf("Expected ErrNoSignonStore, got %v", err)
	}
}

// Test ListSignons - happy path
func TestListSignonsHappyPath(t *testing.T) {
	store := newMockSignonStore()
	service.SetSignonStore(store)
	campaignStore := newMockCampaignStore()
	setupTestCampaign(campaignStore, true)
	service.SetCampaignStore(campaignStore)
	service.SetLocationOptionStore(nil)

	service.CreateSignon(testCampaignIDSignon, "John Doe", "john@example.com", "NYC", false)
	service.CreateSignon(testCampaignIDSignon, "Jane Doe", "jane@example.com", "Boston", false)

	signons, err := service.ListSignons(testCampaignIDSignon, 10, 0)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(signons.Signons) != 2 {
		t.Errorf("Expected 2 signons, got %d", len(signons.Signons))
	}
	if signons.Total != 2 || signons.Limit != 10 || signons.Offset != 0 {
		t.Errorf("Unexpected metadata %+v", signons)
	}
}

// Test ListSignons - pagination defaults
func TestListSignonsPaginationDefaults(t *testing.T) {
	store := newMockSignonStore()
	service.SetSignonStore(store)

	// Test limit <= 0 defaults to 100
	_, err := service.ListSignons(testCampaignIDSignon, -1, 0)
	if err != nil {
		t.Errorf("Expected no error with negative limit, got %v", err)
	}

	// Test offset < 0 defaults to 0
	_, err = service.ListSignons(testCampaignIDSignon, 10, -1)
	if err != nil {
		t.Errorf("Expected no error with negative offset, got %v", err)
	}
}

// Test ListSignons - no store
func TestListSignonsNoStore(t *testing.T) {
	service.SetSignonStore(nil)

	_, err := service.ListSignons(testCampaignIDSignon, 10, 0)

	if err != service.ErrNoSignonStore {
		t.Errorf("Expected ErrNoSignonStore, got %v", err)
	}
}

// Test ListSignons includes total count
func TestListSignonsIncludesTotal(t *testing.T) {
	store := newMockSignonStore()
	service.SetSignonStore(store)
	campaignStore := newMockCampaignStore()
	setupTestCampaign(campaignStore, true)
	service.SetCampaignStore(campaignStore)
	service.SetLocationOptionStore(nil)

	resp, err := service.ListSignons(testCampaignIDSignon, 5, 0)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if resp.Total != 0 {
		t.Errorf("Expected total 0, got %d", resp.Total)
	}

	service.CreateSignon(testCampaignIDSignon, "John Doe", "john@example.com", "NYC", false)
	service.CreateSignon(testCampaignIDSignon, "Jane Doe", "jane@example.com", "Boston", false)

	resp, err = service.ListSignons(testCampaignIDSignon, 5, 0)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if resp.Total != 2 {
		t.Errorf("Expected total 2, got %d", resp.Total)
	}
}

// Test DeleteSignon - happy path
func TestDeleteSignonHappyPath(t *testing.T) {
	store := newMockSignonStore()
	service.SetSignonStore(store)
	campaignStore := newMockCampaignStore()
	setupTestCampaign(campaignStore, true)
	service.SetCampaignStore(campaignStore)
	service.SetLocationOptionStore(nil)

	signon, _ := service.CreateSignon(testCampaignIDSignon, "John Doe", "john@example.com", "NYC", false)

	err := service.DeleteSignon(testCampaignIDSignon, signon.ID)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify it's deleted
	_, err = service.GetSignon(testCampaignIDSignon, signon.ID)
	if err != service.ErrSignonNotFound {
		t.Errorf("Expected signon to be deleted, but GetSignon returned %v", err)
	}
}

// Test DeleteSignon - not found
func TestDeleteSignonNotFound(t *testing.T) {
	store := newMockSignonStore()
	service.SetSignonStore(store)

	err := service.DeleteSignon(testCampaignIDSignon, 999)

	if err != service.ErrSignonNotFound {
		t.Errorf("Expected ErrSignonNotFound, got %v", err)
	}
}

// Test DeleteSignon - no store
func TestDeleteSignonNoStore(t *testing.T) {
	service.SetSignonStore(nil)

	err := service.DeleteSignon(testCampaignIDSignon, 1)

	if err != service.ErrNoSignonStore {
		t.Errorf("Expected ErrNoSignonStore, got %v", err)
	}
}

// Test CreateSignon - valid email formats
func TestCreateSignonValidEmailFormats(t *testing.T) {
	store := newMockSignonStore()
	service.SetSignonStore(store)
	campaignStore := newMockCampaignStore()
	setupTestCampaign(campaignStore, true)
	service.SetCampaignStore(campaignStore)
	service.SetLocationOptionStore(nil)

	validEmails := []string{
		"user@example.com",
		"user.name@example.com",
		"user+tag@example.co.uk",
		"user_name@example-domain.com",
		"123@example.com",
	}

	for i, email := range validEmails {
		signon, err := service.CreateSignon(testCampaignIDSignon, "User"+string(rune(i)), email, "NYC", false)
		if err != nil {
			t.Errorf("Email %q: expected no error, got %v", email, err)
		}
		if signon == nil || signon.Email != email {
			t.Errorf("Email %q: signon not created correctly", email)
		}
	}
}

// Test store error propagation
func TestCreateSignonStoreError(t *testing.T) {
	store := newMockSignonStore()
	store.err = errors.New("database error")
	service.SetSignonStore(store)
	campaignStore := newMockCampaignStore()
	setupTestCampaign(campaignStore, true)
	service.SetCampaignStore(campaignStore)
	service.SetLocationOptionStore(nil)

	_, err := service.CreateSignon(testCampaignIDSignon, "John Doe", "john@example.com", "NYC", false)

	if !errors.Is(err, store.err) {
		t.Errorf("Expected database error, got %v", err)
	}
}
