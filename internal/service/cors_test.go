package service_test

import (
	"cosign/internal/service"
	"errors"
	"testing"
)

// mockCORSStore is a test implementation of CORSStore
type mockCORSStore struct {
	origins map[string]bool
	err     error
}

func newMockCORSStore() *mockCORSStore {
	return &mockCORSStore{
		origins: make(map[string]bool),
	}
}

func (m *mockCORSStore) Add(origin string, createdAt int64) error {
	if m.err != nil {
		return m.err
	}
	m.origins[origin] = true
	return nil
}

func (m *mockCORSStore) List() ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	var origins []string
	for origin := range m.origins {
		origins = append(origins, origin)
	}
	return origins, nil
}

func (m *mockCORSStore) Delete(origin string) error {
	if m.err != nil {
		return m.err
	}
	if !m.origins[origin] {
		return service.ErrCORSOriginNotFound
	}
	delete(m.origins, origin)
	return nil
}

func (m *mockCORSStore) IsAllowed(origin string) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	return m.origins[origin], nil
}

// Test AddCORSOrigin - happy path
func TestAddCORSOriginHappyPath(t *testing.T) {
	store := newMockCORSStore()
	service.SetCORSStore(store)

	err := service.AddCORSOrigin("http://example.com")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify it was added
	allowed, _ := store.IsAllowed("http://example.com")
	if !allowed {
		t.Error("Expected origin to be added to store")
	}
}

// Test AddCORSOrigin - trims whitespace
func TestAddCORSOriginTrimsWhitespace(t *testing.T) {
	store := newMockCORSStore()
	service.SetCORSStore(store)

	err := service.AddCORSOrigin("  http://example.com  ")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify it was trimmed
	allowed, _ := store.IsAllowed("http://example.com")
	if !allowed {
		t.Error("Expected whitespace to be trimmed")
	}
}

// Test AddCORSOrigin - empty origin
func TestAddCORSOriginEmpty(t *testing.T) {
	store := newMockCORSStore()
	service.SetCORSStore(store)

	err := service.AddCORSOrigin("")

	if err != service.ErrOriginNotAllowed {
		t.Errorf("Expected ErrOriginNotAllowed, got %v", err)
	}
}

// Test AddCORSOrigin - whitespace only
func TestAddCORSOriginWhitespaceOnly(t *testing.T) {
	store := newMockCORSStore()
	service.SetCORSStore(store)

	err := service.AddCORSOrigin("   ")

	if err != service.ErrOriginNotAllowed {
		t.Errorf("Expected ErrOriginNotAllowed, got %v", err)
	}
}

// Test AddCORSOrigin - no store
func TestAddCORSOriginNoStore(t *testing.T) {
	service.SetCORSStore(nil)

	err := service.AddCORSOrigin("http://example.com")

	if err != service.ErrNoCORSStore {
		t.Errorf("Expected ErrNoCORSStore, got %v", err)
	}
}

// Test AddCORSOrigin - multiple origins
func TestAddCORSOriginMultiple(t *testing.T) {
	store := newMockCORSStore()
	service.SetCORSStore(store)

	origins := []string{
		"http://example.com",
		"http://localhost:3000",
		"https://app.example.com",
	}

	for _, origin := range origins {
		err := service.AddCORSOrigin(origin)
		if err != nil {
			t.Errorf("Failed to add origin %q: %v", origin, err)
		}
	}

	// Verify all are stored
	list, _ := service.ListCORSOrigins()
	if len(list) != 3 {
		t.Errorf("Expected 3 origins, got %d", len(list))
	}
}

// Test ListCORSOrigins - happy path
func TestListCORSOriginsHappyPath(t *testing.T) {
	store := newMockCORSStore()
	service.SetCORSStore(store)

	service.AddCORSOrigin("http://example.com")
	service.AddCORSOrigin("http://localhost:3000")

	origins, err := service.ListCORSOrigins()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(origins) != 2 {
		t.Errorf("Expected 2 origins, got %d", len(origins))
	}
}

// Test ListCORSOrigins - empty list
func TestListCORSOriginsEmpty(t *testing.T) {
	store := newMockCORSStore()
	service.SetCORSStore(store)

	origins, err := service.ListCORSOrigins()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(origins) != 0 {
		t.Errorf("Expected 0 origins, got %d", len(origins))
	}
}

// Test ListCORSOrigins - no store
func TestListCORSOriginsNoStore(t *testing.T) {
	service.SetCORSStore(nil)

	_, err := service.ListCORSOrigins()

	if err != service.ErrNoCORSStore {
		t.Errorf("Expected ErrNoCORSStore, got %v", err)
	}
}

// Test DeleteCORSOrigin - happy path
func TestDeleteCORSOriginHappyPath(t *testing.T) {
	store := newMockCORSStore()
	service.SetCORSStore(store)

	service.AddCORSOrigin("http://example.com")

	err := service.DeleteCORSOrigin("http://example.com")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify it's deleted
	origins, _ := service.ListCORSOrigins()
	if len(origins) != 0 {
		t.Error("Expected origin to be deleted")
	}
}

// Test DeleteCORSOrigin - nonexistent origin
func TestDeleteCORSOriginNonexistent(t *testing.T) {
	store := newMockCORSStore()
	service.SetCORSStore(store)

	err := service.DeleteCORSOrigin("http://nonexistent.com")

	if err != service.ErrCORSOriginNotFound {
		t.Errorf("Expected ErrCORSOriginNotFound, got %v", err)
	}
}

