package issue

import (
	"atlassian-cli/internal/cmdutil"
	"atlassian-cli/internal/auth"
	"atlassian-cli/internal/config"
	"atlassian-cli/internal/types"
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// newSearchCmd creates the issue search command
func newSearchCmd(tokenManager auth.TokenManager) *cobra.Command {
	var (
		project   string
		jql       string
		assignee  string
		status    string
		issueType string
		limit     int
	)

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search issues using JQL",
		Long: `Search JIRA issues using JQL (JIRA Query Language) or simple filters.

Examples:
  # Search with JQL
  atlassian-cli issue search --jql "assignee = currentUser() AND status = 'In Progress'"
  
  # Search with simple filters
  atlassian-cli issue search --project DEMO --status "In Progress" --assignee john.doe
  
  # Search in default project
  atlassian-cli issue search --status Open --type Bug`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration
			cfg, err := config.LoadConfig(cmdutil.GetConfigPath(cmd))
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Get credentials
			creds, err := tokenManager.Get(context.Background(), cfg.APIEndpoint)
			if err != nil {
				return fmt.Errorf("not authenticated: %w", err)
			}

			// Get JIRA client from factory
			factory := cmdutil.GetFactory(cmd)
			client, err := factory.GetJiraClient(context.Background(), cfg.APIEndpoint, creds.Email, creds.Token)
			if err != nil {
				return fmt.Errorf("failed to get JIRA client: %w", err)
			}

			// Build JQL query
			var finalJQL string
			if jql != "" {
				finalJQL = jql
			} else {
				finalJQL, err = buildJQLFromFilters(cmd, project, assignee, status, issueType)
				if err != nil {
					return err
				}
			}

			// Search issues
			opts := &types.IssueSearchOptions{
				JQL:        finalJQL,
				MaxResults: limit,
				StartAt:    0,
			}

			response, err := client.SearchIssues(context.Background(), opts)
			if err != nil {
				return fmt.Errorf("failed to search issues: %w", err)
			}

			// Convert SearchResponse to ListResponse for output
			listResponse := &types.IssueListResponse{
				Issues:     response.Issues,
				Total:      response.Total,
				StartAt:    response.StartAt,
				MaxResults: response.MaxResults,
			}

			// Output result
			return outputIssueList(cmd, listResponse)
		},
	}

	cmd.Flags().StringVar(&project, "project", "", "JIRA project key (overrides default)")
	cmd.Flags().StringVar(&jql, "jql", "", "JQL query string")
	cmd.Flags().StringVar(&assignee, "assignee", "", "filter by assignee")
	cmd.Flags().StringVar(&status, "status", "", "filter by status")
	cmd.Flags().StringVar(&issueType, "type", "", "filter by issue type")
	cmd.Flags().IntVar(&limit, "limit", 50, "maximum results to return")

	return cmd
}

// buildJQLFromFilters constructs a JQL query from individual filter parameters
func buildJQLFromFilters(cmd *cobra.Command, project, assignee, status, issueType string) (string, error) {
	var conditions []string

	// Resolve project if not provided
	if project == "" {
		resolvedProject, err := config.ResolveProject(cmd)
		if err != nil {
			return "", fmt.Errorf("no project specified and no default configured: %w", err)
		}
		project = resolvedProject
	}

	// Add project condition
	if project != "" {
		conditions = append(conditions, fmt.Sprintf("project = %s", project))
	}

	// Add assignee condition
	if assignee != "" {
		if assignee == "me" || assignee == "currentUser()" {
			conditions = append(conditions, "assignee = currentUser()")
		} else {
			conditions = append(conditions, fmt.Sprintf("assignee = \"%s\"", assignee))
		}
	}

	// Add status condition
	if status != "" {
		conditions = append(conditions, fmt.Sprintf("status = \"%s\"", status))
	}

	// Add issue type condition
	if issueType != "" {
		conditions = append(conditions, fmt.Sprintf("type = \"%s\"", issueType))
	}

	if len(conditions) == 0 {
		return "", fmt.Errorf("no search criteria specified")
	}

	// Join conditions with AND
	jql := strings.Join(conditions, " AND ")

	// Add default ordering
	jql += " ORDER BY created DESC"

	return jql, nil
}
