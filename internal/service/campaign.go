package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"git.sr.ht/~jakintosh/command-go/pkg/wire"
)

type CreateCampaignRequest struct {
	Name string `json:"name"`
}

type UpdateCampaignRequest struct {
	Name            string `json:"name"`
	AllowCustomText *bool  `json:"allow_custom_text"`
}

type CampaignLocationsRequest struct {
	Locations []LocationOption `json:"locations"`
}

type CampaignLocationsResponse struct {
	Locations []LocationOption `json:"locations"`
}

func (s *Service) buildPublicCampaignRouter(mux *http.ServeMux, mw Middleware) {
	mux.HandleFunc("GET /{campaign_id}", mw.cors(s.handleGetCampaign))
	mux.HandleFunc("OPTIONS /{campaign_id}", mw.cors(s.handleGetCampaign))
	mux.HandleFunc("GET /{campaign_id}/locations", mw.cors(s.handleGetCampaignLocations))
	mux.HandleFunc("OPTIONS /{campaign_id}/locations", mw.cors(s.handleGetCampaignLocations))
}

func (s *Service) buildAdminCampaignRouter(mux *http.ServeMux, _ Middleware) {
	mux.HandleFunc("GET /{$}", s.handleListCampaigns)
	mux.HandleFunc("POST /{$}", s.handleCreateCampaign)
	mux.HandleFunc("GET /{campaign_id}", s.handleGetCampaign)
	mux.HandleFunc("PUT /{campaign_id}", s.handleUpdateCampaign)
	mux.HandleFunc("DELETE /{campaign_id}", s.handleDeleteCampaign)
	mux.HandleFunc("GET /{campaign_id}/locations", s.handleGetCampaignLocations)
	mux.HandleFunc("PUT /{campaign_id}/locations", s.handleUpdateCampaignLocations)
}

func (s *Service) CreateCampaign(name string) (*Campaign, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrEmptyCampaignName
	}

	id, err := randomID(16)
	if err != nil {
		return nil, err
	}

	createdAt := s.clock().Unix()
	err = s.store.InsertCampaign(id, name, true, createdAt)
	if err != nil {
		return nil, DatabaseError{Err: err}
	}

	return &Campaign{
		ID:              id,
		Name:            name,
		AllowCustomText: true,
		CreatedAt:       createdAt,
	}, nil
}

func (s *Service) GetCampaign(id string) (*Campaign, error) {
	campaign, err := s.store.GetCampaign(id)
	if err != nil {
		if errors.Is(err, ErrCampaignNotFound) {
			return nil, err
		}
		return nil, DatabaseError{Err: err}
	}
	return campaign, nil
}

func (s *Service) ListCampaigns(limit, offset int) (*Campaigns, error) {
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	campaigns, err := s.store.ListCampaigns(limit, offset)
	if err != nil {
		return nil, DatabaseError{Err: err}
	}

	total, err := s.store.CountCampaigns()
	if err != nil {
		return nil, DatabaseError{Err: err}
	}

	return &Campaigns{
		Campaigns: campaigns,
		Total:     total,
		Limit:     limit,
		Offset:    offset,
	}, nil
}

func (s *Service) UpdateCampaign(id, name string, allowCustomText bool) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrEmptyCampaignName
	}

	err := s.store.UpdateCampaign(id, name, allowCustomText)
	if err != nil {
		if errors.Is(err, ErrCampaignNotFound) {
			return err
		}
		return DatabaseError{Err: err}
	}

	return nil
}

func (s *Service) DeleteCampaign(id string) error {
	err := s.store.DeleteCampaign(id)
	if err != nil {
		if errors.Is(err, ErrCampaignNotFound) {
			return err
		}
		return DatabaseError{Err: err}
	}

	return nil
}

func (s *Service) GetCampaignLocations(campaignID string) ([]LocationOption, error) {
	options, err := s.store.GetCampaignLocations(campaignID)
	if err != nil {
		if errors.Is(err, ErrCampaignNotFound) {
			return nil, err
		}
		return nil, DatabaseError{Err: err}
	}

	locations := make([]LocationOption, 0, len(options))
	for _, opt := range options {
		if opt == nil {
			continue
		}
		locations = append(locations, *opt)
	}

	return locations, nil
}

func (s *Service) SetCampaignLocations(campaignID string, options []LocationOption) error {
	normalized := make([]LocationOption, 0, len(options))
	for idx, loc := range options {
		value := strings.TrimSpace(loc.Value)
		if value == "" {
			return ErrEmptyLocation
		}

		displayOrder := loc.DisplayOrder
		if displayOrder <= 0 {
			displayOrder = idx + 1
		}

		normalized = append(normalized, LocationOption{
			Value:        value,
			DisplayOrder: displayOrder,
		})
	}

	err := s.store.ReplaceCampaignLocations(campaignID, normalized)
	if err != nil {
		if errors.Is(err, ErrCampaignNotFound) {
			return err
		}
		return DatabaseError{Err: err}
	}

	return nil
}

func (s *Service) handleCreateCampaign(w http.ResponseWriter, r *http.Request) {
	var req CreateCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		wire.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	campaign, err := s.CreateCampaign(req.Name)
	if err != nil {
		switch {
		case errors.Is(err, ErrEmptyCampaignName):
			wire.WriteError(w, http.StatusBadRequest, err.Error())
		default:
			wire.WriteError(w, http.StatusInternalServerError, "failed to create campaign")
		}
		return
	}

	wire.WriteData(w, http.StatusCreated, campaign)
}