// Test DeleteCORSOrigin - no store
func TestDeleteCORSOriginNoStore(t *testing.T) {
	service.SetCORSStore(nil)

	err := service.DeleteCORSOrigin("http://example.com")

	if err != service.ErrNoCORSStore {
		t.Errorf("Expected ErrNoCORSStore, got %v", err)
	}
}

// Test IsOriginAllowed - allowed origin
func TestIsOriginAllowedTrue(t *testing.T) {
	store := newMockCORSStore()
	service.SetCORSStore(store)

	service.AddCORSOrigin("http://example.com")

	allowed, err := service.IsOriginAllowed("http://example.com")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !allowed {
		t.Error("Expected origin to be allowed")
	}
}

// Test IsOriginAllowed - disallowed origin
func TestIsOriginAllowedFalse(t *testing.T) {
	store := newMockCORSStore()
	service.SetCORSStore(store)

	service.AddCORSOrigin("http://example.com")

	allowed, err := service.IsOriginAllowed("http://other.com")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if allowed {
		t.Error("Expected origin to not be allowed")
	}
}

// Test IsOriginAllowed - empty origin (non-browser)
func TestIsOriginAllowedEmptyOrigin(t *testing.T) {
	store := newMockCORSStore()
	service.SetCORSStore(store)

	// Empty origin should always be allowed (non-browser request)
	allowed, err := service.IsOriginAllowed("")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !allowed {
		t.Error("Expected empty origin to be allowed (non-browser request)")
	}
}

// Test IsOriginAllowed - no store
func TestIsOriginAllowedNoStore(t *testing.T) {
	service.SetCORSStore(nil)

	_, err := service.IsOriginAllowed("http://example.com")

	if err != service.ErrNoCORSStore {
		t.Errorf("Expected ErrNoCORSStore, got %v", err)
	}
}

// Test IsOriginAllowed - case sensitivity
func TestIsOriginAllowedCaseSensitivity(t *testing.T) {
	store := newMockCORSStore()
	service.SetCORSStore(store)

	service.AddCORSOrigin("http://example.com")

	// Different case should not match
	allowed, err := service.IsOriginAllowed("HTTP://EXAMPLE.COM")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if allowed {
		t.Error("Expected origin matching to be case-sensitive")
	}
}

// Test CORS workflow - add, list, delete
func TestCORSWorkflow(t *testing.T) {
	store := newMockCORSStore()
	service.SetCORSStore(store)

	// Add origins
	service.AddCORSOrigin("http://example.com")
	service.AddCORSOrigin("http://localhost:3000")

	// List
	origins, _ := service.ListCORSOrigins()
	if len(origins) != 2 {
		t.Errorf("Expected 2 origins after adding, got %d", len(origins))
	}

	// Delete one
	service.DeleteCORSOrigin("http://example.com")

	// List again
	origins, _ = service.ListCORSOrigins()
	if len(origins) != 1 {
		t.Errorf("Expected 1 origin after deletion, got %d", len(origins))
	}

	// Verify the right one is left
	allowed, _ := service.IsOriginAllowed("http://localhost:3000")
	if !allowed {
		t.Error("Expected remaining origin to be allowed")
	}

	allowed, _ = service.IsOriginAllowed("http://example.com")
	if allowed {
		t.Error("Expected deleted origin to not be allowed")
	}
}

// Test store error propagation
func TestAddCORSOriginStoreError(t *testing.T) {
	store := newMockCORSStore()
	store.err = errors.New("database error")
	service.SetCORSStore(store)

	err := service.AddCORSOrigin("http://example.com")

	if !errors.Is(err, store.err) {
		t.Errorf("Expected database error, got %v", err)
	}
}

// Test various origin formats
func TestIsOriginAllowedVariousFormats(t *testing.T) {
	store := newMockCORSStore()
	service.SetCORSStore(store)

	testOrigins := []string{
		"http://localhost:3000",
		"https://app.example.com",
		"http://192.168.1.1:8080",
		"https://example.com:443",
	}

	for _, origin := range testOrigins {
		service.AddCORSOrigin(origin)
	}

	for _, origin := range testOrigins {
		allowed, err := service.IsOriginAllowed(origin)
		if err != nil {
			t.Errorf("Origin %q: expected no error, got %v", origin, err)
		}
		if !allowed {
			t.Errorf("Expected origin %q to be allowed", origin)
		}
	}
}

// Test DeleteCORSOrigin with whitespace
func TestDeleteCORSOriginWithWhitespace(t *testing.T) {
	store := newMockCORSStore()
	service.SetCORSStore(store)

	service.AddCORSOrigin("http://example.com")

	// Delete with whitespace - should not match because AddCORSOrigin trims but Delete doesn't
	err := service.DeleteCORSOrigin("  http://example.com  ")

	if err != service.ErrCORSOriginNotFound {
		t.Errorf("Expected ErrCORSOriginNotFound (no trimming in Delete), got %v", err)
	}
}

// Test IsOriginAllowed store error
func TestIsOriginAllowedStoreError(t *testing.T) {
	store := newMockCORSStore()
	store.err = errors.New("database error")
	service.SetCORSStore(store)

	_, err := service.IsOriginAllowed("http://example.com")

	if !errors.Is(err, store.err) {
		t.Errorf("Expected database error, got %v", err)
	}
}
