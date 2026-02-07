package main

import (
	"git.sr.ht/~jakintosh/command-go/pkg/args"
	"git.sr.ht/~jakintosh/command-go/pkg/envs"
	"git.sr.ht/~jakintosh/command-go/pkg/version"
	"git.sr.ht/~jakintosh/command-go/pkg/wire"
)

const (
	BIN_NAME         = "cosign"
	AUTHOR           = "jakintosh"
	DEFAULT_CFG      = "~/.config/cosign"
	API_PREFIX       = "/api/v1"
	DEFAULT_BASE_URL = "http://localhost:8080"
	DEFAULT_PORT     = "8080"
)

func main() {
	root.Parse()
}

var root = &args.Command{
	Name: BIN_NAME,
	Help: "public letter sign-on management system",
	Config: &args.Config{
		Author: AUTHOR,
		HelpOption: &args.HelpOption{
			Short: 'h',
			Long:  "help",
		},
	},
	Subcommands: []*args.Command{
		serveCmd,
		apiCmd,
		envs.Command(DEFAULT_CFG),
		statusCmd,
		version.Command(VersionInfo),
	},
	Options: envs.ConfigOptionsAnd(wire.ClientOptionsAnd(
		args.Option{
			Long: "campaign-id",
			Type: args.OptionTypeParameter,
			Help: "campaign ID for campaign-scoped commands",
		},
		args.Option{
			Short: 'v',
			Long:  "verbose",
			Type:  args.OptionTypeFlag,
			Help:  "use verbose output",
		},
	)...),
}