func (s *Service) handleListCampaigns(w http.ResponseWriter, r *http.Request) {
	limit, offset, malformed := wire.ParsePagination(r)
	if malformed != nil {
		wire.WriteError(w, http.StatusBadRequest, malformed.Error())
		return
	}

	campaigns, err := s.ListCampaigns(limit, offset)
	if err != nil {
		wire.WriteError(w, http.StatusInternalServerError, "failed to list campaigns")
		return
	}

	wire.WriteData(w, http.StatusOK, campaigns)
}

func (s *Service) handleGetCampaign(w http.ResponseWriter, r *http.Request) {
	campaignID := campaignIDFromPath(r)
	if campaignID == "" {
		wire.WriteError(w, http.StatusBadRequest, "campaign id required")
		return
	}

	campaign, err := s.GetCampaign(campaignID)
	if err != nil {
		switch {
		case errors.Is(err, ErrCampaignNotFound):
			wire.WriteError(w, http.StatusNotFound, "campaign not found")
		default:
			wire.WriteError(w, http.StatusInternalServerError, "failed to load campaign")
		}
		return
	}

	wire.WriteData(w, http.StatusOK, campaign)
}

func (s *Service) handleUpdateCampaign(w http.ResponseWriter, r *http.Request) {
	campaignID := campaignIDFromPath(r)
	if campaignID == "" {
		wire.WriteError(w, http.StatusBadRequest, "campaign id required")
		return
	}

	var req UpdateCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		wire.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	existing, err := s.GetCampaign(campaignID)
	if err != nil {
		switch {
		case errors.Is(err, ErrCampaignNotFound):
			wire.WriteError(w, http.StatusNotFound, "campaign not found")
		default:
			wire.WriteError(w, http.StatusInternalServerError, "failed to load campaign")
		}
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = existing.Name
	}

	allowCustomText := existing.AllowCustomText
	if req.AllowCustomText != nil {
		allowCustomText = *req.AllowCustomText
	}

	if err := s.UpdateCampaign(campaignID, name, allowCustomText); err != nil {
		switch {
		case errors.Is(err, ErrCampaignNotFound):
			wire.WriteError(w, http.StatusNotFound, "campaign not found")
		case errors.Is(err, ErrEmptyCampaignName):
			wire.WriteError(w, http.StatusBadRequest, err.Error())
		default:
			wire.WriteError(w, http.StatusInternalServerError, "failed to update campaign")
		}
		return
	}

	updated, err := s.GetCampaign(campaignID)
	if err != nil {
		wire.WriteError(w, http.StatusInternalServerError, "failed to reload campaign")
		return
	}

	wire.WriteData(w, http.StatusOK, updated)
}

func (s *Service) handleDeleteCampaign(w http.ResponseWriter, r *http.Request) {
	campaignID := campaignIDFromPath(r)
	if campaignID == "" {
		wire.WriteError(w, http.StatusBadRequest, "campaign id required")
		return
	}

	if err := s.DeleteCampaign(campaignID); err != nil {
		switch {
		case errors.Is(err, ErrCampaignNotFound):
			wire.WriteError(w, http.StatusNotFound, "campaign not found")
		default:
			wire.WriteError(w, http.StatusInternalServerError, "failed to delete campaign")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Service) handleGetCampaignLocations(w http.ResponseWriter, r *http.Request) {
	campaignID := campaignIDFromPath(r)
	if campaignID == "" {
		wire.WriteError(w, http.StatusBadRequest, "campaign id required")
		return
	}

	locations, err := s.GetCampaignLocations(campaignID)
	if err != nil {
		switch {
		case errors.Is(err, ErrCampaignNotFound):
			wire.WriteError(w, http.StatusNotFound, "campaign not found")
		default:
			wire.WriteError(w, http.StatusInternalServerError, "failed to get campaign locations")
		}
		return
	}

	wire.WriteData(w, http.StatusOK, CampaignLocationsResponse{Locations: locations})
}

func (s *Service) handleUpdateCampaignLocations(w http.ResponseWriter, r *http.Request) {
	campaignID := campaignIDFromPath(r)
	if campaignID == "" {
		wire.WriteError(w, http.StatusBadRequest, "campaign id required")
		return
	}

	var req CampaignLocationsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		wire.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := s.SetCampaignLocations(campaignID, req.Locations); err != nil {
		switch {
		case errors.Is(err, ErrCampaignNotFound):
			wire.WriteError(w, http.StatusNotFound, "campaign not found")
		case errors.Is(err, ErrEmptyLocation):
			wire.WriteError(w, http.StatusBadRequest, err.Error())
		default:
			wire.WriteError(w, http.StatusInternalServerError, "failed to update campaign locations")
		}
		return
	}

	updated, err := s.GetCampaignLocations(campaignID)
	if err != nil {
		wire.WriteError(w, http.StatusInternalServerError, "failed to load updated campaign locations")
		return
	}

	wire.WriteData(w, http.StatusOK, CampaignLocationsResponse{Locations: updated})
}
