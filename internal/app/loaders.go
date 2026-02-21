package app

import "net/http"

func (s *Server) loadCampaignsPage(
	page int,
	name string,
	formError string,
) CampaignsPageView {
	return CampaignsPageView{
		Campaigns: s.loadCampaignsRegion(page, name, formError),
	}
}

func (s *Server) loadCampaignsRegion(
	page int,
	name string,
	formError string,
) CampaignsRegionView {
	return NewCampaignsRegionView(
		s.loadCampaignsTable(page),
		name,
		formError,
	)
}

func (s *Server) loadCampaignsTable(page int) CampaignsTableView {
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * s.pageSize
	campaigns, err := s.listCampaigns(s.pageSize, offset)
	view := NewCampaignsTableView(campaigns, page, err)

	if view.Pagination.TotalPages > 0 && page > view.Pagination.TotalPages {
		return s.loadCampaignsTable(view.Pagination.TotalPages)
	}

	return view
}

func (s *Server) loadSignaturesTable(campaignID string, page int) SignaturesTableView {
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * s.pageSize
	signatures, err := s.listSignatures(campaignID, s.pageSize, offset)
	view := NewSignaturesTableView(campaignID, signatures, page, err)

	if view.Pagination.TotalPages > 0 && page > view.Pagination.TotalPages {
		return s.loadSignaturesTable(campaignID, view.Pagination.TotalPages)
	}

	return view
}

func (s *Server) loadCampaignDetailPage(
	campaignID string,
	state CampaignDetailPageState,
) (CampaignDetailPageView, int) {
	if state.Signatures.Page < 1 {
		state.Signatures.Page = 1
	}

	campaign, campaignErr := s.getCampaign(campaignID)
	campaignView := NewCampaignPanelView(campaign, campaignErr)
	if campaignErr != nil {
		if isNotFoundError(campaignErr) {
			return CampaignDetailPageView{Campaign: campaignView}, http.StatusNotFound
		}
		return CampaignDetailPageView{Campaign: campaignView}, http.StatusBadGateway
	}

	if state.Campaign.Name != "" {
		campaignView.Name = state.Campaign.Name
	}
	if state.Campaign.FormError != "" {
		campaignView.FormError = state.Campaign.FormError
	}

	locations, locationsErr := s.getCampaignLocations(campaignID)
	locationsView := NewLocationsPanelView(
		campaignID,
		campaign.AllowCustomText,
		locations,
		state.Locations,
		locationsErr,
	)

	sigTable := s.loadSignaturesTable(campaignID, state.Signatures.Page)
	sigPanel := NewSignaturesPanelView(campaignID, sigTable, state.Signatures)

	return CampaignDetailPageView{
		Campaign:   campaignView,
		Locations:  locationsView,
		Signatures: sigPanel,
	}, http.StatusOK
}

func (s *Server) loadLocationsPanel(
	campaignID string,
	state LocationsPanelState,
) (LocationsPanelView, int) {
	campaign, campaignErr := s.getCampaign(campaignID)
	if campaignErr != nil {
		view := NewLocationsPanelView(campaignID, false, nil, state, campaignErr)
		if isNotFoundError(campaignErr) {
			return view, http.StatusNotFound
		}
		return view, http.StatusBadGateway
	}

	locations, err := s.getCampaignLocations(campaignID)
	view := NewLocationsPanelView(
		campaignID,
		campaign.AllowCustomText,
		locations,
		state,
		err,
	)
	if err != nil {
		if isNotFoundError(err) {
			return view, http.StatusNotFound
		}
		return view, http.StatusBadGateway
	}

	if state.Mode == "edit" && (state.EditIndex < 0 || state.EditIndex >= len(locations)) {
		view.FormError = "location not found"
		view.Error = ""
		for idx := range view.Rows {
			view.Rows[idx].IsEditing = false
			view.Rows[idx].EditValue = ""
		}
		return view, http.StatusBadRequest
	}

	if state.FormError != "" {
		return view, http.StatusBadRequest
	}

	return view, http.StatusOK
}
