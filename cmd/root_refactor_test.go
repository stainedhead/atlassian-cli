package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlobalFlagsRemoved(t *testing.T) {
	cmd := newRootCmd()

	// These flags should NOT exist on root command
	assert.Nil(t, cmd.PersistentFlags().Lookup("jira-project"), "Global --jira-project flag should be removed")
	assert.Nil(t, cmd.PersistentFlags().Lookup("confluence-space"), "Global --confluence-space flag should be removed")

	// These flags should still exist
	assert.NotNil(t, cmd.PersistentFlags().Lookup("config"), "Config flag should remain")
	assert.NotNil(t, cmd.PersistentFlags().Lookup("output"), "Output flag should remain")
	assert.NotNil(t, cmd.PersistentFlags().Lookup("verbose"), "Verbose flag should remain")
}
