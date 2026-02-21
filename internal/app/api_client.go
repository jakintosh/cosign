package app

import (
	"cosign/internal/service"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

func (s *Server) createCampaign(name string) error {
	body, err := json.Marshal(service.CreateCampaignRequest{Name: name})
	if err != nil {
		return err
	}

	var response service.Campaign
	return s.client.Post("/admin/campaigns", body, &response)
}

func (s *Server) listCampaigns(limit int, offset int) (*service.Campaigns, error) {
	var response service.Campaigns
	path := fmt.Sprintf("/admin/campaigns?limit=%d&offset=%d", limit, offset)
	if err := s.client.Get(path, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *Server) getCampaign(campaignID string) (*service.Campaign, error) {
	var response service.Campaign
	path := "/admin/campaigns/" + url.PathEscape(campaignID)
	if err := s.client.Get(path, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *Server) updateCampaign(campaignID string, name string, allowCustomText bool) error {
	body, err := json.Marshal(service.UpdateCampaignRequest{
		Name:            name,
		AllowCustomText: &allowCustomText,
	})
	if err != nil {
		return err
	}

	var response service.Campaign
	path := "/admin/campaigns/" + url.PathEscape(campaignID)
	return s.client.Put(path, body, &response)
}

func (s *Server) deleteCampaign(campaignID string) error {
	path := "/admin/campaigns/" + url.PathEscape(campaignID)
	return s.client.Delete(path, nil)
}

func (s *Server) getCampaignLocations(campaignID string) ([]service.LocationOption, error) {
	var response service.CampaignLocationsResponse
	path := "/admin/campaigns/" + url.PathEscape(campaignID) + "/locations"
	if err := s.client.Get(path, &response); err != nil {
		return nil, err
	}

	return response.Locations, nil
}

func (s *Server) setCampaignLocations(campaignID string, locations []service.LocationOption) error {
	normalized := make([]service.LocationOption, 0, len(locations))
	for idx, loc := range locations {
		value := strings.TrimSpace(loc.Value)
		if value == "" {
			continue
		}

		normalized = append(normalized, service.LocationOption{
			Value:        value,
			DisplayOrder: idx + 1,
		})
	}

	body, err := json.Marshal(service.CampaignLocationsRequest{Locations: normalized})
	if err != nil {
		return err
	}

	var response service.CampaignLocationsResponse
	path := "/admin/campaigns/" + url.PathEscape(campaignID) + "/locations"
	return s.client.Put(path, body, &response)
}

func (s *Server) listSignatures(campaignID string, limit int, offset int) (*service.Signatures, error) {
	var response service.Signatures
	path := fmt.Sprintf(
		"/admin/campaigns/%s/signatures?limit=%d&offset=%d",
		url.PathEscape(campaignID),
		limit,
		offset,
	)
	if err := s.client.Get(path, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (s *Server) createSignature(campaignID, name, email, location string) error {
	body, err := json.Marshal(service.CreateSignatureRequest{
		Name:     name,
		Email:    email,
		Location: location,
	})
	if err != nil {
		return err
	}

	var response service.Signature
	path := "/campaigns/" + url.PathEscape(campaignID) + "/signatures"
	return s.client.Post(path, body, &response)
}

func (s *Server) deleteSignature(campaignID string, signatureID int64) error {
	path := fmt.Sprintf("/admin/campaigns/%s/signatures/%d", url.PathEscape(campaignID), signatureID)
	return s.client.Delete(path, nil)
}

func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	return errors.Is(err, service.ErrCampaignNotFound) || strings.Contains(msg, "not found")
}
