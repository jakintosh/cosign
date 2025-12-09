package api_test

import (
	"cosign/internal/api"
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

func setupCORS() {
	service.SetAllowedOrigins([]service.AllowedOrigin{{URL: "http://test-default"}})
}

// setupRouter builds a fresh router for testing
func setupRouter() *mux.Router {
	router := mux.NewRouter()
	api.BuildRouter(router.PathPrefix("/api/v1").Subrouter())
	return router
}

// makeTestAuthHeader creates a valid authorization header for admin endpoints
func makeTestAuthHeader(t *testing.T) header {
	token, err := service.CreateAPIKey()
	if err != nil {
		t.Fatal(err)
	}
	auth := header{"Authorization", "Bearer " + token}
	return auth
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
