package api

import (
	"cosign/internal/service"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type addCORSOriginRequest struct {
	Origin string `json:"origin"`
}

// buildAdminCORSRouter builds the admin CORS routes
func buildAdminCORSRouter(r *mux.Router) {
	r.HandleFunc("", withAuth(handleListCORSOrigins)).Methods("GET")
	r.HandleFunc("", withAuth(handleAddCORSOrigin)).Methods("POST")
	r.HandleFunc("/{origin}", withAuth(handleDeleteCORSOrigin)).Methods("DELETE")
}

func handleListCORSOrigins(w http.ResponseWriter, r *http.Request) {
	origins, err := service.ListCORSOrigins()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list CORS origins")
		return
	}

	writeData(w, http.StatusOK, map[string]any{"origins": origins})
}

func handleAddCORSOrigin(w http.ResponseWriter, r *http.Request) {
	var req addCORSOriginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := service.AddCORSOrigin(req.Origin); err != nil {
		if err == service.ErrOriginNotAllowed {
			writeError(w, http.StatusBadRequest, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, "Failed to add CORS origin")
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func handleDeleteCORSOrigin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	origin := vars["origin"]

	if err := service.DeleteCORSOrigin(origin); err != nil {
		if err == service.ErrCORSOriginNotFound {
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, "Failed to delete CORS origin")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
