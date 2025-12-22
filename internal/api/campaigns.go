package api

import (
	"cosign/internal/service"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type CreateCampaignRequest struct {
	Name string `json:"name"`
}

type UpdateCampaignRequest struct {
	Name            string `json:"name"`
	AllowCustomText *bool  `json:"allow_custom_text"`
}

type CreateSignonRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Location string `json:"location"`
}

func buildCampaignPublicRouter(r *mux.Router) {
	r.HandleFunc("/{campaignId}", withCORS(handleGetCampaign)).Methods("GET", "OPTIONS")
	r.HandleFunc("/{campaignId}/locations", withCORS(handleGetCampaignLocations)).Methods("GET", "OPTIONS")
	r.HandleFunc("/{campaignId}/signons", withCORS(handleListSignons)).Methods("GET", "OPTIONS")
	r.HandleFunc("/{campaignId}/signons", withCORSAndRateLimit(handleCreateSignon)).Methods("POST")
}

func buildAdminCampaignRouter(r *mux.Router) {
	r.HandleFunc("", withAuth(handleListCampaigns)).Methods("GET")
	r.HandleFunc("", withAuth(handleCreateCampaign)).Methods("POST")
	r.HandleFunc("/{campaignId}", withAuth(handleGetCampaign)).Methods("GET")
	r.HandleFunc("/{campaignId}", withAuth(handleUpdateCampaign)).Methods("PUT")
	r.HandleFunc("/{campaignId}", withAuth(handleDeleteCampaign)).Methods("DELETE")
	r.HandleFunc("/{campaignId}/locations", withAuth(handleGetCampaignLocations)).Methods("GET")
	r.HandleFunc("/{campaignId}/locations", withAuth(handleUpdateCampaignLocations)).Methods("PUT")
	r.HandleFunc("/{campaignId}/signons", withAuth(handleListSignons)).Methods("GET")
	r.HandleFunc("/{campaignId}/signons/{signonId}", withAuth(handleDeleteSignon)).Methods("DELETE")
}

// Campaign handlers
func handleListCampaigns(w http.ResponseWriter, r *http.Request) {
	limit, offset, malformedQueryErr := parsePaginationQueries(r)
	if malformedQueryErr != nil {
		writeError(w, http.StatusBadRequest, malformedQueryErr.Error())
		return
	}

	campaigns, err := service.ListCampaigns(limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list campaigns")
		return
	}

	writeData(w, http.StatusOK, campaigns)
}

func handleCreateCampaign(w http.ResponseWriter, r *http.Request) {
	var req CreateCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	campaign, err := service.CreateCampaign(req.Name)
	if err != nil {
		if err == service.ErrEmptyCampaignName {
			writeError(w, http.StatusBadRequest, "Campaign name cannot be empty")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to create campaign")
		return
	}

	writeData(w, http.StatusCreated, campaign)
}

func handleGetCampaign(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	campaignId := vars["campaignId"]
	if campaignId == "" {
		writeError(w, http.StatusBadRequest, "Campaign ID required")
		return
	}

	campaign, err := service.GetCampaign(campaignId)
	if err != nil {
		if err == service.ErrCampaignNotFound {
			writeError(w, http.StatusNotFound, "Campaign not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to get campaign")
		return
	}

	writeData(w, http.StatusOK, campaign)
}

func handleUpdateCampaign(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	campaignId := vars["campaignId"]
	if campaignId == "" {
		writeError(w, http.StatusBadRequest, "Campaign ID required")
		return
	}

	var req UpdateCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	campaign, err := service.GetCampaign(campaignId)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get campaign")
		return
	}

	// Use provided name or keep existing
	name := campaign.Name
	if req.Name != "" {
		name = req.Name
	}

	// Use provided allow_custom_text or keep existing
	allowCustomText := campaign.AllowCustomText
	if req.AllowCustomText != nil {
		allowCustomText = *req.AllowCustomText
	}

	err = service.UpdateCampaign(campaignId, name, allowCustomText)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update campaign")
		return
	}

	updated, _ := service.GetCampaign(campaignId)
	writeData(w, http.StatusOK, updated)
}

