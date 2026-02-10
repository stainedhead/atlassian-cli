package page

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

// NewPageCmd creates the page command with subcommands
func NewPageCmd(tokenManager auth.TokenManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "page",
		Short: "Confluence page operations",
		Long:  `Create, read, update, and manage Confluence pages`,
	}

	cmd.AddCommand(newCreateCmd(tokenManager))
	cmd.AddCommand(newGetCmd(tokenManager))
	cmd.AddCommand(newListCmd(tokenManager))
	cmd.AddCommand(newUpdateCmd(tokenManager))
	cmd.AddCommand(newSearchCmd(tokenManager))

	return cmd
}

func newCreateCmd(tokenManager auth.TokenManager) *cobra.Command {
	var (
		spaceKey string
		title    string
		content  string
		parentID string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new Confluence page",
		Long: `Create a new Confluence page with the specified details.

Examples:
  # Create page using default space
  atlassian-cli page create --title "New Page" --content "Page content"
  
  # Override default space
  atlassian-cli page create --confluence-space DOCS --title "API Guide"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig(cmdutil.GetConfigPath(cmd))
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			resolvedSpace, err := config.ResolveSpace(cmd)
			if err != nil {
				return err
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

			req := &types.CreatePageRequest{
				SpaceKey: resolvedSpace,
				Title:    title,
				Content:  content,
				ParentID: parentID,
			}

			page, err := client.CreatePage(context.Background(), req)
			if err != nil {
				return fmt.Errorf("failed to create page: %w", err)
			}

			return outputPage(cmd, page)
		},
	}

	cmd.Flags().StringVar(&spaceKey, "space", "", "Confluence space key (overrides default)")
	cmd.Flags().StringVar(&title, "title", "", "Page title (required)")
	cmd.Flags().StringVar(&content, "content", "", "Page content")
	cmd.Flags().StringVar(&parentID, "parent-id", "", "Parent page ID")

	cmd.MarkFlagRequired("title")

	return cmd
}

func newGetCmd(tokenManager auth.TokenManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <page-id>",
		Short: "Get a Confluence page by ID",
		Long:  `Retrieve detailed information about a Confluence page by its ID`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pageID := args[0]

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

			page, err := client.GetPage(context.Background(), pageID)
			if err != nil {
				return fmt.Errorf("failed to get page: %w", err)
			}

			return outputPage(cmd, page)
		},
	}

	return cmd
}

func newListCmd(tokenManager auth.TokenManager) *cobra.Command {
	var (
		spaceKey   string
		title      string
		maxResults int
		startAt    int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Confluence pages",
		Long: `List Confluence pages with optional filtering.

Examples:
  # List pages in default space
  atlassian-cli page list
  
  # List pages with specific title
  atlassian-cli page list --title "API"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig(cmdutil.GetConfigPath(cmd))
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			var resolvedSpace string
			if spaceKey != "" || cfg.DefaultConfluenceSpace != "" {
				resolvedSpace, err = config.ResolveSpace(cmd)
				if err != nil {
					return err
				}
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

			opts := &types.PageListOptions{
				SpaceKey:   resolvedSpace,
				Title:      title,
				MaxResults: maxResults,
				StartAt:    startAt,
			}

			response, err := client.ListPages(context.Background(), opts)
			if err != nil {
				return fmt.Errorf("failed to list pages: %w", err)
			}

			return outputPageList(cmd, response)
		},
	}

	cmd.Flags().StringVar(&spaceKey, "space", "", "Confluence space key (overrides default)")
	cmd.Flags().StringVar(&title, "title", "", "Filter by title")
	cmd.Flags().IntVar(&maxResults, "max-results", 25, "Maximum number of results")
	cmd.Flags().IntVar(&startAt, "start-at", 0, "Starting index for pagination")

	return cmd
}

func newUpdateCmd(tokenManager auth.TokenManager) *cobra.Command {
	var (
		title   string
		content string
	)

	cmd := &cobra.Command{
		Use:   "update <page-id>",
		Short: "Update a Confluence page",
		Long:  `Update an existing Confluence page with new values`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pageID := args[0]

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

			req := &types.UpdatePageRequest{}
			if title != "" {
				req.Title = &title
			}
			if content != "" {
				req.Content = &content
			}

			page, err := client.UpdatePage(context.Background(), pageID, req)
			if err != nil {
				return fmt.Errorf("failed to update page: %w", err)
			}

			return outputPage(cmd, page)
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "New page title")
	cmd.Flags().StringVar(&content, "content", "", "New page content")

	return cmd
}

func outputPage(cmd *cobra.Command, page *types.Page) error {
	format := cmdutil.GetOutputFormat(cmd)

	switch format {
	case "json":
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		return encoder.Encode(page)
	default: // table
		fmt.Fprintf(cmd.OutOrStdout(), "ID:       %s\n", page.ID)
		fmt.Fprintf(cmd.OutOrStdout(), "Title:    %s\n", page.Title)
		fmt.Fprintf(cmd.OutOrStdout(), "Space:    %s\n", page.SpaceKey)
		fmt.Fprintf(cmd.OutOrStdout(), "Version:  %d\n", page.Version)
		fmt.Fprintf(cmd.OutOrStdout(), "Updated:  %s\n", page.Updated.Format("2006-01-02 15:04:05"))
		if page.Content != "" {
			content := page.Content
			if len(content) > 100 {
				content = content[:100] + "..."
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Content:  %s\n", content)
		}
	}

	return nil
}

func outputPageList(cmd *cobra.Command, response *types.PageListResponse) error {
	format := cmdutil.GetOutputFormat(cmd)

	switch format {
	case "json":
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		return encoder.Encode(response)
	default: // table
		if len(response.Pages) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "No pages found\n")
			return nil
		}

		fmt.Fprintf(cmd.OutOrStdout(), "%-15s %-50s %-10s %-8s\n",
			"ID", "TITLE", "SPACE", "VERSION")
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", strings.Repeat("-", 83))

		for _, page := range response.Pages {
			title := page.Title
			if len(title) > 47 {
				title = title[:47] + "..."
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%-15s %-50s %-10s %-8d\n",
				page.ID, title, page.SpaceKey, page.Version)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "\nShowing %d-%d of %d pages\n",
			response.StartAt+1,
			response.StartAt+len(response.Pages),
			response.Total)
	}

	return nil
}
