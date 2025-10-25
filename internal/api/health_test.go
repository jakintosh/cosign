package api_test

import (
	"net/http"
	"testing"
)

func TestHealthCheckSuccess(t *testing.T) {
	setupTestDB(t)
	setupServices(t)

	router := setupRouter()

	var resp map[string]any
	result := get(router, "/api/v1/health", &resp)

	expectStatus(t, http.StatusOK, result)

	if resp == nil {
		t.Fatal("Expected response data, got nil")
	}

	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatalf("Expected data field to be a map, got %T", resp["data"])
	}

	status, ok := data["status"].(string)
	if !ok || status != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", status)
	}
}
