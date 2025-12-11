package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"git.sr.ht/~jakintosh/command-go/pkg/args"
)

var campaignsCmd = &args.Command{
	Name: "campaigns",
	Help: "Manage campaigns",
	Subcommands: []*args.Command{
		campaignsListCmd,
		campaignsSelectCmd,
		campaignsCreateCmd,
		campaignsGetCmd,
		campaignsUpdateCmd,
		campaignsDeleteCmd,
	},
}

var campaignsListCmd = &args.Command{
	Name: "list",
	Help: "List all campaigns",
	Handler: func(input *args.Input) error {
		response := &struct {
			Campaigns []struct {
				ID              string `json:"id"`
				Name            string `json:"name"`
				AllowCustomText bool   `json:"allow_custom_text"`
				CreatedAt       int64  `json:"created_at"`
			} `json:"campaigns"`
		}{}

		if err := request(input, "GET", "/admin/campaigns", nil, response); err != nil {
			return err
		}

		// Save mapping to config directory
		cfg, err := envConfig(input)
		if err != nil {
			return fmt.Errorf("failed to get environment config: %w", err)
		}

		configDir := filepath.Join(
			os.ExpandEnv("$HOME"),
			".config/cosign",
			cfg.GetActiveEnv(),
		)

		// Create mapping from integers to campaign UUIDs
		mapping := make(map[string]string)
		for i, campaign := range response.Campaigns {
			idx := fmt.Sprintf("%d", i+1)
			mapping[idx] = campaign.ID
		}

		// Save mapping file
		mappingPath := filepath.Join(configDir, "campaign_map.json")
		mappingData, _ := json.MarshalIndent(mapping, "", "  ")
		if err := os.MkdirAll(configDir, 0700); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
		if err := os.WriteFile(mappingPath, mappingData, 0600); err != nil {
			return fmt.Errorf("failed to save campaign mapping: %w", err)
		}

		// Print campaigns with their indices
		for i, campaign := range response.Campaigns {
			fmt.Printf("%d: \"%s\" (%s)\n", i+1, campaign.Name, campaign.ID)
		}

		return nil
	},
}

var campaignsSelectCmd = &args.Command{
	Name: "select",
	Help: "Select active campaign",
	Operands: []args.Operand{
		{Name: "id", Help: "Campaign index (from list command)"},
	},
	Handler: func(input *args.Input) error {
		idx := input.GetOperand("id")
		if idx == "" {
			return fmt.Errorf("campaign index required")
		}

		// Get config and load mapping
		cfg, err := envConfig(input)
		if err != nil {
			return fmt.Errorf("failed to get environment config: %w", err)
		}

		configDir := filepath.Join(
			os.ExpandEnv("$HOME"),
			".config/cosign",
			cfg.GetActiveEnv(),
		)

		mappingPath := filepath.Join(configDir, "campaign_map.json")
		mappingData, err := os.ReadFile(mappingPath)
		if err != nil {
			return fmt.Errorf("campaign mapping not found, run 'cosign api campaigns list' first: %w", err)
		}

		var mapping map[string]string
		if err := json.Unmarshal(mappingData, &mapping); err != nil {
			return fmt.Errorf("failed to parse campaign mapping: %w", err)
		}

		campaignID, ok := mapping[idx]
		if !ok {
			return fmt.Errorf("invalid campaign index: %s", idx)
		}

		// Get campaign details to show name
		response := &struct {
			ID              string `json:"id"`
			Name            string `json:"name"`
			AllowCustomText bool   `json:"allow_custom_text"`
			CreatedAt       int64  `json:"created_at"`
		}{}

		if err := request(input, "GET", fmt.Sprintf("/admin/campaigns/%s", campaignID), nil, response); err != nil {
			return err
		}

		// Save active campaign
		activePath := filepath.Join(configDir, "active_campaign")
		if err := os.MkdirAll(configDir, 0700); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
		if err := os.WriteFile(activePath, []byte(campaignID), 0600); err != nil {
			return fmt.Errorf("failed to save active campaign: %w", err)
		}

		fmt.Printf("Selected \"%s\" (%s)\n", response.Name, response.ID)
		return nil
	},
}

