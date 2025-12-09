package main

import (
	"encoding/json"
	"fmt"

	"git.sr.ht/~jakintosh/command-go/pkg/args"
)

var locationCmd = &args.Command{
	Name: "location",
	Help: "Manage location settings and options",
	Subcommands: []*args.Command{
		locationConfigCmd,
		locationOptionsCmd,
	},
}

var locationConfigCmd = &args.Command{
	Name: "config",
	Help: "Manage location configuration",
	Subcommands: []*args.Command{
		locationConfigGetCmd,
		locationConfigSetCmd,
	},
}

var locationConfigGetCmd = &args.Command{
	Name: "get",
	Help: "Get location configuration",
	Handler: func(input *args.Input) error {
		response := &struct {
			AllowCustomText bool `json:"allow_custom_text"`
		}{}

		if err := request(input, "GET", "/location-config", nil, response); err != nil {
			return err
		}

		return writeJSON(response)
	},
}

var locationConfigSetCmd = &args.Command{
	Name: "set",
	Help: "Update location configuration",
	Options: []args.Option{
		{Long: "allow-custom", Type: args.OptionTypeFlag, Help: "Allow custom text in location field"},
		{Long: "strict", Type: args.OptionTypeFlag, Help: "Only allow preset location options"},
	},
	Handler: func(input *args.Input) error {
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

var locationOptionsCmd = &args.Command{
	Name: "options",
	Help: "Manage location options",
	Subcommands: []*args.Command{
		locationOptionsListCmd,
		locationOptionsAddCmd,
		locationOptionsRemoveCmd,
	},
}

var locationOptionsListCmd = &args.Command{
	Name: "list",
	Help: "List location options",
	Handler: func(input *args.Input) error {
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

var locationOptionsAddCmd = &args.Command{
	Name: "add",
	Help: "Add location option",
	Operands: []args.Operand{
		{Name: "value", Help: "Location value to add"},
	},
	Options: []args.Option{
		{Long: "order", Type: args.OptionTypeParameter, Help: "Display order"},
	},
	Handler: func(input *args.Input) error {
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

var locationOptionsRemoveCmd = &args.Command{
	Name: "remove",
	Help: "Remove location option",
	Operands: []args.Operand{
		{Name: "id", Help: "Location option ID to remove"},
	},
	Handler: func(input *args.Input) error {
		id := input.GetOperand("id")
		path := fmt.Sprintf("/admin/location-config/options/%s", id)

		if err := requestVoid(input, "DELETE", path, nil); err != nil {
			return err
		}

		fmt.Println("Location option removed")
		return nil
	},
}
