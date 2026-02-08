package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"git.sr.ht/~jakintosh/command-go/pkg/wire"
)

var signatureEmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type CreateSignatureRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Location string `json:"location"`
}

func (s *Service) CreateSignature(campaignID, name, email, location string) (*Signature, error) {
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

	if !signatureEmailRegex.MatchString(email) {
		return nil, ErrInvalidEmail
	}

	exists, err := s.store.SignatureEmailExists(campaignID, email)
	if err != nil {
		return nil, DatabaseError{Err: err}
	}
	if exists {
		return nil, ErrDuplicateEmail
	}

	if err := s.validateSignatureLocation(campaignID, location); err != nil {
		return nil, err
	}

	createdAt := s.clock().Unix()
	id, err := s.store.InsertSignature(campaignID, name, email, location, createdAt)
	if err != nil {
		return nil, DatabaseError{Err: err}
	}

	return &Signature{
		ID:        id,
		Name:      name,
		Email:     email,
		Location:  location,
		CreatedAt: createdAt,
	}, nil
}

func (s *Service) GetSignature(campaignID string, id int64) (*Signature, error) {
	signature, err := s.store.GetSignature(campaignID, id)
	if err != nil {
		if errors.Is(err, ErrSignatureNotFound) {
			return nil, err
		}
		return nil, DatabaseError{Err: err}
	}
	return signature, nil
}

func (s *Service) ListSignatures(campaignID string, limit, offset int) (*Signatures, error) {
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	list, err := s.store.ListSignatures(campaignID, limit, offset)
	if err != nil {
		return nil, DatabaseError{Err: err}
	}

	total, err := s.store.CountSignatures(campaignID)
	if err != nil {
		return nil, DatabaseError{Err: err}
	}

	return &Signatures{
		Signatures: list,
		Total:      total,
		Limit:      limit,
		Offset:     offset,
	}, nil
}

func (s *Service) DeleteSignature(campaignID string, id int64) error {
	err := s.store.DeleteSignature(campaignID, id)
	if err != nil {
		if errors.Is(err, ErrSignatureNotFound) {
			return err
		}
		return DatabaseError{Err: err}
	}

	return nil
}

func (s *Service) validateSignatureLocation(campaignID, location string) error {
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

func (s *Service) buildPublicSignatureRouter(mux *http.ServeMux, mw Middleware) {
	mux.HandleFunc("GET /{campaign_id}/signatures", mw.cors(s.handleListSignatures))
	mux.HandleFunc("OPTIONS /{campaign_id}/signatures", mw.cors(s.handleCreateSignature))
	mux.HandleFunc("POST /{campaign_id}/signatures", mw.cors(mw.rateLimit(s.handleCreateSignature)))
}

func (s *Service) buildAdminSignatureRouter(mux *http.ServeMux, _ Middleware) {
	mux.HandleFunc("GET /{campaign_id}/signatures", s.handleListSignatures)
	mux.HandleFunc("DELETE /{campaign_id}/signatures/{signature_id}", s.handleDeleteSignature)
}

func (s *Service) handleCreateSignature(w http.ResponseWriter, r *http.Request) {
	campaignID := campaignIDFromPath(r)
	if campaignID == "" {
		wire.WriteError(w, http.StatusBadRequest, "campaign id required")
		return
	}

	var req CreateSignatureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		wire.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	signature, err := s.CreateSignature(campaignID, req.Name, req.Email, req.Location)
	if err != nil {
		switch {
		case errors.Is(err, ErrEmptyName), errors.Is(err, ErrEmptyEmail), errors.Is(err, ErrEmptyLocation), errors.Is(err, ErrInvalidEmail), errors.Is(err, ErrLocationNotInOptions):
			wire.WriteError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, ErrCampaignNotFound):
			wire.WriteError(w, http.StatusNotFound, "campaign not found")
		case errors.Is(err, ErrDuplicateEmail):
			wire.WriteError(w, http.StatusConflict, err.Error())
		default:
			wire.WriteError(w, http.StatusInternalServerError, "failed to create signature")
		}
		return
	}

	wire.WriteData(w, http.StatusCreated, signature)
}

func (s *Service) handleListSignatures(w http.ResponseWriter, r *http.Request) {
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

	signatures, err := s.ListSignatures(campaignID, limit, offset)
	if err != nil {
		wire.WriteError(w, http.StatusInternalServerError, "failed to list signatures")
		return
	}

	wire.WriteData(w, http.StatusOK, signatures)
}

func (s *Service) handleDeleteSignature(w http.ResponseWriter, r *http.Request) {
	campaignID := campaignIDFromPath(r)
	if campaignID == "" {
		wire.WriteError(w, http.StatusBadRequest, "campaign id required")
		return
	}

	signatureID, err := signatureIDFromPath(r)
	if err != nil {
		wire.WriteError(w, http.StatusBadRequest, "invalid signature id")
		return
	}

	if err := s.DeleteSignature(campaignID, signatureID); err != nil {
		switch {
		case errors.Is(err, ErrSignatureNotFound):
			wire.WriteError(w, http.StatusNotFound, "signature not found")
		default:
			wire.WriteError(w, http.StatusInternalServerError, "failed to delete signature")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
