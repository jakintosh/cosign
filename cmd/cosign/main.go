package main

import (
	"bytes"
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

	"git.sr.ht/~jakintosh/command-go/pkg/args"
)

const (
	BIN_NAME    = "cosign"
	AUTHOR      = "jakintosh"
	VERSION     = "0.1"
	DEFAULT_CFG = "~/.config/cosign"
	DEFAULT_ENV = "default"
	DEFAULT_URL = "http://localhost:8080"
)

func main() {
	root.Parse()
}

var root = &args.Command{
	Name: BIN_NAME,
	Help: "public letter sign-on management system",
	Config: &args.Config{
		Author:  AUTHOR,
		Version: VERSION,
	},
	Subcommands: []*args.Command{
		serveCmd,
		apiCmd,
		envCmd,
		statusCmd,
	},
	Options: []args.Option{
		{
			Long: "url",
			Type: args.OptionTypeParameter,
			Help: "cosign API base URL",
		},
		{
			Long: "env",
			Type: args.OptionTypeParameter,
			Help: "environment name",
		},
		{
			Long: "config-dir",
			Type: args.OptionTypeParameter,
			Help: "configuration directory",
		},
	},
}

func loadCredential(
	name string,
	credsDir string,
) string {
	credPath := filepath.Join(credsDir, name)
	cred, err := os.ReadFile(credPath)
	if err != nil {
		log.Fatalf("failed to load required credential '%s': %v\n", name, err)
	}
	return string(cred)
}

// baseURL resolves the API base URL from options, environment, or config
func baseURL(i *args.Input) string {
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
func baseConfigDir(i *args.Input) string {
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

// activeEnv uses execution environment info to determine which environment is active
func activeEnv(
	i *args.Input,
) string {
	if e := i.GetParameter("env"); e != nil && *e != "" {
		return *e
	}
	if env := os.Getenv("COFFER_ENV"); env != "" {
		return env
	}
	if n, err := loadActiveEnv(i); err == nil && n != "" {
		return n
	}
	return DEFAULT_ENV
}

// generateAPIKey creates a new API key in format {id}.{secret}
func generateAPIKey() (string, error) {
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", err
	}
	return "default." + hex.EncodeToString(keyBytes), nil
}

// request makes an HTTP request and unmarshals the response
func request[T any](
	i *args.Input,
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
	i *args.Input,
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
