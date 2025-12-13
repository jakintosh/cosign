package main

import (
	"encoding/json"
	"fmt"

	"cosign/internal/service"
	"git.sr.ht/~jakintosh/command-go/pkg/args"
)

var corsCmd = &args.Command{
	Name: "cors",
	Help: "Manage CORS origins",
	Subcommands: []*args.Command{
		corsListCmd,
		corsAddCmd,
		corsRemoveCmd,
	},
}

var corsListCmd = &args.Command{
	Name: "list",
	Help: "List CORS origins",
	Handler: func(input *args.Input) error {
		response := &service.AllowedOrigins{}
		if err := request(input, "GET", "/admin/cors", nil, response); err != nil {
			return err
		}

		return writeJSON(response)
	},
}

var corsAddCmd = &args.Command{
	Name: "add",
	Help: "Add CORS origin",
	Operands: []args.Operand{
		{Name: "origin", Help: "Origin to add (e.g., https://example.com)"},
	},
	Handler: func(input *args.Input) error {
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

var corsRemoveCmd = &args.Command{
	Name: "remove",
	Help: "Remove CORS origin",
	Operands: []args.Operand{
		{Name: "origin", Help: "Origin to remove"},
	},
	Handler: func(input *args.Input) error {
		origin := input.GetOperand("origin")
		path := fmt.Sprintf("/admin/cors/%s", origin)

		if err := requestVoid(input, "DELETE", path, nil); err != nil {
			return err
		}

		fmt.Println("CORS origin removed")
		return nil
	},
}
