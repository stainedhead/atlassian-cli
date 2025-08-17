package types

import "time"

// Issue represents a JIRA issue
type Issue struct {
	ID          string    `json:"id"`
	Key         string    `json:"key"`
	Summary     string    `json:"summary"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	IssueType   string    `json:"issueType"`
	Priority    string    `json:"priority"`
	Assignee    string    `json:"assignee"`
	Reporter    string    `json:"reporter"`
	Project     string    `json:"project"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
	Labels      []string  `json:"labels"`
	Components  []string  `json:"components"`
}

// CreateIssueRequest represents a request to create a new issue
type CreateIssueRequest struct {
	Project     string            `json:"project" validate:"required"`
	Summary     string            `json:"summary" validate:"required"`
	Description string            `json:"description"`
	IssueType   string            `json:"issueType" validate:"required"`
	Priority    string            `json:"priority"`
	Assignee    string            `json:"assignee"`
	Labels      []string          `json:"labels"`
	Components  []string          `json:"components"`
	CustomFields map[string]interface{} `json:"customFields"`
}

// UpdateIssueRequest represents a request to update an existing issue
type UpdateIssueRequest struct {
	Summary     *string   `json:"summary,omitempty"`
	Description *string   `json:"description,omitempty"`
	Priority    *string   `json:"priority,omitempty"`
	Assignee    *string   `json:"assignee,omitempty"`
	Status      *string   `json:"status,omitempty"`
	Labels      *[]string `json:"labels,omitempty"`
	Components  *[]string `json:"components,omitempty"`
}

// IssueListOptions represents options for listing issues
type IssueListOptions struct {
	Project    string   `json:"project"`
	Status     []string `json:"status"`
	IssueType  []string `json:"issueType"`
	Assignee   string   `json:"assignee"`
	MaxResults int      `json:"maxResults"`
	StartAt    int      `json:"startAt"`
	JQL        string   `json:"jql"`
}

// IssueListResponse represents the response from listing issues
type IssueListResponse struct {
	Issues     []Issue `json:"issues"`
	Total      int     `json:"total"`
	StartAt    int     `json:"startAt"`
	MaxResults int     `json:"maxResults"`
}