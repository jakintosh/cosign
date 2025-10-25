package api

import (
	"cosign/internal/service"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// buildAdminKeyRouter builds the admin API key routes
func buildAdminKeyRouter(r *mux.Router) {
	r.HandleFunc("", withAuth(handleListAPIKeys)).Methods("GET")
	r.HandleFunc("", withAuth(handleCreateAPIKey)).Methods("POST")
	r.HandleFunc("/{id}", withAuth(handleDeleteAPIKey)).Methods("DELETE")
}

func handleListAPIKeys(w http.ResponseWriter, r *http.Request) {
	keys, err := service.ListAPIKeys()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list API keys")
		return
	}

	writeData(w, http.StatusOK, map[string]any{"keys": keys})
}

type createAPIKeyRequest struct {
	ID string `json:"id,omitempty"`
}

func handleCreateAPIKey(w http.ResponseWriter, r *http.Request) {
	var req createAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Empty body is ok
		req.ID = ""
	}

	fullKey, err := service.CreateAPIKey(req.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create API key")
		return
	}

	writeData(w, http.StatusCreated, map[string]any{
		"key": fullKey,
		"note": "Save this key securely. It will not be shown again.",
	})
}

func handleDeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := service.DeleteAPIKey(id); err != nil {
		if err == service.ErrAPIKeyNotFound {
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, "Failed to delete API key")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
