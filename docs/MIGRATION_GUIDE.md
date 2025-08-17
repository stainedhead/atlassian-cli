# Migration Guide: v1.x to v2.0.0

## Breaking Changes

### Removed Global Flags

The global flags `--jira-project` and `--confluence-space` have been removed from the root command. These have been replaced with command-specific flags for better clarity and consistency.

#### Before (v1.x)
```bash
# Global flags on root command
atlassian-cli --jira-project DEMO issue list
atlassian-cli --confluence-space DEV page list
```

#### After (v2.0.0)
```bash
# Command-specific flags
atlassian-cli issue list --project DEMO
atlassian-cli page list --space DEV
```

### Flag Mapping

| Old Global Flag | New Command-Specific Flag | Commands |
|----------------|---------------------------|----------|
| `--jira-project` | `--project` | `issue`, `project` |
| `--confluence-space` | `--space` | `page`, `space` |

### Configuration Hierarchy Unchanged

The configuration hierarchy remains the same, but now uses command-specific flags:

1. **Command flags** (highest priority): `--project`, `--space`
2. **Environment variables**: `ATLASSIAN_DEFAULT_JIRA_PROJECT`, `ATLASSIAN_DEFAULT_CONFLUENCE_SPACE`
3. **Configuration file**: `default_jira_project`, `default_confluence_space`
4. **Interactive prompts** (lowest priority)

## New Features

### Search Commands

Two new search commands have been added:

#### Issue Search
```bash
# Search with JQL
atlassian-cli issue search --jql "assignee = currentUser() AND status = 'In Progress'"

# Search with simple filters
atlassian-cli issue search --project DEMO --status "Open" --assignee john.doe

# Search in default project
atlassian-cli issue search --status "In Progress" --type Bug
```

#### Page Search
```bash
# Search with CQL
atlassian-cli page search --cql "space = DEV AND type = page"

# Search with text
atlassian-cli page search --space DEV --text "documentation"

# Search by title in default space
atlassian-cli page search --title "API Guide"
```

## Migration Steps

### 1. Update Command Usage

Replace global flags with command-specific flags in your scripts:

```bash
# Find and replace in scripts
sed -i 's/--jira-project/--project/g' your-script.sh
sed -i 's/--confluence-space/--space/g' your-script.sh

# Move flags after the subcommand
sed -i 's/atlassian-cli --project \([A-Z]*\) issue/atlassian-cli issue --project \1/g' your-script.sh
sed -i 's/atlassian-cli --space \([A-Z]*\) page/atlassian-cli page --space \1/g' your-script.sh
```

### 2. Update CI/CD Pipelines

Update your CI/CD pipeline scripts:

```yaml
# Before
- name: Create JIRA issue
  run: atlassian-cli --jira-project ${{ vars.JIRA_PROJECT }} issue create --type Bug --summary "Build failed"

# After  
- name: Create JIRA issue
  run: atlassian-cli issue create --project ${{ vars.JIRA_PROJECT }} --type Bug --summary "Build failed"
```

### 3. Configuration Files Unchanged

Your existing configuration files continue to work without changes:

```yaml
# ~/.atlassian-cli/config.yaml (no changes needed)
default_jira_project: "DEMO"
default_confluence_space: "DEV"
output: "table"
```

### 4. Environment Variables Unchanged

Environment variables continue to work as before:

```bash
# No changes needed
export ATLASSIAN_DEFAULT_JIRA_PROJECT=DEMO
export ATLASSIAN_DEFAULT_CONFLUENCE_SPACE=DEV
```

## Validation

### Test Your Migration

1. **Verify command syntax**:
   ```bash
   # Test issue commands
   atlassian-cli issue list --project DEMO
   atlassian-cli issue create --project DEMO --type Story --summary "Test"
   
   # Test page commands
   atlassian-cli page list --space DEV
   atlassian-cli page create --space DEV --title "Test Page"
   ```

2. **Test default resolution**:
   ```bash
   # Should use configured defaults
   atlassian-cli issue list
   atlassian-cli page list
   ```

3. **Test new search features**:
   ```bash
   # Test issue search
   atlassian-cli issue search --status "In Progress"
   
   # Test page search
   atlassian-cli page search --text "documentation"
   ```

## Troubleshooting

### Common Issues

1. **"unknown flag" errors**:
   - Ensure you've moved flags after the subcommand
   - Use `--project` instead of `--jira-project`
   - Use `--space` instead of `--confluence-space`

2. **"no project/space configured" errors**:
   - Check your configuration: `atlassian-cli config list`
   - Set defaults: `atlassian-cli config set default_jira_project DEMO`
   - Use command flags: `--project DEMO` or `--space DEV`

3. **Scripts not working**:
   - Update flag positions in command lines
   - Test commands manually before updating scripts
   - Use the migration sed commands above

### Getting Help

```bash
# Command help
atlassian-cli issue --help
atlassian-cli page --help

# Specific command help
atlassian-cli issue search --help
atlassian-cli page search --help

# Configuration help
atlassian-cli config --help
```

This migration maintains all existing functionality while providing clearer command structure and powerful new search capabilities.