package types

// Project represents a JIRA project
type Project struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Lead        string `json:"lead"`
	ProjectType string `json:"projectType"`
}

// ProjectListOptions represents options for listing projects
type ProjectListOptions struct {
	MaxResults int `json:"maxResults"`
	StartAt    int `json:"startAt"`
}

// ProjectListResponse represents the response from listing projects
type ProjectListResponse struct {
	Projects   []Project `json:"projects"`
	Total      int       `json:"total"`
	StartAt    int       `json:"startAt"`
	MaxResults int       `json:"maxResults"`
}
