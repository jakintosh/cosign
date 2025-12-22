package main

import (
	"cosign/internal/api"
	"cosign/internal/database"
	"cosign/internal/service"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"git.sr.ht/~jakintosh/command-go/pkg/args"
	"github.com/gorilla/mux"
)

const (
	DB_FILE_PATH          = "/var/lib/" + BIN_NAME
	PORT                  = "8080"
	CORS_ALLOWED_ORIGINS  = "http://localhost:80"
	CREDENTIALS_DIRECTORY = "/etc/" + BIN_NAME
)

func resolveOption(
	i *args.Input,
	opt string,
	env string,
	def string,
) string {
	if v := i.GetParameter(opt); v != nil && *v != "" {
		return *v
	}
	if v := os.Getenv(env); v != "" {
		return v
	}
	return def
}

var serveCmd = &args.Command{
	Name: "serve",
	Help: "run the cosign HTTP API server",
	Options: []args.Option{
		{
			Long: "db-file-path",
			Type: args.OptionTypeParameter,
			Help: "database file path",
		},
		{
			Long: "port",
			Type: args.OptionTypeParameter,
			Help: "port to listen on",
		},
		{
			Long: "cors-allowed-origins",
			Type: args.OptionTypeParameter,
			Help: "comma-separated allowed origins",
		},
		{
			Long: "credentials-directory",
			Type: args.OptionTypeParameter,
			Help: "credentials directory",
		},
	},
	Handler: func(i *args.Input) error {
		// Get options with defaults
		dbPath := resolveOption(i, "db-file-path", "DB_FILE_PATH", DB_FILE_PATH)
		port := ":" + resolveOption(i, "port", "PORT", PORT)
		originsStr := resolveOption(i, "cors-allowed-origins", "CORS_ALLOWED_ORIGINS", CORS_ALLOWED_ORIGINS)
		var origins []string
		if originsStr != "" {
			origins = strings.Split(originsStr, ",")
		}

		credsDir := resolveOption(i, "credentials-directory", "CREDENTIALS_DIRECTORY", CREDENTIALS_DIRECTORY)
		apiKey := loadCredential("api_key", credsDir)

		// Initialize database
		log.Printf("Initializing database at %s...", dbPath)
		if err := database.Init(dbPath, true); err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}

		// Inject stores into service layer
		service.SetCampaignStore(database.NewCampaignStore())
		service.SetSignonStore(database.NewSignonStore())
		service.SetKeyStore(database.NewKeyStore())
		service.SetCORSStore(database.NewCORSStore())

		service.InitCORS(origins)
		service.InitKeys(apiKey)

		// Build router
		r := mux.NewRouter()
		api.BuildRouter(r.PathPrefix("/api/v1").Subrouter())

		// Start server
		log.Printf("Starting server on %s...", port)
		return http.ListenAndServe(port, r)
	},
}
