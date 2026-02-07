package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"cosign/internal/service"
	"git.sr.ht/~jakintosh/command-go/pkg/args"
)

var signonsCmd = &args.Command{
	Name: "signons",
	Help: "manage sign-ons",
	Subcommands: []*args.Command{
		signonsListCmd,
		signonsExportCmd,
	},
}

var signonsListCmd = &args.Command{
	Name: "list",
	Help: "list campaign sign-ons",
	Options: []args.Option{
		{
			Long: "limit",
			Type: args.OptionTypeParameter,
			Help: "page size",
		},
		{
			Long: "offset",
			Type: args.OptionTypeParameter,
			Help: "page offset",
		},
	},
	Handler: func(i *args.Input) error {
		limit := i.GetIntParameterOr("limit", 100)
		offset := i.GetIntParameterOr("offset", 0)

		if limit < 1 {
			return fmt.Errorf("limit must be at least 1")
		}
		if offset < 0 {
			return fmt.Errorf("offset must not be negative")
		}

		id, err := resolveCampaignId(i)
		if err != nil {
			return err
		}

		client, err := resolveClient(i, API_PREFIX)
		if err != nil {
			return err
		}

		var response service.Signons
		path := fmt.Sprintf("/admin/campaigns/%s/signons?limit=%d&offset=%d", id, limit, offset)
		if err := client.Get(path, &response); err != nil {
			return err
		}

		return writeJSON(response)
	},
}

var signonsExportCmd = &args.Command{
	Name: "export",
	Help: "export campaign sign-ons to CSV",
	Options: []args.Option{
		{
			Short: 'o',
			Long:  "output",
			Type:  args.OptionTypeParameter,
			Help:  "output file path",
		},
	},
	Handler: func(i *args.Input) error {
		rawOutput := i.GetParameterOr("output", "signons.csv")

		id, err := resolveCampaignId(i)
		if err != nil {
			return err
		}

		output := strings.TrimSpace(rawOutput)
		if output == "" {
			return fmt.Errorf("output file path required")
		}

		client, err := resolveClient(i, API_PREFIX)
		if err != nil {
			return err
		}

		var response service.Signons
		if err := client.Get("/admin/campaigns/"+id+"/signons", &response); err != nil {
			return err
		}

		file, err := os.Create(output)
		if err != nil {
			return fmt.Errorf("create output file: %w", err)
		}
		defer file.Close()

		writer := csv.NewWriter(file)

		if err := writer.Write([]string{"id", "name", "email", "location", "created_at"}); err != nil {
			return err
		}

		for _, signon := range response.Signons {
			createdAt := time.Unix(signon.CreatedAt, 0).UTC().Format(time.RFC3339)
			row := []string{
				strconv.FormatInt(signon.ID, 10),
				signon.Name,
				signon.Email,
				signon.Location,
				createdAt,
			}
			if err := writer.Write(row); err != nil {
				return err
			}
		}

		writer.Flush()
		if err := writer.Error(); err != nil {
			return err
		}

		fmt.Printf("exported %d sign-ons to %s\n", len(response.Signons), output)
		return nil
	},
}
