package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	cmd "git.sr.ht/~jakintosh/command-go"
)

var apiCmd = &cmd.Command{
	Name: "api",
	Help: "API client commands",
	Subcommands: []*cmd.Command{
		signonsCmd,
		locationConfigCmd,
		locationOptionsCmd,
		keysCmd,
		corsCmd,
	},
}

// Signons commands
var signonsCmd = &cmd.Command{
	Name: "signons",
	Help: "Manage sign-ons",
	Subcommands: []*cmd.Command{
		signonsListCmd,
		signonsExportCmd,
	},
}

var signonsListCmd = &cmd.Command{
	Name: "list",
	Help: "List all sign-ons",
	Options: []cmd.Option{
		{Long: "limit", Type: cmd.OptionTypeParameter, Help: "Limit number of results"},
		{Long: "offset", Type: cmd.OptionTypeParameter, Help: "Offset for pagination"},
	},
	Handler: func(input *cmd.Input) error {
		limit := "100"
		if l := input.GetParameter("limit"); l != nil {
			limit = *l
		}
		offset := "0"
		if o := input.GetParameter("offset"); o != nil {
			offset = *o
		}

		path := fmt.Sprintf("/admin/signons?limit=%s&offset=%s", limit, offset)

		response := &struct {
			Signons []struct {
				ID        int64  `json:"id"`
				Name      string `json:"name"`
				Email     string `json:"email"`
				Location  string `json:"location"`
				CreatedAt int64  `json:"created_at"`
			} `json:"signons"`
		}{}

		if err := request(input, "GET", path, nil, response); err != nil {
			return err
		}

		return writeJSON(response)
	},
}

var signonsExportCmd = &cmd.Command{
	Name: "export",
	Help: "Export sign-ons to CSV",
	Options: []cmd.Option{
		{Short: 'o', Long: "output", Type: cmd.OptionTypeParameter, Help: "Output file path"},
	},
	Handler: func(input *cmd.Input) error {
		output := "signons.csv"
		if o := input.GetParameter("output"); o != nil {
			output = *o
		}

		response := &struct {
			Signons []struct {
				ID        int64  `json:"id"`
				Name      string `json:"name"`
				Email     string `json:"email"`
				Location  string `json:"location"`
				CreatedAt int64  `json:"created_at"`
			} `json:"signons"`
		}{}

		if err := request(input, "GET", "/admin/signons", nil, response); err != nil {
			return err
		}

		// Create CSV file
		file, err := os.Create(output)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		defer writer.Flush()

		// Write header
		if err := writer.Write([]string{"ID", "Name", "Email", "Location", "Created At"}); err != nil {
			return err
		}

		// Write rows
		for _, s := range response.Signons {
			createdAt := time.Unix(s.CreatedAt, 0).Format(time.RFC3339)
			row := []string{
				strconv.FormatInt(s.ID, 10),
				s.Name,
				s.Email,
				s.Location,
				createdAt,
			}
			if err := writer.Write(row); err != nil {
				return err
			}
		}

		fmt.Printf("Exported %d sign-ons to %s\n", len(response.Signons), output)
		return nil
	},
}

// Location config commands
var locationConfigCmd = &cmd.Command{
	Name: "location-config",
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

// Location options commands
var locationOptionsCmd = &cmd.Command{
	Name: "location-options",
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

// Keys commands
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

// CORS commands
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

