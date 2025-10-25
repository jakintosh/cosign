package service_test

import (
	"cosign/internal/service"
	"errors"
	"testing"
)

// mockSignonStore is a test implementation of SignonStore
type mockSignonStore struct {
	signons map[int64]*service.Signon
	emails  map[string]bool
	nextID  int64
	err     error // error to return on any operation
}

func newMockSignonStore() *mockSignonStore {
	return &mockSignonStore{
		signons: make(map[int64]*service.Signon),
		emails:  make(map[string]bool),
		nextID:  1,
	}
}

func (m *mockSignonStore) Insert(name, email, location string, createdAt int64) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	id := m.nextID
	m.nextID++
	m.signons[id] = &service.Signon{
		ID:        id,
		Name:      name,
		Email:     email,
		Location:  location,
		CreatedAt: createdAt,
	}
	m.emails[email] = true
	return id, nil
}

func (m *mockSignonStore) GetByID(id int64) (*service.Signon, error) {
	if m.err != nil {
		return nil, m.err
	}
	signon, exists := m.signons[id]
	if !exists {
		return nil, service.ErrSignonNotFound
	}
	return signon, nil
}

func (m *mockSignonStore) List(limit, offset int) ([]*service.Signon, error) {
	if m.err != nil {
		return nil, m.err
	}
	var signons []*service.Signon
	for i := 0; i < len(m.signons); i++ {
		signons = append(signons, m.signons[int64(i+1)])
	}
	// Simple pagination
	if offset >= len(signons) {
		return []*service.Signon{}, nil
	}
	end := min(offset+limit, len(signons))
	return signons[offset:end], nil
}

func (m *mockSignonStore) Count() (int, error) {
	if m.err != nil {
		return 0, m.err
	}
	return len(m.signons), nil
}

func (m *mockSignonStore) Delete(id int64) error {
	if m.err != nil {
		return m.err
	}
	signon, exists := m.signons[id]
	if !exists {
		return service.ErrSignonNotFound
	}
	delete(m.signons, id)
	delete(m.emails, signon.Email)
	return nil
}

func (m *mockSignonStore) EmailExists(email string) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	return m.emails[email], nil
}

// mockLocationConfigStore for testing location validation
type mockLocationConfigStore struct {
	config  *service.LocationConfig
	options []*service.LocationOption
	err     error
}

func newMockLocationConfigStore(allowCustom bool, options []*service.LocationOption) *mockLocationConfigStore {
	return &mockLocationConfigStore{
		config:  &service.LocationConfig{AllowCustomText: allowCustom},
		options: options,
	}
}

func (m *mockLocationConfigStore) GetConfig() (*service.LocationConfig, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.config, nil
}

func (m *mockLocationConfigStore) SetAllowCustomText(allow bool) error {
	if m.err != nil {
		return m.err
	}
	m.config.AllowCustomText = allow
	return nil
}

func (m *mockLocationConfigStore) GetOptions() ([]*service.LocationOption, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.options, nil
}

