package main

import (
	"encoding/json"
	"fmt"

	cmd "git.sr.ht/~jakintosh/command-go"
)

var locationCmd = &cmd.Command{
	Name: "location",
	Help: "Manage location settings and options",
	Subcommands: []*cmd.Command{
		locationConfigCmd,
		locationOptionsCmd,
	},
}

var locationConfigCmd = &cmd.Command{
	Name: "config",
	Help: "Manage location configuration",
	Subcommands: []*cmd.Command{
		locationConfigGetCmd,
		locationConfigSetCmd,
	},
}

var locationConfigGetCmd = &cmd.Command{
	Name: "get",
	Help: "Get location configuration",
	Handler: func(input *cmd.Input) error {
		response := &struct {
			AllowCustomText bool `json:"allow_custom_text"`
		}{}

		if err := request(input, "GET", "/location-config", nil, response); err != nil {
			return err
		}

		return writeJSON(response)
	},
}

var locationConfigSetCmd = &cmd.Command{
	Name: "set",
	Help: "Update location configuration",
	Options: []cmd.Option{
		{Long: "allow-custom", Type: cmd.OptionTypeFlag, Help: "Allow custom text in location field"},
		{Long: "strict", Type: cmd.OptionTypeFlag, Help: "Only allow preset location options"},
	},
	Handler: func(input *cmd.Input) error {
		allowCustom := input.GetFlag("allow-custom")
		strict := input.GetFlag("strict")

		if allowCustom && strict {
			return fmt.Errorf("cannot use both --allow-custom and --strict")
		}

		payload := map[string]bool{
			"allow_custom_text": !strict,
		}

		body, _ := json.Marshal(payload)
		if err := requestVoid(input, "PUT", "/admin/location-config", body); err != nil {
			return err
		}

		fmt.Println("Location configuration updated")
		return nil
	},
}

var locationOptionsCmd = &cmd.Command{
	Name: "options",
	Help: "Manage location options",
	Subcommands: []*cmd.Command{
		locationOptionsListCmd,
		locationOptionsAddCmd,
		locationOptionsRemoveCmd,
	},
}

var locationOptionsListCmd = &cmd.Command{
	Name: "list",
	Help: "List location options",
	Handler: func(input *cmd.Input) error {
		response := &[]struct {
			ID           int64  `json:"id"`
			Value        string `json:"value"`
			DisplayOrder int    `json:"display_order"`
		}{}

		if err := request(input, "GET", "/admin/location-config/options", nil, response); err != nil {
			return err
		}

		return writeJSON(response)
	},
}

var locationOptionsAddCmd = &cmd.Command{
	Name: "add",
	Help: "Add location option",
	Operands: []cmd.Operand{
		{Name: "value", Help: "Location value to add"},
	},
	Options: []cmd.Option{
		{Long: "order", Type: cmd.OptionTypeParameter, Help: "Display order"},
	},
	Handler: func(input *cmd.Input) error {
		value := input.GetOperand("value")
		order := 0
		if o := input.GetParameter("order"); o != nil {
			fmt.Sscanf(*o, "%d", &order)
		}

		payload := map[string]any{
			"value":         value,
			"display_order": order,
		}

		body, _ := json.Marshal(payload)
		response := &struct {
			ID           int64  `json:"id"`
			Value        string `json:"value"`
			DisplayOrder int    `json:"display_order"`
		}{}

		if err := request(input, "POST", "/admin/location-config/options", body, response); err != nil {
			return err
		}

		return writeJSON(response)
	},
}

var locationOptionsRemoveCmd = &cmd.Command{
	Name: "remove",
	Help: "Remove location option",
	Operands: []cmd.Operand{
		{Name: "id", Help: "Location option ID to remove"},
	},
	Handler: func(input *cmd.Input) error {
		id := input.GetOperand("id")
		path := fmt.Sprintf("/admin/location-config/options/%s", id)

		if err := requestVoid(input, "DELETE", path, nil); err != nil {
			return err
		}

		fmt.Println("Location option removed")
		return nil
	},
}
