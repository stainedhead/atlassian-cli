package project

import (
	"atlassian-cli/internal/auth"
	"atlassian-cli/internal/config"
	"atlassian-cli/internal/jira"
	"atlassian-cli/internal/types"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewProjectCmd creates the project command with subcommands
func NewProjectCmd(tokenManager auth.TokenManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "JIRA project operations",
		Long:  `List and manage JIRA projects`,
	}

	cmd.AddCommand(newListCmd(tokenManager))
	cmd.AddCommand(newGetCmd(tokenManager))

	return cmd
}

func newListCmd(tokenManager auth.TokenManager) *cobra.Command {
	var (
		maxResults int
		startAt    int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List JIRA projects",
		Long: `List JIRA projects.

Examples:
  # List all projects
  atlassian-cli project list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig(viper.GetString("config"))
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			creds, err := tokenManager.Get(context.Background(), cfg.APIEndpoint)
			if err != nil {
				return fmt.Errorf("not authenticated: %w", err)
			}

			client, err := jira.NewAtlassianJiraClient(cfg.APIEndpoint, creds.Email, creds.Token)
			if err != nil {
				return fmt.Errorf("failed to create JIRA client: %w", err)
			}

			opts := &types.ProjectListOptions{
				MaxResults: maxResults,
				StartAt:    startAt,
			}

			response, err := client.ListProjects(context.Background(), opts)
			if err != nil {
				return fmt.Errorf("failed to list projects: %w", err)
			}

			return outputProjectList(cmd, response)
		},
	}

	cmd.Flags().IntVar(&maxResults, "max-results", 50, "Maximum number of results")
	cmd.Flags().IntVar(&startAt, "start-at", 0, "Starting index for pagination")

	return cmd
}

func newGetCmd(tokenManager auth.TokenManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <project-key>",
		Short: "Get a JIRA project by key",
		Long:  `Retrieve detailed information about a JIRA project by its key`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectKey := args[0]

			cfg, err := config.LoadConfig(viper.GetString("config"))
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			creds, err := tokenManager.Get(context.Background(), cfg.APIEndpoint)
			if err != nil {
				return fmt.Errorf("not authenticated: %w", err)
			}

			client, err := jira.NewAtlassianJiraClient(cfg.APIEndpoint, creds.Email, creds.Token)
			if err != nil {
				return fmt.Errorf("failed to create JIRA client: %w", err)
			}

			project, err := client.GetProject(context.Background(), projectKey)
			if err != nil {
				return fmt.Errorf("failed to get project: %w", err)
			}

			return outputProject(cmd, project)
		},
	}

	return cmd
}

func outputProject(cmd *cobra.Command, project *types.Project) error {
	format := viper.GetString("output")

	switch format {
	case "json":
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		return encoder.Encode(project)
	default: // table
		fmt.Fprintf(cmd.OutOrStdout(), "Key:         %s\n", project.Key)
		fmt.Fprintf(cmd.OutOrStdout(), "Name:        %s\n", project.Name)
		fmt.Fprintf(cmd.OutOrStdout(), "Description: %s\n", project.Description)
		fmt.Fprintf(cmd.OutOrStdout(), "Lead:        %s\n", project.Lead)
		fmt.Fprintf(cmd.OutOrStdout(), "Type:        %s\n", project.ProjectType)
	}

	return nil
}

func outputProjectList(cmd *cobra.Command, response *types.ProjectListResponse) error {
	format := viper.GetString("output")

	switch format {
	case "json":
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		return encoder.Encode(response)
	default: // table
		if len(response.Projects) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "No projects found\n")
			return nil
		}

		fmt.Fprintf(cmd.OutOrStdout(), "%-10s %-30s %-15s %-30s\n",
			"KEY", "NAME", "TYPE", "DESCRIPTION")
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", strings.Repeat("-", 85))

		for _, project := range response.Projects {
			name := project.Name
			if len(name) > 27 {
				name = name[:27] + "..."
			}

			description := project.Description
			if len(description) > 27 {
				description = description[:27] + "..."
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%-10s %-30s %-15s %-30s\n",
				project.Key, name, project.ProjectType, description)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "\nShowing %d-%d of %d projects\n",
			response.StartAt+1,
			response.StartAt+len(response.Projects),
			response.Total)
	}

	return nil
}
