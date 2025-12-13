package api_test

import (
	"cosign/internal/util"
	"fmt"
	"net/http"
	"slices"
	"testing"
)

func extractOriginURLs(t *testing.T, data map[string]any) []string {
	t.Helper()
	origins, ok := data["origins"].([]any)
	if !ok {
		t.Fatalf("Expected origins field to be an array, got %T", data["origins"])
	}
	urls := make([]string, 0, len(origins))
	for _, o := range origins {
		if m, ok := o.(map[string]any); ok {
			if url, ok := m["url"].(string); ok {
				urls = append(urls, url)
			}
		}
	}
	return urls
}

// TestListCORSOriginsSuccess tests listing CORS origins with valid authentication
func TestListCORSOriginsSuccess(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	authHeader := makeTestAuthHeader(t)
	var resp map[string]any
	result := get(router, "/api/v1/admin/cors", &resp, authHeader)

	expectStatus(t, http.StatusOK, result)

	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatalf("Expected data field to be a map, got %T", resp["data"])
	}

	origins := extractOriginURLs(t, data)

	// Should have at least the test origin added during setup
	if len(origins) == 0 {
		t.Errorf("Expected at least 1 origin (test origin), got %d", len(origins))
	}
}

// TestListCORSOriginsNoAuth tests listing CORS origins without authentication
func TestListCORSOriginsNoAuth(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	var resp map[string]any
	result := get(router, "/api/v1/admin/cors", &resp)

	expectStatus(t, http.StatusUnauthorized, result)
}

// TestAddCORSOriginSuccess tests adding a valid CORS origin
func TestAddCORSOriginSuccess(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	authHeader := makeTestAuthHeader(t)
	body := `{"origin": "https://example.com"}`
	result := post(router, "/api/v1/admin/cors", body, nil, authHeader)

	expectStatus(t, http.StatusCreated, result)

	// Verify it was added by listing origins
	var listResp map[string]any
	get(router, "/api/v1/admin/cors", &listResp, authHeader)

	data, _ := listResp["data"].(map[string]any)
	origins := extractOriginURLs(t, data)

	// Check that our origin is in the list
	found := false
	for _, o := range origins {
		if o == "https://example.com" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected origin to be added to the list")
	}
}

// TestAddCORSOriginEmptyValue tests adding an empty CORS origin
func TestAddCORSOriginEmptyValue(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	authHeader := makeTestAuthHeader(t)
	body := `{"origin": ""}`
	result := post(router, "/api/v1/admin/cors", body, nil, authHeader)

	expectStatus(t, http.StatusBadRequest, result)
}

// TestAddCORSOriginInvalidJSON tests adding origin with invalid JSON
func TestAddCORSOriginInvalidJSON(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	authHeader := makeTestAuthHeader(t)
	body := `{invalid json`
	result := post(router, "/api/v1/admin/cors", body, nil, authHeader)

	expectStatus(t, http.StatusBadRequest, result)
}

// TestAddCORSOriginNoAuth tests adding CORS origin without authentication
func TestAddCORSOriginNoAuth(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	body := `{"origin": "https://example.com"}`
	result := post(router, "/api/v1/admin/cors", body, nil)

	expectStatus(t, http.StatusUnauthorized, result)
}

// TestDeleteCORSOriginSuccess tests deleting a CORS origin
func TestDeleteCORSOriginSuccess(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	authHeader := makeTestAuthHeader(t)

	// First add an origin (use simpler name to avoid URL encoding issues)
	testOrigin := "example-delete"
	addBody := `{"origin": "` + testOrigin + `"}`
	post(router, "/api/v1/admin/cors", addBody, nil, authHeader)

	// Delete the origin
	result := del(router, "/api/v1/admin/cors/"+testOrigin, nil, authHeader)

	expectStatus(t, http.StatusNoContent, result)

	// Verify it's deleted by listing origins
	var listResp map[string]any
	get(router, "/api/v1/admin/cors", &listResp, authHeader)

	data, _ := listResp["data"].(map[string]any)
	origins := extractOriginURLs(t, data)

	// Check that our origin is not in the list
	for _, o := range origins {
		if o == testOrigin {
			t.Errorf("Expected origin to be deleted, but found it in list")
			return
		}
	}
}

// TestDeleteCORSOriginNotFound tests deleting a non-existent CORS origin
func TestDeleteCORSOriginNotFound(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()
	setupCORS()

	authHeader := makeTestAuthHeader(t)
	result := del(router, "/api/v1/admin/cors/nonexistent-origin", nil, authHeader)

	expectStatus(t, http.StatusNotFound, result)
}

// TestDeleteCORSOriginNoAuth tests deleting a CORS origin without authentication
func TestDeleteCORSOriginNoAuth(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	result := del(router, "/api/v1/admin/cors/example-origin", nil)

	expectStatus(t, http.StatusUnauthorized, result)
}

// TestAddAndListMultipleCORSOrigins tests adding and listing multiple origins
func TestAddAndListMultipleCORSOrigins(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	authHeader := makeTestAuthHeader(t)

	// Add multiple origins (use simple names to avoid URL encoding issues)
	origins := []string{
		"origin-example1",
		"origin-example2",
		"origin-example3",
	}

	for _, origin := range origins {
		body := fmt.Sprintf(`{"origin": "%s"}`, origin)
		post(router, "/api/v1/admin/cors", body, nil, authHeader)
	}

	// List all origins
	var listResp map[string]any
	get(router, "/api/v1/admin/cors", &listResp, authHeader)

	data, _ := listResp["data"].(map[string]any)
	respOrigins := extractOriginURLs(t, data)

	// Should have at least the test origin + our 3 origins = 4 total
	if len(respOrigins) < 4 {
		t.Errorf("Expected at least 4 origins, got %d", len(respOrigins))
	}

	// Verify all our origins are present
	foundCount := 0
	for _, o := range respOrigins {
		if slices.Contains(origins, o) {
			foundCount++
		}
	}

	if foundCount != len(origins) {
		t.Errorf("Expected to find %d origins, found %d", len(origins), foundCount)
	}
}
