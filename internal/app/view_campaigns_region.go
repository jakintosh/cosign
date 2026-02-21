package app

import "net/http"

type CampaignsRegionView struct {
	Table      CampaignsTableView
	Name       string
	FormError  string
	CreatePath string
}

func NewCampaignsRegionView(
	table CampaignsTableView,
	name string,
	formError string,
) CampaignsRegionView {
	return CampaignsRegionView{
		Table:      table,
		Name:       name,
		FormError:  formError,
		CreatePath: "/campaigns",
	}
}

func (r *Renderer) RenderCampaignsRegion(
	w http.ResponseWriter,
	statusCode int,
	view CampaignsRegionView,
) {
	r.renderTemplate(w, statusCode, "campaigns_region", view)
}
