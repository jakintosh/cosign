package api_test

import (
	"net/http"
	"testing"
)

// TestListAPIKeysSuccess tests listing API keys with valid authentication
func TestListAPIKeysSuccess(t *testing.T) {
	setupTestDB(t)
	setupServices(t)

	router := setupRouter()

	authHeader := makeTestAuthHeader(t)
	var resp map[string]any
	result := get(router, "/api/v1/admin/keys", &resp, authHeader)

	expectStatus(t, http.StatusOK, result)

	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatalf("Expected data field to be a map, got %T", resp["data"])
	}

	// Should have a keys array (might be empty or have the test key)
	keys, ok := data["keys"].([]any)
	if !ok {
		t.Fatalf("Expected keys field to be an array, got %T", data["keys"])
	}

	// At minimum should have the test key created during setup
	if len(keys) == 0 {
		t.Errorf("Expected at least 1 key (test key), got %d", len(keys))
	}
}

// TestListAPIKeysNoAuth tests listing API keys without authentication
func TestListAPIKeysNoAuth(t *testing.T) {
	setupTestDB(t)
	setupServices(t)

	router := setupRouter()

	var resp map[string]any
	result := get(router, "/api/v1/admin/keys", &resp)

	expectStatus(t, http.StatusUnauthorized, result)
}

// TestCreateAPIKeySuccess tests creating an API key
func TestCreateAPIKeySuccess(t *testing.T) {
	setupTestDB(t)
	setupServices(t)

	router := setupRouter()

	authHeader := makeTestAuthHeader(t)
	body := `{}`
	var resp map[string]any
	result := post(router, "/api/v1/admin/keys", body, &resp, authHeader)

	expectStatus(t, http.StatusCreated, result)

	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatalf("Expected data field to be a map, got %T", resp["data"])
	}

	// Check for key field
	key, ok := data["key"].(string)
	if !ok || key == "" {
		t.Errorf("Expected key field to be a non-empty string, got %v", key)
	}

	// Check for note field
	note, ok := data["note"].(string)
	if !ok || note == "" {
		t.Errorf("Expected note field to be present, got %v", note)
	}

	// Verify key format: {id}.{secret}
	if len(key) < 3 || key[0] == '.' || key[len(key)-1] == '.' {
		t.Errorf("Expected key to be in format id.secret, got %s", key)
	}
}

// TestCreateAPIKeyWithCustomID tests creating an API key with a custom ID
func TestCreateAPIKeyWithCustomID(t *testing.T) {
	setupTestDB(t)
	setupServices(t)

	router := setupRouter()

	authHeader := makeTestAuthHeader(t)
	body := `{"id": "my-custom-key"}`
	var resp map[string]any
	result := post(router, "/api/v1/admin/keys", body, &resp, authHeader)

	expectStatus(t, http.StatusCreated, result)

	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatalf("Expected data field to be a map, got %T", resp["data"])
	}

	// Check that the returned key starts with our custom ID
	key, ok := data["key"].(string)
	if !ok || key == "" {
		t.Errorf("Expected key field to be a non-empty string, got %v", key)
	}

	// The key should start with "my-custom-key."
	expectedPrefix := "my-custom-key."
	if len(key) < len(expectedPrefix) || key[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("Expected key to start with %s, got %s", expectedPrefix, key)
	}
}

// TestCreateAPIKeyInvalidJSON tests creating a key with invalid JSON (should still work)
func TestCreateAPIKeyInvalidJSON(t *testing.T) {
	setupTestDB(t)
	setupServices(t)

	router := setupRouter()

	authHeader := makeTestAuthHeader(t)
	body := `{invalid json`
	var resp map[string]any
	result := post(router, "/api/v1/admin/keys", body, &resp, authHeader)

	// According to the implementation, invalid JSON for ID is treated as empty ID
	// so this should still create a key successfully
	expectStatus(t, http.StatusCreated, result)
}

// TestCreateAPIKeyNoAuth tests creating an API key without authentication
func TestCreateAPIKeyNoAuth(t *testing.T) {
	setupTestDB(t)
	setupServices(t)

	router := setupRouter()

	body := `{}`
	var resp map[string]any
	result := post(router, "/api/v1/admin/keys", body, &resp)

	expectStatus(t, http.StatusUnauthorized, result)
}

// TestDeleteAPIKeySuccess tests deleting an API key
func TestDeleteAPIKeySuccess(t *testing.T) {
	setupTestDB(t)
	setupServices(t)

	router := setupRouter()

	authHeader := makeTestAuthHeader(t)

	// First create a key
	createBody := `{"id": "key-to-delete"}`
	var createResp map[string]any
	post(router, "/api/v1/admin/keys", createBody, &createResp, authHeader)

	// Delete the key
	result := del(router, "/api/v1/admin/keys/key-to-delete", nil, authHeader)

	expectStatus(t, http.StatusNoContent, result)

	// Verify it's deleted by trying to list and checking it's not there
	var listResp map[string]any
	get(router, "/api/v1/admin/keys", &listResp, authHeader)

	data, _ := listResp["data"].(map[string]any)
	keys, _ := data["keys"].([]any)

	// Find the deleted key
	for _, k := range keys {
		keyData, ok := k.(map[string]any)
		if ok {
			if id, ok := keyData["id"].(string); ok && id == "key-to-delete" {
				t.Errorf("Expected key to be deleted, but found it in list")
				return
			}
		}
	}
}

// TestDeleteAPIKeyNotFound tests deleting a non-existent API key
func TestDeleteAPIKeyNotFound(t *testing.T) {
	setupTestDB(t)
	setupServices(t)

	router := setupRouter()

	authHeader := makeTestAuthHeader(t)
	result := del(router, "/api/v1/admin/keys/nonexistent", nil, authHeader)

	expectStatus(t, http.StatusNotFound, result)
}

// TestDeleteAPIKeyNoAuth tests deleting an API key without authentication
func TestDeleteAPIKeyNoAuth(t *testing.T) {
	setupTestDB(t)
	setupServices(t)

	router := setupRouter()

	result := del(router, "/api/v1/admin/keys/some-key", nil)

	expectStatus(t, http.StatusUnauthorized, result)
}
