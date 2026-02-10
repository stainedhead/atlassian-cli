package cmdutil

import (
	"atlassian-cli/internal/client"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Context keys for passing values through command execution
type contextKey string

const (
	ViperKey   contextKey = "viper"
	FactoryKey contextKey = "factory"
)

// GetViperFromCmd retrieves the viper instance from the command context
func GetViperFromCmd(cmd *cobra.Command) *viper.Viper {
	if v := cmd.Context().Value(ViperKey); v != nil {
		return v.(*viper.Viper)
	}
	// Fallback to global viper for backward compatibility during transition
	return viper.GetViper()
}

// GetConfigPath returns the config file path from the command context
func GetConfigPath(cmd *cobra.Command) string {
	v := GetViperFromCmd(cmd)
	return v.GetString("config")
}

// GetOutputFormat returns the output format from the command context
func GetOutputFormat(cmd *cobra.Command) string {
	v := GetViperFromCmd(cmd)
	return v.GetString("output")
}

// GetFactory retrieves the client factory from the command context
func GetFactory(cmd *cobra.Command) *client.Factory {
	if f := cmd.Context().Value(FactoryKey); f != nil {
		return f.(*client.Factory)
	}
	// If no factory in context, create a new one (shouldn't happen in normal execution)
	return client.NewFactory()
}
