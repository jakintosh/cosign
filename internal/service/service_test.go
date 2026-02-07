package service_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"cosign/internal/service"
	"cosign/internal/testutil"
	"git.sr.ht/~jakintosh/command-go/pkg/cors"
	"git.sr.ht/~jakintosh/command-go/pkg/wire"
)

func authHeader() wire.TestHeader {
	return wire.TestHeader{Key: "Authorization", Value: "Bearer " + testutil.BootstrapToken}
}

func originHeader(origin string) wire.TestHeader {
	return wire.TestHeader{Key: "Origin", Value: origin}
}

func createCampaign(t *testing.T, handler http.Handler, name string) service.Campaign {
	t.Helper()

	body := fmt.Sprintf(`{"name":%q}`, name)
	result := wire.TestPost[service.Campaign](handler, "/admin/campaigns", body, authHeader())
	result.ExpectStatus(t, http.StatusCreated)
	if result.Data.ID == "" {
		t.Fatalf("expected campaign id")
	}

	return result.Data
}

func TestHealthEndpoint(t *testing.T) {
	svc := testutil.SetupService(t)
	handler := svc.BuildRouter()

	result := wire.TestGet[service.HealthResponse](handler, "/health")
	result.ExpectStatus(t, http.StatusOK)
	if result.Data.Status != "healthy" {
		t.Fatalf("expected healthy status, got %#v", result.Data)
	}
}

func TestCampaignLifecycle(t *testing.T) {
	svc := testutil.SetupService(t)
	handler := svc.BuildRouter()

	campaign := createCampaign(t, handler, "Launch")

	listResult := wire.TestGet[service.Campaigns](handler, "/admin/campaigns", authHeader())
	listResult.ExpectStatus(t, http.StatusOK)
	if len(listResult.Data.Campaigns) != 1 {
		t.Fatalf("expected one campaign, got %d", len(listResult.Data.Campaigns))
	}

	updateBody := `{"name":"Launch Updated","allow_custom_text":false}`
	updateResult := wire.TestPut[service.Campaign](handler, "/admin/campaigns/"+campaign.ID, updateBody, authHeader())
	updateResult.ExpectStatus(t, http.StatusOK)
	if updateResult.Data.Name != "Launch Updated" || updateResult.Data.AllowCustomText {
		t.Fatalf("unexpected campaign update response: %+v", updateResult.Data)
	}

	deleteResult := wire.TestDelete[struct{}](handler, "/admin/campaigns/"+campaign.ID, authHeader())
	deleteResult.ExpectStatus(t, http.StatusNoContent)
}

func TestSignonPublicCORSValidation(t *testing.T) {
	svc := testutil.SetupService(t)
	handler := svc.BuildRouter()
	campaign := createCampaign(t, handler, "Petition")

	body := `{"name":"Alice","email":"alice@example.com","location":"NYC"}`

	ok := wire.TestPost[service.Signon](
		handler,
		"/campaigns/"+campaign.ID+"/signons",
		body,
		originHeader("http://test-origin"),
	)
	ok.ExpectStatus(t, http.StatusCreated)
	if got := ok.Headers.Get("Access-Control-Allow-Origin"); got != "http://test-origin" {
		t.Fatalf("expected access-control-allow-origin header for allowed origin, got %q", got)
	}

	notAllowed := wire.TestPost[service.Signon](
		handler,
		"/campaigns/"+campaign.ID+"/signons",
		`{"name":"Bob","email":"bob@example.com","location":"NYC"}`,
		originHeader("http://blocked-origin"),
	)
	notAllowed.ExpectStatus(t, http.StatusCreated)
	if got := notAllowed.Headers.Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected no access-control-allow-origin header for disallowed origin, got %q", got)
	}

	preflightDenied := wire.TestOptions[struct{}](
		handler,
		"/campaigns/"+campaign.ID+"/signons",
		originHeader("http://blocked-origin"),
	)
	preflightDenied.ExpectStatus(t, http.StatusForbidden)
}

func TestSignonStrictLocationValidation(t *testing.T) {
	svc := testutil.SetupService(t)
	handler := svc.BuildRouter()
	campaign := createCampaign(t, handler, "Strict")

	setStrict := wire.TestPut[service.Campaign](
		handler,
		"/admin/campaigns/"+campaign.ID,
		`{"allow_custom_text":false}`,
		authHeader(),
	)
	setStrict.ExpectStatus(t, http.StatusOK)

	setLocs := wire.TestPut[service.CampaignLocationsResponse](
		handler,
		"/admin/campaigns/"+campaign.ID+"/locations",
		`{"locations":[{"value":"NYC","display_order":1}]}`,
		authHeader(),
	)
	setLocs.ExpectStatus(t, http.StatusOK)

	invalid := wire.TestPost[service.Signon](
		handler,
		"/campaigns/"+campaign.ID+"/signons",
		`{"name":"Bob","email":"bob@example.com","location":"Berlin"}`,
		originHeader("http://test-origin"),
	)
	invalid.ExpectStatus(t, http.StatusBadRequest)
	if invalid.Error == nil || !strings.Contains(invalid.Error.Message, "location") {
		t.Fatalf("expected location validation error, got %#v", invalid.Error)
	}
}

func TestSettingsCorsAndKeysRoutes(t *testing.T) {
	svc := testutil.SetupService(t)
	handler := svc.BuildRouter()

	getCors := wire.TestGet[[]cors.AllowedOrigin](handler, "/settings/cors", authHeader())
	getCors.ExpectStatus(t, http.StatusOK)

	setCors := wire.TestPut[struct{}](handler, "/settings/cors", `[{"url":"http://app.example"}]`, authHeader())
	setCors.ExpectStatus(t, http.StatusNoContent)

	createKey := wire.TestPost[string](handler, "/settings/keys", "", authHeader())
	createKey.ExpectStatus(t, http.StatusCreated)
	parts := strings.Split(createKey.Data, ".")
	if len(parts) != 2 {
		t.Fatalf("unexpected key token format: %q", createKey.Data)
	}

	deleteKey := wire.TestDelete[struct{}](handler, "/settings/keys/"+parts[0], authHeader())
	deleteKey.ExpectStatus(t, http.StatusNoContent)
}
