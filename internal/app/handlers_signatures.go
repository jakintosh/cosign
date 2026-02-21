package app

import (
	"net/http"
	"strconv"
	"strings"
)

func (s *Server) handleSignatures(w http.ResponseWriter, r *http.Request) {
	ctx := requestContext(r)
	campaignID := campaignIDFromPath(r)
	if campaignID == "" {
		http.NotFound(w, r)
		return
	}

	page := parsePageQuery(r, "page")
	if ctx.IsHTMX {
		panel := NewSignaturesPanelView(
			campaignID,
			s.loadSignaturesTable(campaignID, page),
			SignaturesPanelState{Page: page},
		)
		s.renderer.RenderSignaturesPanel(w, http.StatusOK, panel)
		return
	}

	locMode, locIndex := parseLocationsMode(r, "mode", "index")
	view, status := s.loadCampaignDetailPage(campaignID, CampaignDetailPageState{
		Locations: LocationsPanelState{
			Mode:      locMode,
			EditIndex: locIndex,
		},
		Signatures: SignaturesPanelState{Page: page},
	})
	s.renderer.RenderCampaignDetailPage(w, status, view)
}

func (s *Server) handleCreateSignature(w http.ResponseWriter, r *http.Request) {
	ctx := requestContext(r)
	campaignID := campaignIDFromPath(r)
	if campaignID == "" {
		http.NotFound(w, r)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.TrimSpace(r.FormValue("email"))
	location := strings.TrimSpace(r.FormValue("location"))

	state := SignaturesPanelState{
		Page:      1,
		Name:      name,
		Email:     email,
		Location:  location,
		FormError: "",
	}

	if name == "" || email == "" || location == "" {
		state.FormError = "name, email, and location are required"
		s.renderSignaturesError(w, r, ctx.IsHTMX, http.StatusBadRequest, campaignID, state)
		return
	}

	if err := s.createSignature(campaignID, name, email, location); err != nil {
		state.FormError = err.Error()
		s.renderSignaturesError(w, r, ctx.IsHTMX, statusFromError(err), campaignID, state)
		return
	}

	if ctx.IsHTMX {
		panel := NewSignaturesPanelView(
			campaignID,
			s.loadSignaturesTable(campaignID, 1),
			SignaturesPanelState{Page: 1},
		)
		s.renderer.RenderSignaturesPanel(w, http.StatusOK, panel)
		return
	}

	http.Redirect(w, r, signaturesPagePath(campaignID, 1), http.StatusSeeOther)
}

func (s *Server) handleDeleteSignature(w http.ResponseWriter, r *http.Request) {
	ctx := requestContext(r)
	campaignID := campaignIDFromPath(r)
	if campaignID == "" {
		http.NotFound(w, r)
		return
	}

	signatureID, err := strconv.ParseInt(strings.TrimSpace(r.PathValue("signature_id")), 10, 64)
	if err != nil {
		s.renderSignaturesError(w, r, ctx.IsHTMX, http.StatusBadRequest, campaignID, SignaturesPanelState{
			Page:      1,
			FormError: "invalid signature id",
		})
		return
	}

	page := max(parsePageQuery(r, "page"), 1)

	if err := s.deleteSignature(campaignID, signatureID); err != nil {
		s.renderSignaturesError(w, r, ctx.IsHTMX, statusFromError(err), campaignID, SignaturesPanelState{
			Page:      page,
			FormError: err.Error(),
		})
		return
	}

	if ctx.IsHTMX {
		panel := NewSignaturesPanelView(
			campaignID,
			s.loadSignaturesTable(campaignID, page),
			SignaturesPanelState{Page: page},
		)
		s.renderer.RenderSignaturesPanel(w, http.StatusOK, panel)
		return
	}

	http.Redirect(w, r, signaturesPagePath(campaignID, page), http.StatusSeeOther)
}

func (s *Server) renderSignaturesError(
	w http.ResponseWriter,
	r *http.Request,
	isHTMX bool,
	statusCode int,
	campaignID string,
	state SignaturesPanelState,
) {
	page := max(state.Page, 1)
	state.Page = page

	if isHTMX {
		panel := NewSignaturesPanelView(
			campaignID,
			s.loadSignaturesTable(campaignID, page),
			state,
		)
		s.renderer.RenderSignaturesPanel(w, http.StatusOK, panel)
		return
	}

	locMode, locIndex := parseLocationsMode(r, "mode", "index")
	detailState := CampaignDetailPageState{
		Locations: LocationsPanelState{
			Mode:      locMode,
			EditIndex: locIndex,
		},
		Signatures: state,
	}

	view, status := s.loadCampaignDetailPage(campaignID, detailState)
	if status == http.StatusOK {
		status = statusCode
	}

	s.renderer.RenderCampaignDetailPage(w, status, view)
}
