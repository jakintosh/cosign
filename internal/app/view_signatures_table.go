package app

import (
	"cosign/internal/service"
	"net/http"
)

type SignatureRowView struct {
	ID         int64
	Name       string
	Email      string
	Location   string
	CreatedAt  string
	DeletePath string
}

type SignaturesTableView struct {
	CampaignID   string
	Signatures   []SignatureRowView
	Pagination   PaginationView
	CurrentPage  int
	Error        string
	PrevPagePath string
	NextPagePath string
}

func NewSignaturesTableView(
	campaignID string,
	response *service.Signatures,
	page int,
	err error,
) SignaturesTableView {
	view := SignaturesTableView{
		CampaignID:  campaignID,
		CurrentPage: page,
	}

	if err != nil {
		view.Error = err.Error()
		return view
	}

	if response == nil {
		view.Error = "failed to load signatures"
		return view
	}

	rows := make([]SignatureRowView, 0, len(response.Signatures))
	for _, signature := range response.Signatures {
		if signature == nil {
			continue
		}

		rows = append(rows, SignatureRowView{
			ID:         signature.ID,
			Name:       signature.Name,
			Email:      signature.Email,
			Location:   signature.Location,
			CreatedAt:  formatUnixTime(signature.CreatedAt),
			DeletePath: campaignDetailPath(campaignID) + "/signatures/" + itoa64(signature.ID),
		})
	}

	view.Signatures = rows
	view.Pagination = NewPaginationView(page, response.Limit, response.Total)

	if view.Pagination.HasPrev {
		view.PrevPagePath = signaturesPagePath(campaignID, view.Pagination.PrevPage)
	}

	if view.Pagination.HasNext {
		view.NextPagePath = signaturesPagePath(campaignID, view.Pagination.NextPage)
	}

	return view
}

func (r *Renderer) RenderSignaturesTable(
	w http.ResponseWriter,
	statusCode int,
	view SignaturesTableView,
) {
	r.renderTemplate(w, statusCode, "signatures_table", view)
}
