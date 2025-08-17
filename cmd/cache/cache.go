package cache

import (
	"atlassian-cli/internal/cache"
	"fmt"

	"github.com/spf13/cobra"
)

// NewCacheCmd creates the cache command with subcommands
func NewCacheCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Cache management operations",
		Long:  `Manage local cache for improved performance`,
	}

	cmd.AddCommand(newClearCmd())
	cmd.AddCommand(newStatusCmd())

	return cmd
}

func newClearCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear all cached data",
		Long:  `Remove all cached entries to force fresh API calls`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := cache.NewCache()
			if err != nil {
				return fmt.Errorf("failed to initialize cache: %w", err)
			}

			if err := c.Clear(); err != nil {
				return fmt.Errorf("failed to clear cache: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "✓ Cache cleared successfully\n")
			return nil
		},
	}

	return cmd
}

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show cache status",
		Long:  `Display information about cached data`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "Cache location: ~/.atlassian-cli/cache/\n")
			fmt.Fprintf(cmd.OutOrStdout(), "Cache enabled: Yes\n")
			fmt.Fprintf(cmd.OutOrStdout(), "Default TTL: 5 minutes\n")
			return nil
		},
	}

	return cmd
}