# Configuration Commands

The `config` command group manages CLI configuration and smart defaults.

## atlassian-cli config set

Set configuration values for smart defaults and preferences.

### Usage

```bash
atlassian-cli config set <key> <value>
```

### Configuration Keys

#### JIRA Settings
- `default_jira_project` - Default JIRA project key (e.g., "DEMO")
- `jira_timeout` - API timeout for JIRA operations (default: "30s")

#### Confluence Settings  
- `default_confluence_space` - Default Confluence space key (e.g., "DEV")
- `confluence_timeout` - API timeout for Confluence operations (default: "30s")

#### Output Settings
- `output` - Default output format: "table", "json", "yaml" (default: "table")
- `no_color` - Disable colored output: "true", "false" (default: "false")

#### Cache Settings
- `cache_ttl` - Cache time-to-live (default: "5m")
- `cache_enabled` - Enable caching: "true", "false" (default: "true")

### Examples

```bash
# Set default JIRA project
atlassian-cli config set default_jira_project DEMO

# Set default Confluence space
atlassian-cli config set default_confluence_space DEV

# Set preferred output format
atlassian-cli config set output json

# Configure timeouts
atlassian-cli config set jira_timeout 60s
atlassian-cli config set confluence_timeout 45s

# Cache configuration
atlassian-cli config set cache_ttl 10m
atlassian-cli config set cache_enabled false
```

## atlassian-cli config get

Get configuration values.

### Usage

```bash
atlassian-cli config get <key>
```

### Examples

```bash
# Get default project
atlassian-cli config get default_jira_project

# Get output format
atlassian-cli config get output

# Get cache settings
atlassian-cli config get cache_ttl
```

### Output

```
default_jira_project: DEMO
```

## atlassian-cli config list

List all configuration settings.

### Usage

```bash
atlassian-cli config list
```

### Examples

```bash
atlassian-cli config list
```

### Output

```
Configuration Settings:
  default_jira_project: DEMO
  default_confluence_space: DEV
  output: table
  cache_ttl: 5m
  cache_enabled: true
  jira_timeout: 30s
  confluence_timeout: 30s
  no_color: false
```

## Configuration Hierarchy

The CLI uses a hierarchical configuration system (highest to lowest priority):

1. **Command flags** - `--jira-project`, `--confluence-space`, `--output`
2. **Environment variables** - `ATLASSIAN_DEFAULT_JIRA_PROJECT`, `ATLASSIAN_DEFAULT_CONFLUENCE_SPACE`
3. **Configuration file** - `~/.atlassian-cli/config.yaml`
4. **Built-in defaults** - Fallback values

### Environment Variables

```bash
# Override defaults with environment variables
export ATLASSIAN_DEFAULT_JIRA_PROJECT=PROD
export ATLASSIAN_DEFAULT_CONFLUENCE_SPACE=DOCS
export ATLASSIAN_OUTPUT=json
export ATLASSIAN_CACHE_TTL=15m

# Commands automatically use these values
atlassian-cli issue list
atlassian-cli page list
```

### Configuration File

The configuration file is stored at `~/.atlassian-cli/config.yaml`:

```yaml
default_jira_project: DEMO
default_confluence_space: DEV
output: table
cache_ttl: 5m
cache_enabled: true
jira_timeout: 30s
confluence_timeout: 30s
no_color: false
```

## Smart Defaults in Action

Once configured, commands become streamlined:

```bash
# Set defaults once
atlassian-cli config set default_jira_project DEMO
atlassian-cli config set default_confluence_space DEV

# Commands automatically use defaults
atlassian-cli issue create --type Story --summary "New feature"
atlassian-cli page create --title "Documentation" --content "<p>Content</p>"

# Override when needed
atlassian-cli issue list --jira-project PROD
atlassian-cli page list --confluence-space DOCS
```