package main

import (
	"bytes"
	"cosign/internal/api"
	"cosign/internal/database"
	"cosign/internal/service"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	cmd "git.sr.ht/~jakintosh/command-go"
	"github.com/gorilla/mux"
)

const (
	BIN_NAME    = "cosign"
	AUTHOR      = "jakintosh"
	VERSION     = "0.1"
	DEFAULT_CFG = "~/.config/cosign"
	DEFAULT_URL = "http://localhost:8080"
	DEFAULT_DB  = "./cosign.db"
)

func main() {
	root.Parse()
}

var root = &cmd.Command{
	Name:    BIN_NAME,
	Author:  AUTHOR,
	Version: VERSION,
	Help:    "public letter sign-on management system",
	Subcommands: []*cmd.Command{
		serveCmd,
		apiCmd,
		envCmd,
		statusCmd,
	},
	Options: []cmd.Option{
		{
			Long: "url",
			Type: cmd.OptionTypeParameter,
			Help: "API base URL",
		},
		{
			Long: "key",
			Type: cmd.OptionTypeParameter,
			Help: "API key for authenticated operations",
		},
		{
			Long: "db",
			Type: cmd.OptionTypeParameter,
			Help: "database file path",
		},
		{
			Long: "config-dir",
			Type: cmd.OptionTypeParameter,
			Help: "configuration directory",
		},
	},
}

var serveCmd = &cmd.Command{
	Name: "serve",
	Help: "Start the HTTP API server",
	Options: []cmd.Option{
		{Long: "db", Type: cmd.OptionTypeParameter, Help: "Database file path"},
		{Long: "port", Type: cmd.OptionTypeParameter, Help: "Port to listen on"},
		{Long: "wal", Type: cmd.OptionTypeFlag, Help: "Enable WAL mode for SQLite"},
		{Long: "bootstrap-key", Type: cmd.OptionTypeParameter, Help: "Create initial API key with this ID"},
	},
	Handler: func(input *cmd.Input) error {
		// Get options with defaults
		dbPath := resolveOption(input, "db", "COSIGN_DB", DEFAULT_DB)
		port := resolveOption(input, "port", "COSIGN_PORT", "8080")
		if !strings.HasPrefix(port, ":") {
			port = ":" + port
		}
		useWAL := input.GetFlag("wal")
		bootstrapKey := input.GetParameter("bootstrap-key")

		// Initialize database
		log.Printf("Initializing database at %s...", dbPath)
		if err := database.Init(dbPath, useWAL); err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}

		// Inject stores into service layer
		service.SetSignonStore(database.NewSignonStore())
		service.SetLocationConfigStore(database.NewLocationConfigStore())
		service.SetKeyStore(database.NewKeyStore())
		service.SetCORSStore(database.NewCORSStore())

		// Create bootstrap key if provided
		if bootstrapKey != nil {
			fullKey, err := service.CreateAPIKey(*bootstrapKey)
			if err != nil {
				log.Printf("Warning: failed to create bootstrap key: %v", err)
			} else {
				log.Printf("Bootstrap API key created: %s", fullKey)
				log.Printf("Save this key securely - it will not be shown again!")
			}
		}

		// Build router
		r := mux.NewRouter()
		api.BuildRouter(r.PathPrefix("/api/v1").Subrouter())

		// Start server
		log.Printf("Starting server on %s...", port)
		return http.ListenAndServe(port, r)
	},
}

// resolveOption resolves a configuration value with priority: CLI > Env > Default
func resolveOption(i *cmd.Input, opt string, env string, def string) string {
	if v := i.GetParameter(opt); v != nil && *v != "" {
		return *v
	}
	if v := os.Getenv(env); v != "" {
		return v
	}
	return def
}

// baseURL resolves the API base URL from options, environment, or config
func baseURL(i *cmd.Input) string {
	u := i.GetParameter("url")
	envVar := os.Getenv("COSIGN_URL")
	cfgURL, _ := loadBaseURL(i)

	var url string
	switch {
	case u != nil && *u != "":
		url = strings.TrimRight(*u, "/")
	case envVar != "":
		url = strings.TrimRight(envVar, "/")
	case cfgURL != "":
		url = strings.TrimRight(cfgURL, "/")
	default:
		url = DEFAULT_URL
	}
	return url + "/api/v1"
}

// baseConfigDir resolves the configuration directory
func baseConfigDir(i *cmd.Input) string {
	dir := DEFAULT_CFG
	if c := i.GetParameter("config-dir"); c != nil && *c != "" {
		dir = *c
	}
	if strings.HasPrefix(dir, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			dir = filepath.Join(home, dir[2:])
		}
	}
	return dir
}

// generateAPIKey creates a new API key in format {id}.{secret}
func generateAPIKey() (string, error) {
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", err
	}
	return "key-" + hex.EncodeToString(keyBytes), nil
}

// request makes an HTTP request and unmarshals the response
func request[T any](
	i *cmd.Input,
	method string,
	path string,
	body []byte,
	response *T,
) error {
	url := baseURL(i) + path

	// create request
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return err
	}

	// set content-type header
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// set authorization header
	if key, err := loadAPIKey(i); err == nil && key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}

	// do request
	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// read body
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	// check for error status
	if res.StatusCode >= 400 {
		return fmt.Errorf("server returned %s", res.Status)
	}

	// if response expected, deserialize
	if response != nil {
		// unmarshal outer APIResponse
		var m map[string]json.RawMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return err
		}

		// unmarshal inner response
		if respData, ok := m["data"]; ok {
			err := json.Unmarshal(respData, &response)
			if err != nil {
				return fmt.Errorf("failed to deserialize response: %v", err)
			}
		}
	}

	return nil
}

// requestVoid makes an HTTP request without expecting a response body
func requestVoid(
	i *cmd.Input,
	method string,
	path string,
	body []byte,
) error {
	return request(i, method, path, body, (*struct{})(nil))
}

// writeJSON writes data to stdout as JSON
func writeJSON(data any) error {
	if data != nil {
		return json.NewEncoder(os.Stdout).Encode(data)
	}
	return nil
}
