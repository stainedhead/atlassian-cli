package confluence

import (
	"atlassian-cli/internal/types"
	"context"
	"fmt"
	"time"
)

// ConfluenceClient defines the interface for Confluence operations
type ConfluenceClient interface {
	CreatePage(ctx context.Context, req *types.CreatePageRequest) (*types.Page, error)
	GetPage(ctx context.Context, id string) (*types.Page, error)
	UpdatePage(ctx context.Context, id string, req *types.UpdatePageRequest) (*types.Page, error)
	ListPages(ctx context.Context, opts *types.PageListOptions) (*types.PageListResponse, error)
	ListSpaces(ctx context.Context, opts *types.SpaceListOptions) (*types.SpaceListResponse, error)
}

// MockConfluenceClient implements ConfluenceClient for demonstration
type MockConfluenceClient struct {
	baseURL string
	email   string
	token   string
}

// NewAtlassianConfluenceClient creates a new Confluence client
func NewAtlassianConfluenceClient(baseURL, email, token string) (*MockConfluenceClient, error) {
	if baseURL == "" || email == "" || token == "" {
		return nil, fmt.Errorf("baseURL, email, and token are required")
	}

	return &MockConfluenceClient{
		baseURL: baseURL,
		email:   email,
		token:   token,
	}, nil
}

// CreatePage creates a new Confluence page
func (c *MockConfluenceClient) CreatePage(ctx context.Context, req *types.CreatePageRequest) (*types.Page, error) {
	// Mock implementation for demonstration
	return &types.Page{
		ID:       "123456",
		Title:    req.Title,
		Type:     "page",
		SpaceKey: req.SpaceKey,
		Content:  req.Content,
		Version:  1,
		Updated:  time.Now(),
	}, nil
}

// GetPage retrieves a Confluence page by ID
func (c *MockConfluenceClient) GetPage(ctx context.Context, id string) (*types.Page, error) {
	// Mock implementation for demonstration
	return &types.Page{
		ID:       id,
		Title:    "Sample Page",
		Type:     "page",
		SpaceKey: "DEMO",
		Content:  "<p>This is sample page content</p>",
		Version:  2,
		Updated:  time.Now(),
	}, nil
}

// UpdatePage updates an existing Confluence page
func (c *MockConfluenceClient) UpdatePage(ctx context.Context, id string, req *types.UpdatePageRequest) (*types.Page, error) {
	// Mock implementation for demonstration
	title := "Updated Page"
	if req.Title != nil {
		title = *req.Title
	}

	content := "<p>Updated content</p>"
	if req.Content != nil {
		content = *req.Content
	}

	return &types.Page{
		ID:       id,
		Title:    title,
		Type:     "page",
		SpaceKey: "DEMO",
		Content:  content,
		Version:  3,
		Updated:  time.Now(),
	}, nil
}

// ListPages lists pages in a space
func (c *MockConfluenceClient) ListPages(ctx context.Context, opts *types.PageListOptions) (*types.PageListResponse, error) {
	// Mock implementation for demonstration
	pages := []types.Page{
		{
			ID:       "123456",
			Title:    "Getting Started",
			Type:     "page",
			SpaceKey: "DEMO",
			Version:  1,
			Updated:  time.Now().Add(-24 * time.Hour),
		},
		{
			ID:       "123457",
			Title:    "API Documentation",
			Type:     "page",
			SpaceKey: "DEMO",
			Version:  2,
			Updated:  time.Now().Add(-12 * time.Hour),
		},
	}

	return &types.PageListResponse{
		Pages:      pages,
		Total:      2,
		StartAt:    0,
		MaxResults: 25,
	}, nil
}

// ListSpaces lists Confluence spaces
func (c *MockConfluenceClient) ListSpaces(ctx context.Context, opts *types.SpaceListOptions) (*types.SpaceListResponse, error) {
	// Mock implementation for demonstration
	spaces := []types.Space{
		{
			ID:          "1",
			Key:         "DEMO",
			Name:        "Demo Space",
			Type:        "global",
			Description: "Demonstration space for testing",
		},
		{
			ID:          "2",
			Key:         "DEV",
			Name:        "Development",
			Type:        "global",
			Description: "Development documentation",
		},
	}

	return &types.SpaceListResponse{
		Spaces:     spaces,
		Total:      2,
		StartAt:    0,
		MaxResults: 25,
	}, nil
}

