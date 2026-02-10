package confluence

import (
	"atlassian-cli/internal/types"
	"context"
	"fmt"
	"strconv"
	"time"

	confluence "github.com/ctreminiom/go-atlassian/confluence"
	"github.com/ctreminiom/go-atlassian/pkg/infra/models"
)

// ConfluenceClient defines the interface for Confluence operations
type ConfluenceClient interface {
	CreatePage(ctx context.Context, req *types.CreatePageRequest) (*types.Page, error)
	GetPage(ctx context.Context, id string) (*types.Page, error)
	UpdatePage(ctx context.Context, id string, req *types.UpdatePageRequest) (*types.Page, error)
	ListPages(ctx context.Context, opts *types.PageListOptions) (*types.PageListResponse, error)
	SearchPages(ctx context.Context, opts *types.PageSearchOptions) (*types.PageSearchResponse, error)
	ListSpaces(ctx context.Context, opts *types.SpaceListOptions) (*types.SpaceListResponse, error)
}

// AtlassianConfluenceClient implements ConfluenceClient using the go-atlassian v1 library
type AtlassianConfluenceClient struct {
	client *confluence.Client
}

// NewAtlassianConfluenceClient creates a new Confluence client using v1 API
func NewAtlassianConfluenceClient(baseURL, email, token string) (*AtlassianConfluenceClient, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if token == "" {
		return nil, fmt.Errorf("token is required")
	}

	// Create the client instance using v1 API
	instance, err := confluence.New(nil, baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create Confluence client: %w", err)
	}

	// Set authentication
	instance.Auth.SetBasicAuth(email, token)

	return &AtlassianConfluenceClient{
		client: instance,
	}, nil
}

// CreatePage creates a new Confluence page
func (c *AtlassianConfluenceClient) CreatePage(ctx context.Context, req *types.CreatePageRequest) (*types.Page, error) {
	if req == nil {
		return nil, fmt.Errorf("create page request cannot be nil")
	}

	// Build the page creation payload
	payload := &models.ContentScheme{
		Type:  "page",
		Title: req.Title,
		Space: &models.SpaceScheme{
			Key: req.SpaceKey,
		},
		Body: &models.BodyScheme{
			Storage: &models.BodyNodeScheme{
				Value:          req.Content,
				Representation: "storage",
			},
		},
	}

	// Set ancestors (parent page) if provided
	if req.ParentID != "" {
		payload.Ancestors = []*models.ContentScheme{
			{
				ID: req.ParentID,
			},
		}
	}

	// Create the page using the v1 API
	result, response, err := c.client.Content.Create(ctx, payload)
	if err != nil {
		if response != nil {
			return nil, fmt.Errorf("failed to create page (status %d): %w", response.StatusCode, err)
		}
		return nil, fmt.Errorf("failed to create page: %w", err)
	}

	// Convert the response to our internal type
	return convertContentSchemeToPage(result), nil
}

// GetPage retrieves a Confluence page by ID
func (c *AtlassianConfluenceClient) GetPage(ctx context.Context, id string) (*types.Page, error) {
	if id == "" {
		return nil, fmt.Errorf("page ID is required")
	}

	// Get the page with body content and version
	result, response, err := c.client.Content.Get(ctx, id, []string{"body.storage", "version", "space"}, 0)
	if err != nil {
		if response != nil {
			return nil, fmt.Errorf("failed to get page (status %d): %w", response.StatusCode, err)
		}
		return nil, fmt.Errorf("failed to get page: %w", err)
	}

	return convertContentSchemeToPage(result), nil
}

