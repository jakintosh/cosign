package main

import (
	"fmt"
	"strings"

	"git.sr.ht/~jakintosh/command-go/pkg/args"
)

var statusCmd = &args.Command{
	Name: "status",
	Help: "show environment and server health",
	Options: []args.Option{
		{
			Long: "verbose",
			Type: args.OptionTypeFlag,
			Help: "show detailed output",
		},
	},
	Handler: func(i *args.Input) error {

		cfg, err := envConfig(i)
		env := DEFAULT_ENV
		key := ""
		if err == nil {
			if cfg.GetActiveEnv() != "" {
				env = cfg.GetActiveEnv()
			}
			key = cfg.GetApiKey()
		}

		base := strings.TrimSuffix(baseURLWithConfig(i, cfg), "/api/v1")

		fmt.Printf("Environment: %s\n", env)
		fmt.Printf("Base URL: %s\n", base)
		if key == "" {
			fmt.Println("API Key: none")
		} else {
			fmt.Println("API Key: present")
		}

		response := &map[string]string{}
		if err := request(i, "GET", "/health", nil, response); err != nil {
			fmt.Println("Health: down")
			if i.GetFlag("verbose") {
				fmt.Printf("  error: %v\n", err)
			}
			return nil
		}

		if data, ok := (*response)["status"]; ok && data == "healthy" {
			fmt.Println("Health: up")
		} else {
			fmt.Printf("Health: down\n")
		}
		if i.GetFlag("verbose") {
			return writeJSON(response)
		}
		return nil
	},
}
