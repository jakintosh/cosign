package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type ErrMalformedQuery struct {
	Query string
}

func (e ErrMalformedQuery) Error() string {
	return fmt.Sprintf("Malformed '%s' Query", e.Query)
}

type APIResponse struct {
	Error *APIError `json:"error,omitempty"`
	Data  any       `json:"data,omitempty"`
}
type APIError struct {
	Message string `json:"message"`
}

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

func parsePaginationQueries(
	r *http.Request,
) (
	int,
	int,
	*ErrMalformedQuery,
) {
	limitQ := r.URL.Query().Get("limit")
	offsetQ := r.URL.Query().Get("offset")

	limit := 100
	if limitQ != "" {
		var err error
		limit, err = strconv.Atoi(limitQ)
		if err != nil {
			return 0, 0, &ErrMalformedQuery{"limit"}
		}
	}

	offset := 0
	if offsetQ != "" {
		var err error
		offset, err = strconv.Atoi(offsetQ)
		if err != nil {
			return 0, 0, &ErrMalformedQuery{"offset"}
		}
	}
	return limit, offset, nil
}

func writeJSON(
	w http.ResponseWriter,
	status int,
	v any,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func writeData(
	w http.ResponseWriter,
	status int,
	data any,
) {
	writeJSON(w, status, APIResponse{Data: data})
}

func writeError(
	w http.ResponseWriter,
	status int,
	message string,
) {
	writeJSON(w, status, APIResponse{
		Error: &APIError{Message: message},
	})
}
