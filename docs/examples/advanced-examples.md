# Advanced Usage Examples

This guide covers advanced usage patterns and workflows for power users.

## Batch Operations

### Creating Multiple Issues

```bash
# Create multiple issues from a CSV file
while IFS=',' read -r type summary description assignee; do
  atlassian-cli issue create \
    --type "$type" \
    --summary "$summary" \
    --description "$description" \
    --assignee "$assignee"
done < issues.csv

# Create issues with JSON output for further processing
for story in "User Login" "User Registration" "Password Reset"; do
  atlassian-cli issue create \
    --type Story \
    --summary "$story" \
    --output json >> created_issues.json
done
```

### Bulk Issue Updates

```bash
# Update all issues assigned to a user
atlassian-cli issue list \
  --assignee john.doe \
  --output json | \
jq -r '.[].key' | \
while read -r issue_key; do
  atlassian-cli issue update "$issue_key" \
    --assignee jane.smith \
    --add-labels "reassigned"
done

# Bulk status transitions
atlassian-cli issue list \
  --status "To Do" \
  --type Story \
  --output json | \
jq -r '.[].key' | \
while read -r issue_key; do
  atlassian-cli issue update "$issue_key" --status "In Progress"
done
```

## Advanced JQL Queries

### Complex Filtering

```bash
# Issues created in the last sprint
atlassian-cli issue list \
  --jql "project = DEMO AND created >= -14d AND type = Story ORDER BY priority DESC"

# Issues with specific labels and components
atlassian-cli issue list \
  --jql "project = DEMO AND labels IN (urgent, security) AND component = Backend"

# Issues assigned to team members
atlassian-cli issue list \
  --jql "project = DEMO AND assignee IN (john.doe, jane.smith, bob.wilson) AND status != Done"

# Overdue issues
atlassian-cli issue list \
  --jql "project = DEMO AND duedate < now() AND status NOT IN (Done, Cancelled)"
```

### Reporting Queries

```bash
# Sprint burndown data
atlassian-cli issue list \
  --jql "project = DEMO AND sprint in openSprints() AND type != Epic" \
  --output json | \
jq '[.[] | {key, summary, status, storyPoints}]'

# Team workload analysis
for user in john.doe jane.smith bob.wilson; do
  echo "=== $user ==="
  atlassian-cli issue list \
    --assignee "$user" \
    --status "In Progress" \
    --output json | \
  jq -r 'length as $count | "Active issues: \($count)"'
done
```

## Multi-Environment Workflows

### Environment-Specific Configuration

```bash
# Production environment
export ATLASSIAN_DEFAULT_JIRA_PROJECT=PROD
export ATLASSIAN_DEFAULT_CONFLUENCE_SPACE=PROD-DOCS
export ATLASSIAN_OUTPUT=json

# Development environment  
export ATLASSIAN_DEFAULT_JIRA_PROJECT=DEV
export ATLASSIAN_DEFAULT_CONFLUENCE_SPACE=DEV-DOCS
export ATLASSIAN_OUTPUT=table

# Use environment-specific commands
atlassian-cli issue list --status Critical  # Uses PROD project
```

### Profile-Based Switching

```bash
# Create environment profiles
atlassian-cli config set prod_jira_project PROD
atlassian-cli config set prod_confluence_space PROD-DOCS
atlassian-cli config set dev_jira_project DEV
atlassian-cli config set dev_confluence_space DEV-DOCS

# Switch contexts with environment variables
switch_to_prod() {
  export ATLASSIAN_DEFAULT_JIRA_PROJECT=$(atlassian-cli config get prod_jira_project)
  export ATLASSIAN_DEFAULT_CONFLUENCE_SPACE=$(atlassian-cli config get prod_confluence_space)
  echo "Switched to production environment"
}

switch_to_dev() {
  export ATLASSIAN_DEFAULT_JIRA_PROJECT=$(atlassian-cli config get dev_jira_project)
  export ATLASSIAN_DEFAULT_CONFLUENCE_SPACE=$(atlassian-cli config get dev_confluence_space)
  echo "Switched to development environment"
}
```

## Integration with Git Workflows

### Git Hooks Integration

```bash
# .git/hooks/commit-msg
#!/bin/bash
# Extract issue key from commit message
issue_key=$(grep -o '[A-Z]\+-[0-9]\+' "$1" | head -1)

if [ -n "$issue_key" ]; then
  # Add comment to JIRA issue
  commit_msg=$(cat "$1")
  atlassian-cli issue update "$issue_key" \
    --add-comment "Commit: $commit_msg"
fi
```

### Branch-Based Issue Management

```bash
# Create branch and issue together
create_feature_branch() {
  local summary="$1"
  local issue_type="${2:-Story}"
  
  # Create JIRA issue
  issue_key=$(atlassian-cli issue create \
    --type "$issue_type" \
    --summary "$summary" \
    --output json | jq -r '.key')
  
  # Create git branch
  branch_name="feature/${issue_key,,}-$(echo "$summary" | tr ' ' '-' | tr '[:upper:]' '[:lower:]')"
  git checkout -b "$branch_name"
  
  echo "Created issue $issue_key and branch $branch_name"
}

# Usage
create_feature_branch "Implement OAuth2 authentication" "Story"
```

