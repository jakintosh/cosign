package main

import (
	"git.sr.ht/~jakintosh/command-go/pkg/args"
)

var apiCmd = &args.Command{
	Name: "api",
	Help: "API client commands",
	Subcommands: []*args.Command{
		campaignsCmd,
		signonsCmd,
		locationCmd,
		keysCmd,
		corsCmd,
	},
}
