# Shell Completion

The `completion` command generates shell completion scripts for enhanced CLI experience.

## atlassian-cli completion

Generate shell completion scripts for bash, zsh, fish, and PowerShell.

### Usage

```bash
atlassian-cli completion <shell>
```

### Supported Shells

- `bash` - Bash completion
- `zsh` - Zsh completion  
- `fish` - Fish completion
- `powershell` - PowerShell completion

## Installation Instructions

### Bash

#### Option 1: Source directly (temporary)
```bash
source <(atlassian-cli completion bash)
```

#### Option 2: Add to ~/.bashrc (permanent)
```bash
echo 'source <(atlassian-cli completion bash)' >> ~/.bashrc
source ~/.bashrc
```

#### Option 3: System-wide installation
```bash
atlassian-cli completion bash | sudo tee /etc/bash_completion.d/atlassian-cli
```

### Zsh

#### Option 1: Source directly (temporary)
```bash
source <(atlassian-cli completion zsh)
```

#### Option 2: Add to ~/.zshrc (permanent)
```bash
echo 'source <(atlassian-cli completion zsh)' >> ~/.zshrc
source ~/.zshrc
```

#### Option 3: Install to completion directory
```bash
atlassian-cli completion zsh > "${fpath[1]}/_atlassian-cli"
```

### Fish

#### Option 1: Source directly (temporary)
```bash
atlassian-cli completion fish | source
```

#### Option 2: Install permanently
```bash
atlassian-cli completion fish > ~/.config/fish/completions/atlassian-cli.fish
```

### PowerShell

#### Add to PowerShell profile
```powershell
atlassian-cli completion powershell | Out-String | Invoke-Expression
```

#### Or save to profile file
```powershell
atlassian-cli completion powershell >> $PROFILE
```

## Completion Features

### Command Completion
- All commands and subcommands
- Global flags and command-specific flags
- Help text for each option

### Dynamic Value Completion
- JIRA project keys (from configured instances)
- Confluence space keys
- Issue types and statuses
- User names and email addresses
- Output formats (json, table, yaml)

### Smart Context Completion
- Issue keys with project prefix
- Page IDs and titles
- Configuration keys and values

## Examples

### Command Completion
```bash
# Type and press TAB
atlassian-cli <TAB>
# Shows: auth, config, issue, page, project, space, cache, completion

atlassian-cli issue <TAB>
# Shows: create, get, list, update

atlassian-cli issue create --<TAB>
# Shows: --type, --summary, --description, --assignee, --priority, etc.
```

### Value Completion
```bash
# Project completion (if configured)
atlassian-cli issue list --jira-project <TAB>
# Shows: DEMO, PROD, TEST (your configured projects)

# Output format completion
atlassian-cli issue list --output <TAB>
# Shows: json, table, yaml

# Status completion
atlassian-cli issue list --status <TAB>
# Shows: "To Do", "In Progress", "Done", "Review", etc.
```

### Configuration Completion
```bash
# Configuration key completion
atlassian-cli config set <TAB>
# Shows: default_jira_project, default_confluence_space, output, cache_ttl, etc.

# Configuration value completion for specific keys
atlassian-cli config set output <TAB>
# Shows: json, table, yaml
```

## Troubleshooting

### Completion Not Working

1. **Verify installation**:
   ```bash
   # Check if completion is loaded
   complete -p atlassian-cli
   ```

2. **Reload shell configuration**:
   ```bash
   # Bash
   source ~/.bashrc
   
   # Zsh
   source ~/.zshrc
   
   # Fish
   source ~/.config/fish/config.fish
   ```

3. **Check shell compatibility**:
   ```bash
   # Verify shell version
   echo $SHELL
   bash --version
   zsh --version
   fish --version
   ```

### Slow Completion

If completion is slow, it may be due to network calls for dynamic completion:

1. **Configure caching**:
   ```bash
   atlassian-cli config set cache_enabled true
   atlassian-cli config set cache_ttl 10m
   ```

2. **Use offline mode** (if available):
   ```bash
   export ATLASSIAN_CLI_OFFLINE=true
   ```

### Missing Dynamic Completion

Dynamic completion requires authentication and configuration:

1. **Authenticate first**:
   ```bash
   atlassian-cli auth login --server <url> --email <email> --token <token>
   ```

2. **Set defaults**:
   ```bash
   atlassian-cli config set default_jira_project DEMO
   atlassian-cli config set default_confluence_space DEV
   ```

## Advanced Configuration

### Custom Completion Scripts

You can extend completion by creating custom scripts:

```bash
# ~/.atlassian-cli-completion-custom.sh
_atlassian_cli_custom() {
    # Add custom completion logic here
    local cur="${COMP_WORDS[COMP_CWORD]}"
    
    # Example: complete with custom project list
    if [[ "$cur" == "--jira-project="* ]]; then
        local projects="DEMO PROD TEST STAGING"
        COMPREPLY=($(compgen -W "$projects" -- "${cur#--jira-project=}"))
        return 0
    fi
}

# Register custom completion
complete -F _atlassian_cli_custom atlassian-cli
```

### Environment-Specific Completion

Set up different completion for different environments:

```bash
# ~/.bashrc
if [[ "$ENVIRONMENT" == "production" ]]; then
    # Production-specific completion
    export ATLASSIAN_DEFAULT_JIRA_PROJECT=PROD
    source <(atlassian-cli completion bash)
else
    # Development completion
    export ATLASSIAN_DEFAULT_JIRA_PROJECT=DEV
    source <(atlassian-cli completion bash)
fi
```