## Confluence Automation

### Documentation Generation

```bash
# Generate API documentation from OpenAPI spec
generate_api_docs() {
  local spec_file="$1"
  local space="${2:-$(atlassian-cli config get default_confluence_space)}"
  
  # Convert OpenAPI to HTML
  swagger-codegen generate -i "$spec_file" -l html2 -o /tmp/api-docs
  
  # Create Confluence page
  atlassian-cli page create \
    --confluence-space "$space" \
    --title "API Documentation $(date +%Y-%m-%d)" \
    --content "$(cat /tmp/api-docs/index.html)"
}
```

### Meeting Notes Automation

```bash
# Create weekly meeting notes template
create_meeting_notes() {
  local week_of="$1"
  local space="${2:-TEAM}"
  
  local content="<h2>Weekly Team Meeting - $week_of</h2>
<h3>Attendees</h3>
<ul>
<li>Team Member 1</li>
<li>Team Member 2</li>
</ul>
<h3>Agenda</h3>
<ol>
<li>Sprint Review</li>
<li>Blockers Discussion</li>
<li>Next Week Planning</li>
</ol>
<h3>Action Items</h3>
<table>
<tr><th>Action</th><th>Owner</th><th>Due Date</th></tr>
<tr><td></td><td></td><td></td></tr>
</table>"

  atlassian-cli page create \
    --confluence-space "$space" \
    --title "Team Meeting Notes - $week_of" \
    --content "$content"
}

# Usage
create_meeting_notes "$(date +%Y-%m-%d)"
```

## Performance Optimization

### Caching Strategies

```bash
# Pre-warm cache for common operations
warm_cache() {
  echo "Warming cache..."
  atlassian-cli project list >/dev/null 2>&1
  atlassian-cli space list >/dev/null 2>&1
  atlassian-cli issue list --limit 10 >/dev/null 2>&1
  echo "Cache warmed"
}

# Cache status monitoring
monitor_cache() {
  while true; do
    atlassian-cli cache status
    sleep 300  # Check every 5 minutes
  done
}
```

### Parallel Processing

```bash
# Process multiple projects in parallel
process_projects_parallel() {
  local projects=("PROJ1" "PROJ2" "PROJ3" "PROJ4")
  
  for project in "${projects[@]}"; do
    {
      echo "Processing $project..."
      atlassian-cli issue list \
        --jira-project "$project" \
        --status "In Progress" \
        --output json > "${project}_issues.json"
      echo "Completed $project"
    } &
  done
  
  wait  # Wait for all background jobs to complete
  echo "All projects processed"
}
```

## Error Handling and Retry Logic

### Robust Scripting

```bash
# Retry wrapper function
retry_command() {
  local max_attempts=3
  local delay=5
  local attempt=1
  
  while [ $attempt -le $max_attempts ]; do
    if "$@"; then
      return 0
    else
      echo "Attempt $attempt failed. Retrying in $delay seconds..."
      sleep $delay
      ((attempt++))
      delay=$((delay * 2))  # Exponential backoff
    fi
  done
  
  echo "Command failed after $max_attempts attempts"
  return 1
}

# Usage
retry_command atlassian-cli issue create \
  --type Bug \
  --summary "Critical production issue" \
  --priority Highest
```

### Validation and Error Recovery

```bash
# Validate configuration before operations
validate_setup() {
  # Check authentication
  if ! atlassian-cli auth status >/dev/null 2>&1; then
    echo "Error: Not authenticated. Run 'atlassian-cli auth login' first."
    return 1
  fi
  
  # Check default project
  if [ -z "$(atlassian-cli config get default_jira_project)" ]; then
    echo "Warning: No default JIRA project configured."
    echo "Set one with: atlassian-cli config set default_jira_project PROJECT"
  fi
  
  # Test connectivity
  if ! atlassian-cli project list --limit 1 >/dev/null 2>&1; then
    echo "Error: Cannot connect to JIRA. Check your configuration."
    return 1
  fi
  
  echo "Setup validation passed"
  return 0
}

# Safe issue creation with validation
safe_create_issue() {
  validate_setup || return 1
  
  atlassian-cli issue create "$@" || {
    echo "Issue creation failed. Checking for duplicates..."
    # Add duplicate detection logic here
    return 1
  }
}
```

## Monitoring and Alerting

### Issue Tracking Dashboard

```bash
# Generate daily status report
daily_report() {
  local project="${1:-$(atlassian-cli config get default_jira_project)}"
  local date=$(date +%Y-%m-%d)
  
  echo "=== Daily Report for $project - $date ==="
  echo
  
  echo "Critical Issues:"
  atlassian-cli issue list \
    --jira-project "$project" \
    --priority Critical \
    --status "To Do,In Progress" \
    --output table
  
  echo
  echo "New Issues (Last 24h):"
  atlassian-cli issue list \
    --jql "project = $project AND created >= -1d" \
    --output table
  
  echo
  echo "Completed Issues (Last 24h):"
  atlassian-cli issue list \
    --jql "project = $project AND status = Done AND resolved >= -1d" \
    --output table
}

# Schedule with cron: 0 9 * * * /path/to/daily_report.sh
```

These advanced examples demonstrate the power and flexibility of the Atlassian CLI for complex workflows, automation, and integration scenarios.