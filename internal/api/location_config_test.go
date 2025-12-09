package api_test

import (
	"cosign/internal/util"
	"fmt"
	"net/http"
	"testing"
)

// ===== PUBLIC ENDPOINTS =====

// TestGetLocationConfigSuccess tests retrieving location config with public endpoint
func TestGetLocationConfigSuccess(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	var resp map[string]any
	corsHeader := header{"Origin", "http://test-origin"}
	result := get(router, "/api/v1/location-config", &resp, corsHeader)

	expectStatus(t, http.StatusOK, result)

	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatalf("Expected data field to be a map, got %T", resp["data"])
	}

	config, ok := data["config"].(map[string]any)
	if !ok {
		t.Fatalf("Expected config field to be a map, got %T", data["config"])
	}

	// Check that allow_custom_text is present
	if _, ok := config["allow_custom_text"]; !ok {
		t.Errorf("Expected allow_custom_text field in config")
	}

	// Check that options field exists
	if _, ok := data["options"]; !ok {
		t.Errorf("Expected options field in response")
	}
}

// TestGetLocationConfigCORSForbidden tests retrieving location config with disallowed origin
func TestGetLocationConfigCORSForbidden(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	var resp map[string]any
	corsHeader := header{"Origin", "http://disallowed-origin"}
	result := get(router, "/api/v1/location-config", &resp, corsHeader)

	expectStatus(t, http.StatusForbidden, result)
}

// ===== ADMIN ENDPOINTS =====

// TestGetLocationConfigAdminSuccess tests retrieving location config with admin auth
func TestGetLocationConfigAdminSuccess(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	authHeader := makeTestAuthHeader(t)
	var resp map[string]any
	result := get(router, "/api/v1/admin/location-config", &resp, authHeader)

	expectStatus(t, http.StatusOK, result)

	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatalf("Expected data field to be a map, got %T", resp["data"])
	}

	config, ok := data["config"].(map[string]any)
	if !ok {
		t.Fatalf("Expected config field to be a map, got %T", data["config"])
	}

	if _, ok := config["allow_custom_text"]; !ok {
		t.Errorf("Expected allow_custom_text field in config")
	}
}

// TestGetLocationConfigAdminNoAuth tests retrieving location config without auth
func TestGetLocationConfigAdminNoAuth(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	var resp map[string]any
	result := get(router, "/api/v1/admin/location-config", &resp)

	expectStatus(t, http.StatusUnauthorized, result)
}

// TestUpdateLocationConfigEnableCustomText tests enabling custom text
func TestUpdateLocationConfigEnableCustomText(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	authHeader := makeTestAuthHeader(t)
	body := `{"allow_custom_text": true}`
	result := put(router, "/api/v1/admin/location-config", body, nil, authHeader)

	expectStatus(t, http.StatusNoContent, result)
}

// TestUpdateLocationConfigDisableCustomText tests disabling custom text
func TestUpdateLocationConfigDisableCustomText(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	authHeader := makeTestAuthHeader(t)
	body := `{"allow_custom_text": false}`
	result := put(router, "/api/v1/admin/location-config", body, nil, authHeader)

	expectStatus(t, http.StatusNoContent, result)
}

// TestUpdateLocationConfigInvalidJSON tests updating with invalid JSON
func TestUpdateLocationConfigInvalidJSON(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	authHeader := makeTestAuthHeader(t)
	body := `{invalid json`
	result := put(router, "/api/v1/admin/location-config", body, nil, authHeader)

	expectStatus(t, http.StatusBadRequest, result)
}

// TestUpdateLocationConfigNoAuth tests updating without authentication
func TestUpdateLocationConfigNoAuth(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	body := `{"allow_custom_text": true}`
	result := put(router, "/api/v1/admin/location-config", body, nil)

	expectStatus(t, http.StatusUnauthorized, result)
}

// TestListLocationOptionsSuccess tests listing location options
func TestListLocationOptionsSuccess(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	authHeader := makeTestAuthHeader(t)
	var resp map[string]any
	result := get(router, "/api/v1/admin/location-config/options", &resp, authHeader)

	expectStatus(t, http.StatusOK, result)

	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatalf("Expected data field to be a map, got %T", resp["data"])
	}

	// Handle both nil and empty array cases
	var options []any
	if o, ok := data["options"].([]any); ok {
		options = o
	}

	// Should be empty initially
	if len(options) != 0 {
		t.Errorf("Expected 0 options initially, got %d", len(options))
	}
}

// TestListLocationOptionsNoAuth tests listing options without authentication
func TestListLocationOptionsNoAuth(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	var resp map[string]any
	result := get(router, "/api/v1/admin/location-config/options", &resp)

	expectStatus(t, http.StatusUnauthorized, result)
}

// TestAddLocationOptionSuccess tests adding a location option
func TestAddLocationOptionSuccess(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	authHeader := makeTestAuthHeader(t)
	body := `{"value": "New York", "display_order": 1}`
	var resp map[string]any
	result := post(router, "/api/v1/admin/location-config/options", body, &resp, authHeader)

	expectStatus(t, http.StatusCreated, result)

	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatalf("Expected data field to be a map, got %T", resp["data"])
	}

	if _, ok := data["id"]; !ok {
		t.Errorf("Expected id field in response")
	}
}

