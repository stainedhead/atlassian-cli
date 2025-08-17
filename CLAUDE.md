# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an Atlassian CLI tool project for JIRA and Confluence automation. The project is currently in early planning/research phase with comprehensive documentation but no Go code implementation yet.

## Project Structure

- `documentation/` - Contains product research and architectural planning
  - `product-research.md` - Comprehensive guide for building Atlassian CLI with Go, including API patterns, architecture decisions, and implementation examples
- `prompts/` - Contains initial project setup prompts
- `bin/` - Binary output directory (excluded from git)
- `tests/` - Test directory (excluded from git)

## Development Commands

Since this is a new Go project without implementation yet, standard Go commands will be used:

- **Build**: `go build -o bin/atlassian-cli`
- **Test**: `go test ./...`
- **Lint**: `golangci-lint run` (assuming golangci-lint will be used)
- **Clean**: `rm -rf bin/`

## Architecture & Implementation Guide

The `documentation/product-research.md` file contains a comprehensive blueprint for implementation:

### Recommended Tech Stack
- **CLI Framework**: Cobra for command structure and argument parsing
- **Atlassian SDK**: go-atlassian (github.com/ctreminiom/go-atlassian/v2) for unified JIRA/Confluence API access
- **Configuration**: Viper for hierarchical configuration management
- **Authentication**: OS keychain integration for secure token storage
- **Output**: Multi-format support (JSON, table, YAML)

### Command Structure
```
atlassian-cli
├── issue (JIRA issue management)
│   ├── create, get, list, update, comment, link, transition
├── sprint (Agile/Sprint management) 
│   ├── create, list, start, complete, issues, move
├── project (Project operations)
│   ├── list, get, components, versions
├── page (Confluence pages)
│   ├── create, get, list, update, delete, history
├── space (Confluence spaces)
│   ├── list, get, content, permissions
└── auth (Authentication)
    ├── login, logout, status, switch
```

### Authentication Pattern
- API token-based authentication (email + token)
- Secure credential storage using OS keychain
- Support for multiple Atlassian instances/profiles

### Key Implementation Patterns
- Service-oriented API client architecture
- Interface-based design for testability
- Comprehensive error handling with actionable suggestions
- Progress indicators for long-running operations
- Intelligent output formatting based on terminal capabilities

## Next Steps for Implementation

1. Initialize Go module: `go mod init atlassian-cli`
2. Add core dependencies (Cobra, Viper, go-atlassian)
3. Implement basic command structure following Cobra patterns
4. Set up authentication and configuration management
5. Implement core JIRA issue operations
6. Add Confluence page management
7. Implement comprehensive testing strategy

## Testing Strategy

- Unit tests with mocked API clients
- Integration tests with HTTP mocking
- CLI command testing using Cobra's testing utilities
- Separate test packages for different domains (issue, page, auth)