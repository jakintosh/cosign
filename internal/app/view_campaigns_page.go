package app

import "net/http"

type CampaignsPageView struct {
	Campaigns CampaignsRegionView
}

func (r *Renderer) RenderCampaignsPage(
	w http.ResponseWriter,
	statusCode int,
	view CampaignsPageView,
) {
	r.renderTemplate(w, statusCode, "campaigns_page", view)
}
