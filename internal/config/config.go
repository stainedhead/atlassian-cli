package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"atlassian-cli/internal/types"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

var validate = validator.New()

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configFile string) (*types.Config, error) {
	v := viper.New()

	// Set config file if provided
	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		// Default config locations
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}

		v.AddConfigPath(filepath.Join(home, ".atlassian-cli"))
		v.AddConfigPath(".")
		v.SetConfigType("yaml")
		v.SetConfigName("config")
	}

	// Environment variable configuration
	v.SetEnvPrefix("ATLASSIAN")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// Set defaults
	v.SetDefault("timeout", 30*time.Second)
	v.SetDefault("output", "table")
	v.SetDefault("default_jira_project", "")
	v.SetDefault("default_confluence_space", "")
	v.SetDefault("debug", false)
	v.SetDefault("verbose", false)

	// Read config file if it exists
	if err := v.ReadInConfig(); err != nil {
		// Config file not found is not an error
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config types.Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("config unmarshal failed: %w", err)
	}

	// Validate config
	if err := validate.Struct(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// SaveConfig saves configuration to file
func SaveConfig(configFile string, config *types.Config) error {
	// Ensure directory exists
	dir := filepath.Dir(configFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	v := viper.New()
	v.SetConfigFile(configFile)
	v.SetConfigType("yaml")

	// Set all config values
	v.Set("api_endpoint", config.APIEndpoint)
	v.Set("email", config.Email)
	v.Set("token", config.Token)
	v.Set("default_jira_project", config.DefaultJiraProject)
	v.Set("default_confluence_space", config.DefaultConfluenceSpace)
	v.Set("timeout", config.Timeout)
	v.Set("output", config.Output)
	v.Set("debug", config.Debug)
	v.Set("verbose", config.Verbose)

	if err := v.WriteConfig(); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetDefaultConfigPath returns the default configuration file path
func GetDefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(home, ".atlassian-cli")
	return filepath.Join(configDir, "config.yaml"), nil
}