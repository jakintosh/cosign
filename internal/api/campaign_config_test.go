package api_test

import (
	"cosign/internal/util"
	"net/http"
	"testing"
)

// TestGetCampaignLocationsPublic tests retrieving campaign locations publicly.
func TestGetCampaignLocationsPublic(t *testing.T) {
	util.SetupTestDB(t)
	router := setupRouter()
	campaignID := createTestCampaign(t)

	var resp map[string]any
	result := get(router, "/api/v1/campaigns/"+campaignID+"/locations", &resp)
	expectStatus(t, http.StatusOK, result)

	locs, ok := resp["data"].([]any)
	if !ok {
		t.Fatalf("expected data array, got %T", resp["data"])
	}
	if len(locs) != 0 {
		t.Errorf("expected no locations by default")
	}
}

// TestUpdateCampaignLocationsAdmin updates locations in full.
func TestUpdateCampaignLocationsAdmin(t *testing.T) {
	util.SetupTestDB(t)
	router := setupRouter()
	campaignID := createTestCampaign(t)

	authHeader := makeTestAuthHeader(t)
	body := `[{"value":"NYC","display_order":1}]`
	var resp map[string]any
	result := put(router, "/api/v1/admin/campaigns/"+campaignID+"/locations", body, &resp, authHeader)
	expectStatus(t, http.StatusOK, result)

	locs, ok := resp["data"].([]any)
	if !ok || len(locs) != 1 {
		t.Fatalf("expected 1 location, got %+v", resp["data"])
	}
	loc, _ := locs[0].(map[string]any)
	if loc["value"] != "NYC" {
		t.Errorf("unexpected location value %+v", loc)
	}
}

// TestUpdateCampaignLocationsInvalidJSON ensures bad payload rejected.
func TestUpdateCampaignLocationsInvalidJSON(t *testing.T) {
	util.SetupTestDB(t)
	router := setupRouter()
	campaignID := createTestCampaign(t)

	authHeader := makeTestAuthHeader(t)
	body := `{invalid`
	var resp map[string]any
	result := put(router, "/api/v1/admin/campaigns/"+campaignID+"/locations", body, &resp, authHeader)
	expectStatus(t, http.StatusBadRequest, result)
}
