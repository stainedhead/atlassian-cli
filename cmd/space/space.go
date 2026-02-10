package space

import (
	"atlassian-cli/internal/cmdutil"
	"atlassian-cli/internal/auth"
	"atlassian-cli/internal/config"
	"atlassian-cli/internal/types"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
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
		cursor     string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Confluence spaces",
		Long: `List Confluence spaces with optional filtering.

Examples:
  # List all spaces
  atlassian-cli space list

  # List only personal spaces
  atlassian-cli space list --type personal

  # Use cursor-based pagination
  atlassian-cli space list --cursor "eyJsaW1pdCI6MjUsIm9mZnNldCI6MjV9"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig(cmdutil.GetConfigPath(cmd))
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			creds, err := tokenManager.Get(context.Background(), cfg.APIEndpoint)
			if err != nil {
				return fmt.Errorf("not authenticated: %w", err)
			}

			factory := cmdutil.GetFactory(cmd)
			client, err := factory.GetConfluenceClient(context.Background(), cfg.APIEndpoint, creds.Email, creds.Token)
			if err != nil {
				return fmt.Errorf("failed to get Confluence client: %w", err)
			}

			opts := &types.SpaceListOptions{
				Type:       spaceType,
				MaxResults: maxResults,
				StartAt:    startAt,
				Cursor:     cursor,
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
	cmd.Flags().IntVar(&startAt, "start-at", 0, "Starting index for pagination (deprecated, use --cursor)")
	cmd.Flags().StringVar(&cursor, "cursor", "", "Cursor for pagination (preferred over --start-at)")

	return cmd
}

func outputSpaceList(cmd *cobra.Command, response *types.SpaceListResponse) error {
	format := cmdutil.GetOutputFormat(cmd)

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

		if response.NextCursor != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "\nNext cursor: %s\n", response.NextCursor)
			fmt.Fprintf(cmd.OutOrStdout(), "Use --cursor \"%s\" to fetch the next page\n", response.NextCursor)
		}
	}

	return nil
}
