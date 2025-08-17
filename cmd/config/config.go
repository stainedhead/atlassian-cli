package config

import (
	"atlassian-cli/internal/config"
	"atlassian-cli/internal/types"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewConfigCmd creates the config command with subcommands
func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management",
		Long:  `Manage CLI configuration settings including default projects and spaces`,
	}

	// Add subcommands
	cmd.AddCommand(newSetCmd())
	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newListCmd())

	return cmd
}

// newSetCmd creates the config set command
func newSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Long: `Set a configuration value.

Available keys:
  default_jira_project      - Default JIRA project key
  default_confluence_space  - Default Confluence space key
  output                    - Default output format (json, table, yaml)
  timeout                   - Request timeout (e.g., 30s, 1m)

Examples:
  atlassian-cli config set default_jira_project DEMO
  atlassian-cli config set default_confluence_space DEV
  atlassian-cli config set output json`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]

			// Get config file path
			configPath, err := getConfigPath()
			if err != nil {
				return fmt.Errorf("failed to get config path: %w", err)
			}

			// Load existing config or create new one
			cfg, err := config.LoadConfig(configPath)
			if err != nil {
				// If config doesn't exist, create a new one
				cfg = &types.Config{
					Output: "table",
				}
			}

			// Set the value
			switch key {
			case "default_jira_project":
				cfg.DefaultJiraProject = value
			case "default_confluence_space":
				cfg.DefaultConfluenceSpace = value
			case "output":
				if value != "json" && value != "table" && value != "yaml" {
					return fmt.Errorf("invalid output format: %s (must be json, table, or yaml)", value)
				}
				cfg.Output = value
			case "api_endpoint":
				cfg.APIEndpoint = value
			case "email":
				cfg.Email = value
			default:
				return fmt.Errorf("unknown configuration key: %s", key)
			}

			// Save the config
			if err := config.SaveConfig(configPath, cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "✓ Set %s = %s\n", key, value)
			return nil
		},
	}

	return cmd
}

// newGetCmd creates the config get command
func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Long:  `Get the current value of a configuration setting`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			// Get config file path
			configPath, err := getConfigPath()
			if err != nil {
				return fmt.Errorf("failed to get config path: %w", err)
			}

			// Load config
			cfg, err := config.LoadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Get the value
			var value string
			switch key {
			case "default_jira_project":
				value = cfg.DefaultJiraProject
			case "default_confluence_space":
				value = cfg.DefaultConfluenceSpace
			case "output":
				value = cfg.Output
			case "api_endpoint":
				value = cfg.APIEndpoint
			case "email":
				value = cfg.Email
			default:
				return fmt.Errorf("unknown configuration key: %s", key)
			}

			if value == "" {
				fmt.Fprintf(cmd.OutOrStdout(), "%s is not set\n", key)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\n", value)
			}

			return nil
		},
	}

	return cmd
}

// newListCmd creates the config list command
func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all configuration values",
		Long:  `Display all current configuration settings`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get config file path
			configPath, err := getConfigPath()
			if err != nil {
				return fmt.Errorf("failed to get config path: %w", err)
			}

			// Load config
			cfg, err := config.LoadConfig(configPath)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "No configuration file found\n")
				return nil
			}

			// Display all settings
			fmt.Fprintf(cmd.OutOrStdout(), "Configuration settings:\n\n")
			
			if cfg.APIEndpoint != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "api_endpoint:             %s\n", cfg.APIEndpoint)
			}
			if cfg.Email != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "email:                    %s\n", cfg.Email)
			}
			if cfg.DefaultJiraProject != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "default_jira_project:     %s\n", cfg.DefaultJiraProject)
			}
			if cfg.DefaultConfluenceSpace != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "default_confluence_space: %s\n", cfg.DefaultConfluenceSpace)
			}
			
			fmt.Fprintf(cmd.OutOrStdout(), "output:                   %s\n", cfg.Output)
			fmt.Fprintf(cmd.OutOrStdout(), "timeout:                  %s\n", cfg.Timeout)
			fmt.Fprintf(cmd.OutOrStdout(), "debug:                    %t\n", cfg.Debug)
			fmt.Fprintf(cmd.OutOrStdout(), "verbose:                  %t\n", cfg.Verbose)

			fmt.Fprintf(cmd.OutOrStdout(), "\nConfig file: %s\n", configPath)

			return nil
		},
	}

	return cmd
}

// getConfigPath returns the configuration file path
func getConfigPath() (string, error) {
	// Check if config file is specified via flag
	if configFile := viper.GetString("config"); configFile != "" {
		return configFile, nil
	}

	// Use default path
	return config.GetDefaultConfigPath()
}