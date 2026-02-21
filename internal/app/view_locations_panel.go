package app

import (
	"cosign/internal/service"
	"net/http"
)

type LocationRowView struct {
	Index      int
	Value      string
	IsEditing  bool
	EditValue  string
	UpdatePath string
	DeletePath string
	EditPath   string
	CancelPath string
}

type LocationsPanelState struct {
	Mode       string
	EditIndex  int
	DraftValue string
	FormError  string
}

type LocationsPanelView struct {
	CampaignID         string
	Rows               []LocationRowView
	AllowCustomText    bool
	Error              string
	FormError          string
	NewMode            bool
	NewValue           string
	UpdateSettingsPath string
	CreatePath         string
	NewPath            string
	CancelNewPath      string
}

func NewLocationsPanelView(
	campaignID string,
	allowCustomText bool,
	locations []service.LocationOption,
	state LocationsPanelState,
	err error,
) LocationsPanelView {
	locationsPath := campaignLocationsPath(campaignID)

	view := LocationsPanelView{
		CampaignID:         campaignID,
		AllowCustomText:    allowCustomText,
		FormError:          state.FormError,
		UpdateSettingsPath: locationsPath + "/settings",
		CreatePath:         locationsPath,
		NewPath:            locationsPath + "?mode=new",
		CancelNewPath:      locationsPath,
	}

	if err != nil {
		view.Error = err.Error()
		return view
	}

	rows := make([]LocationRowView, 0, len(locations))
	for idx, loc := range locations {
		rowBasePath := locationsPath + "/" + itoa(idx)
		row := LocationRowView{
			Index:      idx,
			Value:      loc.Value,
			UpdatePath: rowBasePath,
			DeletePath: rowBasePath,
			EditPath:   locationsPath + "?mode=edit&index=" + itoa(idx),
			CancelPath: locationsPath,
		}

		if state.Mode == "edit" && idx == state.EditIndex {
			row.IsEditing = true
			row.EditValue = loc.Value
			if state.DraftValue != "" {
				row.EditValue = state.DraftValue
			}
		}

		rows = append(rows, row)
	}

	view.Rows = rows
	if state.Mode == "new" {
		view.NewMode = true
		view.NewValue = state.DraftValue
	}

	return view
}

func (r *Renderer) RenderLocationsPanel(
	w http.ResponseWriter,
	statusCode int,
	view LocationsPanelView,
) {
	r.renderTemplate(w, statusCode, "locations_panel", view)
}
