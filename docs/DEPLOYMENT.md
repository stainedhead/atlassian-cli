# Deployment Guide

This guide covers deployment strategies and distribution methods for the Atlassian CLI.

## Release Process

### 1. Automated Release Pipeline

The project includes automated release scripts for consistent, reproducible builds:

```bash
# Create a new release
make release

# Or use the release script directly
./scripts/release.sh
```

### 2. Manual Release Steps

If you need to create a release manually:

```bash
# 1. Update version
export VERSION="v1.2.0"

# 2. Tag the release
git tag -a "$VERSION" -m "Release $VERSION"
git push origin "$VERSION"

# 3. Build all platforms
make build-all

# 4. Create GitHub release
gh release create "$VERSION" \
  --title "Atlassian CLI $VERSION" \
  --notes "Release notes here" \
  dist/*
```

## Distribution Methods

### 1. GitHub Releases

Primary distribution method with pre-built binaries:

- **Linux**: `atlassian-cli-linux-amd64`, `atlassian-cli-linux-arm64`
- **macOS**: `atlassian-cli-darwin-amd64`, `atlassian-cli-darwin-arm64`
- **Windows**: `atlassian-cli-windows-amd64.exe`

### 2. Installation Script

Quick installation via curl:

```bash
# Install latest version
curl -sSL https://raw.githubusercontent.com/your-org/atlassian-cli/main/scripts/install.sh | bash

# Install specific version
curl -sSL https://raw.githubusercontent.com/your-org/atlassian-cli/main/scripts/install.sh | VERSION=v1.2.0 bash
```

### 3. Package Managers

#### Homebrew (macOS/Linux)

```ruby
# Formula: atlassian-cli.rb
class AtlassianCli < Formula
  desc "Command-line interface for JIRA and Confluence"
  homepage "https://github.com/your-org/atlassian-cli"
  url "https://github.com/your-org/atlassian-cli/archive/v1.2.0.tar.gz"
  sha256 "..."
  license "MIT"

  depends_on "go" => :build

  def install
    system "make", "build"
    bin.install "bin/atlassian-cli"
    
    # Install shell completions
    generate_completions_from_executable(bin/"atlassian-cli", "completion")
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/atlassian-cli --version")
  end
end
```

#### Chocolatey (Windows)

```xml
<!-- atlassian-cli.nuspec -->
<?xml version="1.0" encoding="utf-8"?>
<package xmlns="http://schemas.microsoft.com/packaging/2015/06/nuspec.xsd">
  <metadata>
    <id>atlassian-cli</id>
    <version>1.2.0</version>
    <title>Atlassian CLI</title>
    <authors>Your Organization</authors>
    <description>Command-line interface for JIRA and Confluence</description>
    <projectUrl>https://github.com/your-org/atlassian-cli</projectUrl>
    <licenseUrl>https://github.com/your-org/atlassian-cli/blob/main/LICENSE</licenseUrl>
    <requireLicenseAcceptance>false</requireLicenseAcceptance>
    <tags>atlassian jira confluence cli</tags>
  </metadata>
  <files>
    <file src="atlassian-cli.exe" target="tools" />
  </files>
</package>
```

#### Snap (Linux)

```yaml
# snapcraft.yaml
name: atlassian-cli
version: '1.2.0'
summary: Command-line interface for JIRA and Confluence
description: |
  Atlassian CLI streamlines development workflows by providing intuitive
  access to JIRA and Confluence operations with smart defaults.

grade: stable
confinement: strict

parts:
  atlassian-cli:
    plugin: go
    source: .
    build-snaps: [go]
    
apps:
  atlassian-cli:
    command: bin/atlassian-cli
    plugs: [network, home]
```

### 4. Docker Images

#### Multi-stage Dockerfile

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN make build

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/bin/atlassian-cli /usr/local/bin/
ENTRYPOINT ["atlassian-cli"]
```

#### Docker Hub Deployment

```bash
# Build and push Docker image
docker build -t your-org/atlassian-cli:latest .
docker build -t your-org/atlassian-cli:v1.2.0 .

docker push your-org/atlassian-cli:latest
docker push your-org/atlassian-cli:v1.2.0
```

## Enterprise Deployment

### 1. Internal Package Repository

For enterprise environments, host binaries in internal repositories:

```bash
# Artifactory/Nexus deployment
curl -u admin:password \
  -T dist/atlassian-cli-linux-amd64 \
  "https://artifactory.company.com/artifactory/tools/atlassian-cli/v1.2.0/"
