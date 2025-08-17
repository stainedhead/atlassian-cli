# Atlassian CLI Documentation

Welcome to the comprehensive documentation for the Atlassian CLI - a modern command-line interface for JIRA and Confluence that transforms REST API complexity into intuitive developer workflows.

## Quick Navigation

### 📚 Getting Started
- [Installation & Setup](examples/getting-started.md) - Get up and running in minutes
- [Command Reference](commands/README.md) - Complete command documentation
- [Configuration Guide](commands/config.md) - Smart defaults and configuration

### 🚀 Usage Examples
- [Basic Usage](examples/getting-started.md#basic-usage) - Common daily workflows
- [Advanced Examples](examples/advanced-examples.md) - Power user patterns
- [CI/CD Integration](examples/ci-cd-integration.md) - Automation workflows

### 🔧 Development & Deployment
- [Development Guide](development/README.md) - Contributing and development setup
- [Deployment Guide](DEPLOYMENT.md) - Distribution and enterprise deployment
- [Architecture Overview](development/architecture.md) - Technical design

## Core Features

### 🎯 Smart Default Configuration
Eliminate repetitive parameter specification with hierarchical configuration:

```bash
# Set once
atlassian-cli config set default_jira_project DEMO
atlassian-cli config set default_confluence_space DEV

# Use everywhere
atlassian-cli issue create --type Story --summary "New feature"
atlassian-cli page create --title "Documentation"

# Override when needed
atlassian-cli issue list --jira-project PROD --status Critical
```

### 🔒 Secure Authentication
- API token-based authentication with secure storage
- Multiple instance support
- Session-based credential management
- Environment variable support for CI/CD

### 📊 Multi-Format Output
- **JSON** for scripting and automation
- **Table** for human-readable output  
- **YAML** for configuration management
- Consistent formatting across all commands

### ⚡ Developer-Focused Workflows
- **JIRA**: Complete issue lifecycle management
- **Confluence**: Full page and space operations
- **Enterprise**: Caching, audit logging, retry logic
- **Integration**: Shell completion and CI/CD examples

## Command Categories

### Authentication & Configuration
- [`atlassian-cli auth`](commands/auth.md) - Manage authentication
- [`atlassian-cli config`](commands/config.md) - Configure defaults and preferences

### JIRA Operations
- [`atlassian-cli issue`](commands/issue.md) - Issue management with smart defaults
- [`atlassian-cli issue search`](commands/issue.md#search) - JQL-based issue search
- [`atlassian-cli project`](commands/project.md) - Project operations and metadata

### Confluence Operations  
- [`atlassian-cli page`](commands/page.md) - Page CRUD operations
- [`atlassian-cli page search`](commands/page.md#search) - CQL-based page search
- [`atlassian-cli space`](commands/space.md) - Space management

### Enterprise Features
- [`atlassian-cli cache`](commands/cache.md) - Performance optimization
- [`atlassian-cli completion`](commands/completion.md) - Shell integration

## Configuration Hierarchy

The CLI uses intelligent configuration resolution:

1. **Command flags** (highest priority)
2. **Environment variables** 
3. **Configuration file**
4. **Interactive prompts** (lowest priority)

### Environment Variables

```bash
export ATLASSIAN_DEFAULT_JIRA_PROJECT=DEMO
export ATLASSIAN_DEFAULT_CONFLUENCE_SPACE=DEV
export ATLASSIAN_OUTPUT=json
export ATLASSIAN_CACHE_TTL=10m
```

### Configuration File

Location: `~/.atlassian-cli/config.yaml`

```yaml
default_jira_project: DEMO
default_confluence_space: DEV
output: table
cache_ttl: 5m
cache_enabled: true
jira_timeout: 30s
confluence_timeout: 30s
```

## Integration Examples

### Git Workflow Integration

```bash
# Extract issue key from branch name
issue_key=$(git branch --show-current | grep -o '[A-Z]\+-[0-9]\+')

# Update issue with commit
atlassian-cli issue update "$issue_key" \
  --add-comment "Implemented feature in commit $(git rev-parse --short HEAD)"
```

### CI/CD Pipeline Integration

```yaml
# GitHub Actions example
- name: Update JIRA Issues
  run: |
    git log --oneline ${{ github.event.before }}..${{ github.sha }} | \
    grep -o '[A-Z]\+-[0-9]\+' | sort -u | \
    while read issue_key; do
      atlassian-cli issue update "$issue_key" \
        --add-comment "Deployed in build ${{ github.run_number }}"
    done
```

### Scripting and Automation

```bash
# Daily standup report
atlassian-cli issue list \
  --assignee "$(whoami)" \
  --status "In Progress" \
  --output json | \
jq -r '.[] | "- \(.key): \(.summary)"'

# Bulk issue updates
cat issue_list.txt | while read issue_key; do
  atlassian-cli issue update "$issue_key" \
    --add-labels "release-v1.2.0" \
    --priority "High"
done
```

## Performance & Reliability

### Intelligent Caching
- 5-minute default TTL for API responses
- Automatic cache invalidation
- Manual cache management commands

### Retry Logic
- Exponential backoff with jitter
- Configurable timeout settings
- Graceful error handling

### Enterprise Security
- Audit logging for compliance
- Structured event tracking
- No credential persistence to disk

## Support & Community

### Documentation
- [Command Reference](commands/README.md) - Detailed command documentation
- [Examples](examples/) - Real-world usage patterns
- [Development](development/) - Contributing guidelines

### Getting Help

```bash
# Command help
atlassian-cli --help
atlassian-cli issue --help
atlassian-cli issue create --help

# Configuration help
atlassian-cli config list
atlassian-cli auth status
```

### Troubleshooting

Common issues and solutions:

1. **Authentication Problems**
   ```bash
   atlassian-cli auth status
   atlassian-cli auth login --server <url> --email <email> --token <token>
   ```

2. **Configuration Issues**
   ```bash
   atlassian-cli config list
   atlassian-cli config set default_jira_project PROJECT
   ```

3. **Performance Issues**
   ```bash
   atlassian-cli cache status
   atlassian-cli cache clear
   ```

## What's Next?

The Atlassian CLI is production-ready with all core features implemented:

- ✅ **Phase 1**: Foundation & Core JIRA Operations
- ✅ **Phase 2**: Confluence Integration & Advanced Features  
- ✅ **Phase 3**: Enterprise Features & Polish
- ✅ **Phase 4**: Documentation & Distribution

The CLI provides a solid foundation for JIRA and Confluence automation with room for future enhancements based on user feedback and evolving requirements.

---

**Ready to get started?** Check out the [Getting Started Guide](examples/getting-started.md) or jump straight to the [Command Reference](commands/README.md).