package app

import (
	"cosign/internal/service"
	"testing"
)

func TestNewLocationsPanelViewBuildsRowActionPaths(t *testing.T) {
	view := NewLocationsPanelView(
		"cmp-1",
		false,
		[]service.LocationOption{{Value: "Boston"}},
		LocationsPanelState{Mode: "edit", EditIndex: 0, DraftValue: "Somerville"},
		nil,
	)

	if len(view.Rows) != 1 {
		t.Fatalf("expected one row, got %d", len(view.Rows))
	}

	row := view.Rows[0]
	if row.UpdatePath != "/campaigns/cmp-1/locations/0" {
		t.Fatalf("unexpected update path: %q", row.UpdatePath)
	}
	if row.DeletePath != "/campaigns/cmp-1/locations/0" {
		t.Fatalf("unexpected delete path: %q", row.DeletePath)
	}
	if row.EditPath != "/campaigns/cmp-1/locations?mode=edit&index=0" {
		t.Fatalf("unexpected edit path: %q", row.EditPath)
	}
	if !row.IsEditing {
		t.Fatalf("expected row to be in editing mode")
	}
	if row.EditValue != "Somerville" {
		t.Fatalf("unexpected edit value: %q", row.EditValue)
	}
}

func TestNewLocationsPanelViewBuildsNewModePaths(t *testing.T) {
	view := NewLocationsPanelView(
		"cmp-1",
		true,
		nil,
		LocationsPanelState{Mode: "new", DraftValue: "New York"},
		nil,
	)

	if !view.NewMode {
		t.Fatalf("expected new mode")
	}
	if view.NewValue != "New York" {
		t.Fatalf("unexpected new value: %q", view.NewValue)
	}
	if view.NewPath != "/campaigns/cmp-1/locations?mode=new" {
		t.Fatalf("unexpected new path: %q", view.NewPath)
	}
	if view.CancelNewPath != "/campaigns/cmp-1/locations" {
		t.Fatalf("unexpected cancel new path: %q", view.CancelNewPath)
	}
}
