# Issue Commands

The `issue` command group manages JIRA issues with smart project defaults.

## atlassian-cli issue create

Create new JIRA issues with automatic project resolution.

### Usage

```bash
atlassian-cli issue create --type <type> --summary <summary> [flags]
```

### Required Flags

- `--type` - Issue type (Story, Bug, Task, Epic, etc.)
- `--summary` - Issue summary/title

### Optional Flags

- `--description` - Issue description
- `--assignee` - Assignee username or email
- `--priority` - Priority (Highest, High, Medium, Low, Lowest)
- `--labels` - Comma-separated labels
- `--components` - Comma-separated component names
- `--jira-project` - Override default JIRA project

### Examples

```bash
# Basic issue creation (uses default project)
atlassian-cli issue create \
  --type Story \
  --summary "Implement user authentication"

# Full issue with all fields
atlassian-cli issue create \
  --type Bug \
  --summary "Login page crashes on mobile" \
  --description "The login page crashes when accessed from mobile devices" \
  --assignee john.doe \
  --priority High \
  --labels "mobile,critical" \
  --components "Frontend,Authentication"

# Override default project
atlassian-cli issue create \
  --jira-project PROD \
  --type Task \
  --summary "Update production database"
```

### Output

```json
{
  "key": "DEMO-123",
  "id": "10001",
  "summary": "Implement user authentication",
  "status": "To Do",
  "assignee": "john.doe",
  "created": "2024-01-15T10:30:00.000Z",
  "url": "https://company.atlassian.net/browse/DEMO-123"
}
```

## atlassian-cli issue get

Retrieve detailed information about a specific issue.

### Usage

```bash
atlassian-cli issue get <issue-key>
```

### Examples

```bash
# Get issue details
atlassian-cli issue get DEMO-123

# Get with JSON output
atlassian-cli issue get DEMO-123 --output json
```

### Output (Table Format)

```
Issue: DEMO-123
Summary: Implement user authentication
Type: Story
Status: In Progress
Priority: Medium
Assignee: john.doe
Reporter: jane.smith
Created: 2024-01-15 10:30:00
Updated: 2024-01-16 14:20:00
Description:
  Add OAuth2 authentication to the application to improve security
  and user experience.

Components: Authentication, Frontend
Labels: security, oauth2
```

## atlassian-cli issue list

List and search issues with filtering and JQL support.

### Usage

```bash
atlassian-cli issue list [flags]
```

### Optional Flags

- `--status` - Filter by status (comma-separated)
- `--assignee` - Filter by assignee
- `--type` - Filter by issue type
- `--priority` - Filter by priority
- `--labels` - Filter by labels
- `--jql` - Custom JQL query
- `--limit` - Maximum number of results (default: 50)
- `--jira-project` - Override default JIRA project

### Examples

```bash
# List all issues in default project
atlassian-cli issue list

# Filter by status
atlassian-cli issue list --status "In Progress,Review"

# Filter by assignee
atlassian-cli issue list --assignee john.doe

# Multiple filters
atlassian-cli issue list \
  --status "To Do" \
  --type Story \
  --priority High

# Custom JQL query
atlassian-cli issue list \
  --jql "project = DEMO AND created >= -7d ORDER BY created DESC"

# Override project and limit results
atlassian-cli issue list \
  --jira-project PROD \
  --status Critical \
  --limit 10
```

### Output (Table Format)

```
KEY      SUMMARY                          TYPE    STATUS      ASSIGNEE    PRIORITY
DEMO-123 Implement user authentication    Story   In Progress john.doe    Medium
DEMO-124 Fix mobile login crash          Bug     To Do       jane.smith  High
DEMO-125 Update API documentation        Task    Review      bob.wilson  Low
```

## atlassian-cli issue update

Update existing issues with field modifications.

### Usage

```bash
atlassian-cli issue update <issue-key> [flags]
```

### Optional Flags

- `--summary` - Update summary
- `--description` - Update description
- `--assignee` - Change assignee
- `--priority` - Change priority
- `--status` - Change status (triggers workflow transition)
- `--labels` - Update labels (replaces existing)
- `--add-labels` - Add labels (preserves existing)
- `--remove-labels` - Remove specific labels
- `--components` - Update components

### Examples

```bash
# Update assignee and priority
atlassian-cli issue update DEMO-123 \
  --assignee jane.smith \
  --priority High

# Update summary and description
atlassian-cli issue update DEMO-123 \
  --summary "Enhanced user authentication" \
  --description "Implement OAuth2 with multi-factor authentication"

# Manage labels
atlassian-cli issue update DEMO-123 \
  --add-labels "security,enhancement" \
  --remove-labels "draft"

# Transition status
atlassian-cli issue update DEMO-123 --status "In Progress"

# Multiple field updates
atlassian-cli issue update DEMO-123 \
  --assignee john.doe \
  --priority Critical \
  --add-labels "urgent" \
  --components "Backend,Security"
```

### Output

```json
{
  "key": "DEMO-123",
  "updated": true,
  "fields_changed": ["assignee", "priority", "labels"],
  "url": "https://company.atlassian.net/browse/DEMO-123"
}
```

## Smart Project Resolution

All issue commands automatically resolve the JIRA project using this hierarchy:

1. `--jira-project` flag (highest priority)
2. `ATLASSIAN_DEFAULT_JIRA_PROJECT` environment variable
3. `default_jira_project` configuration setting
4. Interactive prompt (if no default configured)

### Configuration Example

```bash
# Set default project once
atlassian-cli config set default_jira_project DEMO

# All commands use DEMO project automatically
atlassian-cli issue create --type Story --summary "New feature"
atlassian-cli issue list --status "In Progress"

# Override when needed
atlassian-cli issue list --jira-project PROD --status Critical
```

## JQL (JIRA Query Language) Support

The `--jql` flag supports full JQL syntax for advanced filtering:

```bash
# Recent issues
atlassian-cli issue list --jql "created >= -7d ORDER BY created DESC"

# Issues assigned to current user
atlassian-cli issue list --jql "assignee = currentUser()"

# Complex query
atlassian-cli issue list --jql "project = DEMO AND status IN ('To Do', 'In Progress') AND priority >= High AND created >= -30d ORDER BY priority DESC, created ASC"
```

## Error Handling

The CLI provides clear error messages with actionable suggestions:

```bash
# Missing project configuration
$ atlassian-cli issue create --type Story --summary "Test"
Error: No JIRA project specified. Use one of:
  1. Set default: atlassian-cli config set default_jira_project DEMO
  2. Use flag: --jira-project DEMO
  3. Set environment: export ATLASSIAN_DEFAULT_JIRA_PROJECT=DEMO

# Invalid issue key
$ atlassian-cli issue get INVALID-123
Error: Issue 'INVALID-123' not found
Suggestion: Check the issue key format (PROJECT-NUMBER)

# Authentication required
$ atlassian-cli issue list
Error: Not authenticated with any Atlassian instance
Run: atlassian-cli auth login --server <url> --email <email> --token <token>
```