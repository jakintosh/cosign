package main

import (
	"encoding/json"
	"fmt"

	cmd "git.sr.ht/~jakintosh/command-go"
)

var keysCmd = &cmd.Command{
	Name: "keys",
	Help: "Manage API keys",
	Subcommands: []*cmd.Command{
		keysCreateCmd,
		keysDeleteCmd,
	},
}

var keysCreateCmd = &cmd.Command{
	Name: "create",
	Help: "Create new API key",
	Operands: []cmd.Operand{
		{Name: "id", Help: "Optional key ID (generated if not provided)"},
	},
	Handler: func(input *cmd.Input) error {
		id := input.GetOperand("id")

		payload := map[string]string{}
		if id != "" {
			payload["id"] = id
		}

		body, _ := json.Marshal(payload)
		response := &struct {
			ID        string `json:"id"`
			Secret    string `json:"secret"`
			CreatedAt int64  `json:"created_at"`
		}{}

		if err := request(input, "POST", "/admin/keys", body, response); err != nil {
			return err
		}

		return writeJSON(response)
	},
}

var keysDeleteCmd = &cmd.Command{
	Name: "delete",
	Help: "Delete API key",
	Operands: []cmd.Operand{
		{Name: "id", Help: "Key ID to delete"},
	},
	Handler: func(input *cmd.Input) error {
		id := input.GetOperand("id")
		path := fmt.Sprintf("/admin/keys/%s", id)

		if err := requestVoid(input, "DELETE", path, nil); err != nil {
			return err
		}

		fmt.Println("API key deleted")
		return nil
	},
}
