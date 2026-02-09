package types

import "time"

// Page represents a Confluence page
type Page struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	Type     string    `json:"type"`
	SpaceKey string    `json:"spaceKey"`
	Content  string    `json:"content"`
	Version  int       `json:"version"`
	Updated  time.Time `json:"updated"`
}

// CreatePageRequest represents a request to create a new page
type CreatePageRequest struct {
	SpaceKey string `json:"spaceKey" validate:"required"`
	Title    string `json:"title" validate:"required"`
	Content  string `json:"content"`
	ParentID string `json:"parentId"`
}

// UpdatePageRequest represents a request to update an existing page
type UpdatePageRequest struct {
	Title   *string `json:"title,omitempty"`
	Content *string `json:"content,omitempty"`
}

// PageListOptions represents options for listing pages
type PageListOptions struct {
	SpaceKey   string `json:"spaceKey"`
	Title      string `json:"title"`
	MaxResults int    `json:"maxResults"`
	StartAt    int    `json:"startAt"` // Retained for backward compatibility
	Cursor     string `json:"cursor"`  // Cursor-based pagination for Confluence v2
}

// PageListResponse represents the response from listing pages
type PageListResponse struct {
	Pages      []Page `json:"pages"`
	Total      int    `json:"total"`
	StartAt    int    `json:"startAt"`
	MaxResults int    `json:"maxResults"`
	NextCursor string `json:"nextCursor"` // Cursor for next page of results (Confluence v2)
}

// Space represents a Confluence space
type Space struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// SpaceListOptions represents options for listing spaces
type SpaceListOptions struct {
	Type       string `json:"type"`
	MaxResults int    `json:"maxResults"`
	StartAt    int    `json:"startAt"`
}

// SpaceListResponse represents the response from listing spaces
type SpaceListResponse struct {
	Spaces     []Space `json:"spaces"`
	Total      int     `json:"total"`
	StartAt    int     `json:"startAt"`
	MaxResults int     `json:"maxResults"`
}

// PageSearchOptions represents options for searching pages with CQL
type PageSearchOptions struct {
	CQL        string `json:"cql"`
	MaxResults int    `json:"maxResults"`
	StartAt    int    `json:"startAt"`
}

// PageSearchResponse represents the response from searching pages
type PageSearchResponse struct {
	Pages      []Page `json:"pages"`
	Total      int    `json:"total"`
	StartAt    int    `json:"startAt"`
	MaxResults int    `json:"maxResults"`
}
