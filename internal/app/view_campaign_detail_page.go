package app

import "net/http"

type CampaignDetailPageState struct {
	Campaign   CampaignPanelState
	Locations  LocationsPanelState
	Signatures SignaturesPanelState
}

type CampaignDetailPageView struct {
	Campaign   CampaignPanelView
	Locations  LocationsPanelView
	Signatures SignaturesPanelView
}

func (r *Renderer) RenderCampaignDetailPage(
	w http.ResponseWriter,
	statusCode int,
	view CampaignDetailPageView,
) {
	r.renderTemplate(w, statusCode, "campaign_detail_page", view)
}
