package main

import (
	"os"

	"atlassian-cli/cmd"
)

// version will be set by build process
var version = "dev"

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