// UpdatePage updates an existing Confluence page
func (c *AtlassianConfluenceClient) UpdatePage(ctx context.Context, id string, req *types.UpdatePageRequest) (*types.Page, error) {
	if id == "" {
		return nil, fmt.Errorf("page ID is required")
	}
	if req == nil {
		return nil, fmt.Errorf("update page request cannot be nil")
	}

	// Get current page to retrieve version and existing data
	currentPage, response, err := c.client.Content.Get(ctx, id, []string{"body.storage", "version", "space"}, 0)
	if err != nil {
		if response != nil {
			return nil, fmt.Errorf("failed to get current page (status %d): %w", response.StatusCode, err)
		}
		return nil, fmt.Errorf("failed to get current page: %w", err)
	}

	// Build update payload with incremented version
	payload := &models.ContentScheme{
		Type:  "page",
		Title: currentPage.Title,
		Version: &models.ContentVersionScheme{
			Number: currentPage.Version.Number + 1,
		},
	}

	// Set title if provided
	if req.Title != nil {
		payload.Title = *req.Title
	}

	// Set content if provided
	if req.Content != nil {
		payload.Body = &models.BodyScheme{
			Storage: &models.BodyNodeScheme{
				Value:          *req.Content,
				Representation: "storage",
			},
		}
	} else if currentPage.Body != nil && currentPage.Body.Storage != nil {
		// Keep existing content if not updating
		payload.Body = currentPage.Body
	}

	// Update the page
	result, response, err := c.client.Content.Update(ctx, id, payload)
	if err != nil {
		if response != nil {
			return nil, fmt.Errorf("failed to update page (status %d): %w", response.StatusCode, err)
		}
		return nil, fmt.Errorf("failed to update page: %w", err)
	}

	return convertContentSchemeToPage(result), nil
}

// ListPages lists pages in a space
func (c *AtlassianConfluenceClient) ListPages(ctx context.Context, opts *types.PageListOptions) (*types.PageListResponse, error) {
	if opts == nil {
		opts = &types.PageListOptions{}
	}

	// Set defaults
	maxResults := opts.MaxResults
	if maxResults <= 0 {
		maxResults = 25
	}

	startAt := opts.StartAt
	if startAt < 0 {
		startAt = 0
	}

	// Note: Confluence v1 API doesn't support cursor pagination natively
	// Cursor is ignored for now; we use offset-based pagination
	if opts.Cursor != "" {
		// Could decode cursor and extract offset, but for simplicity we ignore it
		// and rely on startAt parameter
	}

	// Build query options
	options := &models.GetContentOptionsScheme{}

	// Add space filter if provided
	if opts.SpaceKey != "" {
		options.SpaceKey = opts.SpaceKey
	}

	// Add title filter if provided
	if opts.Title != "" {
		options.Title = opts.Title
	}

	// Get pages
	result, response, err := c.client.Content.Gets(ctx, options, startAt, maxResults)
	if err != nil {
		if response != nil {
			return nil, fmt.Errorf("failed to list pages (status %d): %w", response.StatusCode, err)
		}
		return nil, fmt.Errorf("failed to list pages: %w", err)
	}

	// Convert results - filter by type=page since API doesn't support type filter directly
	pages := make([]types.Page, 0, len(result.Results))
	for _, contentScheme := range result.Results {
		if contentScheme.Type == "page" {
			pages = append(pages, *convertContentSchemeToPage(contentScheme))
		}
	}

	// Generate next cursor if there are more results
	// Note: v1 API doesn't return total count, so we use result.Size as an indicator
	var nextCursor string
	if result.Size == maxResults {
		// More results likely available
		nextCursor = strconv.Itoa(startAt + maxResults)
	}

	// Estimate total based on what we know
	total := startAt + result.Size
	if result.Size < maxResults {
		// This is the last page
		total = startAt + result.Size
	} else {
		// More pages exist, we don't know the exact total
		total = startAt + result.Size + 1 // At least one more
	}

	return &types.PageListResponse{
		Pages:      pages,
		Total:      total,
		StartAt:    startAt,
		MaxResults: maxResults,
		NextCursor: nextCursor,
	}, nil
}

