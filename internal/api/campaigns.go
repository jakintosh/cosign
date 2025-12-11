package api

import (
	"cosign/internal/service"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// buildCampaignPublicRouter builds public campaign routes
func buildCampaignPublicRouter(r *mux.Router) {
	r.HandleFunc("/{uuid}", withCORSAndCampaign(handleGetCampaignPublic)).Methods("GET", "OPTIONS")
	r.HandleFunc("/{uuid}/signons", withCORSAndCampaign(handleListSignonsPublic)).Methods("GET", "OPTIONS")
	r.HandleFunc("/{uuid}/signons", withCORSAndRateLimit(withCampaign(handleCreateSignonPublic))).Methods("POST", "OPTIONS")
	r.HandleFunc("/{uuid}/config", withCORSAndCampaign(handleGetLocationConfig)).Methods("GET", "OPTIONS")
	r.HandleFunc("/{uuid}/options", withCORSAndCampaign(handleListLocationOptions)).Methods("GET", "OPTIONS")
}

// buildAdminCampaignRouter builds admin campaign routes
func buildAdminCampaignRouter(r *mux.Router) {
	// Campaign CRUD
	r.HandleFunc("", withAuth(handleListCampaigns)).Methods("GET")
	r.HandleFunc("", withAuth(handleCreateCampaign)).Methods("POST")
	r.HandleFunc("/{uuid}", withAuth(withCampaign(handleGetCampaign))).Methods("GET")
	r.HandleFunc("/{uuid}", withAuth(withCampaign(handleUpdateCampaign))).Methods("PUT")
	r.HandleFunc("/{uuid}", withAuth(withCampaign(handleDeleteCampaign))).Methods("DELETE")

	// Campaign-scoped signons
	r.HandleFunc("/{uuid}/signons", withAuth(withCampaign(handleListSignonsAdmin))).Methods("GET")
	r.HandleFunc("/{uuid}/signons/{id}", withAuth(withCampaign(handleDeleteSignon))).Methods("DELETE")

	// Campaign-scoped config and options
	r.HandleFunc("/{uuid}/config", withAuth(withCampaign(handleGetLocationConfigAdmin))).Methods("GET")
	r.HandleFunc("/{uuid}/config", withAuth(withCampaign(handleUpdateLocationConfig))).Methods("PUT")
	r.HandleFunc("/{uuid}/options", withAuth(withCampaign(handleListLocationOptionsAdmin))).Methods("GET")
	r.HandleFunc("/{uuid}/options", withAuth(withCampaign(handleAddLocationOption))).Methods("POST")
	r.HandleFunc("/{uuid}/options/{id}", withAuth(withCampaign(handleUpdateLocationOption))).Methods("PUT")
	r.HandleFunc("/{uuid}/options/{id}", withAuth(withCampaign(handleDeleteLocationOption))).Methods("DELETE")
}

// withCampaign extracts campaign UUID from path and validates it exists
func withCampaign(next func(w http.ResponseWriter, r *http.Request, campaignID string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		campaignID := vars["uuid"]

		if campaignID == "" {
			writeError(w, http.StatusBadRequest, "Campaign ID required")
			return
		}

		// Verify campaign exists
		_, err := service.GetCampaign(campaignID)
		if err != nil {
			if err == service.ErrCampaignNotFound {
				writeError(w, http.StatusNotFound, "Campaign not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "Failed to verify campaign")
			return
		}

		next(w, r, campaignID)
	}
}

// Campaign handlers
func handleListCampaigns(w http.ResponseWriter, r *http.Request) {
	limit := 100
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	campaigns, err := service.ListCampaigns(limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list campaigns")
		return
	}

	count, err := service.CountCampaigns()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to count campaigns")
		return
	}

	response := map[string]interface{}{
		"campaigns": campaigns,
		"total":     count,
		"limit":     limit,
		"offset":    offset,
	}
	writeData(w, http.StatusOK, response)
}

func handleCreateCampaign(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}

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

func handleGetCampaign(w http.ResponseWriter, r *http.Request, campaignID string) {
	campaign, err := service.GetCampaign(campaignID)
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

func handleGetCampaignPublic(w http.ResponseWriter, r *http.Request, campaignID string) {
	campaign, err := service.GetCampaign(campaignID)
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

func handleUpdateCampaign(w http.ResponseWriter, r *http.Request, campaignID string) {
	var req struct {
		Name            string `json:"name"`
		AllowCustomText *bool  `json:"allow_custom_text"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	campaign, err := service.GetCampaign(campaignID)
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

	err = service.UpdateCampaign(campaignID, name, allowCustomText)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update campaign")
		return
	}

	updated, _ := service.GetCampaign(campaignID)
	writeData(w, http.StatusOK, updated)
}

func handleDeleteCampaign(w http.ResponseWriter, r *http.Request, campaignID string) {
	err := service.DeleteCampaign(campaignID)
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
func handleCreateSignonPublic(w http.ResponseWriter, r *http.Request, campaignID string) {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Location string `json:"location"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	signon, err := service.CreateSignon(campaignID, req.Name, req.Email, req.Location, false)
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

func handleListSignonsPublic(w http.ResponseWriter, r *http.Request, campaignID string) {
	limit := 100
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	signons, err := service.ListSignons(campaignID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list signons")
		return
	}

	count, err := service.CountSignons(campaignID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to count signons")
		return
	}

	response := map[string]interface{}{
		"signons": signons,
		"total":   count,
		"limit":   limit,
		"offset":  offset,
	}
	writeData(w, http.StatusOK, response)
}

func handleListSignonsAdmin(w http.ResponseWriter, r *http.Request, campaignID string) {
	limit := 100
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	signons, err := service.ListSignons(campaignID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list signons")
		return
	}

	count, err := service.CountSignons(campaignID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to count signons")
		return
	}

	response := map[string]interface{}{
		"signons": signons,
		"total":   count,
		"limit":   limit,
		"offset":  offset,
	}
	writeData(w, http.StatusOK, response)
}

func handleDeleteSignon(w http.ResponseWriter, r *http.Request, campaignID string) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid signon ID")
		return
	}

	err = service.DeleteSignon(campaignID, id)
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
func handleGetLocationConfig(w http.ResponseWriter, r *http.Request, campaignID string) {
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

	response := map[string]interface{}{
		"config":  config,
		"options": options,
	}
	writeData(w, http.StatusOK, response)
}

func handleGetLocationConfigAdmin(w http.ResponseWriter, r *http.Request, campaignID string) {
	config, err := service.GetLocationConfig(campaignID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get location config")
		return
	}

	response := map[string]interface{}{
		"config": config,
	}
	writeData(w, http.StatusOK, response)
}

func handleUpdateLocationConfig(w http.ResponseWriter, r *http.Request, campaignID string) {
	var req struct {
		AllowCustomText bool `json:"allow_custom_text"`
	}

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

func handleListLocationOptions(w http.ResponseWriter, r *http.Request, campaignID string) {
	options, err := service.GetLocationOptions(campaignID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get location options")
		return
	}

	writeData(w, http.StatusOK, options)
}

func handleListLocationOptionsAdmin(w http.ResponseWriter, r *http.Request, campaignID string) {
	options, err := service.GetLocationOptions(campaignID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get location options")
		return
	}

	response := map[string]interface{}{
		"options": options,
	}
	writeData(w, http.StatusOK, response)
}

func handleAddLocationOption(w http.ResponseWriter, r *http.Request, campaignID string) {
	var req struct {
		Value        string `json:"value"`
		DisplayOrder int    `json:"display_order"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	id, err := service.AddLocationOption(campaignID, req.Value, req.DisplayOrder)
	if err != nil {
		if err == service.ErrEmptyLocation {
			writeError(w, http.StatusBadRequest, "Location value cannot be empty")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to add location option")
		return
	}

	response := map[string]interface{}{
		"id":             id,
		"value":          req.Value,
		"display_order":  req.DisplayOrder,
	}
	writeData(w, http.StatusCreated, response)
}

func handleUpdateLocationOption(w http.ResponseWriter, r *http.Request, campaignID string) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid option ID")
		return
	}

	var req struct {
		Value        string `json:"value"`
		DisplayOrder int    `json:"display_order"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err = service.UpdateLocationOption(campaignID, id, req.Value, req.DisplayOrder)
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

func handleDeleteLocationOption(w http.ResponseWriter, r *http.Request, campaignID string) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid option ID")
		return
	}

	err = service.DeleteLocationOption(campaignID, id)
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

// Helper to apply both CORS and campaign validation
func withCORSAndCampaign(next func(w http.ResponseWriter, r *http.Request, campaignID string)) http.HandlerFunc {
	return withCORS(withCampaign(next))
}

// Helper to apply both CORS, rate limiting, and campaign validation
func withCORSAndRateLimit(next http.HandlerFunc) http.HandlerFunc {
	return withCORS(withRateLimit(next))
}
