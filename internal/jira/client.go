package jira

import (
	"atlassian-cli/internal/cache"
	"atlassian-cli/internal/types"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ctreminiom/go-atlassian/jira/v3"
	"github.com/ctreminiom/go-atlassian/pkg/infra/models"
)

// JiraClient defines the interface for JIRA operations
type JiraClient interface {
	CreateIssue(ctx context.Context, req *types.CreateIssueRequest) (*types.Issue, error)
	GetIssue(ctx context.Context, key string) (*types.Issue, error)
	UpdateIssue(ctx context.Context, key string, req *types.UpdateIssueRequest) (*types.Issue, error)
	ListIssues(ctx context.Context, opts *types.IssueListOptions) (*types.IssueListResponse, error)
	ListProjects(ctx context.Context, opts *types.ProjectListOptions) (*types.ProjectListResponse, error)
	GetProject(ctx context.Context, key string) (*types.Project, error)
}

// AtlassianJiraClient implements JiraClient using the go-atlassian library
type AtlassianJiraClient struct {
	client *v3.Client
}

// NewAtlassianJiraClient creates a new JIRA client
func NewAtlassianJiraClient(baseURL, email, token string) (*AtlassianJiraClient, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if token == "" {
		return nil, fmt.Errorf("token is required")
	}

	// Create the client instance
	instance, err := v3.New(nil, baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create JIRA client: %w", err)
	}

	// Set authentication
	instance.Auth.SetBasicAuth(email, token)

	return &AtlassianJiraClient{
		client: instance,
	}, nil
}

// CreateIssue creates a new JIRA issue
func (c *AtlassianJiraClient) CreateIssue(ctx context.Context, req *types.CreateIssueRequest) (*types.Issue, error) {
	if req == nil {
		return nil, fmt.Errorf("create issue request cannot be nil")
	}

	// Build the issue creation payload
	fields := &models.IssueFieldsScheme{
		Summary: req.Summary,
		Project: &models.ProjectScheme{
			Key: req.Project,
		},
		IssueType: &models.IssueTypeScheme{
			Name: req.IssueType,
		},
	}

	// Set optional fields
	if req.Priority != "" {
		fields.Priority = &models.PriorityScheme{
			Name: req.Priority,
		}
	}

	if req.Assignee != "" {
		fields.Assignee = &models.UserScheme{
			Name: req.Assignee,
		}
	}

	if len(req.Labels) > 0 {
		fields.Labels = req.Labels
	}

	if len(req.Components) > 0 {
		components := make([]*models.ComponentScheme, len(req.Components))
		for i, comp := range req.Components {
			components[i] = &models.ComponentScheme{
				Name: comp,
			}
		}
		fields.Components = components
	}

	payload := &models.IssueScheme{
		Fields: fields,
	}

	// Create the issue
	result, _, err := c.client.Issue.Create(ctx, payload, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}

	// Convert the response to our internal type
	issue := &types.Issue{
		ID:      result.ID,
		Key:     result.Key,
		Project: req.Project,
		Summary: req.Summary,
	}

	return issue, nil
}

// GetIssue retrieves a JIRA issue by key
func (c *AtlassianJiraClient) GetIssue(ctx context.Context, key string) (*types.Issue, error) {
	if key == "" {
		return nil, fmt.Errorf("issue key cannot be empty")
	}

	// Get the issue
	result, _, err := c.client.Issue.Get(ctx, key, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue %s: %w", key, err)
	}

	// Convert the response to our internal type
	issue := convertAtlassianIssue(result)
	return issue, nil
}

// UpdateIssue updates an existing JIRA issue
func (c *AtlassianJiraClient) UpdateIssue(ctx context.Context, key string, req *types.UpdateIssueRequest) (*types.Issue, error) {
	if key == "" {
		return nil, fmt.Errorf("issue key cannot be empty")
	}
	if req == nil {
		return nil, fmt.Errorf("update issue request cannot be nil")
	}

	// Build the fields for the update
	fields := &models.IssueFieldsScheme{}

	if req.Summary != nil {
		fields.Summary = *req.Summary
	}

	if req.Priority != nil {
		fields.Priority = &models.PriorityScheme{
			Name: *req.Priority,
		}
	}

	if req.Assignee != nil {
		fields.Assignee = &models.UserScheme{
			Name: *req.Assignee,
		}
	}

	if req.Labels != nil {
		fields.Labels = *req.Labels
	}

	if req.Components != nil {
		components := make([]*models.ComponentScheme, len(*req.Components))
		for i, comp := range *req.Components {
			components[i] = &models.ComponentScheme{
				Name: comp,
			}
		}
		fields.Components = components
	}

	payload := &models.IssueScheme{
		Fields: fields,
	}

	// Update the issue
	_, err := c.client.Issue.Update(ctx, key, true, payload, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to update issue %s: %w", key, err)
	}

	// Handle status transition if needed
	if req.Status != nil {
		// Note: Status updates require transitions, which is more complex
		// For now, we'll return an error if status update is attempted
		return nil, fmt.Errorf("status updates not yet implemented - use transitions")
	}

	// Fetch and return the updated issue
	return c.GetIssue(ctx, key)
}

