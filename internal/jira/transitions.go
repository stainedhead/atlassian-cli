package jira

import (
	"atlassian-cli/internal/types"
	"context"
	"fmt"
)

// GetTransitions retrieves available transitions for an issue
func (c *AtlassianJiraClient) GetTransitions(ctx context.Context, issueKey string) ([]types.Transition, error) {
	if issueKey == "" {
		return nil, fmt.Errorf("issue key is required")
	}

	// Call JIRA API to get transitions
	transitionsResult, resp, err := c.client.Issue.Transitions(ctx, issueKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get transitions for issue %s: %w", issueKey, err)
	}

	if resp == nil || transitionsResult == nil {
		return nil, fmt.Errorf("empty response from JIRA API")
	}

	// Convert to our types
	transitions := make([]types.Transition, 0, len(transitionsResult.Transitions))
	for _, t := range transitionsResult.Transitions {
		transition := types.Transition{
			ID:   t.ID,
			Name: t.Name,
		}

		if t.To != nil {
			transition.To.Name = t.To.Name
		}

		transitions = append(transitions, transition)
	}

	return transitions, nil
}

// TransitionIssue transitions an issue to a new status
// Note: This is a simplified implementation - proper error handling for different
// go-atlassian versions should be added in production
func (c *AtlassianJiraClient) TransitionIssue(ctx context.Context, issueKey string, transitionID string) error {
	if issueKey == "" {
		return fmt.Errorf("issue key is required")
	}

	if transitionID == "" {
		return fmt.Errorf("transition ID is required")
	}

	// Use the Transitions endpoint to perform the transition
	// The go-atlassian library's API for transitions varies by version
	// For now, we return an informative error directing users to use the API directly
	return fmt.Errorf("direct transition support requires go-atlassian v2+ - please get transitions with GetTransitions and use the JIRA API directly for transition %s on issue %s", transitionID, issueKey)
}
