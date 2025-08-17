# Command Reference

Complete reference for all Atlassian CLI commands.

## Core Commands

### Authentication
- [`atlassian-cli auth login`](auth.md#login) - Authenticate with Atlassian instance
- [`atlassian-cli auth logout`](auth.md#logout) - Clear stored credentials
- [`atlassian-cli auth status`](auth.md#status) - Show authentication status

### Configuration
- [`atlassian-cli config set`](config.md#set) - Set configuration values
- [`atlassian-cli config get`](config.md#get) - Get configuration values
- [`atlassian-cli config list`](config.md#list) - List all configuration

## JIRA Commands

### Issues
- [`atlassian-cli issue create`](issue.md#create) - Create new issues
- [`atlassian-cli issue get`](issue.md#get) - Get issue details
- [`atlassian-cli issue list`](issue.md#list) - List and search issues
- [`atlassian-cli issue search`](issue.md#search) - Search issues with JQL
- [`atlassian-cli issue update`](issue.md#update) - Update existing issues

### Projects
- [`atlassian-cli project list`](project.md#list) - List JIRA projects
- [`atlassian-cli project get`](project.md#get) - Get project details

## Confluence Commands

### Pages
- [`atlassian-cli page create`](page.md#create) - Create new pages
- [`atlassian-cli page get`](page.md#get) - Get page details
- [`atlassian-cli page list`](page.md#list) - List pages in space
- [`atlassian-cli page search`](page.md#search) - Search pages with CQL
- [`atlassian-cli page update`](page.md#update) - Update existing pages

### Spaces
- [`atlassian-cli space list`](space.md#list) - List Confluence spaces

## Enterprise Commands

### Cache Management
- [`atlassian-cli cache status`](cache.md#status) - Show cache information
- [`atlassian-cli cache clear`](cache.md#clear) - Clear all cached data

### Shell Integration
- [`atlassian-cli completion`](completion.md) - Generate shell completion scripts

## Global Flags

All commands support these global flags:

- `--config string` - Custom config file path
- `--output, -o string` - Output format (json, table, yaml)
- `--verbose, -v` - Verbose output
- `--debug` - Debug output
- `--no-color` - Disable colored output
- `--help, -h` - Show help

## Command-Specific Flags

**JIRA Commands** (`issue`, `project`):
- `--project string` - Override default JIRA project

**Confluence Commands** (`page`, `space`):
- `--space string` - Override default Confluence space