package auth

import (
	"atlassian-cli/internal/auth"
	"atlassian-cli/internal/types"
	"context"
	"fmt"
	"net/mail"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

// NewAuthCmd creates the auth command with subcommands
func NewAuthCmd(tokenManager auth.TokenManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication commands",
		Long:  `Manage authentication with Atlassian instances`,
	}

	// Add subcommands
	cmd.AddCommand(newLoginCmd(tokenManager))
	cmd.AddCommand(newLogoutCmd(tokenManager))
	cmd.AddCommand(newStatusCmd(tokenManager))

	return cmd
}

// newLoginCmd creates the login command
func newLoginCmd(tokenManager auth.TokenManager) *cobra.Command {
	var (
		serverURL string
		email     string
		token     string
		noStore   bool
	)

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Atlassian instance",
		Long: `Authenticate with an Atlassian instance using email and API token.

Create an API token at: https://id.atlassian.com/manage/api-tokens

Example:
  atlassian-cli auth login --server https://your-domain.atlassian.net --email user@example.com --token your-api-token`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateAuthFlags(serverURL, email, token); err != nil {
				return err
			}

			creds := &types.AuthCredentials{
				ServerURL: serverURL,
				Email:     email,
				Token:     token,
			}

			if !noStore {
				if err := tokenManager.Store(context.Background(), creds); err != nil {
					return fmt.Errorf("failed to store credentials: %w", err)
				}
			}

			fmt.Fprintf(cmd.OutOrStdout(), "✓ Authenticated as %s\n", email)
			if !noStore {
				fmt.Fprintf(cmd.OutOrStdout(), "  Credentials stored securely\n")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&serverURL, "server", "", "Atlassian instance URL (required)")
	cmd.Flags().StringVar(&email, "email", "", "User email (required)")
	cmd.Flags().StringVar(&token, "token", "", "API token (required)")
	cmd.Flags().BoolVar(&noStore, "no-store", false, "Don't store credentials")

	cmd.MarkFlagRequired("server")
	cmd.MarkFlagRequired("email")
	cmd.MarkFlagRequired("token")

	return cmd
}

// newLogoutCmd creates the logout command
func newLogoutCmd(tokenManager auth.TokenManager) *cobra.Command {
	var serverURL string

	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Clear stored credentials",
		Long:  `Remove stored authentication credentials for the specified server`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if serverURL == "" {
				return fmt.Errorf("server URL is required")
			}

			if err := tokenManager.Delete(context.Background(), serverURL); err != nil {
				return fmt.Errorf("failed to delete credentials: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "✓ Logged out from %s\n", serverURL)
			return nil
		},
	}

	cmd.Flags().StringVar(&serverURL, "server", "", "Atlassian instance URL (required)")
	cmd.MarkFlagRequired("server")

	return cmd
}

// newStatusCmd creates the status command
func newStatusCmd(tokenManager auth.TokenManager) *cobra.Command {
	var serverURL string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		Long:  `Display current authentication status for the specified server`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if serverURL == "" {
				return fmt.Errorf("server URL is required")
			}

			creds, err := tokenManager.Get(context.Background(), serverURL)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "Not authenticated for %s\n", serverURL)
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "✓ Authenticated as %s for %s\n", creds.Email, serverURL)
			return nil
		},
	}

	cmd.Flags().StringVar(&serverURL, "server", "", "Atlassian instance URL (required)")
	cmd.MarkFlagRequired("server")

	return cmd
}

// validateAuthFlags validates authentication flags
func validateAuthFlags(serverURL, email, token string) error {
	if serverURL == "" {
		return fmt.Errorf("server URL is required")
	}

	// Validate URL format
	if _, err := url.Parse(serverURL); err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}

	// Ensure URL has scheme
	if !strings.HasPrefix(serverURL, "http://") && !strings.HasPrefix(serverURL, "https://") {
		return fmt.Errorf("server URL must include protocol (https://)")
	}

	if email == "" {
		return fmt.Errorf("email is required")
	}

	// Validate email format
	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("invalid email format: %w", err)
	}

	if token == "" {
		return fmt.Errorf("API token is required")
	}

	return nil
}