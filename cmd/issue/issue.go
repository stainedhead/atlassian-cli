package issue

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

// NewIssueCmd creates the issue command with subcommands
func NewIssueCmd(tokenManager auth.TokenManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue",
		Short: "JIRA issue operations",
		Long:  `Create, read, update, and manage JIRA issues`,
	}

	// Add subcommands
	cmd.AddCommand(newCreateCmd(tokenManager))
	cmd.AddCommand(newGetCmd(tokenManager))
	cmd.AddCommand(newListCmd(tokenManager))
	cmd.AddCommand(newUpdateCmd(tokenManager))
	cmd.AddCommand(newSearchCmd(tokenManager))

	return cmd
}

// newCreateCmd creates the issue create command
func newCreateCmd(tokenManager auth.TokenManager) *cobra.Command {
	var (
		project     string
		summary     string
		description string
		issueType   string
		priority    string
		assignee    string
		labels      []string
		components  []string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new JIRA issue",
		Long: `Create a new JIRA issue with the specified details.

Examples:
  # Create issue using default project
  atlassian-cli issue create --type Story --summary "New feature"
  
  # Override default project
  atlassian-cli issue create --jira-project MYPROJ --type Bug --summary "Fix issue"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration
			cfg, err := config.LoadConfig(viper.GetString("config"))
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Resolve project using smart defaults
			resolvedProject, err := config.ResolveProject(cmd)
			if err != nil {
				return err
			}

			// Get credentials
			creds, err := tokenManager.Get(context.Background(), cfg.APIEndpoint)
			if err != nil {
				return fmt.Errorf("not authenticated: %w", err)
			}

			// Create JIRA client
			client, err := jira.NewAtlassianJiraClient(cfg.APIEndpoint, creds.Email, creds.Token)
			if err != nil {
				return fmt.Errorf("failed to create JIRA client: %w", err)
			}

			// Create issue request
			req := &types.CreateIssueRequest{
				Project:     resolvedProject,
				Summary:     summary,
				Description: description,
				IssueType:   issueType,
				Priority:    priority,
				Assignee:    assignee,
				Labels:      labels,
				Components:  components,
			}

			// Create the issue
			issue, err := client.CreateIssue(context.Background(), req)
			if err != nil {
				return fmt.Errorf("failed to create issue: %w", err)
			}

			// Output result
			return outputIssue(cmd, issue)
		},
	}

	cmd.Flags().StringVar(&project, "project", "", "JIRA project key (overrides default)")
	cmd.Flags().StringVar(&summary, "summary", "", "Issue summary (required)")
	cmd.Flags().StringVar(&description, "description", "", "Issue description")
	cmd.Flags().StringVar(&issueType, "type", "Task", "Issue type")
	cmd.Flags().StringVar(&priority, "priority", "", "Issue priority")
	cmd.Flags().StringVar(&assignee, "assignee", "", "Issue assignee")
	cmd.Flags().StringSliceVar(&labels, "labels", nil, "Issue labels")
	cmd.Flags().StringSliceVar(&components, "components", nil, "Issue components")

	cmd.MarkFlagRequired("summary")

	return cmd
}

// newGetCmd creates the issue get command
func newGetCmd(tokenManager auth.TokenManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <issue-key>",
		Short: "Get a JIRA issue by key",
		Long:  `Retrieve detailed information about a JIRA issue by its key`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			issueKey := args[0]

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

			// Create JIRA client
			client, err := jira.NewAtlassianJiraClient(cfg.APIEndpoint, creds.Email, creds.Token)
			if err != nil {
				return fmt.Errorf("failed to create JIRA client: %w", err)
			}

			// Get the issue
			issue, err := client.GetIssue(context.Background(), issueKey)
			if err != nil {
				return fmt.Errorf("failed to get issue: %w", err)
			}

			// Output result
			return outputIssue(cmd, issue)
		},
	}

	return cmd
}

// newListCmd creates the issue list command
func newListCmd(tokenManager auth.TokenManager) *cobra.Command {
	var (
		project    string
		status     []string
		issueType  []string
		assignee   string
		maxResults int
		startAt    int
		jql        string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List JIRA issues",
		Long: `List JIRA issues with optional filtering.

Examples:
  # List issues in default project
  atlassian-cli issue list
  
  # List issues with specific status
  atlassian-cli issue list --status "In Progress,Done"
  
  # Use custom JQL
  atlassian-cli issue list --jql "project = DEMO AND assignee = currentUser()"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration
			cfg, err := config.LoadConfig(viper.GetString("config"))
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Resolve project using smart defaults (only if no JQL provided)
			var resolvedProject string
			if jql == "" {
				resolvedProject, err = config.ResolveProject(cmd)
				if err != nil {
					return err
				}
			}

			// Get credentials
			creds, err := tokenManager.Get(context.Background(), cfg.APIEndpoint)
			if err != nil {
				return fmt.Errorf("not authenticated: %w", err)
			}

			// Create JIRA client
			client, err := jira.NewAtlassianJiraClient(cfg.APIEndpoint, creds.Email, creds.Token)
			if err != nil {
				return fmt.Errorf("failed to create JIRA client: %w", err)
			}

			// Create list options
			opts := &types.IssueListOptions{
				Project:    resolvedProject,
				Status:     status,
				IssueType:  issueType,
				Assignee:   assignee,
				MaxResults: maxResults,
				StartAt:    startAt,
				JQL:        jql,
			}

			// List issues
			response, err := client.ListIssues(context.Background(), opts)
			if err != nil {
				return fmt.Errorf("failed to list issues: %w", err)
			}

			// Output result
			return outputIssueList(cmd, response)
		},
	}

	cmd.Flags().StringVar(&project, "project", "", "JIRA project key (overrides default)")
	cmd.Flags().StringSliceVar(&status, "status", nil, "Filter by status")
	cmd.Flags().StringSliceVar(&issueType, "type", nil, "Filter by issue type")
	cmd.Flags().StringVar(&assignee, "assignee", "", "Filter by assignee")
	cmd.Flags().IntVar(&maxResults, "max-results", 50, "Maximum number of results")
	cmd.Flags().IntVar(&startAt, "start-at", 0, "Starting index for pagination")
	cmd.Flags().StringVar(&jql, "jql", "", "Custom JQL query")

	return cmd
}

