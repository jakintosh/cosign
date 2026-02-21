package app

import "net/http"

type SignaturesPanelState struct {
	Page      int
	Name      string
	Email     string
	Location  string
	FormError string
}

type SignaturesPanelView struct {
	CampaignID string
	Name       string
	Email      string
	Location   string
	FormError  string
	CreatePath string
	Table      SignaturesTableView
}

func NewSignaturesPanelView(
	campaignID string,
	table SignaturesTableView,
	state SignaturesPanelState,
) SignaturesPanelView {
	return SignaturesPanelView{
		CampaignID: campaignID,
		Name:       state.Name,
		Email:      state.Email,
		Location:   state.Location,
		FormError:  state.FormError,
		CreatePath: campaignDetailPath(campaignID) + "/signatures",
		Table:      table,
	}
}

func (r *Renderer) RenderSignaturesPanel(
	w http.ResponseWriter,
	statusCode int,
	view SignaturesPanelView,
) {
	r.renderTemplate(w, statusCode, "signatures_panel", view)
}
