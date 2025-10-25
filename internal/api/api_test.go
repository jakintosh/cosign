package api

import (
	"cosign/internal/database"
	"cosign/internal/service"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

// httpResult holds the result of an HTTP request
type httpResult struct {
	Code  int
	Error error
}

// header represents an HTTP header
type header struct {
	key   string
	value string
}

// setupTestDB initializes an in-memory SQLite database for testing
func setupTestDB(t *testing.T) {
	err := database.Init(":memory:", false)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
}

// setupServices initializes all service stores with the database
func setupServices(t *testing.T) {
	// Set up all store implementations
	service.SetSignonStore(database.NewSignonStore())
	service.SetLocationConfigStore(database.NewLocationConfigStore())
	service.SetKeyStore(database.NewKeyStore())
	service.SetCORSStore(database.NewCORSStore())

	// Add a test CORS origin
	err := service.AddCORSOrigin("http://test-origin")
	if err != nil {
		t.Fatalf("Failed to add test CORS origin: %v", err)
	}

	// Create a test API key
	testKey, err := service.CreateAPIKey("test-key")
	if err != nil {
		t.Fatalf("Failed to create test API key: %v", err)
	}

	// Store the test key for use in tests
	// We can recreate it from the pattern: it will be "test-key.{secret}"
	_ = testKey // The actual key is stored in the database
}

// setupRouter builds a fresh router for testing
func setupRouter() *mux.Router {
	router := mux.NewRouter()
	BuildRouter(router.PathPrefix("/api/v1").Subrouter())
	return router
}

// makeTestAuthHeader creates a valid authorization header for admin endpoints
func makeTestAuthHeader(t *testing.T) header {
	token, err := service.CreateAPIKey("")
	if err != nil {
		t.Fatalf("Failed to create test API key: %v", err)
	}
	return header{"Authorization", "Bearer " + token}
}

// expectStatus validates that the HTTP result has the expected status code
func expectStatus(t *testing.T, expectedCode int, result httpResult) {
	if result.Code != expectedCode {
		t.Errorf("Expected status %d, got %d: %v", expectedCode, result.Code, result.Error)
	}
}

// get performs an HTTP GET request
func get(
	router *mux.Router,
	url string,
	response any,
	headers ...header,
) httpResult {
	req := httptest.NewRequest("GET", url, nil)
	res := httptest.NewRecorder()

	for _, h := range headers {
		req.Header.Set(h.key, h.value)
	}

	router.ServeHTTP(res, req)

	if response != nil && res.Body.Len() > 0 {
		if err := json.Unmarshal(res.Body.Bytes(), response); err != nil {
			return httpResult{
				Code:  res.Code,
				Error: fmt.Errorf("failed to decode JSON: %v\n%s", err, res.Body.String()),
			}
		}
	}

	return httpResult{Code: res.Code}
}

// post performs an HTTP POST request
func post(
	router *mux.Router,
	url string,
	body string,
	response any,
	headers ...header,
) httpResult {
	req := httptest.NewRequest("POST", url, strings.NewReader(body))
	res := httptest.NewRecorder()

	for _, h := range headers {
		req.Header.Set(h.key, h.value)
	}

	router.ServeHTTP(res, req)

	if response != nil && res.Body.Len() > 0 {
		if err := json.Unmarshal(res.Body.Bytes(), response); err != nil {
			return httpResult{
				Code:  res.Code,
				Error: fmt.Errorf("failed to decode JSON: %v\n%s", err, res.Body.String()),
			}
		}
	}

	return httpResult{Code: res.Code}
}

// put performs an HTTP PUT request
func put(
	router *mux.Router,
	url string,
	body string,
	response any,
	headers ...header,
) httpResult {
	req := httptest.NewRequest("PUT", url, strings.NewReader(body))
	res := httptest.NewRecorder()

	for _, h := range headers {
		req.Header.Set(h.key, h.value)
	}

	router.ServeHTTP(res, req)

	if response != nil && res.Body.Len() > 0 {
		if err := json.Unmarshal(res.Body.Bytes(), response); err != nil {
			return httpResult{
				Code:  res.Code,
				Error: fmt.Errorf("failed to decode JSON: %v\n%s", err, res.Body.String()),
			}
		}
	}

	return httpResult{Code: res.Code}
}

// del performs an HTTP DELETE request
func del(
	router *mux.Router,
	url string,
	response any,
	headers ...header,
) httpResult {
	req := httptest.NewRequest("DELETE", url, nil)
	res := httptest.NewRecorder()

	for _, h := range headers {
		req.Header.Set(h.key, h.value)
	}

	router.ServeHTTP(res, req)

	if response != nil && res.Body.Len() > 0 {
		if err := json.Unmarshal(res.Body.Bytes(), response); err != nil {
			return httpResult{
				Code:  res.Code,
				Error: fmt.Errorf("failed to decode JSON: %v\n%s", err, res.Body.String()),
			}
		}
	}

	return httpResult{Code: res.Code}
}
