package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cosign/internal/service"
	"git.sr.ht/~jakintosh/command-go/pkg/args"
)

var campaignsCmd = &args.Command{
	Name: "campaign",
	Help: "Manage campaigns",
	Subcommands: []*args.Command{
		campaignsListCmd,
		campaignsSelectCmd,
		campaignsCreateCmd,
		campaignsGetCmd,
		campaignsSetCmd,
		campaignsDeleteCmd,
		campaignsLocationsCmd,
	},
	Options: []args.Option{
		{
			Long: "uuid",
			Type: args.OptionTypeParameter,
			Help: "Campaign UUID (overrides active selection)",
		},
	},
}

var campaignsListCmd = &args.Command{
	Name: "list",
	Help: "List all campaigns",
	Handler: func(i *args.Input) error {
		response := service.Campaigns{}
		if err := request(i, "GET", "/admin/campaigns", nil, &response); err != nil {
			return err
		}

		// Save mapping to config directory
		cfg, err := envConfig(i)
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
		{
			Name: "id",
			Help: "Campaign index (from list command)",
		},
	},
	Handler: func(i *args.Input) error {
		idx := i.GetOperand("id")
		if idx == "" {
			return fmt.Errorf("campaign index required")
		}

		// Get config and load mapping
		cfg, err := envConfig(i)
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
		response := &service.Campaign{}

		if err := request(i, "GET", fmt.Sprintf("/admin/campaigns/%s", campaignID), nil, response); err != nil {
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
		{
			Name: "name",
			Help: "Campaign name",
		},
	},
	Handler: func(i *args.Input) error {
		name := i.GetOperand("name")
		if name == "" {
			return fmt.Errorf("campaign name required")
		}

		body, err := json.Marshal(
			map[string]string{
				"name": name,
			},
		)
		if err != nil {
			return err
		}

		response := service.Campaign{}
		if err := request(i, "POST", "/admin/campaigns", body, &response); err != nil {
			return err
		}

		return writeJSON(response)
	},
}

var campaignsGetCmd = &args.Command{
	Name: "get",
	Help: "Get campaign details",
	Handler: func(i *args.Input) error {
		uuid, err := getActiveCampaign(i)
		if err != nil {
			return err
		}

		response := &service.Campaign{}
		if err := request(i, "GET", fmt.Sprintf("/admin/campaigns/%s", uuid), nil, response); err != nil {
			return err
		}

		return writeJSON(response)
	},
}

var campaignsSetCmd = &args.Command{
	Name: "set",
	Help: "Update campaign name and strictness",
	Operands: []args.Operand{
		{
			Name: "name",
			Help: "New campaign name",
		},
	},
	Options: []args.Option{
		{
			Long: "strict",
			Type: args.OptionTypeFlag,
			Help: "Disallow custom location text",
		},
		{
			Long: "allow-custom",
			Type: args.OptionTypeFlag,
			Help: "Allow custom location text",
		},
	},
	Handler: func(i *args.Input) error {
		uuid, err := getActiveCampaign(i)
		if err != nil {
			return err
		}

		name := strings.TrimSpace(i.GetOperand("name"))
		if name == "" {
			return fmt.Errorf("campaign name required")
		}

		payload := map[string]any{
			"name": name,
		}

		if i.GetFlag("strict") && i.GetFlag("allow-custom") {
			return fmt.Errorf("use only one of --strict or --allow-custom")
		}
		if i.GetFlag("strict") {
			payload["allow_custom_text"] = false
		}
		if i.GetFlag("allow-custom") {
			payload["allow_custom_text"] = true
		}

		body, _ := json.Marshal(payload)

		response := &service.Campaign{}
		if err := request(i, "PUT", fmt.Sprintf("/admin/campaigns/%s", uuid), body, response); err != nil {
			return err
		}

		return writeJSON(response)
	},
}

var campaignsDeleteCmd = &args.Command{
	Name: "delete",
	Help: "Delete campaign",
	Handler: func(i *args.Input) error {
		uuid, err := getActiveCampaign(i)
		if err != nil {
			return err
		}

		if err := requestVoid(i, "DELETE", fmt.Sprintf("/admin/campaigns/%s", uuid), nil); err != nil {
			return err
		}

		fmt.Println("Campaign deleted")
		return nil
	},
}

var campaignsLocationsCmd = &args.Command{
	Name: "locations",
	Help: "Manage campaign locations",
	Subcommands: []*args.Command{
		campaignsLocationsGetCmd,
		campaignsLocationsSetCmd,
	},
}

var campaignsLocationsGetCmd = &args.Command{
	Name: "get",
	Help: "Get campaign locations",
	Handler: func(i *args.Input) error {
		campaignID, err := getActiveCampaign(i)
		if err != nil {
			return err
		}

		locs, err := fetchCampaignLocations(i, campaignID)
		if err != nil {
			return err
		}

		return writeJSON(locs)
	},
}

var campaignsLocationsSetCmd = &args.Command{
	Name: "set",
	Help: "Replace campaign locations",
	Options: []args.Option{
		{
			Long: "location",
			Type: args.OptionTypeArray,
			Help: "Comma-separated location values in desired order",
		},
	},
	Handler: func(i *args.Input) error {
		campaignID, err := getActiveCampaign(i)
		if err != nil {
			return err
		}

		locations := i.GetArray("location")
		locs := make([]service.LocationOption, 0, len(locations))
		for idx, raw := range locations {
			value := strings.TrimSpace(raw)
			if value == "" {
				continue
			}
			locs = append(locs, service.LocationOption{
				Value:        value,
				DisplayOrder: idx + 1,
			})
		}

		if err := putCampaignLocations(i, campaignID, locs); err != nil {
			return err
		}

		updated, err := fetchCampaignLocations(i, campaignID)
		if err != nil {
			return err
		}

		return writeJSON(updated)
	},
}

func fetchCampaignLocations(
	input *args.Input,
	campaignID string,
) ([]service.LocationOption, error) {
	var response []service.LocationOption
	path := fmt.Sprintf("/admin/campaigns/%s/locations", campaignID)
	if err := request(input, "GET", path, nil, &response); err != nil {
		return nil, err
	}
	return response, nil
}

func putCampaignLocations(
	input *args.Input,
	campaignID string,
	locs []service.LocationOption,
) error {
	body, _ := json.Marshal(locs)
	path := fmt.Sprintf("/admin/campaigns/%s/locations", campaignID)
	return requestVoid(input, "PUT", path, body)
}

// Helper function to get active campaign UUID
func getActiveCampaign(
	input *args.Input,
) (
	string,
	error,
) {

	// check uuid option
	if uuid := input.GetParameter("uuid"); uuid != nil && *uuid != "" {
		return *uuid, nil
	}

	// check CAMPAIGN_UUID env var
	if uuid := os.Getenv("CAMPAIGN_UUID"); uuid != "" {
		return uuid, nil
	}

	// check config file
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
