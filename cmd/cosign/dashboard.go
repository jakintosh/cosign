package main

import (
	"cosign/internal/app"
	"fmt"
	"log"
	"strings"

	"git.sr.ht/~jakintosh/command-go/pkg/args"
	"git.sr.ht/~jakintosh/command-go/pkg/wire"
)

const (
	DEFAULT_DASHBOARD_PORT      = "3000"
	DEFAULT_DASHBOARD_CREDS_DIR = "./secrets"
	DEFAULT_DASHBOARD_KEY_FILE  = "api_key"
)

var dashboardCmd = &args.Command{
	Name: "dashboard",
	Help: "run the admin dashboard web UI",
	Options: []args.Option{
		{
			Long: "port",
			Type: args.OptionTypeParameter,
			Help: "dashboard port",
		},
		{
			Long: "api-base-url",
			Type: args.OptionTypeParameter,
			Help: "base URL for the API service",
		},
		{
			Long: "api-prefix",
			Type: args.OptionTypeParameter,
			Help: "API path prefix",
		},
		{
			Long: "credentials-directory",
			Type: args.OptionTypeParameter,
			Help: "directory containing API key",
		},
		{
			Long: "api-key-file",
			Type: args.OptionTypeParameter,
			Help: "api key filename",
		},
	},
	Handler: func(i *args.Input) error {
		rawPort := resolveOption(i, "port", "COSIGN_DASHBOARD_PORT", DEFAULT_DASHBOARD_PORT)
		rawAPIBaseURL := resolveOption(i, "api-base-url", "COSIGN_DASHBOARD_API_BASE_URL", DEFAULT_BASE_URL)
		rawAPIPrefix := resolveOption(i, "api-prefix", "COSIGN_DASHBOARD_API_PREFIX", API_PREFIX)
		rawCredentialsDir := resolveOption(i, "credentials-directory", "COSIGN_DASHBOARD_CREDENTIALS_DIRECTORY", DEFAULT_DASHBOARD_CREDS_DIR)
		rawAPIKeyFile := resolveOption(i, "api-key-file", "COSIGN_DASHBOARD_API_KEY_FILE", DEFAULT_DASHBOARD_KEY_FILE)

		port, err := normalizePort(rawPort)
		if err != nil {
			return err
		}

		credentialsDir := strings.TrimSpace(rawCredentialsDir)
		if credentialsDir == "" {
			return fmt.Errorf("credentials directory required")
		}

		apiKeyFile := strings.TrimSpace(rawAPIKeyFile)
		if apiKeyFile == "" {
			return fmt.Errorf("api key file required")
		}

		apiPrefix := strings.TrimSpace(rawAPIPrefix)
		if apiPrefix == "" {
			return fmt.Errorf("api prefix required")
		}
		if !strings.HasPrefix(apiPrefix, "/") {
			apiPrefix = "/" + apiPrefix
		}

		apiBaseURL := strings.TrimSpace(rawAPIBaseURL)
		if apiBaseURL == "" {
			return fmt.Errorf("api base URL required")
		}

		apiKey, err := loadCredential(apiKeyFile, credentialsDir)
		if err != nil {
			return err
		}

		client := wire.Client{
			BaseURL: strings.TrimRight(apiBaseURL, "/") + apiPrefix,
			APIKey:  apiKey,
		}

		dashboard, err := app.New(app.Options{Client: client, PageSize: 10})
		if err != nil {
			return fmt.Errorf("initialize dashboard: %w", err)
		}

		log.Printf("Starting dashboard on %s...", port)
		return dashboard.Serve(port)
	},
}
