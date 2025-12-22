package util

import (
	"testing"
	"time"

	"cosign/internal/database"
	"cosign/internal/service"
)

const STRIPE_TEST_KEY = "whsec_test"

func MakeDate(
	year int,
	month int,
	day int,
) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

func MakeDateUnix(
	year int,
	month int,
	day int,
) int64 {
	return MakeDate(year, month, day).Unix()
}

func MakeDate3339(
	year int,
	month int,
	day int,
) string {
	return MakeDate(year, month, day).Format(time.RFC3339)
}

func SetupTestDB(t *testing.T) {

	// Set up all store implementations
	database.Init(":memory:", false)
	service.SetCampaignStore(database.NewCampaignStore())
	service.SetSignonStore(database.NewSignonStore())
	service.SetKeyStore(database.NewKeyStore())
	service.SetCORSStore(database.NewCORSStore())

	// Seed default allowed origins so CORS validation passes in tests
	if err := service.InitCORS([]string{
		"http://test-default",
		"http://test-origin",
	}); err != nil {
		t.Fatalf("failed to seed CORS origins: %v", err)
	}

	t.Cleanup(func() {
		service.SetCampaignStore(nil)
		service.SetSignonStore(nil)
		service.SetKeyStore(nil)
		service.SetCORSStore(nil)
	})
}