func (m *mockLocationConfigStore) AddOption(value string, displayOrder int) (int64, error) {
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

func (m *mockLocationConfigStore) UpdateOption(id int64, value string, displayOrder int) error {
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

func (m *mockLocationConfigStore) DeleteOption(id int64) error {
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

// Test CreateSignon - happy path
func TestCreateSignonHappyPath(t *testing.T) {
	store := newMockSignonStore()
	service.SetSignonStore(store)
	service.SetLocationConfigStore(nil) // Allow any location

	signon, err := service.CreateSignon("John Doe", "john@example.com", "New York", false)

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
	service.SetLocationConfigStore(nil)

	signon, err := service.CreateSignon("  Jane Doe  ", "  jane@example.com  ", "  Boston  ", false)

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

	_, err := service.CreateSignon("", "john@example.com", "NYC", false)

	if err != service.ErrEmptyName {
		t.Errorf("Expected ErrEmptyName, got %v", err)
	}
}

// Test CreateSignon - empty email
func TestCreateSignonEmptyEmail(t *testing.T) {
	store := newMockSignonStore()
	service.SetSignonStore(store)

	_, err := service.CreateSignon("John Doe", "", "NYC", false)

	if err != service.ErrEmptyEmail {
		t.Errorf("Expected ErrEmptyEmail, got %v", err)
	}
}

// Test CreateSignon - empty location
func TestCreateSignonEmptyLocation(t *testing.T) {
	store := newMockSignonStore()
	service.SetSignonStore(store)

	_, err := service.CreateSignon("John Doe", "john@example.com", "", false)

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
		_, err := service.CreateSignon("John Doe", email, "NYC", false)
		if err != service.ErrInvalidEmail {
			t.Errorf("Email %q: expected ErrInvalidEmail, got %v", email, err)
		}
	}
}

// Test CreateSignon - duplicate email not allowed
func TestCreateSignonDuplicateEmailNotAllowed(t *testing.T) {
	store := newMockSignonStore()
	service.SetSignonStore(store)
	service.SetLocationConfigStore(nil)

	// Create first signon
	_, err := service.CreateSignon("John Doe", "john@example.com", "NYC", false)
	if err != nil {
		t.Fatalf("Failed to create first signon: %v", err)
	}

	// Try to create duplicate
	_, err = service.CreateSignon("Jane Doe", "john@example.com", "Boston", false)
	if err != service.ErrDuplicateEmail {
		t.Errorf("Expected ErrDuplicateEmail, got %v", err)
	}
}

// Test CreateSignon - duplicate email allowed
func TestCreateSignonDuplicateEmailAllowed(t *testing.T) {
	store := newMockSignonStore()
	service.SetSignonStore(store)
	service.SetLocationConfigStore(nil)

	// Create first signon
	_, err := service.CreateSignon("John Doe", "john@example.com", "NYC", true)
	if err != nil {
		t.Fatalf("Failed to create first signon: %v", err)
	}

	// Create duplicate (should succeed)
	signon, err := service.CreateSignon("Jane Doe", "john@example.com", "Boston", true)
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

	locationStore := newMockLocationConfigStore(false, []*service.LocationOption{
		{ID: 1, Value: "New York", DisplayOrder: 1},
		{ID: 2, Value: "Boston", DisplayOrder: 2},
	})
	service.SetLocationConfigStore(locationStore)

	// Valid preset location
	signon, err := service.CreateSignon("John Doe", "john@example.com", "New York", false)
	if err != nil {
		t.Errorf("Expected no error for valid preset location, got %v", err)
	}
	if signon == nil {
		t.Fatal("Expected signon, got nil")
	}

	// Invalid custom location
	_, err = service.CreateSignon("Jane Doe", "jane@example.com", "Chicago", false)
	if err != service.ErrLocationNotInOptions {
		t.Errorf("Expected ErrLocationNotInOptions for custom location, got %v", err)
	}
}

// Test CreateSignon - location validation with custom text allowed
func TestCreateSignonLocationValidationCustomAllowed(t *testing.T) {
	store := newMockSignonStore()
	service.SetSignonStore(store)

	locationStore := newMockLocationConfigStore(true, []*service.LocationOption{})
	service.SetLocationConfigStore(locationStore)

	// Custom location should be allowed
	signon, err := service.CreateSignon("John Doe", "john@example.com", "Anywhere", false)
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

	_, err := service.CreateSignon("John Doe", "john@example.com", "NYC", false)

	if err != service.ErrNoSignonStore {
		t.Errorf("Expected ErrNoSignonStore, got %v", err)
	}
}

// Test GetSignon - happy path
func TestGetSignonHappyPath(t *testing.T) {
	store := newMockSignonStore()
	service.SetSignonStore(store)
	service.SetLocationConfigStore(nil)

	created, _ := service.CreateSignon("John Doe", "john@example.com", "NYC", false)

	signon, err := service.GetSignon(created.ID)

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

	_, err := service.GetSignon(999)

	if err != service.ErrSignonNotFound {
		t.Errorf("Expected ErrSignonNotFound, got %v", err)
	}
}

// Test GetSignon - no store
func TestGetSignonNoStore(t *testing.T) {
	service.SetSignonStore(nil)

	_, err := service.GetSignon(1)

	if err != service.ErrNoSignonStore {
		t.Errorf("Expected ErrNoSignonStore, got %v", err)
	}
}

// Test ListSignons - happy path
func TestListSignonsHappyPath(t *testing.T) {
	store := newMockSignonStore()
	service.SetSignonStore(store)
	service.SetLocationConfigStore(nil)

	service.CreateSignon("John Doe", "john@example.com", "NYC", false)
	service.CreateSignon("Jane Doe", "jane@example.com", "Boston", false)

	signons, err := service.ListSignons(10, 0)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(signons) != 2 {
		t.Errorf("Expected 2 signons, got %d", len(signons))
	}
}

// Test ListSignons - pagination defaults
func TestListSignonsPaginationDefaults(t *testing.T) {
	store := newMockSignonStore()
	service.SetSignonStore(store)

	// Test limit <= 0 defaults to 100
	_, err := service.ListSignons(-1, 0)
	if err != nil {
		t.Errorf("Expected no error with negative limit, got %v", err)
	}

	// Test offset < 0 defaults to 0
	_, err = service.ListSignons(10, -1)
	if err != nil {
		t.Errorf("Expected no error with negative offset, got %v", err)
	}
}

// Test ListSignons - no store
func TestListSignonsNoStore(t *testing.T) {
	service.SetSignonStore(nil)

	_, err := service.ListSignons(10, 0)

	if err != service.ErrNoSignonStore {
		t.Errorf("Expected ErrNoSignonStore, got %v", err)
	}
}

// Test CountSignons - happy path
func TestCountSignonsHappyPath(t *testing.T) {
	store := newMockSignonStore()
	service.SetSignonStore(store)
	service.SetLocationConfigStore(nil)

	service.CreateSignon("John Doe", "john@example.com", "NYC", false)
	service.CreateSignon("Jane Doe", "jane@example.com", "Boston", false)

	count, err := service.CountSignons()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

// Test CountSignons - empty
func TestCountSignonsEmpty(t *testing.T) {
	store := newMockSignonStore()
	service.SetSignonStore(store)

	count, err := service.CountSignons()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}
}

// Test CountSignons - no store
func TestCountSignonsNoStore(t *testing.T) {
	service.SetSignonStore(nil)

	_, err := service.CountSignons()

	if err != service.ErrNoSignonStore {
		t.Errorf("Expected ErrNoSignonStore, got %v", err)
	}
}

// Test DeleteSignon - happy path
func TestDeleteSignonHappyPath(t *testing.T) {
	store := newMockSignonStore()
	service.SetSignonStore(store)
	service.SetLocationConfigStore(nil)

	signon, _ := service.CreateSignon("John Doe", "john@example.com", "NYC", false)

	err := service.DeleteSignon(signon.ID)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify it's deleted
	_, err = service.GetSignon(signon.ID)
	if err != service.ErrSignonNotFound {
		t.Errorf("Expected signon to be deleted, but GetSignon returned %v", err)
	}
}

// Test DeleteSignon - not found
func TestDeleteSignonNotFound(t *testing.T) {
	store := newMockSignonStore()
	service.SetSignonStore(store)

	err := service.DeleteSignon(999)

	if err != service.ErrSignonNotFound {
		t.Errorf("Expected ErrSignonNotFound, got %v", err)
	}
}

// Test DeleteSignon - no store
func TestDeleteSignonNoStore(t *testing.T) {
	service.SetSignonStore(nil)

	err := service.DeleteSignon(1)

	if err != service.ErrNoSignonStore {
		t.Errorf("Expected ErrNoSignonStore, got %v", err)
	}
}

// Test CreateSignon - valid email formats
func TestCreateSignonValidEmailFormats(t *testing.T) {
	store := newMockSignonStore()
	service.SetSignonStore(store)
	service.SetLocationConfigStore(nil)

	validEmails := []string{
		"user@example.com",
		"user.name@example.com",
		"user+tag@example.co.uk",
		"user_name@example-domain.com",
		"123@example.com",
	}

	for i, email := range validEmails {
		signon, err := service.CreateSignon("User"+string(rune(i)), email, "NYC", false)
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
	service.SetLocationConfigStore(nil)

	_, err := service.CreateSignon("John Doe", "john@example.com", "NYC", false)

	if !errors.Is(err, store.err) {
		t.Errorf("Expected database error, got %v", err)
	}
}