func handleDeleteCampaign(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	campaignId := vars["campaignId"]
	if campaignId == "" {
		writeError(w, http.StatusBadRequest, "Campaign ID required")
		return
	}

	err := service.DeleteCampaign(campaignId)
	if err != nil {
		if err == service.ErrCampaignNotFound {
			writeError(w, http.StatusNotFound, "Campaign not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to delete campaign")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleCreateSignon(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	campaignId := vars["campaignId"]
	if campaignId == "" {
		writeError(w, http.StatusBadRequest, "Campaign ID required")
		return
	}

	var req CreateSignonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	signon, err := service.CreateSignon(campaignId, req.Name, req.Email, req.Location, false)
	if err != nil {
		status := http.StatusBadRequest
		switch err {
		case service.ErrEmptyName, service.ErrEmptyEmail, service.ErrEmptyLocation:
			writeError(w, status, err.Error())
		case service.ErrInvalidEmail:
			writeError(w, status, err.Error())
		case service.ErrLocationNotInOptions:
			writeError(w, status, err.Error())
		case service.ErrDuplicateEmail:
			writeError(w, http.StatusConflict, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "Failed to create signon")
		}
		return
	}

	writeData(w, http.StatusCreated, signon)
}

func handleListSignons(w http.ResponseWriter, r *http.Request) {
	limit, offset, malformedQueryErr := parsePaginationQueries(r)
	if malformedQueryErr != nil {
		writeError(w, http.StatusBadRequest, malformedQueryErr.Error())
		return
	}

	vars := mux.Vars(r)
	campaignId := vars["campaignId"]
	if campaignId == "" {
		writeError(w, http.StatusBadRequest, "Campaign ID required")
		return
	}

	signons, err := service.ListSignons(campaignId, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list signons")
		return
	}

	writeData(w, http.StatusOK, signons)
}

func handleDeleteSignon(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	campaignId := vars["campaignId"]
	if campaignId == "" {
		writeError(w, http.StatusBadRequest, "Campaign ID required")
		return
	}
	signonIdStr := vars["signonId"]

	signonId, err := strconv.ParseInt(signonIdStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid signon ID")
		return
	}

	err = service.DeleteSignon(campaignId, signonId)
	if err != nil {
		if err == service.ErrSignonNotFound {
			writeError(w, http.StatusNotFound, "Signon not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to delete signon")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleGetCampaignLocations(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	campaignID := vars["campaignId"]
	if campaignID == "" {
		writeError(w, http.StatusBadRequest, "Campaign ID required")
		return
	}

	locations, err := service.GetCampaignLocations(campaignID)
	if err != nil {
		if err == service.ErrCampaignNotFound {
			writeError(w, http.StatusNotFound, "Campaign not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to get campaign locations")
		return
	}

	writeData(w, http.StatusOK, locations)
}

func handleUpdateCampaignLocations(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	campaignID := vars["campaignId"]
	if campaignID == "" {
		writeError(w, http.StatusBadRequest, "Campaign ID required")
		return
	}

	var req []service.LocationOption
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := service.SetCampaignLocations(campaignID, req); err != nil {
		if err == service.ErrEmptyLocation {
			writeError(w, http.StatusBadRequest, "Location value cannot be empty")
			return
		}
		if err == service.ErrCampaignNotFound {
			writeError(w, http.StatusNotFound, "Campaign not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to update campaign locations")
		return
	}

	updated, err := service.GetCampaignLocations(campaignID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to load updated campaign locations")
		return
	}

	writeData(w, http.StatusOK, updated)
}

// Helper to apply both CORS and rate limiting
func withCORSAndRateLimit(next http.HandlerFunc) http.HandlerFunc {
	return withCORS(withRateLimit(next))
}
