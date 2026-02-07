package main

import (
	"cosign/internal/database"
	"cosign/internal/service"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"git.sr.ht/~jakintosh/command-go/pkg/args"
	"git.sr.ht/~jakintosh/command-go/pkg/cors"
	"git.sr.ht/~jakintosh/command-go/pkg/keys"
)

const (
	DEFAULT_CREDS_DIR       = "/etc/cosign"
	DEFAULT_DB_PATH         = "/var/lib/cosign/cosign.db"
	DEFAULT_ALLOWED_ORIGINS = "http://localhost:3000"
)

func resolveOption(
	i *args.Input,
	opt string,
	env string,
	def string,
) string {
	rawParameter := i.GetParameter(opt)
	rawEnv := os.Getenv(env)

	if rawParameter != nil {
		parameter := *rawParameter
		if parameter != "" {
			return parameter
		}
	}

	if rawEnv != "" {
		return rawEnv
	}

	return def
}

func normalizePort(
	raw string,
) (
	string,
	error,
) {
	port := strings.TrimSpace(raw)
	port = strings.TrimPrefix(port, ":")
	if port == "" {
		return "", fmt.Errorf("port required")
	}

	value, err := strconv.Atoi(port)
	if err != nil || value < 1 || value > 65535 {
		return "", fmt.Errorf("invalid port %q", raw)
	}

	return ":" + strconv.Itoa(value), nil
}

func parseCSVValues(
	raw string,
) []string {
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			continue
		}
		values = append(values, value)
	}

	return values
}

func loadCredential(
	name string,
	credsDir string,
) (
	string,
	error,
) {
	path := filepath.Join(credsDir, name)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("load credential %q: %w", name, err)
	}

	value := strings.TrimSpace(string(data))
	if value == "" {
		return "", fmt.Errorf("credential %q is empty", name)
	}

	return value, nil
}

var serveCmd = &args.Command{
	Name: "serve",
	Help: "run the cosign HTTP API server",
	Options: []args.Option{
		{
			Long: "db-path",
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
		// read inputs
		rawDBPath := resolveOption(i, "db-path", "COSIGN_DB_PATH", DEFAULT_DB_PATH)
		rawPort := resolveOption(i, "port", "COSIGN_PORT", DEFAULT_PORT)
		rawOrigins := resolveOption(i, "cors-allowed-origins", "COSIGN_CORS_ALLOWED_ORIGINS", DEFAULT_ALLOWED_ORIGINS)
		rawCredentialsDirectory := resolveOption(i, "credentials-directory", "COSIGN_CREDENTIALS_DIRECTORY", DEFAULT_CREDS_DIR)

		// validate inputs
		dbPath := strings.TrimSpace(rawDBPath)
		if dbPath == "" {
			return fmt.Errorf("database path required")
		}

		credentialsDirectory := strings.TrimSpace(rawCredentialsDirectory)
		if credentialsDirectory == "" {
			return fmt.Errorf("credentials directory required")
		}

		origins := parseCSVValues(rawOrigins)

		port, err := normalizePort(rawPort)
		if err != nil {
			return err
		}

		bootstrapToken, err := loadCredential("api_key", credentialsDirectory)
		if err != nil {
			return err
		}

		// init db
		log.Printf("Initializing database at %s...", dbPath)
		dbOpts := database.Options{
			Path: dbPath,
			WAL:  true,
		}
		db, err := database.Open(dbOpts)
		if err != nil {
			return fmt.Errorf("initialize database: %w", err)
		}
		defer db.Close()

		// init service
		svcOpts := service.Options{
			Store: db,
			KeysOptions: &keys.Options{
				Store:          db.KeysStore,
				BootstrapToken: bootstrapToken,
			},
			CORSOptions: &cors.Options{
				Store:          db.CORSStore,
				InitialOrigins: origins,
			},
			HealthCheck: db.HealthCheck,
		}
		svc, err := service.New(svcOpts)
		if err != nil {
			return fmt.Errorf("initialize service: %w", err)
		}

		log.Printf("Starting server on %s...", port)
		return svc.Serve(port, API_PREFIX)
	},
}
