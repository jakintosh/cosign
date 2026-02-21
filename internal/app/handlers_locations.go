package app

import (
	"cosign/internal/service"
	"net/http"
	"strings"
)

func (s *Server) handleLocations(w http.ResponseWriter, r *http.Request) {
	ctx := requestContext(r)
	campaignID := campaignIDFromPath(r)
	if campaignID == "" {
		http.NotFound(w, r)
		return
	}

	mode, editIndex := parseLocationsMode(r, "mode", "index")
	state := LocationsPanelState{
		Mode:      mode,
		EditIndex: editIndex,
	}

	if ctx.IsHTMX {
		panel, status := s.loadLocationsPanel(campaignID, state)
		s.renderer.RenderLocationsPanel(w, status, panel)
		return
	}

	view, status := s.loadCampaignDetailPage(campaignID, CampaignDetailPageState{
		Locations: state,
		Signatures: SignaturesPanelState{
			Page: parsePageQuery(r, "page"),
		},
	})
	s.renderer.RenderCampaignDetailPage(w, status, view)
}

func (s *Server) handleCreateLocation(w http.ResponseWriter, r *http.Request) {
	ctx := requestContext(r)
	campaignID := campaignIDFromPath(r)
	if campaignID == "" {
		http.NotFound(w, r)
		return
	}

	value := strings.TrimSpace(r.FormValue("value"))
	if value == "" {
		s.renderLocationsError(w, r, ctx.IsHTMX, http.StatusBadRequest, campaignID, LocationsPanelState{
			Mode:       "new",
			EditIndex:  -1,
			DraftValue: value,
			FormError:  "location cannot be empty",
		})
		return
	}

	locations, err := s.getCampaignLocations(campaignID)
	if err != nil {
		s.renderLocationsError(w, r, ctx.IsHTMX, statusFromError(err), campaignID, LocationsPanelState{
			Mode:       "new",
			EditIndex:  -1,
			DraftValue: value,
			FormError:  err.Error(),
		})
		return
	}

	locations = append(locations, service.LocationOption{Value: value})
	if err := s.setCampaignLocations(campaignID, locations); err != nil {
		s.renderLocationsError(w, r, ctx.IsHTMX, statusFromError(err), campaignID, LocationsPanelState{
			Mode:       "new",
			EditIndex:  -1,
			DraftValue: value,
			FormError:  err.Error(),
		})
		return
	}

	if ctx.IsHTMX {
		panel, status := s.loadLocationsPanel(campaignID, LocationsPanelState{EditIndex: -1})
		s.renderer.RenderLocationsPanel(w, status, panel)
		return
	}

	http.Redirect(w, r, campaignLocationsPath(campaignID), http.StatusSeeOther)
}

func (s *Server) handleUpdateLocation(w http.ResponseWriter, r *http.Request) {
	ctx := requestContext(r)
	campaignID := campaignIDFromPath(r)
	if campaignID == "" {
		http.NotFound(w, r)
		return
	}

	index, err := parsePathIndex(r, "index")
	if err != nil {
		s.renderLocationsError(w, r, ctx.IsHTMX, http.StatusBadRequest, campaignID, LocationsPanelState{
			EditIndex: -1,
			FormError: "invalid location index",
		})
		return
	}

	value := strings.TrimSpace(r.FormValue("value"))
	if value == "" {
		s.renderLocationsError(w, r, ctx.IsHTMX, http.StatusBadRequest, campaignID, LocationsPanelState{
			Mode:       "edit",
			EditIndex:  index,
			DraftValue: value,
			FormError:  "location cannot be empty",
		})
		return
	}

	locations, err := s.getCampaignLocations(campaignID)
	if err != nil {
		s.renderLocationsError(w, r, ctx.IsHTMX, statusFromError(err), campaignID, LocationsPanelState{
			Mode:       "edit",
			EditIndex:  index,
			DraftValue: value,
			FormError:  err.Error(),
		})
		return
	}

	if index < 0 || index >= len(locations) {
		s.renderLocationsError(w, r, ctx.IsHTMX, http.StatusBadRequest, campaignID, LocationsPanelState{
			EditIndex: -1,
			FormError: "location not found",
		})
		return
	}

	locations[index].Value = value
	if err := s.setCampaignLocations(campaignID, locations); err != nil {
		s.renderLocationsError(w, r, ctx.IsHTMX, statusFromError(err), campaignID, LocationsPanelState{
			Mode:       "edit",
			EditIndex:  index,
			DraftValue: value,
			FormError:  err.Error(),
		})
		return
	}

	if ctx.IsHTMX {
		panel, status := s.loadLocationsPanel(campaignID, LocationsPanelState{EditIndex: -1})
		s.renderer.RenderLocationsPanel(w, status, panel)
		return
	}

	http.Redirect(w, r, campaignLocationsPath(campaignID), http.StatusSeeOther)
}

