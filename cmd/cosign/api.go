package main

import (
	"encoding/json"
	"os"
	"strings"

	"git.sr.ht/~jakintosh/command-go/pkg/args"
	cors "git.sr.ht/~jakintosh/command-go/pkg/cors/cmd"
	"git.sr.ht/~jakintosh/command-go/pkg/envs"
	keys "git.sr.ht/~jakintosh/command-go/pkg/keys/cmd"
	"git.sr.ht/~jakintosh/command-go/pkg/wire"
)

var settingsCmd = &args.Command{
	Name: "settings",
	Help: "manage API settings",
	Subcommands: []*args.Command{
		keys.Command(DEFAULT_CFG, API_PREFIX+"/settings"),
		cors.Command(DEFAULT_CFG, API_PREFIX+"/settings"),
	},
}

var apiCmd = &args.Command{
	Name: "api",
	Help: "API client commands",
	Subcommands: []*args.Command{
		campaignCmd,
		signonsCmd,
		settingsCmd,
	},
}

func resolveClient(
	i *args.Input,
	pathPrefix string,
) (
	wire.Client,
	error,
) {
	client, err := envs.ResolveClient(i, DEFAULT_CFG, pathPrefix)
	if err != nil {
		return wire.Client{}, err
	}

	if strings.HasPrefix(client.BaseURL, "/") {
		client.BaseURL = DEFAULT_BASE_URL + client.BaseURL
	}

	return client, nil
}

func writeJSON(v any) error {
	encoder := json.NewEncoder(os.Stdout)
	return encoder.Encode(v)
}
