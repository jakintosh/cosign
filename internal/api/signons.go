package api

import (
	"cosign/internal/service"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// buildSignonRouter builds the public signon routes
func buildSignonRouter(r *mux.Router) {
	r.HandleFunc("", withCORS(withRateLimit(handleCreateSignon))).Methods("POST", "OPTIONS")
	r.HandleFunc("", withCORS(handleListSignons)).Methods("GET", "OPTIONS")
}

// buildAdminSignonRouter builds the admin signon routes
func buildAdminSignonRouter(r *mux.Router) {
	r.HandleFunc("", withAuth(handleListSignons)).Methods("GET")
	r.HandleFunc("/{id}", withAuth(handleDeleteSignon)).Methods("DELETE")
}

type createSignonRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Location string `json:"location"`
}

func handleCreateSignon(w http.ResponseWriter, r *http.Request) {
	var req createSignonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	signon, err := service.CreateSignon(req.Name, req.Email, req.Location, false)
	if err != nil {
		switch err {
		case service.ErrEmptyName, service.ErrEmptyEmail, service.ErrEmptyLocation:
			writeError(w, http.StatusBadRequest, err.Error())
		case service.ErrInvalidEmail:
			writeError(w, http.StatusBadRequest, err.Error())
		case service.ErrDuplicateEmail:
			writeError(w, http.StatusConflict, err.Error())
		case service.ErrLocationNotInOptions:
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "Failed to create sign-on")
		}
		return
	}

	writeData(w, http.StatusCreated, signon)
}

func handleListSignons(w http.ResponseWriter, r *http.Request) {
	// Parse pagination params
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 100
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	signons, err := service.ListSignons(limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list sign-ons")
		return
	}

	// Also get total count
	count, err := service.CountSignons()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to count sign-ons")
		return
	}

	response := map[string]any{
		"signons": signons,
		"total":   count,
		"limit":   limit,
		"offset":  offset,
	}

	writeData(w, http.StatusOK, response)
}

func handleDeleteSignon(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	if err := service.DeleteSignon(id); err != nil {
		if err == service.ErrSignonNotFound {
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, "Failed to delete sign-on")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
