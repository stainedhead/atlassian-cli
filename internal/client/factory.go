package client

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"atlassian-cli/internal/confluence"
	"atlassian-cli/internal/jira"
)

// ClientKey uniquely identifies a client instance
type ClientKey struct {
	ServerURL string
	Email     string
}

// Factory manages shared client instances with connection pooling
type Factory struct {
	jiraClients       map[ClientKey]jira.JiraClient
	confluenceClients map[ClientKey]confluence.ConfluenceClient
	mu                sync.RWMutex
	httpClient        *http.Client
}

// NewFactory creates a new client factory with shared HTTP transport
func NewFactory() *Factory {
	// Create shared HTTP client with connection pooling
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	return &Factory{
		jiraClients:       make(map[ClientKey]jira.JiraClient),
		confluenceClients: make(map[ClientKey]confluence.ConfluenceClient),
		httpClient:        httpClient,
	}
}

// GetJiraClient returns a cached or new JIRA client
func (f *Factory) GetJiraClient(ctx context.Context, serverURL, email, token string) (jira.JiraClient, error) {
	key := ClientKey{
		ServerURL: serverURL,
		Email:     email,
	}

	// Check cache with read lock
	f.mu.RLock()
	if client, exists := f.jiraClients[key]; exists {
		f.mu.RUnlock()
		return client, nil
	}
	f.mu.RUnlock()

	// Create new client with write lock
	f.mu.Lock()
	defer f.mu.Unlock()

	// Double-check after acquiring write lock (another goroutine might have created it)
	if client, exists := f.jiraClients[key]; exists {
		return client, nil
	}

	// Create new JIRA client
	client, err := jira.NewAtlassianJiraClient(serverURL, email, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create JIRA client: %w", err)
	}

	f.jiraClients[key] = client
	return client, nil
}

// GetConfluenceClient returns a cached or new Confluence client
func (f *Factory) GetConfluenceClient(ctx context.Context, serverURL, email, token string) (confluence.ConfluenceClient, error) {
	key := ClientKey{
		ServerURL: serverURL,
		Email:     email,
	}

	// Check cache with read lock
	f.mu.RLock()
	if client, exists := f.confluenceClients[key]; exists {
		f.mu.RUnlock()
		return client, nil
	}
	f.mu.RUnlock()

	// Create new client with write lock
	f.mu.Lock()
	defer f.mu.Unlock()

	// Double-check after acquiring write lock
	if client, exists := f.confluenceClients[key]; exists {
		return client, nil
	}

	// Create new Confluence client
	client, err := confluence.NewAtlassianConfluenceClient(serverURL, email, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create Confluence client: %w", err)
	}

	f.confluenceClients[key] = client
	return client, nil
}

// ClearCache removes all cached clients
func (f *Factory) ClearCache() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.jiraClients = make(map[ClientKey]jira.JiraClient)
	f.confluenceClients = make(map[ClientKey]confluence.ConfluenceClient)
}

// GetHTTPClient returns the shared HTTP client for custom requests
func (f *Factory) GetHTTPClient() *http.Client {
	return f.httpClient
}
