package app

import (
	"net/http"
	"strings"
)

func (s *Server) handleCampaigns(w http.ResponseWriter, r *http.Request) {
	ctx := requestContext(r)
	page := parsePageQuery(r, "page")

	if ctx.IsHTMX {
		view := s.loadCampaignsRegion(page, "", "")
		s.renderer.RenderCampaignsRegion(w, http.StatusOK, view)
		return
	}

	view := s.loadCampaignsPage(page, "", "")
	s.renderer.RenderCampaignsPage(w, http.StatusOK, view)
}

func (s *Server) handleCreateCampaign(w http.ResponseWriter, r *http.Request) {
	ctx := requestContext(r)
	page := max(parsePageQuery(r, "page"), 1)

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		s.renderCampaignsError(w, ctx, http.StatusBadRequest, page, "campaign name cannot be empty", name)
		return
	}

	if err := s.createCampaign(name); err != nil {
		s.renderCampaignsError(w, ctx, statusFromError(err), page, err.Error(), name)
		return
	}

	if ctx.IsHTMX {
		view := s.loadCampaignsRegion(1, "", "")
		s.renderer.RenderCampaignsRegion(w, http.StatusOK, view)
		return
	}

	http.Redirect(w, r, "/campaigns", http.StatusSeeOther)
}

func (s *Server) handleDeleteCampaign(w http.ResponseWriter, r *http.Request) {
	ctx := requestContext(r)
	page := max(parsePageQuery(r, "page"), 1)

	campaignID := campaignIDFromPath(r)
	if campaignID == "" {
		s.renderCampaignsError(w, ctx, http.StatusBadRequest, page, "campaign id required", "")
		return
	}

	if err := s.deleteCampaign(campaignID); err != nil {
		s.renderCampaignsError(w, ctx, statusFromError(err), page, err.Error(), "")
		return
	}

	if ctx.IsHTMX {
		view := s.loadCampaignsRegion(page, "", "")
		s.renderer.RenderCampaignsRegion(w, http.StatusOK, view)
		return
	}

	http.Redirect(w, r, "/campaigns", http.StatusSeeOther)
}

func (s *Server) renderCampaignsError(
	w http.ResponseWriter,
	ctx RequestContext,
	statusCode int,
	page int,
	formError string,
	name string,
) {
	if ctx.IsHTMX {
		view := s.loadCampaignsRegion(page, name, formError)
		s.renderer.RenderCampaignsRegion(w, statusCode, view)
		return
	}

	view := s.loadCampaignsPage(page, name, formError)
	s.renderer.RenderCampaignsPage(w, statusCode, view)
}