// newUpdateCmd creates the issue update command
func newUpdateCmd(tokenManager auth.TokenManager) *cobra.Command {
	var (
		summary     string
		description string
		priority    string
		assignee    string
		labels      []string
		components  []string
	)

	cmd := &cobra.Command{
		Use:   "update <issue-key>",
		Short: "Update a JIRA issue",
		Long:  `Update an existing JIRA issue with new values`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			issueKey := args[0]

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

			// Create JIRA client
			client, err := jira.NewAtlassianJiraClient(cfg.APIEndpoint, creds.Email, creds.Token)
			if err != nil {
				return fmt.Errorf("failed to create JIRA client: %w", err)
			}

			// Build update request
			req := &types.UpdateIssueRequest{}
			
			if summary != "" {
				req.Summary = &summary
			}
			if description != "" {
				req.Description = &description
			}
			if priority != "" {
				req.Priority = &priority
			}
			if assignee != "" {
				req.Assignee = &assignee
			}
			if len(labels) > 0 {
				req.Labels = &labels
			}
			if len(components) > 0 {
				req.Components = &components
			}

			// Update the issue
			issue, err := client.UpdateIssue(context.Background(), issueKey, req)
			if err != nil {
				return fmt.Errorf("failed to update issue: %w", err)
			}

			// Output result
			return outputIssue(cmd, issue)
		},
	}

	cmd.Flags().StringVar(&summary, "summary", "", "New issue summary")
	cmd.Flags().StringVar(&description, "description", "", "New issue description")
	cmd.Flags().StringVar(&priority, "priority", "", "New issue priority")
	cmd.Flags().StringVar(&assignee, "assignee", "", "New issue assignee")
	cmd.Flags().StringSliceVar(&labels, "labels", nil, "New issue labels")
	cmd.Flags().StringSliceVar(&components, "components", nil, "New issue components")

	return cmd
}

// outputIssue outputs a single issue in the configured format
func outputIssue(cmd *cobra.Command, issue *types.Issue) error {
	format := viper.GetString("output")
	
	switch format {
	case "json":
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		return encoder.Encode(issue)
	case "yaml":
		// TODO: Implement YAML output
		return fmt.Errorf("YAML output not yet implemented")
	default: // table
		fmt.Fprintf(cmd.OutOrStdout(), "Key:         %s\n", issue.Key)
		fmt.Fprintf(cmd.OutOrStdout(), "Summary:     %s\n", issue.Summary)
		fmt.Fprintf(cmd.OutOrStdout(), "Status:      %s\n", issue.Status)
		fmt.Fprintf(cmd.OutOrStdout(), "Type:        %s\n", issue.IssueType)
		fmt.Fprintf(cmd.OutOrStdout(), "Priority:    %s\n", issue.Priority)
		fmt.Fprintf(cmd.OutOrStdout(), "Assignee:    %s\n", issue.Assignee)
		fmt.Fprintf(cmd.OutOrStdout(), "Reporter:    %s\n", issue.Reporter)
		fmt.Fprintf(cmd.OutOrStdout(), "Project:     %s\n", issue.Project)
		if len(issue.Labels) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "Labels:      %s\n", strings.Join(issue.Labels, ", "))
		}
		if len(issue.Components) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "Components:  %s\n", strings.Join(issue.Components, ", "))
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Created:     %s\n", issue.Created.Format("2006-01-02 15:04:05"))
		fmt.Fprintf(cmd.OutOrStdout(), "Updated:     %s\n", issue.Updated.Format("2006-01-02 15:04:05"))
	}
	
	return nil
}

// outputIssueList outputs a list of issues in the configured format
func outputIssueList(cmd *cobra.Command, response *types.IssueListResponse) error {
	format := viper.GetString("output")
	
	switch format {
	case "json":
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		return encoder.Encode(response)
	case "yaml":
		// TODO: Implement YAML output
		return fmt.Errorf("YAML output not yet implemented")
	default: // table
		if len(response.Issues) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "No issues found\n")
			return nil
		}

		// Print header
		fmt.Fprintf(cmd.OutOrStdout(), "%-12s %-50s %-15s %-10s %-15s\n", 
			"KEY", "SUMMARY", "STATUS", "TYPE", "ASSIGNEE")
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", strings.Repeat("-", 102))

		// Print issues
		for _, issue := range response.Issues {
			summary := issue.Summary
			if len(summary) > 47 {
				summary = summary[:47] + "..."
			}
			
			fmt.Fprintf(cmd.OutOrStdout(), "%-12s %-50s %-15s %-10s %-15s\n",
				issue.Key, summary, issue.Status, issue.IssueType, issue.Assignee)
		}

		// Print summary
		fmt.Fprintf(cmd.OutOrStdout(), "\nShowing %d-%d of %d issues\n",
			response.StartAt+1, 
			response.StartAt+len(response.Issues), 
			response.Total)
	}
	
	return nil
}