var campaignsCreateCmd = &args.Command{
	Name: "create",
	Help: "Create a new campaign",
	Operands: []args.Operand{
		{Name: "name", Help: "Campaign name"},
	},
	Handler: func(input *args.Input) error {
		name := input.GetOperand("name")
		if name == "" {
			return fmt.Errorf("campaign name required")
		}

		payload := map[string]string{"name": name}
		body, _ := json.Marshal(payload)

		response := &struct {
			ID              string `json:"id"`
			Name            string `json:"name"`
			AllowCustomText bool   `json:"allow_custom_text"`
			CreatedAt       int64  `json:"created_at"`
		}{}

		if err := request(input, "POST", "/admin/campaigns", body, response); err != nil {
			return err
		}

		return writeJSON(response)
	},
}

var campaignsGetCmd = &args.Command{
	Name: "get",
	Help: "Get campaign details",
	Operands: []args.Operand{
		{Name: "uuid", Help: "Campaign UUID (optional, uses active campaign if not specified)"},
	},
	Handler: func(input *args.Input) error {
		uuid := input.GetOperand("uuid")
		if uuid == "" {
			var err error
			uuid, err = getActiveCampaign(input)
			if err != nil {
				return err
			}
		}

		response := &struct {
			ID              string `json:"id"`
			Name            string `json:"name"`
			AllowCustomText bool   `json:"allow_custom_text"`
			CreatedAt       int64  `json:"created_at"`
		}{}

		if err := request(input, "GET", fmt.Sprintf("/admin/campaigns/%s", uuid), nil, response); err != nil {
			return err
		}

		return writeJSON(response)
	},
}

var campaignsUpdateCmd = &args.Command{
	Name: "update",
	Help: "Update campaign",
	Operands: []args.Operand{
		{Name: "uuid", Help: "Campaign UUID (optional, uses active campaign if not specified)"},
	},
	Options: []args.Option{
		{Long: "name", Type: args.OptionTypeParameter, Help: "New campaign name"},
	},
	Handler: func(input *args.Input) error {
		uuid := input.GetOperand("uuid")
		if uuid == "" {
			var err error
			uuid, err = getActiveCampaign(input)
			if err != nil {
				return err
			}
		}

		payload := make(map[string]interface{})
		if name := input.GetParameter("name"); name != nil {
			payload["name"] = *name
		}

		body, _ := json.Marshal(payload)

		response := &struct {
			ID              string `json:"id"`
			Name            string `json:"name"`
			AllowCustomText bool   `json:"allow_custom_text"`
			CreatedAt       int64  `json:"created_at"`
		}{}

		if err := request(input, "PUT", fmt.Sprintf("/admin/campaigns/%s", uuid), body, response); err != nil {
			return err
		}

		return writeJSON(response)
	},
}

var campaignsDeleteCmd = &args.Command{
	Name: "delete",
	Help: "Delete campaign",
	Operands: []args.Operand{
		{Name: "uuid", Help: "Campaign UUID"},
	},
	Handler: func(input *args.Input) error {
		uuid := input.GetOperand("uuid")
		if uuid == "" {
			return fmt.Errorf("campaign UUID required")
		}

		if err := requestVoid(input, "DELETE", fmt.Sprintf("/admin/campaigns/%s", uuid), nil); err != nil {
			return err
		}

		fmt.Println("Campaign deleted")
		return nil
	},
}

// Helper function to get active campaign UUID
func getActiveCampaign(input *args.Input) (string, error) {
	// Check --campaign-uuid flag
	if uuid := input.GetParameter("campaign-uuid"); uuid != nil && *uuid != "" {
		return *uuid, nil
	}

	// Check CAMPAIGN_UUID env var
	if uuid := os.Getenv("CAMPAIGN_UUID"); uuid != "" {
		return uuid, nil
	}

	// Check config file
	cfg, err := envConfig(input)
	if err != nil {
		return "", fmt.Errorf("failed to get environment config: %w", err)
	}

	configDir := filepath.Join(
		os.ExpandEnv("$HOME"),
		".config/cosign",
		cfg.GetActiveEnv(),
	)

	activePath := filepath.Join(configDir, "active_campaign")
	data, err := os.ReadFile(activePath)
	if err != nil {
		return "", fmt.Errorf("no campaign selected. Run 'cosign api campaigns list' then 'cosign api campaigns select <id>'")
	}

	return string(data), nil
}
