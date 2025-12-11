package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// APIError represents an error in the API response
type APIError struct {
	Message string `json:"message"`
}

// APIResponse is the standard response format
type APIResponse struct {
	Error *APIError `json:"error,omitempty"`
	Data  any       `json:"data,omitempty"`
}

// BuildRouter builds the main API router with all routes
func BuildRouter(r *mux.Router) {

	buildHealthRouter(r.PathPrefix("/health").Subrouter())

	// Campaign-scoped public routes
	buildCampaignPublicRouter(r.PathPrefix("/campaigns").Subrouter())

	// Admin routes
	admin := r.PathPrefix("/admin").Subrouter()
	buildAdminCampaignRouter(admin.PathPrefix("/campaigns").Subrouter())
	buildAdminKeysRouter(admin.PathPrefix("/keys").Subrouter())
	buildAdminCORSRouter(admin.PathPrefix("/cors").Subrouter())
}

// writeJSON writes a JSON response with the given status code
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// writeData writes a successful response with data
func writeData(w http.ResponseWriter, status int, data any) {
	writeJSON(w, status, APIResponse{Data: data})
}

// writeError writes an error response
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, APIResponse{
		Error: &APIError{Message: message},
	})
}
