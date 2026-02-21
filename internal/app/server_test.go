package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWithMethodOverridePatch(t *testing.T) {
	var method string
	handler := withMethodOverride(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		method = r.Method
	}))

	req := httptest.NewRequest(
		http.MethodPost,
		"/campaigns/cmp-1",
		strings.NewReader("_method=PATCH&name=Updated"),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler.ServeHTTP(httptest.NewRecorder(), req)

	if method != http.MethodPatch {
		t.Fatalf("expected method %q, got %q", http.MethodPatch, method)
	}
}

func TestWithMethodOverrideDeletePreservesFormValues(t *testing.T) {
	var (
		method string
		page   string
	)

	handler := withMethodOverride(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		method = r.Method
		page = r.FormValue("page")
	}))

	req := httptest.NewRequest(
		http.MethodPost,
		"/campaigns/cmp-1/signatures/42",
		strings.NewReader("_method=DELETE&page=3"),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler.ServeHTTP(httptest.NewRecorder(), req)

	if method != http.MethodDelete {
		t.Fatalf("expected method %q, got %q", http.MethodDelete, method)
	}
	if page != "3" {
		t.Fatalf("expected page %q, got %q", "3", page)
	}
}

func TestWithMethodOverrideIgnoresUnknownMethod(t *testing.T) {
	var method string
	handler := withMethodOverride(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		method = r.Method
	}))

	req := httptest.NewRequest(
		http.MethodPost,
		"/campaigns/cmp-1",
		strings.NewReader("_method=PUT"),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler.ServeHTTP(httptest.NewRecorder(), req)

	if method != http.MethodPost {
		t.Fatalf("expected method %q, got %q", http.MethodPost, method)
	}
}
