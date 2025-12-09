package api_test

import (
	"cosign/internal/util"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

// ===== PUBLIC ENDPOINTS =====

// TestCreateSignonSuccess tests creating a sign-on with valid data
func TestCreateSignonSuccess(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	body := `{
		"name": "John Doe",
		"email": "john@example.com",
		"location": "New York"
	}`

	var resp map[string]any
	corsHeader := header{"Origin", "http://test-origin"}
	result := post(router, "/api/v1/signons", body, &resp, corsHeader)

	expectStatus(t, http.StatusCreated, result)

	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatalf("Expected data field to be a map, got %T", resp["data"])
	}

	if data["name"] != "John Doe" {
		t.Errorf("Expected name 'John Doe', got %v", data["name"])
	}
	if data["email"] != "john@example.com" {
		t.Errorf("Expected email 'john@example.com', got %v", data["email"])
	}
	if data["location"] != "New York" {
		t.Errorf("Expected location 'New York', got %v", data["location"])
	}
}

// TestCreateSignonEmptyName tests creating a sign-on with empty name
func TestCreateSignonEmptyName(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	body := `{
		"name": "",
		"email": "john@example.com",
		"location": "New York"
	}`

	var resp map[string]any
	corsHeader := header{"Origin", "http://test-origin"}
	result := post(router, "/api/v1/signons", body, &resp, corsHeader)

	expectStatus(t, http.StatusBadRequest, result)
}

// TestCreateSignonEmptyEmail tests creating a sign-on with empty email
func TestCreateSignonEmptyEmail(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	body := `{
		"name": "John Doe",
		"email": "",
		"location": "New York"
	}`

	var resp map[string]any
	corsHeader := header{"Origin", "http://test-origin"}
	result := post(router, "/api/v1/signons", body, &resp, corsHeader)

	expectStatus(t, http.StatusBadRequest, result)
}

// TestCreateSignonEmptyLocation tests creating a sign-on with empty location
func TestCreateSignonEmptyLocation(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	body := `{
		"name": "John Doe",
		"email": "john@example.com",
		"location": ""
	}`

	var resp map[string]any
	corsHeader := header{"Origin", "http://test-origin"}
	result := post(router, "/api/v1/signons", body, &resp, corsHeader)

	expectStatus(t, http.StatusBadRequest, result)
}

// TestCreateSignonInvalidEmail tests creating a sign-on with invalid email format
func TestCreateSignonInvalidEmail(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	body := `{
		"name": "John Doe",
		"email": "not-an-email",
		"location": "New York"
	}`

	var resp map[string]any
	corsHeader := header{"Origin", "http://test-origin"}
	result := post(router, "/api/v1/signons", body, &resp, corsHeader)

	expectStatus(t, http.StatusBadRequest, result)
}

// TestCreateSignonDuplicateEmail tests creating a sign-on with duplicate email
func TestCreateSignonDuplicateEmail(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	corsHeader := header{"Origin", "http://test-origin"}

	// First sign-on
	body1 := `{
		"name": "John Doe",
		"email": "john@example.com",
		"location": "New York"
	}`
	var resp1 map[string]any
	result1 := post(router, "/api/v1/signons", body1, &resp1, corsHeader)
	expectStatus(t, http.StatusCreated, result1)

	// Second sign-on with same email
	body2 := `{
		"name": "Jane Doe",
		"email": "john@example.com",
		"location": "Los Angeles"
	}`
	var resp2 map[string]any
	result2 := post(router, "/api/v1/signons", body2, &resp2, corsHeader)

	expectStatus(t, http.StatusConflict, result2)
}

// TestCreateSignonCORSForbidden tests creating a sign-on with disallowed CORS origin
func TestCreateSignonCORSForbidden(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	body := `{
		"name": "John Doe",
		"email": "john@example.com",
		"location": "New York"
	}`

	var resp map[string]any
	corsHeader := header{"Origin", "http://disallowed-origin"}
	result := post(router, "/api/v1/signons", body, &resp, corsHeader)

	expectStatus(t, http.StatusForbidden, result)
}

// TestCreateSignonInvalidJSON tests creating a sign-on with invalid JSON
func TestCreateSignonInvalidJSON(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	body := `{invalid json`

	var resp map[string]any
	corsHeader := header{"Origin", "http://test-origin"}
	result := post(router, "/api/v1/signons", body, &resp, corsHeader)

	expectStatus(t, http.StatusBadRequest, result)
}

// TestListSignonsSuccess tests listing sign-ons with default pagination
func TestListSignonsSuccess(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	// Create some sign-ons
	corsHeader := header{"Origin", "http://test-origin"}
	for i := 1; i <= 3; i++ {
		body := `{
			"name": "Person` + fmt.Sprintf("%d", i) + `",
			"email": "person` + fmt.Sprintf("%d", i) + `@example.com",
			"location": "City` + fmt.Sprintf("%d", i) + `"
		}`
		var resp map[string]any
		post(router, "/api/v1/signons", body, &resp, corsHeader)
	}

	// List sign-ons
	var listResp map[string]any
	result := get(router, "/api/v1/signons", &listResp, corsHeader)

	expectStatus(t, http.StatusOK, result)

	data, ok := listResp["data"].(map[string]any)
	if !ok {
		t.Fatalf("Expected data field to be a map, got %T", listResp["data"])
	}

	// Handle both nil and empty array cases
	var signons []any
	if s, ok := data["signons"].([]any); ok {
		signons = s
	}

	if len(signons) != 3 {
		t.Errorf("Expected 3 sign-ons, got %d", len(signons))
	}

	total, ok := data["total"].(float64)
	if !ok || int(total) != 3 {
		t.Errorf("Expected total 3, got %v", total)
	}
}

