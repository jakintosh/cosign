package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"cosign/internal/service"
	"git.sr.ht/~jakintosh/command-go/pkg/args"
)

var campaignCmd = &args.Command{
	Name: "campaign",
	Help: "manage campaigns",
	Subcommands: []*args.Command{
		campaignListCmd,
		campaignCreateCmd,
		campaignGetCmd,
		campaignUpdateCmd,
		campaignDeleteCmd,
		campaignLocationsCmd,
	},
}

var campaignListCmd = &args.Command{
	Name: "list",
	Help: "list campaigns",
	Handler: func(i *args.Input) error {
		client, err := resolveClient(i, API_PREFIX)
		if err != nil {
			return err
		}

		var response service.Campaigns
		if err := client.Get("/admin/campaigns", &response); err != nil {
			return err
		}

		return writeJSON(response)
	},
}

var campaignCreateCmd = &args.Command{
	Name: "create",
	Help: "create campaign",
	Operands: []args.Operand{
		{
			Name: "name",
			Help: "campaign name",
		},
	},
	Handler: func(i *args.Input) error {
		// get input
		name := i.GetOperand("name")

		// validate input
		name = strings.TrimSpace(name)
		if name == "" {
			return fmt.Errorf("campaign name required")
		}

		// setup client
		client, err := resolveClient(i, API_PREFIX)
		if err != nil {
			return err
		}

		// build request
		req := service.CreateCampaignRequest{
			Name: name,
		}
		body, err := json.Marshal(req)
		if err != nil {
			return err
		}

		// post campaign
		var response service.Campaign
		if err := client.Post("/admin/campaigns", body, &response); err != nil {
			return err
		}

		return writeJSON(response)
	},
}

var campaignGetCmd = &args.Command{
	Name: "get",
	Help: "get campaign",
	Handler: func(i *args.Input) error {
		id, err := resolveCampaignId(i)
		if err != nil {
			return err
		}

		client, err := resolveClient(i, API_PREFIX)
		if err != nil {
			return err
		}

		var response service.Campaign
		if err := client.Get("/admin/campaigns/"+id, &response); err != nil {
			return err
		}

		return writeJSON(response)
	},
}

var campaignUpdateCmd = &args.Command{
	Name: "update",
	Help: "update campaign",
	Operands: []args.Operand{
		{
			Name: "name",
			Help: "new campaign name",
		},
	},
	Options: []args.Option{
		{
			Long: "strict",
			Type: args.OptionTypeFlag,
			Help: "disallow custom location text",
		},
		{
			Long: "allow-custom",
			Type: args.OptionTypeFlag,
			Help: "allow custom location text",
		},
	},
	Handler: func(i *args.Input) error {
		// get input
		strict := i.GetFlag("strict")
		allowCustom := i.GetFlag("allow-custom")
		name := i.GetOperand("name")
		id, err := resolveCampaignId(i)
		if err != nil {
			return err
		}

		// validate input
		name = strings.TrimSpace(name)
		if name == "" {
			return fmt.Errorf("campaign name required")
		}

		if strict && allowCustom {
			return fmt.Errorf("use only one of --strict or --allow-custom")
		}

		// setup client
		client, err := resolveClient(i, API_PREFIX)
		if err != nil {
			return err
		}

		// build request
		payload := service.UpdateCampaignRequest{
			Name: name,
		}
		if strict {
			allowCustomText := false
			payload.AllowCustomText = &allowCustomText
		}
		if allowCustom {
			allowCustomText := true
			payload.AllowCustomText = &allowCustomText
		}
		body, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		// send request
		var response service.Campaign
		if err := client.Put("/admin/campaigns/"+id, body, &response); err != nil {
			return err
		}

		return writeJSON(response)
	},
}

var campaignDeleteCmd = &args.Command{
	Name: "delete",
	Help: "delete campaign",
	Handler: func(i *args.Input) error {
		id, err := resolveCampaignId(i)
		if err != nil {
			return err
		}

		client, err := resolveClient(i, API_PREFIX)
		if err != nil {
			return err
		}

		if err := client.Delete("/admin/campaigns/"+id, nil); err != nil {
			return err
		}

		fmt.Println("campaign deleted")
		return nil
	},
}

var campaignLocationsCmd = &args.Command{
	Name: "locations",
	Help: "manage campaign locations",
	Subcommands: []*args.Command{
		campaignLocationsGetCmd,
		campaignLocationsSetCmd,
	},
}

var campaignLocationsGetCmd = &args.Command{
	Name: "get",
	Help: "get campaign locations",
	Handler: func(i *args.Input) error {
		id, err := resolveCampaignId(i)
		if err != nil {
			return err
		}

		client, err := resolveClient(i, API_PREFIX)
		if err != nil {
			return err
		}

		var response service.CampaignLocationsResponse
		if err := client.Get("/admin/campaigns/"+id+"/locations", &response); err != nil {
			return err
		}

		return writeJSON(response)
	},
}

var campaignLocationsSetCmd = &args.Command{
	Name: "set",
	Help: "replace campaign locations",
	Options: []args.Option{{
		Long: "location",
		Type: args.OptionTypeArray,
		Help: "location values in desired display order",
	}},
	Handler: func(i *args.Input) error {
		values := i.GetArray("location")

		id, err := resolveCampaignId(i)
		if err != nil {
			return err
		}

		locations := make([]service.LocationOption, 0, len(values))
		for idx, value := range values {
			trimmed := strings.TrimSpace(value)
			if trimmed == "" {
				continue
			}
			locations = append(locations, service.LocationOption{
				Value:        trimmed,
				DisplayOrder: idx + 1,
			})
		}

		client, err := resolveClient(i, API_PREFIX)
		if err != nil {
			return err
		}

		req := service.CampaignLocationsRequest{
			Locations: locations,
		}
		body, err := json.Marshal(req)
		if err != nil {
			return err
		}

		var response service.CampaignLocationsResponse
		if err := client.Put("/admin/campaigns/"+id+"/locations", body, &response); err != nil {
			return err
		}

		return writeJSON(response)
	},
}

func resolveCampaignId(
	i *args.Input,
) (
	string,
	error,
) {
	opt := i.GetParameterOr("campaign-id", "")
	opt = strings.TrimSpace(opt)
	if opt != "" {
		return opt, nil
	}

	env := os.Getenv("COSIGN_CAMPAIGN_ID")
	env = strings.TrimSpace(env)
	if env != "" {
		return env, nil
	}

	return "", fmt.Errorf("campaign id required; set --campaign-id or COSIGN_CAMPAIGN_ID")
}