func (s *Server) handleDeleteLocation(w http.ResponseWriter, r *http.Request) {
	ctx := requestContext(r)
	campaignID := campaignIDFromPath(r)
	if campaignID == "" {
		http.NotFound(w, r)
		return
	}

	index, err := parsePathIndex(r, "index")
	if err != nil {
		s.renderLocationsError(w, r, ctx.IsHTMX, http.StatusBadRequest, campaignID, LocationsPanelState{
			EditIndex: -1,
			FormError: "invalid location index",
		})
		return
	}

	locations, err := s.getCampaignLocations(campaignID)
	if err != nil {
		s.renderLocationsError(w, r, ctx.IsHTMX, statusFromError(err), campaignID, LocationsPanelState{
			EditIndex: -1,
			FormError: err.Error(),
		})
		return
	}

	if index < 0 || index >= len(locations) {
		s.renderLocationsError(w, r, ctx.IsHTMX, http.StatusBadRequest, campaignID, LocationsPanelState{
			EditIndex: -1,
			FormError: "location not found",
		})
		return
	}

	updated := make([]service.LocationOption, 0, len(locations)-1)
	updated = append(updated, locations[:index]...)
	updated = append(updated, locations[index+1:]...)

	if err := s.setCampaignLocations(campaignID, updated); err != nil {
		s.renderLocationsError(w, r, ctx.IsHTMX, statusFromError(err), campaignID, LocationsPanelState{
			EditIndex: -1,
			FormError: err.Error(),
		})
		return
	}

	if ctx.IsHTMX {
		panel, status := s.loadLocationsPanel(campaignID, LocationsPanelState{EditIndex: -1})
		s.renderer.RenderLocationsPanel(w, status, panel)
		return
	}

	http.Redirect(w, r, campaignLocationsPath(campaignID), http.StatusSeeOther)
}

func (s *Server) handleUpdateLocationsSettings(w http.ResponseWriter, r *http.Request) {
	ctx := requestContext(r)
	campaignID := campaignIDFromPath(r)
	if campaignID == "" {
		http.NotFound(w, r)
		return
	}

	allowCustomText := r.FormValue("allow_custom_text") == "on"

	campaign, err := s.getCampaign(campaignID)
	if err != nil {
		s.renderLocationsError(w, r, ctx.IsHTMX, statusFromError(err), campaignID, LocationsPanelState{
			EditIndex: -1,
			FormError: err.Error(),
		})
		return
	}

	if err := s.updateCampaign(campaignID, campaign.Name, allowCustomText); err != nil {
		s.renderLocationsError(w, r, ctx.IsHTMX, statusFromError(err), campaignID, LocationsPanelState{
			EditIndex: -1,
			FormError: err.Error(),
		})
		return
	}

	if ctx.IsHTMX {
		panel, status := s.loadLocationsPanel(campaignID, LocationsPanelState{EditIndex: -1})
		s.renderer.RenderLocationsPanel(w, status, panel)
		return
	}

	http.Redirect(w, r, campaignLocationsPath(campaignID), http.StatusSeeOther)
}

func (s *Server) renderLocationsError(
	w http.ResponseWriter,
	r *http.Request,
	isHTMX bool,
	statusCode int,
	campaignID string,
	state LocationsPanelState,
) {
	if isHTMX {
		panel, status := s.loadLocationsPanel(campaignID, state)
		if status == http.StatusOK {
			status = statusCode
		}
		s.renderer.RenderLocationsPanel(w, status, panel)
		return
	}

	detailState := CampaignDetailPageState{
		Locations: state,
		Signatures: SignaturesPanelState{
			Page: parsePageQuery(r, "page"),
		},
	}

	view, status := s.loadCampaignDetailPage(campaignID, detailState)
	if status == http.StatusOK {
		status = statusCode
	}

	s.renderer.RenderCampaignDetailPage(w, status, view)
}
