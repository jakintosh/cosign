package app

import (
	"cosign/internal/service"
	"testing"
)

func TestNewSignaturesTableViewBuildsDeleteAndPagerPaths(t *testing.T) {
	response := &service.Signatures{
		Signatures: []*service.Signature{
			{
				ID:        42,
				Name:      "Alice",
				Email:     "alice@example.com",
				Location:  "Boston",
				CreatedAt: 1736802000,
			},
		},
		Total: 25,
		Limit: 10,
	}

	view := NewSignaturesTableView("cmp-1", response, 2, nil)

	if len(view.Signatures) != 1 {
		t.Fatalf("expected one signature row, got %d", len(view.Signatures))
	}

	row := view.Signatures[0]
	if row.DeletePath != "/campaigns/cmp-1/signatures/42" {
		t.Fatalf("unexpected delete path: %q", row.DeletePath)
	}

	if view.PrevPagePath != "/campaigns/cmp-1/signatures?page=1" {
		t.Fatalf("unexpected previous page path: %q", view.PrevPagePath)
	}
	if view.NextPagePath != "/campaigns/cmp-1/signatures?page=3" {
		t.Fatalf("unexpected next page path: %q", view.NextPagePath)
	}
}

func TestNewSignaturesPanelViewCarriesFormState(t *testing.T) {
	panel := NewSignaturesPanelView(
		"cmp-1",
		SignaturesTableView{},
		SignaturesPanelState{
			Name:      "Alice",
			Email:     "alice@example.com",
			Location:  "Boston",
			FormError: "duplicate email",
		},
	)

	if panel.CreatePath != "/campaigns/cmp-1/signatures" {
		t.Fatalf("unexpected create path: %q", panel.CreatePath)
	}
	if panel.FormError != "duplicate email" {
		t.Fatalf("unexpected form error: %q", panel.FormError)
	}
}
