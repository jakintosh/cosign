package service_test

import (
	"testing"

	"cosign/internal/service"
)

// mockCORSStore is a test implementation of CORSStore
type mockCORSStore struct {
	origins []service.AllowedOrigin
	err     error
}

func newMockCORSStore() *mockCORSStore {
	return &mockCORSStore{
		origins: []service.AllowedOrigin{},
	}
}

func (m *mockCORSStore) CountOrigins() (int, error) {
	return len(m.origins), nil
}
func (m *mockCORSStore) GetOrigins() ([]service.AllowedOrigin, error) {
	return m.origins, nil
}
func (m *mockCORSStore) SetOrigins(origins []service.AllowedOrigin) error {
	m.origins = origins
	return nil
}

func TestSetAndGetAllowedOrigins(t *testing.T) {

	store := newMockCORSStore()
	service.SetCORSStore(store)

	origins := []service.AllowedOrigin{
		{URL: "http://one"},
		{URL: "https://two"},
	}
	if err := service.SetAllowedOrigins(origins); err != nil {
		t.Fatal(err)
	}

	list, err := service.GetAllowedOrigins()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 || list[0].URL != "http://one" || list[1].URL != "https://two" {
		t.Fatalf("unexpected origins %+v", list)
	}
}

func TestSetAllowedOriginsInvalid(t *testing.T) {

	store := newMockCORSStore()
	service.SetCORSStore(store)

	origins := []service.AllowedOrigin{
		{URL: "ftp://bad"},
	}
	err := service.SetAllowedOrigins(origins)
	if err == nil {
		t.Fatalf("expected validation error")
	}
}
