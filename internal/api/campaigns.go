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

type UpdateLocationConfigRequest struct {
	AllowCustomText bool `json:"allow_custom_text"`
}

type LocationOptionRequest struct {
	Value        string `json:"value"`
	DisplayOrder int    `json:"display_order"`
}

// buildCampaignPublicRouter builds public campaign routes
func buildCampaignPublicRouter(r *mux.Router) {
	r.HandleFunc("/{campaignId}", withCORS(handleGetCampaignPublic)).Methods("GET", "OPTIONS")
	r.HandleFunc("/{campaignId}/signons", withCORS(handleListSignonsPublic)).Methods("GET", "OPTIONS")
	r.HandleFunc("/{campaignId}/signons", withCORSAndRateLimit(handleCreateSignonPublic)).Methods("POST", "OPTIONS")
	r.HandleFunc("/{campaignId}/config", withCORS(handleGetLocationConfig)).Methods("GET", "OPTIONS")
	r.HandleFunc("/{campaignId}/options", withCORS(handleListLocationOptions)).Methods("GET", "OPTIONS")
}

// buildAdminCampaignRouter builds admin campaign routes
func buildAdminCampaignRouter(r *mux.Router) {
	// Campaign CRUD
	r.HandleFunc("", withAuth(handleListCampaigns)).Methods("GET")
	r.HandleFunc("", withAuth(handleCreateCampaign)).Methods("POST")
	r.HandleFunc("/{campaignId}", withAuth(handleGetCampaign)).Methods("GET")
	r.HandleFunc("/{campaignId}", withAuth(handleUpdateCampaign)).Methods("PUT")
	r.HandleFunc("/{campaignId}", withAuth(handleDeleteCampaign)).Methods("DELETE")

	// Campaign-scoped signons
	r.HandleFunc("/{campaignId}/signons", withAuth(handleListSignonsAdmin)).Methods("GET")
	r.HandleFunc("/{campaignId}/signons/{signonId}", withAuth(handleDeleteSignon)).Methods("DELETE")

	// Campaign-scoped config and options
	r.HandleFunc("/{campaignId}/config", withAuth(handleGetLocationConfigAdmin)).Methods("GET")
	r.HandleFunc("/{campaignId}/config", withAuth(handleUpdateLocationConfig)).Methods("PUT")
	r.HandleFunc("/{campaignId}/options", withAuth(handleListLocationOptionsAdmin)).Methods("GET")
	r.HandleFunc("/{campaignId}/options", withAuth(handleAddLocationOption)).Methods("POST")
	r.HandleFunc("/{campaignId}/options/{optionId}", withAuth(handleUpdateLocationOption)).Methods("PUT")
	r.HandleFunc("/{campaignId}/options/{optionId}", withAuth(handleDeleteLocationOption)).Methods("DELETE")
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

func handleGetCampaignPublic(w http.ResponseWriter, r *http.Request) {
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

// Signon handlers (campaign-scoped)
func handleCreateSignonPublic(w http.ResponseWriter, r *http.Request) {
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

func handleListSignonsPublic(w http.ResponseWriter, r *http.Request) {
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

func handleListSignonsAdmin(w http.ResponseWriter, r *http.Request) {
	limit, offset, malformedQueryErr := parsePaginationQueries(r)
	if malformedQueryErr != nil {
		writeError(w, http.StatusBadRequest, malformedQueryErr.Error())
		return
	}

	vars := mux.Vars(r)
	campaignID := vars["campaignId"]

	if campaignID == "" {
		writeError(w, http.StatusBadRequest, "Campaign ID required")
		return
	}

	signons, err := service.ListSignons(campaignID, limit, offset)
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

// Location config and options handlers (campaign-scoped)
func handleGetLocationConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	campaignID := vars["campaignId"]
	if campaignID == "" {
		writeError(w, http.StatusBadRequest, "Campaign ID required")
		return
	}

	config, err := service.GetLocationConfig(campaignID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get location config")
		return
	}

	options, err := service.GetLocationOptions(campaignID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get location options")
		return
	}

	response := map[string]any{
		"config":  config,
		"options": options,
	}
	writeData(w, http.StatusOK, response)
}

func handleGetLocationConfigAdmin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	campaignID := vars["campaignId"]
	if campaignID == "" {
		writeError(w, http.StatusBadRequest, "Campaign ID required")
		return
	}

	config, err := service.GetLocationConfig(campaignID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get location config")
		return
	}

	response := map[string]any{
		"config": config,
	}
	writeData(w, http.StatusOK, response)
}

func handleUpdateLocationConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	campaignID := vars["campaignId"]
	if campaignID == "" {
		writeError(w, http.StatusBadRequest, "Campaign ID required")
		return
	}

	var req UpdateLocationConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := service.SetAllowCustomText(campaignID, req.AllowCustomText)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update location config")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleListLocationOptions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	campaignID := vars["campaignId"]
	if campaignID == "" {
		writeError(w, http.StatusBadRequest, "Campaign ID required")
		return
	}

	options, err := service.GetLocationOptions(campaignID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get location options")
		return
	}

	writeData(w, http.StatusOK, options)
}

func handleListLocationOptionsAdmin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	campaignID := vars["campaignId"]
	if campaignID == "" {
		writeError(w, http.StatusBadRequest, "Campaign ID required")
		return
	}

	options, err := service.GetLocationOptions(campaignID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get location options")
		return
	}

	response := map[string]any{
		"options": options,
	}
	writeData(w, http.StatusOK, response)
}

func handleAddLocationOption(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	campaignID := vars["campaignId"]
	if campaignID == "" {
		writeError(w, http.StatusBadRequest, "Campaign ID required")
		return
	}

	var req LocationOptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	locationId, err := service.AddLocationOption(campaignID, req.Value, req.DisplayOrder)
	if err != nil {
		if err == service.ErrEmptyLocation {
			writeError(w, http.StatusBadRequest, "Location value cannot be empty")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to add location option")
		return
	}

	response := map[string]any{
		"id":            locationId,
		"value":         req.Value,
		"display_order": req.DisplayOrder,
	}
	writeData(w, http.StatusCreated, response)
}

func handleUpdateLocationOption(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	campaignID := vars["campaignId"]
	if campaignID == "" {
		writeError(w, http.StatusBadRequest, "Campaign ID required")
		return
	}
	optionIdStr := vars["optionId"]

	optionId, err := strconv.ParseInt(optionIdStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid option ID")
		return
	}

	var req LocationOptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err = service.UpdateLocationOption(campaignID, optionId, req.Value, req.DisplayOrder)
	if err != nil {
		if err == service.ErrLocationOptionNotFound {
			writeError(w, http.StatusNotFound, "Location option not found")
			return
		}
		if err == service.ErrEmptyLocation {
			writeError(w, http.StatusBadRequest, "Location value cannot be empty")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to update location option")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleDeleteLocationOption(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	campaignID := vars["campaignId"]
	if campaignID == "" {
		writeError(w, http.StatusBadRequest, "Campaign ID required")
		return
	}
	optionIdStr := vars["optionId"]

	optionId, err := strconv.ParseInt(optionIdStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid option ID")
		return
	}

	err = service.DeleteLocationOption(campaignID, optionId)
	if err != nil {
		if err == service.ErrLocationOptionNotFound {
			writeError(w, http.StatusNotFound, "Location option not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to delete location option")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper to apply both CORS and rate limiting
func withCORSAndRateLimit(next http.HandlerFunc) http.HandlerFunc {
	return withCORS(withRateLimit(next))
}
