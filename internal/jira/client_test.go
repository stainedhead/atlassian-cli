package jira

import (
	"atlassian-cli/internal/types"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockJiraClient is a mock implementation of JiraClient for testing
type MockJiraClient struct {
	mock.Mock
}

func (m *MockJiraClient) CreateIssue(ctx context.Context, req *types.CreateIssueRequest) (*types.Issue, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Issue), args.Error(1)
}

func (m *MockJiraClient) GetIssue(ctx context.Context, key string) (*types.Issue, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Issue), args.Error(1)
}

func (m *MockJiraClient) UpdateIssue(ctx context.Context, key string, req *types.UpdateIssueRequest) (*types.Issue, error) {
	args := m.Called(ctx, key, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Issue), args.Error(1)
}

func (m *MockJiraClient) ListIssues(ctx context.Context, opts *types.IssueListOptions) (*types.IssueListResponse, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.IssueListResponse), args.Error(1)
}

func (m *MockJiraClient) ListProjects(ctx context.Context, opts *types.ProjectListOptions) (*types.ProjectListResponse, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.ProjectListResponse), args.Error(1)
}

func (m *MockJiraClient) GetProject(ctx context.Context, key string) (*types.Project, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Project), args.Error(1)
}

func (m *MockJiraClient) SearchIssues(ctx context.Context, opts *types.IssueSearchOptions) (*types.IssueSearchResponse, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.IssueSearchResponse), args.Error(1)
}

func (m *MockJiraClient) GetTransitions(ctx context.Context, issueKey string) ([]types.Transition, error) {
	args := m.Called(ctx, issueKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]types.Transition), args.Error(1)
}

func (m *MockJiraClient) TransitionIssue(ctx context.Context, issueKey string, transitionID string) error {
	args := m.Called(ctx, issueKey, transitionID)
	return args.Error(0)
}

func TestMockJiraClient(t *testing.T) {
	// Test that the mock client implements the JiraClient interface
	var client JiraClient = &MockJiraClient{}
	assert.NotNil(t, client)
}
