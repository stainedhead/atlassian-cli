# Authentication Commands

The `auth` command group manages authentication with Atlassian instances.

## atlassian-cli auth login

Authenticate with an Atlassian instance using API token.

### Usage

```bash
atlassian-cli auth login --server <url> --email <email> --token <token>
```

### Flags

- `--server` (required) - Atlassian instance URL (e.g., https://company.atlassian.net)
- `--email` (required) - Your Atlassian account email
- `--token` (required) - API token from id.atlassian.com/manage/api-tokens

### Examples

```bash
# Basic authentication
atlassian-cli auth login \
  --server https://company.atlassian.net \
  --email user@company.com \
  --token abcd1234efgh5678

# Interactive mode (prompts for missing values)
atlassian-cli auth login --server https://company.atlassian.net
```

### Notes

- API tokens are stored securely in memory during the session
- Tokens are validated during login to ensure they work
- Multiple instances can be configured by running login multiple times

## atlassian-cli auth logout

Clear stored authentication credentials.

### Usage

```bash
atlassian-cli auth logout [--server <url>]
```

### Flags

- `--server` (optional) - Specific server to logout from. If omitted, clears all credentials.

### Examples

```bash
# Logout from specific server
atlassian-cli auth logout --server https://company.atlassian.net

# Logout from all servers
atlassian-cli auth logout
```

## atlassian-cli auth status

Show current authentication status.

### Usage

```bash
atlassian-cli auth status [--server <url>]
```

### Flags

- `--server` (optional) - Check status for specific server

### Examples

```bash
# Check status for specific server
atlassian-cli auth status --server https://company.atlassian.net

# Check status for all configured servers
atlassian-cli auth status
```

### Output

```
Server: https://company.atlassian.net
Status: Authenticated
Email: user@company.com
Last Login: 2024-01-15 10:30:00
```

## Security

- API tokens are never stored to disk
- Credentials are cleared when the CLI session ends
- All API communication uses HTTPS
- Token validation occurs during login to prevent invalid credentials