package main

import (
	"cosign/internal/service"
	"fmt"
	"strings"

	"git.sr.ht/~jakintosh/command-go/pkg/args"
	"git.sr.ht/~jakintosh/command-go/pkg/envs"
)

var statusCmd = &args.Command{
	Name: "status",
	Help: "show active environment and server health",
	Handler: func(i *args.Input) error {
		verbose := i.GetFlag("verbose")

		cfg, err := envs.LoadConfig(i, DEFAULT_CFG)
		if err != nil {
			return err
		}

		activeEnv := cfg.ResolveActiveEnvName(i)
		client := cfg.ResolveClient(i, API_PREFIX)
		if strings.HasPrefix(client.BaseURL, "/") {
			client.BaseURL = DEFAULT_BASE_URL + client.BaseURL
		}

		fmt.Printf("Environment: %s\n", activeEnv)
		fmt.Printf("Base URL: %s\n", client.BaseURL)
		if client.APIKey == "" {
			fmt.Println("API Key: none configured")
		} else {
			fmt.Println("API Key: present")
		}

		response := service.HealthResponse{}
		if err := client.Get("/health", &response); err != nil {
			fmt.Println("Health: down")
			if verbose {
				fmt.Printf("  error: %v\n", err)
			}
			return nil
		}

		if response.Status == "healthy" {
			fmt.Println("Health: up")
		} else {
			fmt.Println("Health: down")
		}

		if verbose {
			return writeJSON(response)
		}

		return nil
	},
}
