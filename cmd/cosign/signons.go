package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	cmd "git.sr.ht/~jakintosh/command-go/pkg/args"
)

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
