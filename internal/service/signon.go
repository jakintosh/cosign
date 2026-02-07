package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"git.sr.ht/~jakintosh/command-go/pkg/wire"
)

var signonEmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type CreateSignonRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Location string `json:"location"`
}

func (s *Service) CreateSignon(campaignID, name, email, location string) (*Signon, error) {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(email)
	location = strings.TrimSpace(location)

	if name == "" {
		return nil, ErrEmptyName
	}
	if email == "" {
		return nil, ErrEmptyEmail
	}
	if location == "" {
		return nil, ErrEmptyLocation
	}

	if !signonEmailRegex.MatchString(email) {
		return nil, ErrInvalidEmail
	}

	exists, err := s.store.SignonEmailExists(campaignID, email)
	if err != nil {
		return nil, DatabaseError{Err: err}
	}
	if exists {
		return nil, ErrDuplicateEmail
	}

	if err := s.validateSignonLocation(campaignID, location); err != nil {
		return nil, err
	}

	createdAt := s.clock().Unix()
	id, err := s.store.InsertSignon(campaignID, name, email, location, createdAt)
	if err != nil {
		return nil, DatabaseError{Err: err}
	}

	return &Signon{
		ID:        id,
		Name:      name,
		Email:     email,
		Location:  location,
		CreatedAt: createdAt,
	}, nil
}

func (s *Service) GetSignon(campaignID string, id int64) (*Signon, error) {
	signon, err := s.store.GetSignon(campaignID, id)
	if err != nil {
		if errors.Is(err, ErrSignonNotFound) {
			return nil, err
		}
		return nil, DatabaseError{Err: err}
	}
	return signon, nil
}

func (s *Service) ListSignons(campaignID string, limit, offset int) (*Signons, error) {
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	list, err := s.store.ListSignons(campaignID, limit, offset)
	if err != nil {
		return nil, DatabaseError{Err: err}
	}

	total, err := s.store.CountSignons(campaignID)
	if err != nil {
		return nil, DatabaseError{Err: err}
	}

	return &Signons{
		Signons: list,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	}, nil
}

func (s *Service) DeleteSignon(campaignID string, id int64) error {
	err := s.store.DeleteSignon(campaignID, id)
	if err != nil {
		if errors.Is(err, ErrSignonNotFound) {
			return err
		}
		return DatabaseError{Err: err}
	}

	return nil
}

func (s *Service) validateSignonLocation(campaignID, location string) error {
	campaign, err := s.GetCampaign(campaignID)
	if err != nil {
		if errors.Is(err, ErrCampaignNotFound) {
			return ErrCampaignNotFound
		}
		return err
	}

	if campaign.AllowCustomText {
		return nil
	}

	options, err := s.GetCampaignLocations(campaignID)
	if err != nil {
		return err
	}

	if len(options) == 0 {
		return nil
	}

	for _, opt := range options {
		if opt.Value == location {
			return nil
		}
	}

	return ErrLocationNotInOptions
}

func (s *Service) buildPublicSignonRouter(mux *http.ServeMux, mw Middleware) {
	mux.HandleFunc("GET /{campaign_id}/signons", mw.cors(s.handleListSignons))
	mux.HandleFunc("OPTIONS /{campaign_id}/signons", mw.cors(s.handleCreateSignon))
	mux.HandleFunc("POST /{campaign_id}/signons", mw.cors(mw.rateLimit(s.handleCreateSignon)))
}

func (s *Service) buildAdminSignonRouter(mux *http.ServeMux, _ Middleware) {
	mux.HandleFunc("GET /{campaign_id}/signons", s.handleListSignons)
	mux.HandleFunc("DELETE /{campaign_id}/signons/{signon_id}", s.handleDeleteSignon)
}

func (s *Service) handleCreateSignon(w http.ResponseWriter, r *http.Request) {
	campaignID := campaignIDFromPath(r)
	if campaignID == "" {
		wire.WriteError(w, http.StatusBadRequest, "campaign id required")
		return
	}

	var req CreateSignonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		wire.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	signon, err := s.CreateSignon(campaignID, req.Name, req.Email, req.Location)
	if err != nil {
		switch {
		case errors.Is(err, ErrEmptyName), errors.Is(err, ErrEmptyEmail), errors.Is(err, ErrEmptyLocation), errors.Is(err, ErrInvalidEmail), errors.Is(err, ErrLocationNotInOptions):
			wire.WriteError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, ErrCampaignNotFound):
			wire.WriteError(w, http.StatusNotFound, "campaign not found")
		case errors.Is(err, ErrDuplicateEmail):
			wire.WriteError(w, http.StatusConflict, err.Error())
		default:
			wire.WriteError(w, http.StatusInternalServerError, "failed to create sign-on")
		}
		return
	}

	wire.WriteData(w, http.StatusCreated, signon)
}

func (s *Service) handleListSignons(w http.ResponseWriter, r *http.Request) {
	campaignID := campaignIDFromPath(r)
	if campaignID == "" {
		wire.WriteError(w, http.StatusBadRequest, "campaign id required")
		return
	}

	limit, offset, malformed := wire.ParsePagination(r)
	if malformed != nil {
		wire.WriteError(w, http.StatusBadRequest, malformed.Error())
		return
	}

	signons, err := s.ListSignons(campaignID, limit, offset)
	if err != nil {
		wire.WriteError(w, http.StatusInternalServerError, "failed to list sign-ons")
		return
	}

	wire.WriteData(w, http.StatusOK, signons)
}

func (s *Service) handleDeleteSignon(w http.ResponseWriter, r *http.Request) {
	campaignID := campaignIDFromPath(r)
	if campaignID == "" {
		wire.WriteError(w, http.StatusBadRequest, "campaign id required")
		return
	}

	signonID, err := signonIDFromPath(r)
	if err != nil {
		wire.WriteError(w, http.StatusBadRequest, "invalid sign-on id")
		return
	}

	if err := s.DeleteSignon(campaignID, signonID); err != nil {
		switch {
		case errors.Is(err, ErrSignonNotFound):
			wire.WriteError(w, http.StatusNotFound, "sign-on not found")
		default:
			wire.WriteError(w, http.StatusInternalServerError, "failed to delete sign-on")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
