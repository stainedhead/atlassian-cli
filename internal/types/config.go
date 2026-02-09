package types

import "time"

// Config represents the application configuration
type Config struct {
	APIEndpoint            string        `mapstructure:"api_endpoint"`
	Email                  string        `mapstructure:"email"`
	Token                  string        `mapstructure:"token"`
	DefaultJiraProject     string        `mapstructure:"default_jira_project"`
	DefaultConfluenceSpace string        `mapstructure:"default_confluence_space"`
	Timeout                time.Duration `mapstructure:"timeout"`
	Output                 string        `mapstructure:"output" validate:"oneof=json table yaml"`
	Debug                  bool          `mapstructure:"debug"`
	Verbose                bool          `mapstructure:"verbose"`
}

// Profile represents a configuration profile for different environments
type Profile struct {
	Name                   string `mapstructure:"name" validate:"required"`
	APIEndpoint            string `mapstructure:"api_endpoint" validate:"required,url"`
	Email                  string `mapstructure:"email" validate:"required,email"`
	Token                  string `mapstructure:"token" validate:"required"`
	DefaultJiraProject     string `mapstructure:"default_jira_project"`
	DefaultConfluenceSpace string `mapstructure:"default_confluence_space"`
	Active                 bool   `mapstructure:"active"`
}

// AuthCredentials represents stored authentication credentials
type AuthCredentials struct {
	ServerURL string `json:"server_url" validate:"required,url"`
	Email     string `json:"email" validate:"required,email"`
	Token     string `json:"token" validate:"required"`
}

// UserInfo represents an authenticated Atlassian user profile
type UserInfo struct {
	AccountID   string `json:"accountId" validate:"required"`
	DisplayName string `json:"displayName" validate:"required"`
	Email       string `json:"emailAddress"`
	Active      bool   `json:"active"`
}
