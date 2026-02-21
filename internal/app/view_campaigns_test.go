package app

import (
	"cosign/internal/service"
	"testing"
)

func TestNewCampaignsTableViewBuildsRowAndPagerPaths(t *testing.T) {
	response := &service.Campaigns{
		Campaigns: []*service.Campaign{
			{
				ID:              "cmp-1",
				Name:            "Launch",
				AllowCustomText: true,
				CreatedAt:       1736802000,
			},
		},
		Total: 25,
		Limit: 10,
	}

	view := NewCampaignsTableView(response, 2, nil)

	if len(view.Campaigns) != 1 {
		t.Fatalf("expected one row, got %d", len(view.Campaigns))
	}

	row := view.Campaigns[0]
	if row.DetailPath != "/campaigns/cmp-1" {
		t.Fatalf("unexpected detail path: %q", row.DetailPath)
	}
	if row.DeletePath != "/campaigns/cmp-1" {
		t.Fatalf("unexpected delete path: %q", row.DeletePath)
	}

	if view.PrevPagePath != "/campaigns?page=1" {
		t.Fatalf("unexpected previous page path: %q", view.PrevPagePath)
	}
	if view.NextPagePath != "/campaigns?page=3" {
		t.Fatalf("unexpected next page path: %q", view.NextPagePath)
	}
}

func TestNewCampaignPanelViewSetsMutationPaths(t *testing.T) {
	campaign := &service.Campaign{
		ID:              "cmp-1",
		Name:            "Launch",
		AllowCustomText: true,
		CreatedAt:       1736802000,
	}

	view := NewCampaignPanelView(campaign, nil)

	if view.UpdatePath != "/campaigns/cmp-1" {
		t.Fatalf("unexpected update path: %q", view.UpdatePath)
	}
	if view.DeletePath != "/campaigns/cmp-1" {
		t.Fatalf("unexpected delete path: %q", view.DeletePath)
	}
}
