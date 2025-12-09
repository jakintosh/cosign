package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"git.sr.ht/~jakintosh/command-go/pkg/args"
)

// Config represents the client configuration file
type Config struct {
	ActiveEnv string               `json:"activeEnv"`
	Envs      map[string]EnvConfig `json:"envs"`
}

// EnvConfig represents configuration for a specific environment
type EnvConfig struct {
	BaseURL string `json:"baseUrl"`
	APIKey  string `json:"apiKey"`
}

// configPath returns the path to the config file
func configPath(i *args.Input) string {
	return filepath.Join(baseConfigDir(i), "config.json")
}

// loadConfig loads the configuration from disk
func loadConfig(i *args.Input) (*Config, error) {
	path := configPath(i)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				ActiveEnv: "default",
				Envs:      map[string]EnvConfig{},
			}, nil
		}
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Envs == nil {
		cfg.Envs = map[string]EnvConfig{}
	}
	if cfg.ActiveEnv == "" {
		cfg.ActiveEnv = "default"
	}
	return &cfg, nil
}

// saveConfig saves the configuration to disk
func saveConfig(i *args.Input, cfg *Config) error {
	dir := baseConfigDir(i)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	path := configPath(i)
	return os.WriteFile(path, data, 0o600)
}

// loadActiveEnv returns the name of the active environment
func loadActiveEnv(i *args.Input) (string, error) {
	cfg, err := loadConfig(i)
	if err != nil {
		return "", err
	}
	return cfg.ActiveEnv, nil
}

// saveActiveEnv sets the active environment
func saveActiveEnv(i *args.Input, name string) error {
	cfg, err := loadConfig(i)
	if err != nil {
		return err
	}
	if _, ok := cfg.Envs[name]; !ok {
		return fmt.Errorf("environment '%s' does not exist", name)
	}
	cfg.ActiveEnv = name
	return saveConfig(i, cfg)
}

// loadAPIKey returns the API key for the active environment
func loadAPIKey(i *args.Input) (string, error) {
	cfg, err := loadConfig(i)
	if err != nil {
		return "", err
	}
	env := activeEnv(i)
	if e, ok := cfg.Envs[env]; ok {
		return strings.TrimSpace(e.APIKey), nil
	}
	return "", nil
}

// saveAPIKey saves the API key for the active environment
func saveAPIKey(i *args.Input, key string) error {
	cfg, err := loadConfig(i)
	if err != nil {
		return err
	}
	env := activeEnv(i)
	ec := cfg.Envs[env]
	ec.APIKey = key
	cfg.Envs[env] = ec
	return saveConfig(i, cfg)
}

// deleteAPIKey removes the API key for the active environment
func deleteAPIKey(i *args.Input) error {
	cfg, err := loadConfig(i)
	if err != nil {
		return err
	}
	env := activeEnv(i)
	if ec, ok := cfg.Envs[env]; ok {
		ec.APIKey = ""
		cfg.Envs[env] = ec
	}
	return saveConfig(i, cfg)
}

// loadBaseURL returns the base URL for the active environment
func loadBaseURL(i *args.Input) (string, error) {
	cfg, err := loadConfig(i)
	if err != nil {
		return "", err
	}
	env := activeEnv(i)
	if e, ok := cfg.Envs[env]; ok {
		return strings.TrimSpace(e.BaseURL), nil
	}
	return "", nil
}

// saveBaseURL saves the base URL for the active environment
func saveBaseURL(i *args.Input, url string) error {
	cfg, err := loadConfig(i)
	if err != nil {
		return err
	}
	env := activeEnv(i)
	ec := cfg.Envs[env]
	ec.BaseURL = url
	cfg.Envs[env] = ec
	return saveConfig(i, cfg)
}

// deleteBaseURL removes the base URL for the active environment
func deleteBaseURL(i *args.Input) error {
	cfg, err := loadConfig(i)
	if err != nil {
		return err
	}
	env := activeEnv(i)
	if ec, ok := cfg.Envs[env]; ok {
		ec.BaseURL = ""
		cfg.Envs[env] = ec
	}
	return saveConfig(i, cfg)
}
