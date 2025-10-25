package main

import (
	cmd "git.sr.ht/~jakintosh/command-go"
)

var apiCmd = &cmd.Command{
	Name: "api",
	Help: "API client commands",
	Subcommands: []*cmd.Command{
		signonsCmd,
		locationCmd,
		keysCmd,
		corsCmd,
	},
}
