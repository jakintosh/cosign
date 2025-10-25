package api

import (
	"cosign/internal/service"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// buildLocationConfigRouter builds the public location config routes
func buildLocationConfigRouter(r *mux.Router) {
	r.HandleFunc("", withCORS(handleGetLocationConfig)).Methods("GET", "OPTIONS")
}

// buildAdminLocationConfigRouter builds the admin location config routes
func buildAdminLocationConfigRouter(r *mux.Router) {
	r.HandleFunc("", withAuth(handleGetLocationConfig)).Methods("GET")
	r.HandleFunc("", withAuth(handleUpdateLocationConfig)).Methods("PUT")
	r.HandleFunc("/options", withAuth(handleListLocationOptions)).Methods("GET")
	r.HandleFunc("/options", withAuth(handleAddLocationOption)).Methods("POST")
	r.HandleFunc("/options/{id}", withAuth(handleUpdateLocationOption)).Methods("PUT")
	r.HandleFunc("/options/{id}", withAuth(handleDeleteLocationOption)).Methods("DELETE")
}

func handleGetLocationConfig(w http.ResponseWriter, r *http.Request) {
	config, options, err := service.GetLocationConfigWithOptions()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get location config")
		return
	}

	response := map[string]any{
		"config":  config,
		"options": options,
	}

	writeData(w, http.StatusOK, response)
}

type updateLocationConfigRequest struct {
	AllowCustomText bool `json:"allow_custom_text"`
}

func handleUpdateLocationConfig(w http.ResponseWriter, r *http.Request) {
	var req updateLocationConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := service.SetAllowCustomText(req.AllowCustomText); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update location config")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleListLocationOptions(w http.ResponseWriter, r *http.Request) {
	options, err := service.GetLocationOptions()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list location options")
		return
	}

	writeData(w, http.StatusOK, map[string]any{"options": options})
}

type addLocationOptionRequest struct {
	Value        string `json:"value"`
	DisplayOrder int    `json:"display_order"`
}

func handleAddLocationOption(w http.ResponseWriter, r *http.Request) {
	var req addLocationOptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	id, err := service.AddLocationOption(req.Value, req.DisplayOrder)
	if err != nil {
		if err == service.ErrEmptyLocation {
			writeError(w, http.StatusBadRequest, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, "Failed to add location option")
		}
		return
	}

	writeData(w, http.StatusCreated, map[string]any{"id": id})
}

type updateLocationOptionRequest struct {
	Value        string `json:"value"`
	DisplayOrder int    `json:"display_order"`
}

func handleUpdateLocationOption(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var req updateLocationOptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := service.UpdateLocationOption(id, req.Value, req.DisplayOrder); err != nil {
		switch err {
		case service.ErrLocationOptionNotFound:
			writeError(w, http.StatusNotFound, err.Error())
		case service.ErrEmptyLocation:
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "Failed to update location option")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleDeleteLocationOption(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	if err := service.DeleteLocationOption(id); err != nil {
		if err == service.ErrLocationOptionNotFound {
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, "Failed to delete location option")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