// SearchPages searches pages using CQL
func (c *AtlassianConfluenceClient) SearchPages(ctx context.Context, opts *types.PageSearchOptions) (*types.PageSearchResponse, error) {
	if opts == nil {
		return nil, fmt.Errorf("search options cannot be nil")
	}

	if opts.CQL == "" {
		return nil, fmt.Errorf("CQL query is required")
	}

	// Set defaults
	maxResults := opts.MaxResults
	if maxResults <= 0 {
		maxResults = 25
	}

	startAt := opts.StartAt
	if startAt < 0 {
		startAt = 0
	}

	// Build search options
	searchOptions := &models.SearchContentOptions{
		Limit: maxResults,
		Start: startAt,
	}

	// Execute CQL search
	result, response, err := c.client.Search.Content(ctx, opts.CQL, searchOptions)
	if err != nil {
		if response != nil {
			return nil, fmt.Errorf("failed to search pages (status %d): %w", response.StatusCode, err)
		}
		return nil, fmt.Errorf("failed to search pages: %w", err)
	}

	// Convert results
	pages := make([]types.Page, 0, len(result.Results))
	for _, searchResult := range result.Results {
		if searchResult.Content != nil {
			pages = append(pages, *convertContentSchemeToPage(searchResult.Content))
		}
	}

	return &types.PageSearchResponse{
		Pages:      pages,
		Total:      result.TotalSize,
		StartAt:    startAt,
		MaxResults: maxResults,
	}, nil
}

// ListSpaces lists Confluence spaces
func (c *AtlassianConfluenceClient) ListSpaces(ctx context.Context, opts *types.SpaceListOptions) (*types.SpaceListResponse, error) {
	if opts == nil {
		opts = &types.SpaceListOptions{}
	}

	// Set defaults
	maxResults := opts.MaxResults
	if maxResults <= 0 {
		maxResults = 25
	}

	startAt := opts.StartAt
	if startAt < 0 {
		startAt = 0
	}

	// Note: Cursor is ignored for v1 API (uses offset-based pagination)
	if opts.Cursor != "" {
		// Could decode cursor and extract offset, but for simplicity we ignore it
	}

	// Build query options
	options := &models.GetSpacesOptionScheme{}

	// Add type filter if provided
	if opts.Type != "" {
		options.SpaceType = opts.Type
	}

	// Get spaces
	result, response, err := c.client.Space.Gets(ctx, options, startAt, maxResults)
	if err != nil {
		if response != nil {
			return nil, fmt.Errorf("failed to list spaces (status %d): %w", response.StatusCode, err)
		}
		return nil, fmt.Errorf("failed to list spaces: %w", err)
	}

	// Convert results
	spaces := make([]types.Space, 0, len(result.Results))
	for _, spaceScheme := range result.Results {
		space := types.Space{
			ID:          strconv.Itoa(spaceScheme.ID),
			Key:         spaceScheme.Key,
			Name:        spaceScheme.Name,
			Type:        spaceScheme.Type,
			Description: "", // v1 API SpaceScheme doesn't include description by default
		}

		spaces = append(spaces, space)
	}

	// Generate next cursor if there are more results
	var nextCursor string
	if result.Size == maxResults {
		// More results likely available
		nextCursor = strconv.Itoa(startAt + maxResults)
	}

	// Estimate total based on what we know
	total := startAt + result.Size
	if result.Size < maxResults {
		// This is the last page
		total = startAt + result.Size
	} else {
		// More pages exist, we don't know the exact total
		total = startAt + result.Size + 1 // At least one more
	}

	return &types.SpaceListResponse{
		Spaces:     spaces,
		Total:      total,
		StartAt:    startAt,
		MaxResults: maxResults,
		NextCursor: nextCursor,
	}, nil
}

// convertContentSchemeToPage converts go-atlassian ContentScheme to our internal Page type
func convertContentSchemeToPage(scheme *models.ContentScheme) *types.Page {
	if scheme == nil {
		return nil
	}

	page := &types.Page{
		ID:    scheme.ID,
		Title: scheme.Title,
		Type:  scheme.Type,
	}

	// Extract space key if available
	if scheme.Space != nil {
		page.SpaceKey = scheme.Space.Key
	}

	// Extract content if available
	if scheme.Body != nil && scheme.Body.Storage != nil {
		page.Content = scheme.Body.Storage.Value
	}

	// Extract version if available
	if scheme.Version != nil {
		page.Version = scheme.Version.Number

		// Parse updated time if available
		if scheme.Version.When != "" {
			if t, err := time.Parse(time.RFC3339, scheme.Version.When); err == nil {
				page.Updated = t
			}
		}
	}

	return page
}