// TestAddLocationOptionEmptyValue tests adding option with empty value
func TestAddLocationOptionEmptyValue(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	authHeader := makeTestAuthHeader(t)
	body := `{"value": "", "display_order": 1}`
	var resp map[string]any
	result := post(router, "/api/v1/admin/location-config/options", body, &resp, authHeader)

	expectStatus(t, http.StatusBadRequest, result)
}

// TestAddLocationOptionInvalidJSON tests adding option with invalid JSON
func TestAddLocationOptionInvalidJSON(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	authHeader := makeTestAuthHeader(t)
	body := `{invalid json`
	var resp map[string]any
	result := post(router, "/api/v1/admin/location-config/options", body, &resp, authHeader)

	expectStatus(t, http.StatusBadRequest, result)
}

// TestAddLocationOptionNoAuth tests adding option without authentication
func TestAddLocationOptionNoAuth(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	body := `{"value": "New York", "display_order": 1}`
	var resp map[string]any
	result := post(router, "/api/v1/admin/location-config/options", body, &resp)

	expectStatus(t, http.StatusUnauthorized, result)
}

// TestUpdateLocationOptionSuccess tests updating a location option
func TestUpdateLocationOptionSuccess(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	authHeader := makeTestAuthHeader(t)

	// First add an option
	addBody := `{"value": "New York", "display_order": 1}`
	var addResp map[string]any
	post(router, "/api/v1/admin/location-config/options", addBody, &addResp, authHeader)

	addData, _ := addResp["data"].(map[string]any)
	optionID := formatID(addData["id"])

	// Update the option
	updateBody := `{"value": "Boston", "display_order": 2}`
	result := put(router, fmt.Sprintf("/api/v1/admin/location-config/options/%s", optionID), updateBody, nil, authHeader)

	expectStatus(t, http.StatusNoContent, result)
}

// TestUpdateLocationOptionInvalidID tests updating with invalid ID
func TestUpdateLocationOptionInvalidID(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	authHeader := makeTestAuthHeader(t)
	body := `{"value": "Boston", "display_order": 2}`
	result := put(router, "/api/v1/admin/location-config/options/invalid", body, nil, authHeader)

	expectStatus(t, http.StatusBadRequest, result)
}

// TestUpdateLocationOptionNotFound tests updating non-existent option
func TestUpdateLocationOptionNotFound(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	authHeader := makeTestAuthHeader(t)
	body := `{"value": "Boston", "display_order": 2}`
	result := put(router, "/api/v1/admin/location-config/options/999", body, nil, authHeader)

	expectStatus(t, http.StatusNotFound, result)
}

// TestUpdateLocationOptionEmptyValue tests updating with empty value
func TestUpdateLocationOptionEmptyValue(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	authHeader := makeTestAuthHeader(t)

	// First add an option
	addBody := `{"value": "New York", "display_order": 1}`
	var addResp map[string]any
	post(router, "/api/v1/admin/location-config/options", addBody, &addResp, authHeader)

	addData, _ := addResp["data"].(map[string]any)
	optionID := formatID(addData["id"])

	// Try to update with empty value
	updateBody := `{"value": "", "display_order": 2}`
	result := put(router, fmt.Sprintf("/api/v1/admin/location-config/options/%s", optionID), updateBody, nil, authHeader)

	expectStatus(t, http.StatusBadRequest, result)
}

// TestUpdateLocationOptionInvalidJSON tests updating with invalid JSON
func TestUpdateLocationOptionInvalidJSON(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	authHeader := makeTestAuthHeader(t)
	body := `{invalid json`
	result := put(router, "/api/v1/admin/location-config/options/1", body, nil, authHeader)

	expectStatus(t, http.StatusBadRequest, result)
}

// TestUpdateLocationOptionNoAuth tests updating without authentication
func TestUpdateLocationOptionNoAuth(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	body := `{"value": "Boston", "display_order": 2}`
	result := put(router, "/api/v1/admin/location-config/options/1", body, nil)

	expectStatus(t, http.StatusUnauthorized, result)
}

// TestDeleteLocationOptionSuccess tests deleting a location option
func TestDeleteLocationOptionSuccess(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	authHeader := makeTestAuthHeader(t)

	// First add an option
	addBody := `{"value": "New York", "display_order": 1}`
	var addResp map[string]any
	post(router, "/api/v1/admin/location-config/options", addBody, &addResp, authHeader)

	addData, _ := addResp["data"].(map[string]any)
	optionID := formatID(addData["id"])

	// Delete the option
	result := del(router, fmt.Sprintf("/api/v1/admin/location-config/options/%s", optionID), nil, authHeader)

	expectStatus(t, http.StatusNoContent, result)
}

// TestDeleteLocationOptionInvalidID tests deleting with invalid ID
func TestDeleteLocationOptionInvalidID(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	authHeader := makeTestAuthHeader(t)
	result := del(router, "/api/v1/admin/location-config/options/invalid", nil, authHeader)

	expectStatus(t, http.StatusBadRequest, result)
}

// TestDeleteLocationOptionNotFound tests deleting non-existent option
func TestDeleteLocationOptionNotFound(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	authHeader := makeTestAuthHeader(t)
	result := del(router, "/api/v1/admin/location-config/options/999", nil, authHeader)

	expectStatus(t, http.StatusNotFound, result)
}

// TestDeleteLocationOptionNoAuth tests deleting without authentication
func TestDeleteLocationOptionNoAuth(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	result := del(router, "/api/v1/admin/location-config/options/1", nil)

	expectStatus(t, http.StatusUnauthorized, result)
}
