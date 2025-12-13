package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"cosign/internal/service"
	"github.com/gorilla/mux"
)

func buildAdminCORSRouter(
	r *mux.Router,
) {
	r.HandleFunc("", withAuth(handleGetCORS)).Methods("GET")
	r.HandleFunc("", withAuth(handlePostCORS)).Methods("POST")
	r.HandleFunc("/{origin}", withAuth(handleDeleteCORS)).Methods("DELETE")
}

func handleGetCORS(
	w http.ResponseWriter,
	r *http.Request,
) {
	origins, err := service.GetAllowedOrigins()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	writeData(w, http.StatusOK, origins)
}

func handlePostCORS(
	w http.ResponseWriter,
	r *http.Request,
) {
	var payload struct {
		Origin string `json:"origin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "Malformed JSON")
		return
	}

	origin := strings.TrimSpace(payload.Origin)
	if origin == "" {
		writeError(w, http.StatusBadRequest, "Origin is required")
		return
	}

	if err := service.AddAllowedOrigin(origin); err != nil {
		if errors.Is(err, service.ErrCORSOriginNotFound) {
			writeError(w, http.StatusBadRequest, "Invalid Origin URL")
		} else {
			writeError(w, http.StatusInternalServerError, "Internal Server Error")
		}
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func handleDeleteCORS(
	w http.ResponseWriter,
	r *http.Request,
) {
	origin := mux.Vars(r)["origin"]
	origin = strings.TrimSpace(origin)
	if origin == "" {
		writeError(w, http.StatusBadRequest, "Origin is required")
		return
	}

	if err := service.DeleteAllowedOrigin(origin); err != nil {
		if errors.Is(err, service.ErrCORSOriginNotFound) {
			writeError(w, http.StatusNotFound, "Origin not found")
		} else {
			writeError(w, http.StatusInternalServerError, "Internal Server Error")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
