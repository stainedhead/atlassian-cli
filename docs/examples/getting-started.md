# Getting Started Guide

This guide will help you get up and running with the Atlassian CLI in minutes.

## Installation

### Download Binary

Download the latest release for your platform:

```bash
# Linux
curl -L https://github.com/your-org/atlassian-cli/releases/latest/download/atlassian-cli-linux-amd64 -o atlassian-cli
chmod +x atlassian-cli
sudo mv atlassian-cli /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/your-org/atlassian-cli/releases/latest/download/atlassian-cli-darwin-amd64 -o atlassian-cli
chmod +x atlassian-cli
sudo mv atlassian-cli /usr/local/bin/

# macOS (Apple Silicon)
curl -L https://github.com/your-org/atlassian-cli/releases/latest/download/atlassian-cli-darwin-arm64 -o atlassian-cli
chmod +x atlassian-cli
sudo mv atlassian-cli /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/your-org/atlassian-cli.git
cd atlassian-cli
make build
sudo cp bin/atlassian-cli /usr/local/bin/
```

## Quick Setup

### 1. Authentication

First, create an API token at [id.atlassian.com/manage/api-tokens](https://id.atlassian.com/manage/api-tokens), then:

```bash
atlassian-cli auth login \
  --server https://your-domain.atlassian.net \
  --email your-email@example.com \
  --token your-api-token
```

### 2. Configure Defaults

Set up default project and space to streamline your workflow:

```bash
# Set default JIRA project
atlassian-cli config set default_jira_project DEMO

# Set default Confluence space
atlassian-cli config set default_confluence_space DEV

# Set preferred output format
atlassian-cli config set output table
```

### 3. Verify Setup

```bash
# Check authentication
atlassian-cli auth status --server https://your-domain.atlassian.net

# List your configuration
atlassian-cli config list

# Test JIRA access
atlassian-cli project list

# Test Confluence access
atlassian-cli space list
```

## Basic Usage

### Working with Issues

```bash
# Create a new issue (uses default project)
atlassian-cli issue create \
  --type Story \
  --summary "Implement user authentication" \
  --description "Add OAuth2 authentication to the application"

# List issues in default project
atlassian-cli issue list --status "In Progress"

# Get specific issue
atlassian-cli issue get DEMO-123

# Update an issue
atlassian-cli issue update DEMO-123 \
  --assignee john.doe \
  --priority High
```

### Working with Pages

```bash
# Create a new page (uses default space)
atlassian-cli page create \
  --title "API Documentation" \
  --content "<p>This page contains API documentation</p>"

# List pages in default space
atlassian-cli page list

# Get specific page
atlassian-cli page get 123456

# Update a page
atlassian-cli page update 123456 \
  --title "Updated API Documentation" \
  --content "<p>Updated content</p>"
```

### Override Defaults

When you need to work with different projects or spaces:

```bash
# Work with a different project
atlassian-cli issue list --jira-project PROD --status Critical

# Work with a different space
atlassian-cli page list --confluence-space DOCS

# Use different output format
atlassian-cli project list --output json
```

## Shell Completion

Enable shell completion for a better experience:

### Bash

```bash
# Add to ~/.bashrc
source <(atlassian-cli completion bash)

# Or install system-wide
atlassian-cli completion bash | sudo tee /etc/bash_completion.d/atlassian-cli
```

### Zsh

```bash
# Add to ~/.zshrc
source <(atlassian-cli completion zsh)

# Or install to completion directory
atlassian-cli completion zsh > "${fpath[1]}/_atlassian-cli"
```

### Fish

```bash
atlassian-cli completion fish | source

# Or install permanently
atlassian-cli completion fish > ~/.config/fish/completions/atlassian-cli.fish
```

## Next Steps

- Explore the [Command Reference](../commands/README.md) for detailed documentation
- Check out [Advanced Usage Examples](advanced-examples.md)
- Learn about [Enterprise Features](enterprise-features.md)
- Set up [CI/CD Integration](ci-cd-integration.md)