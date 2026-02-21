package app

import (
	"cosign/internal/service"
	"net/http"
)

type CampaignPanelView struct {
	ID              string
	Name            string
	AllowCustomText bool
	CreatedAt       string
	FormError       string
	UpdatePath      string
	DeletePath      string
}

type CampaignPanelState struct {
	Name      string
	FormError string
}

func NewCampaignPanelView(
	campaign *service.Campaign,
	err error,
) CampaignPanelView {
	if err != nil {
		return CampaignPanelView{FormError: err.Error()}
	}

	if campaign == nil {
		return CampaignPanelView{FormError: "campaign not found"}
	}

	path := campaignDetailPath(campaign.ID)

	return CampaignPanelView{
		ID:              campaign.ID,
		Name:            campaign.Name,
		AllowCustomText: campaign.AllowCustomText,
		CreatedAt:       formatUnixTime(campaign.CreatedAt),
		UpdatePath:      path,
		DeletePath:      path,
	}
}

func (r *Renderer) RenderCampaignPanel(
	w http.ResponseWriter,
	statusCode int,
	view CampaignPanelView,
) {
	r.renderTemplate(w, statusCode, "campaign_panel", view)
}
