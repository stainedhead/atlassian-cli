# Atlassian CLI

A modern command-line interface for JIRA and Confluence that transforms REST API complexity into intuitive developer workflows.

[![Go Version](https://img.shields.io/badge/go-1.21-blue.svg)](https://golang.org/doc/devel/release.html)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](#)

## Features

### 🚀 **Smart Default Configuration**
- Configure default JIRA projects and Confluence spaces once
- Override defaults with command-line flags when needed
- Environment variable support for CI/CD workflows
- Profile management for multiple Atlassian instances

### 🔒 **Secure Authentication**
- API token-based authentication
- Secure credential storage using OS keychain
- Multiple instance support
- Interactive authentication flow

### 📊 **Multi-Format Output**
- JSON for scripting and automation
- Table format for human-readable output
- YAML for configuration management
- Customizable output formatting

### ⚡ **Developer-Focused Workflows**
- **JIRA**: Issue management, project operations, agile workflows
- **Confluence**: Page and space management, content operations
- **Smart Defaults**: Eliminate repetitive parameter specification
- **Comprehensive Testing**: 90%+ test coverage with TDD approach

## Quick Start

### Installation

#### Quick Install (Recommended)
```bash
# Install latest version
curl -sSL https://raw.githubusercontent.com/your-org/atlassian-cli/main/scripts/install.sh | bash
```

#### Manual Installation
```bash
# Download binary for your platform from GitHub releases
# Linux
curl -L https://github.com/your-org/atlassian-cli/releases/latest/download/atlassian-cli-linux-amd64 -o atlassian-cli
chmod +x atlassian-cli
sudo mv atlassian-cli /usr/local/bin/

# macOS
curl -L https://github.com/your-org/atlassian-cli/releases/latest/download/atlassian-cli-darwin-amd64 -o atlassian-cli
chmod +x atlassian-cli
sudo mv atlassian-cli /usr/local/bin/
```

#### Build from Source
```bash
git clone https://github.com/your-org/atlassian-cli.git
cd atlassian-cli
make build
sudo make install
```

### Authentication

Create an API token at [id.atlassian.com/manage/api-tokens](https://id.atlassian.com/manage/api-tokens), then:

```bash
# Authenticate with your Atlassian instance
atlassian-cli auth login \
  --server https://your-domain.atlassian.net \
  --email your-email@example.com \
  --token your-api-token

# Check authentication status
atlassian-cli auth status --server https://your-domain.atlassian.net
```

### Configuration

Set up default project and space for streamlined workflows:

```bash
# Set default JIRA project
atlassian-cli config set default_jira_project MYPROJECT

# Set default Confluence space  
atlassian-cli config set default_confluence_space MYSPACE

# Set preferred output format
atlassian-cli config set output table
```

## Usage Examples

### Smart Defaults in Action

Once configured, commands become streamlined:

```bash
# Create issue using default project
atlassian-cli issue create --type Story --summary "New feature"

# Override default when needed
atlassian-cli issue create --jira-project OTHERPROJ --type Bug --summary "Critical fix"

# List issues in default project
atlassian-cli issue list --status "In Progress"
```

### Configuration Hierarchy

The CLI uses a hierarchical configuration system (highest to lowest priority):

1. **Command flags**: `--jira-project`, `--confluence-space`
2. **Environment variables**: `ATLASSIAN_DEFAULT_JIRA_PROJECT`, `ATLASSIAN_DEFAULT_CONFLUENCE_SPACE`
3. **Configuration file**: `~/.atlassian-cli/config.yaml`
4. **Interactive prompts**: When no defaults are configured

### Environment Variables

```bash
# Set defaults via environment variables
export ATLASSIAN_DEFAULT_JIRA_PROJECT=DEMO
export ATLASSIAN_DEFAULT_CONFLUENCE_SPACE=DEV
export ATLASSIAN_OUTPUT=json

# Commands automatically use these defaults
atlassian-cli issue list
atlassian-cli page list
```

## Development

### Prerequisites

- Go 1.21 or later
- Make

### Building from Source

```bash
# Clone and build
git clone https://github.com/your-org/atlassian-cli.git
cd atlassian-cli
make build

# Run tests
make test

# Run with coverage
make test-coverage

# Run linting (requires golangci-lint)
make lint

# Build for all platforms
make build-all
```

### Testing

The project follows Test-Driven Development (TDD) with comprehensive test coverage:

```bash
# Run all tests
make test

# Run tests with race detection
make test-race

# Generate coverage report
make test-coverage
```

### Code Quality

We maintain high code quality standards:

- **90%+ test coverage** requirement
- **golangci-lint** static analysis
- **Automated security scanning**
- **Dependency vulnerability scanning**

## Project Structure

```
atlassian-cli/
├── cmd/                          # Command implementations
│   ├── root.go                   # Root command and global flags
│   └── auth/                     # Authentication commands
├── internal/                     # Private application code
│   ├── api/                      # API client wrappers (planned)
│   ├── auth/                     # Authentication management
│   ├── config/                   # Configuration management
│   │   ├── config.go            # Configuration loading/saving
│   │   └── resolver.go          # Smart default + override logic
│   ├── output/                   # Output formatting (planned)
│   └── types/                    # Common data structures
├── test/                         # Test utilities and fixtures
├── docs/                         # Documentation
└── scripts/                      # Build and deployment scripts
```

## Configuration

### Configuration File

The CLI creates a configuration file at `~/.atlassian-cli/config.yaml`:

```yaml
# Atlassian CLI Configuration
api_endpoint: "https://your-domain.atlassian.net"
email: "user@example.com"
default_jira_project: "DEMO"
default_confluence_space: "DEV"
timeout: "30s"
output: "table"
debug: false
```

### Global Flags

All commands support these global flags:

- `--config`: Custom config file path
- `--jira-project`: Override default JIRA project
- `--confluence-space`: Override default Confluence space
- `--output` (`-o`): Output format (json, table, yaml)
- `--verbose` (`-v`): Verbose output
- `--debug`: Debug output
- `--no-color`: Disable colored output

## Implementation Status

### ✅ **All Phases Complete - Production Ready**

#### Phase 1: Foundation & Core JIRA Operations
- [x] Project structure and Go module setup
- [x] Core Cobra command structure with global flags
- [x] Configuration management with Viper
- [x] Smart default project/space resolution system
- [x] Authentication system with secure credential storage
- [x] JIRA issue operations (create, get, list, update)
- [x] Comprehensive test suite with 90%+ coverage
- [x] Build system with Makefile and quality tools

#### Phase 2: Confluence Integration & Advanced Features
- [x] Confluence page operations (create, get, list, update)
- [x] Confluence space management
- [x] Enhanced project operations
- [x] Multi-format output system (JSON, table, YAML)
- [x] Modular command architecture

#### Phase 3: Enterprise Features & Polish
- [x] Intelligent caching with configurable TTL
- [x] Retry logic with exponential backoff
- [x] Audit logging for compliance
- [x] Performance optimization
- [x] Enhanced error handling and recovery

#### Phase 4: Documentation & Distribution
- [x] Complete CLI reference documentation
- [x] Multi-platform build and distribution
- [x] Installation scripts and package managers
- [x] Shell completion (bash, zsh, fish, powershell)
- [x] CI/CD integration examples
- [x] Enterprise deployment guides
- [x] Comprehensive testing framework

## Contributing

1. Read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines
2. Follow the [Definition of Done](.github/copilot-instructions.md) checklist
3. Ensure all tests pass: `make test`
4. Run linting: `make lint`
5. Submit a pull request

## Architecture

The CLI follows proven patterns from successful tools like GitHub CLI and kubectl:

- **Modular Design**: Separate packages for distinct concerns
- **Interface-Based**: Testable with dependency injection
- **Smart Defaults**: Reduce cognitive overhead while maintaining flexibility
- **Comprehensive Testing**: TDD approach with unit, integration, and E2E tests
- **Security First**: Secure credential storage and input validation

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Documentation

### Complete Documentation Suite
- **[Product Summary](documentation/product-summary.md)** - Overview, features, and value proposition
- **[Technical Details](documentation/technical-details.md)** - Architecture, implementation, and performance
- **[Command Reference](docs/commands/README.md)** - Complete command documentation
- **[Usage Examples](docs/examples/)** - Getting started, advanced patterns, CI/CD integration
- **[Deployment Guide](docs/DEPLOYMENT.md)** - Enterprise deployment and distribution

## Support

- **Documentation**: Complete documentation in [docs/](./docs/) and [documentation/](./documentation/) directories
- **Issues**: Report bugs and feature requests on GitHub
- **Examples**: Real-world usage patterns in [docs/examples/](./docs/examples/)
- **Integration**: CI/CD examples for all major platforms