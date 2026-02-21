package app

import (
	"net/http"
	"strings"
)

func (s *Server) handleCampaignDetailPage(w http.ResponseWriter, r *http.Request) {
	campaignID := campaignIDFromPath(r)
	if campaignID == "" {
		http.NotFound(w, r)
		return
	}

	locMode, locIndex := parseLocationsMode(r, "mode", "index")
	state := CampaignDetailPageState{
		Locations: LocationsPanelState{
			Mode:      locMode,
			EditIndex: locIndex,
		},
		Signatures: SignaturesPanelState{
			Page: parsePageQuery(r, "page"),
		},
	}

	view, status := s.loadCampaignDetailPage(campaignID, state)
	s.renderer.RenderCampaignDetailPage(w, status, view)
}

func (s *Server) handleUpdateCampaign(w http.ResponseWriter, r *http.Request) {
	ctx := requestContext(r)
	campaignID := campaignIDFromPath(r)
	if campaignID == "" {
		http.NotFound(w, r)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	existing, err := s.getCampaign(campaignID)
	if err != nil {
		s.renderCampaignUpdateError(w, r, ctx, statusFromError(err), campaignID, name, err.Error())
		return
	}

	if name == "" {
		s.renderCampaignUpdateError(w, r, ctx, http.StatusBadRequest, campaignID, name, "campaign name cannot be empty")
		return
	}

	if err := s.updateCampaign(campaignID, name, existing.AllowCustomText); err != nil {
		s.renderCampaignUpdateError(w, r, ctx, statusFromError(err), campaignID, name, err.Error())
		return
	}

	if ctx.IsHTMX {
		campaign, err := s.getCampaign(campaignID)
		panel := NewCampaignPanelView(campaign, err)
		status := http.StatusOK
		if err != nil {
			status = statusFromError(err)
		}
		s.renderer.RenderCampaignPanel(w, status, panel)
		return
	}

	http.Redirect(w, r, campaignDetailPath(campaignID), http.StatusSeeOther)
}

func (s *Server) renderCampaignUpdateError(
	w http.ResponseWriter,
	r *http.Request,
	ctx RequestContext,
	statusCode int,
	campaignID string,
	name string,
	formError string,
) {
	if ctx.IsHTMX {
		campaign, err := s.getCampaign(campaignID)
		panel := NewCampaignPanelView(campaign, err)
		if name != "" {
			panel.Name = name
		}
		panel.FormError = formError
		s.renderer.RenderCampaignPanel(w, statusCode, panel)
		return
	}

	locMode, locIndex := parseLocationsMode(r, "mode", "index")
	state := CampaignDetailPageState{
		Campaign: CampaignPanelState{
			Name:      name,
			FormError: formError,
		},
		Locations: LocationsPanelState{
			Mode:      locMode,
			EditIndex: locIndex,
		},
		Signatures: SignaturesPanelState{
			Page: parsePageQuery(r, "page"),
		},
	}

	view, status := s.loadCampaignDetailPage(campaignID, state)
	if status == http.StatusOK {
		status = statusCode
	}

	s.renderer.RenderCampaignDetailPage(w, status, view)
}
