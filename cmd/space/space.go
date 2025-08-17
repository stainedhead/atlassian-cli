package space

import (
	"atlassian-cli/internal/auth"
	"atlassian-cli/internal/config"
	"atlassian-cli/internal/confluence"
	"atlassian-cli/internal/types"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewSpaceCmd creates the space command with subcommands
func NewSpaceCmd(tokenManager auth.TokenManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "space",
		Short: "Confluence space operations",
		Long:  `List and manage Confluence spaces`,
	}

	cmd.AddCommand(newListCmd(tokenManager))

	return cmd
}

func newListCmd(tokenManager auth.TokenManager) *cobra.Command {
	var (
		spaceType  string
		maxResults int
		startAt    int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Confluence spaces",
		Long: `List Confluence spaces with optional filtering.

Examples:
  # List all spaces
  atlassian-cli space list
  
  # List only personal spaces
  atlassian-cli space list --type personal`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig(viper.GetString("config"))
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			creds, err := tokenManager.Get(context.Background(), cfg.APIEndpoint)
			if err != nil {
				return fmt.Errorf("not authenticated: %w", err)
			}

			client, err := confluence.NewAtlassianConfluenceClient(cfg.APIEndpoint, creds.Email, creds.Token)
			if err != nil {
				return fmt.Errorf("failed to create Confluence client: %w", err)
			}

			opts := &types.SpaceListOptions{
				Type:       spaceType,
				MaxResults: maxResults,
				StartAt:    startAt,
			}

			response, err := client.ListSpaces(context.Background(), opts)
			if err != nil {
				return fmt.Errorf("failed to list spaces: %w", err)
			}

			return outputSpaceList(cmd, response)
		},
	}

	cmd.Flags().StringVar(&spaceType, "type", "", "Filter by space type (global, personal)")
	cmd.Flags().IntVar(&maxResults, "max-results", 25, "Maximum number of results")
	cmd.Flags().IntVar(&startAt, "start-at", 0, "Starting index for pagination")

	return cmd
}

func outputSpaceList(cmd *cobra.Command, response *types.SpaceListResponse) error {
	format := viper.GetString("output")

	switch format {
	case "json":
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		return encoder.Encode(response)
	default: // table
		if len(response.Spaces) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "No spaces found\n")
			return nil
		}

		fmt.Fprintf(cmd.OutOrStdout(), "%-10s %-30s %-10s %-30s\n",
			"KEY", "NAME", "TYPE", "DESCRIPTION")
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", strings.Repeat("-", 80))

		for _, space := range response.Spaces {
			name := space.Name
			if len(name) > 27 {
				name = name[:27] + "..."
			}

			description := space.Description
			if len(description) > 27 {
				description = description[:27] + "..."
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%-10s %-30s %-10s %-30s\n",
				space.Key, name, space.Type, description)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "\nShowing %d-%d of %d spaces\n",
			response.StartAt+1,
			response.StartAt+len(response.Spaces),
			response.Total)
	}

	return nil
}