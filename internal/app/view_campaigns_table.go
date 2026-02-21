package app

import (
	"cosign/internal/service"
	"net/http"
)

type CampaignRowView struct {
	ID              string
	Name            string
	AllowCustomText bool
	CreatedAt       string
	DetailPath      string
	DeletePath      string
}

type CampaignsTableView struct {
	Campaigns    []CampaignRowView
	Pagination   PaginationView
	CurrentPage  int
	Error        string
	PrevPagePath string
	NextPagePath string
}

func NewCampaignsTableView(
	response *service.Campaigns,
	page int,
	err error,
) CampaignsTableView {
	view := CampaignsTableView{CurrentPage: page}

	if err != nil {
		view.Error = err.Error()
		return view
	}

	if response == nil {
		view.Error = "failed to load campaigns"
		return view
	}

	rows := make([]CampaignRowView, 0, len(response.Campaigns))
	for _, campaign := range response.Campaigns {
		if campaign == nil {
			continue
		}

		rows = append(rows, CampaignRowView{
			ID:              campaign.ID,
			Name:            campaign.Name,
			AllowCustomText: campaign.AllowCustomText,
			CreatedAt:       formatUnixTime(campaign.CreatedAt),
			DetailPath:      campaignDetailPath(campaign.ID),
			DeletePath:      campaignDetailPath(campaign.ID),
		})
	}

	view.Campaigns = rows
	view.Pagination = NewPaginationView(page, response.Limit, response.Total)

	if view.Pagination.HasPrev {
		view.PrevPagePath = campaignsPagePath(view.Pagination.PrevPage)
	}

	if view.Pagination.HasNext {
		view.NextPagePath = campaignsPagePath(view.Pagination.NextPage)
	}

	return view
}

func (r *Renderer) RenderCampaignsTable(
	w http.ResponseWriter,
	statusCode int,
	view CampaignsTableView,
) {
	r.renderTemplate(w, statusCode, "campaigns_table", view)
}