```

### 2. Configuration Management

#### Ansible Playbook

```yaml
# playbooks/install-atlassian-cli.yml
---
- name: Install Atlassian CLI
  hosts: all
  become: yes
  
  vars:
    atlassian_cli_version: "v1.2.0"
    atlassian_cli_url: "https://github.com/your-org/atlassian-cli/releases/download/{{ atlassian_cli_version }}"
    
  tasks:
    - name: Download Atlassian CLI
      get_url:
        url: "{{ atlassian_cli_url }}/atlassian-cli-linux-amd64"
        dest: "/usr/local/bin/atlassian-cli"
        mode: '0755'
        
    - name: Install shell completion
      shell: |
        atlassian-cli completion bash > /etc/bash_completion.d/atlassian-cli
        atlassian-cli completion zsh > /usr/share/zsh/site-functions/_atlassian-cli
```

#### Chef Cookbook

```ruby
# cookbooks/atlassian-cli/recipes/default.rb
remote_file '/usr/local/bin/atlassian-cli' do
  source "https://github.com/your-org/atlassian-cli/releases/download/v#{node['atlassian_cli']['version']}/atlassian-cli-linux-amd64"
  mode '0755'
  action :create
end

execute 'install-bash-completion' do
  command 'atlassian-cli completion bash > /etc/bash_completion.d/atlassian-cli'
  creates '/etc/bash_completion.d/atlassian-cli'
end
```

### 3. Security Considerations

#### Binary Verification

```bash
# Generate checksums during build
sha256sum dist/* > dist/checksums.txt

# Verify downloads
sha256sum -c checksums.txt
```

#### Code Signing

```bash
# Sign binaries (macOS)
codesign -s "Developer ID Application: Your Name" dist/atlassian-cli-darwin-amd64

# Sign binaries (Windows)
signtool sign /f certificate.p12 /p password dist/atlassian-cli-windows-amd64.exe
```

## Monitoring and Updates

### 1. Update Notifications

Built-in update checking:

```go
// internal/update/checker.go
func CheckForUpdates() (*Version, error) {
    resp, err := http.Get("https://api.github.com/repos/your-org/atlassian-cli/releases/latest")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var release struct {
        TagName string `json:"tag_name"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
        return nil, err
    }
    
    return ParseVersion(release.TagName)
}
```

### 2. Telemetry (Optional)

Anonymous usage statistics:

```go
// internal/telemetry/collector.go
type Event struct {
    Command   string    `json:"command"`
    Version   string    `json:"version"`
    OS        string    `json:"os"`
    Timestamp time.Time `json:"timestamp"`
}

func (c *Collector) Track(command string) {
    if !c.enabled {
        return
    }
    
    event := Event{
        Command:   command,
        Version:   version.Current,
        OS:        runtime.GOOS,
        Timestamp: time.Now(),
    }
    
    // Send asynchronously
    go c.send(event)
}
```

### 3. Health Monitoring

```bash
# Health check endpoint for monitoring
atlassian-cli auth status --output json | jq '.status'

# Version check
atlassian-cli --version

# Configuration validation
atlassian-cli config list --output json
```

## Rollback Procedures

### 1. Version Rollback

```bash
# Install specific version
curl -L https://github.com/your-org/atlassian-cli/releases/download/v1.1.0/atlassian-cli-linux-amd64 \
  -o /usr/local/bin/atlassian-cli
chmod +x /usr/local/bin/atlassian-cli
```

### 2. Configuration Backup

```bash
# Backup configuration
cp ~/.atlassian-cli/config.yaml ~/.atlassian-cli/config.yaml.backup

# Restore configuration
cp ~/.atlassian-cli/config.yaml.backup ~/.atlassian-cli/config.yaml
```

## Performance Optimization

### 1. Binary Size Optimization

```bash
# Build with optimizations
go build -ldflags="-s -w" -trimpath

# Use UPX compression (optional)
upx --best dist/atlassian-cli-*
```

### 2. Startup Performance

```go
// Lazy loading of heavy dependencies
var jiraClient *JiraClient

func getJiraClient() *JiraClient {
    if jiraClient == nil {
        jiraClient = NewJiraClient()
    }
    return jiraClient
}
```

### 3. Caching Strategy

```bash
# Configure aggressive caching for CI/CD
export ATLASSIAN_CACHE_TTL=1h
export ATLASSIAN_CACHE_ENABLED=true
```

This deployment guide provides comprehensive strategies for distributing and managing the Atlassian CLI across different environments and platforms.