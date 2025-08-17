package page

import (
	"atlassian-cli/internal/auth"
	"atlassian-cli/internal/config"
	"atlassian-cli/internal/confluence"
	"atlassian-cli/internal/types"
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// newSearchCmd creates the page search command
func newSearchCmd(tokenManager auth.TokenManager) *cobra.Command {
	var (
		space     string
		cql       string
		text      string
		title     string
		pageType  string
		limit     int
	)

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search pages using CQL",
		Long: `Search Confluence pages using CQL (Confluence Query Language) or simple filters.

Examples:
  # Search with CQL
  atlassian-cli page search --cql "space = DEV AND type = page"
  
  # Search with text
  atlassian-cli page search --space DEV --text "documentation"
  
  # Search by title in default space
  atlassian-cli page search --title "API Guide"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration
			cfg, err := config.LoadConfig(viper.GetString("config"))
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Get credentials
			creds, err := tokenManager.Get(context.Background(), cfg.APIEndpoint)
			if err != nil {
				return fmt.Errorf("not authenticated: %w", err)
			}

			// Create Confluence client
			client, err := confluence.NewAtlassianConfluenceClient(cfg.APIEndpoint, creds.Email, creds.Token)
			if err != nil {
				return fmt.Errorf("failed to create Confluence client: %w", err)
			}

			// Build CQL query
			var finalCQL string
			if cql != "" {
				finalCQL = cql
			} else {
				finalCQL, err = buildCQLFromFilters(cmd, space, text, title, pageType)
				if err != nil {
					return err
				}
			}

			// Search pages
			opts := &types.PageSearchOptions{
				CQL:        finalCQL,
				MaxResults: limit,
				StartAt:    0,
			}

			response, err := client.SearchPages(context.Background(), opts)
			if err != nil {
				return fmt.Errorf("failed to search pages: %w", err)
			}

			// Output result
			return outputPageList(cmd, response)
		},
	}

	cmd.Flags().StringVar(&space, "space", "", "Confluence space key (overrides default)")
	cmd.Flags().StringVar(&cql, "cql", "", "CQL query string")
	cmd.Flags().StringVar(&text, "text", "", "search in page content")
	cmd.Flags().StringVar(&title, "title", "", "search in page title")
	cmd.Flags().StringVar(&pageType, "type", "", "filter by content type (page, blogpost)")
	cmd.Flags().IntVar(&limit, "limit", 25, "maximum results to return")

	return cmd
}

// buildCQLFromFilters constructs a CQL query from individual filter parameters
func buildCQLFromFilters(cmd *cobra.Command, space, text, title, pageType string) (string, error) {
	var conditions []string

	// Resolve space if not provided
	if space == "" {
		resolvedSpace, err := config.ResolveSpace(cmd)
		if err != nil {
			return "", fmt.Errorf("no space specified and no default configured: %w", err)
		}
		space = resolvedSpace
	}

	// Add space condition
	if space != "" {
		conditions = append(conditions, fmt.Sprintf("space = %s", space))
	}

	// Add content type condition
	if pageType != "" {
		conditions = append(conditions, fmt.Sprintf("type = %s", pageType))
	} else {
		// Default to pages only
		conditions = append(conditions, "type = page")
	}

	// Add text search condition
	if text != "" {
		conditions = append(conditions, fmt.Sprintf("text ~ \"%s\"", text))
	}

	// Add title search condition
	if title != "" {
		conditions = append(conditions, fmt.Sprintf("title ~ \"%s\"", title))
	}

	if len(conditions) == 0 {
		return "", fmt.Errorf("no search criteria specified")
	}

	// Join conditions with AND
	cql := strings.Join(conditions, " AND ")
	
	// Add default ordering
	cql += " ORDER BY lastModified DESC"

	return cql, nil
}