// ListIssues lists JIRA issues based on the provided options
func (c *AtlassianJiraClient) ListIssues(ctx context.Context, opts *types.IssueListOptions) (*types.IssueListResponse, error) {
	if opts == nil {
		opts = &types.IssueListOptions{}
	}

	// Build JQL query
	jql := opts.JQL
	if jql == "" {
		jql = buildJQLFromOptions(opts)
	}

	// Set default values
	maxResults := opts.MaxResults
	if maxResults <= 0 {
		maxResults = 50
	}

	startAt := opts.StartAt
	if startAt < 0 {
		startAt = 0
	}

	// Search for issues
	result, _, err := c.client.Issue.Search.Get(ctx, jql, nil, nil, maxResults, startAt, "")
	if err != nil {
		return nil, fmt.Errorf("failed to list issues: %w", err)
	}

	// Convert the response to our internal type
	issues := make([]types.Issue, len(result.Issues))
	for i, atlassianIssue := range result.Issues {
		issues[i] = *convertAtlassianIssue(atlassianIssue)
	}

	response := &types.IssueListResponse{
		Issues:     issues,
		Total:      result.Total,
		StartAt:    result.StartAt,
		MaxResults: result.MaxResults,
	}

	return response, nil
}

// convertAtlassianIssue converts a go-atlassian issue to our internal type
func convertAtlassianIssue(issue *models.IssueScheme) *types.Issue {
	result := &types.Issue{
		ID:  issue.ID,
		Key: issue.Key,
	}

	if issue.Fields != nil {
		result.Summary = issue.Fields.Summary

		if issue.Fields.Status != nil {
			result.Status = issue.Fields.Status.Name
		}

		if issue.Fields.IssueType != nil {
			result.IssueType = issue.Fields.IssueType.Name
		}

		if issue.Fields.Priority != nil {
			result.Priority = issue.Fields.Priority.Name
		}

		if issue.Fields.Assignee != nil {
			result.Assignee = issue.Fields.Assignee.DisplayName
		}

		if issue.Fields.Reporter != nil {
			result.Reporter = issue.Fields.Reporter.DisplayName
		}

		if issue.Fields.Project != nil {
			result.Project = issue.Fields.Project.Key
		}

		if issue.Fields.Created != "" {
			if created, err := time.Parse(time.RFC3339, issue.Fields.Created); err == nil {
				result.Created = created
			}
		}

		if issue.Fields.Updated != "" {
			if updated, err := time.Parse(time.RFC3339, issue.Fields.Updated); err == nil {
				result.Updated = updated
			}
		}

		if len(issue.Fields.Labels) > 0 {
			result.Labels = issue.Fields.Labels
		}

		if len(issue.Fields.Components) > 0 {
			components := make([]string, len(issue.Fields.Components))
			for i, comp := range issue.Fields.Components {
				components[i] = comp.Name
			}
			result.Components = components
		}
	}

	return result
}

// buildJQLFromOptions builds a JQL query from the list options
func buildJQLFromOptions(opts *types.IssueListOptions) string {
	var clauses []string

	if opts.Project != "" {
		clauses = append(clauses, fmt.Sprintf("project = %s", opts.Project))
	}

	if len(opts.Status) > 0 {
		statuses := make([]string, len(opts.Status))
		for i, status := range opts.Status {
			statuses[i] = fmt.Sprintf(`"%s"`, status)
		}
		clauses = append(clauses, fmt.Sprintf("status IN (%s)", strings.Join(statuses, ",")))
	}

	if len(opts.IssueType) > 0 {
		types := make([]string, len(opts.IssueType))
		for i, issueType := range opts.IssueType {
			types[i] = fmt.Sprintf(`"%s"`, issueType)
		}
		clauses = append(clauses, fmt.Sprintf("issuetype IN (%s)", strings.Join(types, ",")))
	}

	if opts.Assignee != "" {
		clauses = append(clauses, fmt.Sprintf("assignee = %s", opts.Assignee))
	}

	if len(clauses) == 0 {
		return "ORDER BY created DESC"
	}

	return strings.Join(clauses, " AND ") + " ORDER BY created DESC"
}

// ListProjects lists JIRA projects with caching
func (c *AtlassianJiraClient) ListProjects(ctx context.Context, opts *types.ProjectListOptions) (*types.ProjectListResponse, error) {
	// Try cache first
	cacheKey := "projects_list"
	var cached types.ProjectListResponse
	if cache, err := cache.NewCache(); err == nil {
		if found, _ := cache.Get(cacheKey, &cached); found {
			return &cached, nil
		}
	}

	// Mock implementation for demonstration
	projects := []types.Project{
		{
			ID:          "10000",
			Key:         "DEMO",
			Name:        "Demo Project",
			Description: "Demonstration project for testing",
			Lead:        "admin",
			ProjectType: "software",
		},
		{
			ID:          "10001",
			Key:         "DEV",
			Name:        "Development",
			Description: "Development project",
			Lead:        "dev-lead",
			ProjectType: "software",
		},
	}

	response := &types.ProjectListResponse{
		Projects:   projects,
		Total:      2,
		StartAt:    0,
		MaxResults: 50,
	}

	// Cache the result
	if cache, err := cache.NewCache(); err == nil {
		cache.Set(cacheKey, response, 5*time.Minute)
	}

	return response, nil
}

// GetProject retrieves a JIRA project by key
func (c *AtlassianJiraClient) GetProject(ctx context.Context, key string) (*types.Project, error) {
	// Mock implementation for demonstration
	return &types.Project{
		ID:          "10000",
		Key:         key,
		Name:        "Demo Project",
		Description: "Demonstration project for testing",
		Lead:        "admin",
		ProjectType: "software",
	}, nil
}