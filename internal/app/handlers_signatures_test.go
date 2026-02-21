package app

import (
	"cosign/internal/service"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"git.sr.ht/~jakintosh/command-go/pkg/wire"
)

func TestHandleCreateSignatureHTMXErrorRendersPanel(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/campaigns/cmp-1/signatures":
			wire.WriteError(w, http.StatusBadRequest, service.ErrLocationNotInOptions.Error())
		case r.Method == http.MethodGet && r.URL.Path == "/admin/campaigns/cmp-1/signatures":
			wire.WriteData(w, http.StatusOK, service.Signatures{Limit: 10})
		default:
			wire.WriteError(w, http.StatusNotFound, "not found")
		}
	}))
	defer backend.Close()

	server, err := New(Options{
		Client: wire.Client{BaseURL: backend.URL},
	})
	if err != nil {
		t.Fatalf("new dashboard server: %v", err)
	}

	form := url.Values{
		"name":     {"Alice"},
		"email":    {"alice@example.com"},
		"location": {"Berlin"},
	}

	req := httptest.NewRequest(
		http.MethodPost,
		"/campaigns/cmp-1/signatures",
		strings.NewReader(form.Encode()),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")

	res := httptest.NewRecorder()
	server.BuildRouter().ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}

	body := res.Body.String()
	if !strings.Contains(body, `id="signatures-panel"`) {
		t.Fatalf("expected rendered signatures panel, got body: %q", body)
	}
	if !strings.Contains(body, service.ErrLocationNotInOptions.Error()) {
		t.Fatalf("expected error message in panel, got body: %q", body)
	}
}
