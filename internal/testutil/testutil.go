package testutil

import (
	"testing"

	"cosign/internal/database"
	"cosign/internal/service"
	"git.sr.ht/~jakintosh/command-go/pkg/cors"
	"git.sr.ht/~jakintosh/command-go/pkg/keys"
)

const BootstrapToken = "default.0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func SetupService(t *testing.T) *service.Service {
	t.Helper()

	db, err := database.Open(database.Options{Path: ":memory:", WAL: false})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close test database: %v", err)
		}
	})

	svc, err := service.New(service.Options{
		Store: db,
		KeysOptions: &keys.Options{
			Store:          db.KeysStore,
			BootstrapToken: BootstrapToken,
		},
		CORSOptions: &cors.Options{
			Store:          db.CORSStore,
			InitialOrigins: []string{"http://test-origin"},
		},
		HealthCheck: db.HealthCheck,
	})
	if err != nil {
		t.Fatalf("create test service: %v", err)
	}

	return svc
}