// TestListSignonsEmptyList tests listing sign-ons when none exist
func TestListSignonsEmptyList(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	var listResp map[string]any
	corsHeader := header{"Origin", "http://test-origin"}
	result := get(router, "/api/v1/signons", &listResp, corsHeader)

	expectStatus(t, http.StatusOK, result)

	data, ok := listResp["data"].(map[string]any)
	if !ok {
		t.Fatalf("Expected data field to be a map, got %T", listResp["data"])
	}

	// Handle both nil and empty array cases
	var signons []any
	if s, ok := data["signons"].([]any); ok {
		signons = s
	}

	if len(signons) != 0 {
		t.Errorf("Expected 0 sign-ons, got %d", len(signons))
	}
}

// TestListSignonsWithPagination tests listing sign-ons with custom limit and offset
func TestListSignonsWithPagination(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	// Create 5 sign-ons
	corsHeader := header{"Origin", "http://test-origin"}
	for i := 1; i <= 5; i++ {
		body := `{"name": "Person", "email": "person` + fmt.Sprintf("%d", i) + `@example.com", "location": "City"}`
		var resp map[string]any
		post(router, "/api/v1/signons", body, &resp, corsHeader)
	}

	// List with limit=2, offset=1
	var listResp map[string]any
	result := get(router, "/api/v1/signons?limit=2&offset=1", &listResp, corsHeader)

	expectStatus(t, http.StatusOK, result)

	data, ok := listResp["data"].(map[string]any)
	if !ok {
		t.Fatalf("Expected data field to be a map, got %T", listResp["data"])
	}

	// Handle both nil and empty array cases
	var signons []any
	if s, ok := data["signons"].([]any); ok {
		signons = s
	}

	if len(signons) != 2 {
		t.Errorf("Expected 2 sign-ons (limit), got %d", len(signons))
	}

	limit, ok := data["limit"].(float64)
	if !ok || int(limit) != 2 {
		t.Errorf("Expected limit 2 in response, got %v", limit)
	}

	offset, ok := data["offset"].(float64)
	if !ok || int(offset) != 1 {
		t.Errorf("Expected offset 1 in response, got %v", offset)
	}
}

// TestListSignonsCORSValidation tests listing sign-ons with CORS validation
func TestListSignonsCORSValidation(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	var listResp map[string]any
	corsHeader := header{"Origin", "http://disallowed-origin"}
	result := get(router, "/api/v1/signons", &listResp, corsHeader)

	expectStatus(t, http.StatusForbidden, result)
}

// ===== ADMIN ENDPOINTS =====

// TestListSignonsAdminSuccess tests listing sign-ons with admin authentication
func TestListSignonsAdminSuccess(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	// Create a sign-on
	corsHeader := header{"Origin", "http://test-origin"}
	body := `{"name": "John", "email": "john@example.com", "location": "NYC"}`
	var resp map[string]any
	post(router, "/api/v1/signons", body, &resp, corsHeader)

	// List with auth
	authHeader := makeTestAuthHeader(t)
	var listResp map[string]any
	result := get(router, "/api/v1/admin/signons", &listResp, authHeader)

	expectStatus(t, http.StatusOK, result)
}

// TestListSignonsAdminNoAuth tests listing sign-ons without authentication
func TestListSignonsAdminNoAuth(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	var listResp map[string]any
	result := get(router, "/api/v1/admin/signons", &listResp)

	expectStatus(t, http.StatusUnauthorized, result)
}

// TestDeleteSignonSuccess tests deleting a sign-on
func TestDeleteSignonSuccess(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	// Create a sign-on
	corsHeader := header{"Origin", "http://test-origin"}
	body := `{"name": "John", "email": "john@example.com", "location": "NYC"}`
	var createResp map[string]any
	post(router, "/api/v1/signons", body, &createResp, corsHeader)

	createdData, _ := createResp["data"].(map[string]any)
	signonID := createdData["id"]

	// Delete with auth
	authHeader := makeTestAuthHeader(t)
	result := del(router, "/api/v1/admin/signons/"+formatID(signonID), nil, authHeader)

	expectStatus(t, http.StatusNoContent, result)
}

// TestDeleteSignonNotFound tests deleting a non-existent sign-on
func TestDeleteSignonNotFound(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	authHeader := makeTestAuthHeader(t)
	result := del(router, "/api/v1/admin/signons/999", nil, authHeader)

	expectStatus(t, http.StatusNotFound, result)
}

// TestDeleteSignonInvalidID tests deleting a sign-on with invalid ID
func TestDeleteSignonInvalidID(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	authHeader := makeTestAuthHeader(t)
	result := del(router, "/api/v1/admin/signons/invalid", nil, authHeader)

	expectStatus(t, http.StatusBadRequest, result)
}

// TestDeleteSignonNoAuth tests deleting a sign-on without authentication
func TestDeleteSignonNoAuth(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	result := del(router, "/api/v1/admin/signons/1", nil)

	expectStatus(t, http.StatusUnauthorized, result)
}

// formatID converts a JSON ID (float64) to string for use in URLs
func formatID(id any) string {
	if f, ok := id.(float64); ok {
		return json.Number(fmt.Sprintf("%.0f", f)).String()
	}
	return ""
}
