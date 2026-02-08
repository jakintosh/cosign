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

var signaturesCmd = &args.Command{
	Name: "signatures",
	Help: "manage signatures",
	Subcommands: []*args.Command{
		signaturesListCmd,
		signaturesExportCmd,
	},
}

var signaturesListCmd = &args.Command{
	Name: "list",
	Help: "list campaign signatures",
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

		var response service.Signatures
		path := fmt.Sprintf("/admin/campaigns/%s/signatures?limit=%d&offset=%d", id, limit, offset)
		if err := client.Get(path, &response); err != nil {
			return err
		}

		return writeJSON(response)
	},
}

var signaturesExportCmd = &args.Command{
	Name: "export",
	Help: "export campaign signatures to CSV",
	Options: []args.Option{
		{
			Short: 'o',
			Long:  "output",
			Type:  args.OptionTypeParameter,
			Help:  "output file path",
		},
	},
	Handler: func(i *args.Input) error {
		rawOutput := i.GetParameterOr("output", "signatures.csv")

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

		var response service.Signatures
		if err := client.Get("/admin/campaigns/"+id+"/signatures", &response); err != nil {
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

		for _, signature := range response.Signatures {
			createdAt := time.Unix(signature.CreatedAt, 0).UTC().Format(time.RFC3339)
			row := []string{
				strconv.FormatInt(signature.ID, 10),
				signature.Name,
				signature.Email,
				signature.Location,
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

		fmt.Printf("exported %d signatures to %s\n", len(response.Signatures), output)
		return nil
	},
}
