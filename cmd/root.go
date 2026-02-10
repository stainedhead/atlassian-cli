package cmd

import (
	"context"
	"fmt"
	"os"

	"atlassian-cli/cmd/auth"
	"atlassian-cli/cmd/cache"
	"atlassian-cli/cmd/config"
	"atlassian-cli/cmd/issue"
	"atlassian-cli/cmd/page"
	"atlassian-cli/cmd/project"
	"atlassian-cli/cmd/space"
	authManager "atlassian-cli/internal/auth"
	"atlassian-cli/internal/cmdutil"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile      string
	outputFormat string
	verbose      bool
	debug        bool
	version      = "dev" // Will be set by build process
)

// newRootCmd creates the root command
func newRootCmd() *cobra.Command {
	// Create local viper instance
	v := viper.New()

	cmd := &cobra.Command{
		Use:   "atlassian-cli",
		Short: "Developer toolkit for JIRA and Confluence",
		Long: `Atlassian CLI is a command-line tool that streamlines development workflows
by providing intuitive access to JIRA and Confluence operations.

Features:
• Smart default configuration for projects and spaces
• Secure credential management with OS keychain integration
• Multi-format output (JSON, table, YAML)
• Comprehensive JIRA issue and project management
• Full Confluence page and space operations
• Enterprise-grade reliability with caching and retry logic`,
		Version: version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Initialize config with local viper instance
			if err := initializeConfigWithViper(v); err != nil {
				return err
			}
			// Store viper in context for subcommands
			ctx := context.WithValue(cmd.Context(), cmdutil.ViperKey, v)
			cmd.SetContext(ctx)
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return fmt.Errorf("unknown command %q", args[0])
		},
	}

	// Global persistent flags
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.atlassian-cli/config.yaml)")
	cmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "output format (json, table, yaml)")
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	cmd.PersistentFlags().BoolVar(&debug, "debug", false, "debug output")
	// Global project/space flags removed - use command-specific flags instead
	cmd.PersistentFlags().Bool("no-color", false, "disable colored output")

	// Bind flags to local viper instance
	v.BindPFlag("output", cmd.PersistentFlags().Lookup("output"))
	v.BindPFlag("verbose", cmd.PersistentFlags().Lookup("verbose"))
	v.BindPFlag("debug", cmd.PersistentFlags().Lookup("debug"))
	// Viper bindings for global project/space flags removed

	// Add subcommands
	tokenManager := createTokenManager()
	cmd.AddCommand(auth.NewAuthCmd(tokenManager))
	cmd.AddCommand(issue.NewIssueCmd(tokenManager))
	cmd.AddCommand(project.NewProjectCmd(tokenManager))
	cmd.AddCommand(page.NewPageCmd(tokenManager))
	cmd.AddCommand(space.NewSpaceCmd(tokenManager))
	cmd.AddCommand(config.NewConfigCmd())
	cmd.AddCommand(cache.NewCacheCmd())
	cmd.AddCommand(newCompletionCmd())

	return cmd
}

// Execute is the main entry point for the CLI
func Execute() error {
	return newRootCmd().Execute()
}

// initializeConfigWithViper reads in config file and ENV variables with a specific viper instance
func initializeConfigWithViper(v *viper.Viper) error {
	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		// Search config in home directory and current directory
		v.AddConfigPath(home + "/.atlassian-cli")
		v.AddConfigPath(".")
		v.SetConfigType("yaml")
		v.SetConfigName("config")
	}

	// Environment variable configuration
	v.SetEnvPrefix("ATLASSIAN")
	v.AutomaticEnv()

	// Set defaults
	v.SetDefault("timeout", "30s")
	v.SetDefault("output", "table")
	v.SetDefault("default_jira_project", "")
	v.SetDefault("default_confluence_space", "")
	v.SetDefault("debug", false)
	v.SetDefault("verbose", false)

	// Read config file if it exists
	if err := v.ReadInConfig(); err != nil {
		// Config file not found is not an error
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}

	return nil
}

// newCompletionCmd creates the completion command
func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: `To load completions:

Bash:
  $ source <(atlassian-cli completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ atlassian-cli completion bash > /etc/bash_completion.d/atlassian-cli
  # macOS:
  $ atlassian-cli completion bash > /usr/local/etc/bash_completion.d/atlassian-cli

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc
  # To load completions for each session, execute once:
  $ atlassian-cli completion zsh > "${fpath[1]}/_atlassian-cli"
  # You will need to start a new shell for this setup to take effect.

fish:
  $ atlassian-cli completion fish | source
  # To load completions for each session, execute once:
  $ atlassian-cli completion fish > ~/.config/fish/completions/atlassian-cli.fish

PowerShell:
  PS> atlassian-cli completion powershell | Out-String | Invoke-Expression
  # To load completions for every new session, run:
  PS> atlassian-cli completion powershell > atlassian-cli.ps1
  # and source this file from your PowerShell profile.
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(cmd.OutOrStdout())
			case "zsh":
				cmd.Root().GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
			case "powershell":
				cmd.Root().GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
			}
		},
	}

	return cmd
}

// createTokenManager creates a token manager with tiered fallback
// Priority: Keychain -> Encrypted File -> Memory
func createTokenManager() authManager.TokenManager {
	// Try OS keychain first
	keychainManager := authManager.NewKeychainTokenManager()
	
	// Test if keychain is available by attempting a no-op operation
	// We try to get a non-existent key to see if keychain access works
	ctx := context.Background()
	_, err := keychainManager.Get(ctx, "test-availability")
	
	// If the error is "not found", keychain is working
	// If the error is about platform support or access, keychain is not available
	if err != nil && err.Error() == "credentials not found for server: test-availability" {
		// Keychain is available
		if debug || verbose {
			fmt.Fprintf(os.Stderr, "Using OS keychain for credential storage\n")
		}
		return keychainManager
	}
	
	// Try encrypted file fallback
	encryptedManager, err := authManager.NewEncryptedFileTokenManager("")
	if err == nil {
		if debug || verbose {
			fmt.Fprintf(os.Stderr, "OS keychain unavailable, using encrypted file storage\n")
		}
		return encryptedManager
	}
	
	// Fallback to memory (with warning)
	fmt.Fprintf(os.Stderr, "Warning: Using in-memory credential storage. Credentials will not persist across sessions.\n")
	fmt.Fprintf(os.Stderr, "         Run 'auth login' in each session to authenticate.\n")
	return authManager.NewMemoryTokenManager()
}
