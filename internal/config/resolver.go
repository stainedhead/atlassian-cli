package config

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	ErrNoProjectConfigured = errors.New("no JIRA project configured")
	ErrNoSpaceConfigured   = errors.New("no Confluence space configured")
)

// ResolveProject resolves JIRA project using command flag > env var > config > error
func ResolveProject(cmd *cobra.Command) (string, error) {
	// 1. Command-specific --project flag (highest priority)
	if project, _ := cmd.Flags().GetString("project"); project != "" {
		return project, nil
	}

	// 2. Environment variable
	if project := os.Getenv("ATLASSIAN_DEFAULT_JIRA_PROJECT"); project != "" {
		return project, nil
	}

	// 3. Configuration file
	if project := viper.GetString("default_jira_project"); project != "" {
		return project, nil
	}

	// 4. No configuration found
	return "", ErrNoProjectConfigured
}

// ResolveSpace resolves Confluence space using command flag > env var > config > error
func ResolveSpace(cmd *cobra.Command) (string, error) {
	// 1. Command-specific --space flag (highest priority)
	if space, _ := cmd.Flags().GetString("space"); space != "" {
		return space, nil
	}

	// 2. Environment variable
	if space := os.Getenv("ATLASSIAN_DEFAULT_CONFLUENCE_SPACE"); space != "" {
		return space, nil
	}

	// 3. Configuration file
	if space := viper.GetString("default_confluence_space"); space != "" {
		return space, nil
	}

	// 4. No configuration found
	return "", ErrNoSpaceConfigured
}

