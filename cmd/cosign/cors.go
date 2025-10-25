package main

import (
	"encoding/json"
	"fmt"

	cmd "git.sr.ht/~jakintosh/command-go"
)

var corsCmd = &cmd.Command{
	Name: "cors",
	Help: "Manage CORS origins",
	Subcommands: []*cmd.Command{
		corsListCmd,
		corsAddCmd,
		corsRemoveCmd,
	},
}

var corsListCmd = &cmd.Command{
	Name: "list",
	Help: "List CORS origins",
	Handler: func(input *cmd.Input) error {
		response := &[]string{}

		if err := request(input, "GET", "/admin/cors", nil, response); err != nil {
			return err
		}

		return writeJSON(response)
	},
}

var corsAddCmd = &cmd.Command{
	Name: "add",
	Help: "Add CORS origin",
	Operands: []cmd.Operand{
		{Name: "origin", Help: "Origin to add (e.g., https://example.com)"},
	},
	Handler: func(input *cmd.Input) error {
		origin := input.GetOperand("origin")

		payload := map[string]string{
			"origin": origin,
		}

		body, _ := json.Marshal(payload)
		if err := requestVoid(input, "POST", "/admin/cors", body); err != nil {
			return err
		}

		fmt.Println("CORS origin added")
		return nil
	},
}

var corsRemoveCmd = &cmd.Command{
	Name: "remove",
	Help: "Remove CORS origin",
	Operands: []cmd.Operand{
		{Name: "origin", Help: "Origin to remove"},
	},
	Handler: func(input *cmd.Input) error {
		origin := input.GetOperand("origin")
		path := fmt.Sprintf("/admin/cors/%s", origin)

		if err := requestVoid(input, "DELETE", path, nil); err != nil {
			return err
		}

		fmt.Println("CORS origin removed")
		return nil
	},
}
