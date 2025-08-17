# Atlassian CLI - Technical Details

## Architecture Overview

### System Design

The Atlassian CLI follows a modular, layered architecture designed for maintainability, testability, and extensibility.

```
┌─────────────────────────────────────────────────────────────┐
│                    Command Layer (cmd/)                     │
├─────────────────────────────────────────────────────────────┤
│                  Business Logic (internal/)                 │
├─────────────────────────────────────────────────────────────┤
│                External Dependencies (pkg/)                  │
└─────────────────────────────────────────────────────────────┘
```

### Core Components

#### Command Layer (`cmd/`)
- **Root Command**: Global flags, configuration initialization, subcommand registration
- **Auth Commands**: Authentication workflow management
- **Issue Commands**: JIRA issue lifecycle operations
- **Config Commands**: Configuration management and smart defaults
- **Page/Space Commands**: Confluence content operations

#### Business Logic (`internal/`)
- **Authentication**: Token management and credential validation
- **Configuration**: Hierarchical configuration resolution
- **API Clients**: JIRA and Confluence API abstractions
- **Output Formatting**: Multi-format output rendering
- **Caching**: Intelligent response caching with TTL
- **Retry Logic**: Exponential backoff with jitter

#### External Interface (`pkg/`)
- **Public API**: Stable interfaces for external integration
- **Types**: Shared data structures and constants

## Technology Stack

### Core Technologies
- **Language**: Go 1.21+
- **CLI Framework**: Cobra (command structure and parsing)
- **Configuration**: Viper (hierarchical configuration management)
- **HTTP Client**: Standard library with custom retry logic
- **Testing**: Go testing package with testify assertions

### Dependencies
```go
// Core dependencies
github.com/spf13/cobra      // CLI framework
github.com/spf13/viper      // Configuration management
github.com/stretchr/testify // Testing assertions
gopkg.in/yaml.v3           // YAML processing

// Optional dependencies
github.com/olekukonko/tablewriter // Table formatting
```

### Build System
- **Make**: Primary build automation
- **Go Modules**: Dependency management
- **golangci-lint**: Static analysis and linting
- **gosec**: Security vulnerability scanning

## Configuration System

### Hierarchical Resolution

The configuration system implements a four-tier hierarchy:

```go
type ConfigResolver struct {
    flags       map[string]interface{}
    environment map[string]string
    configFile  map[string]interface{}
    defaults    map[string]interface{}
}

func (r *ConfigResolver) Resolve(key string) (interface{}, error) {
    // 1. Command flags (highest priority)
    if val, exists := r.flags[key]; exists {
        return val, nil
    }
    
    // 2. Environment variables
    if val, exists := r.environment[key]; exists {
        return val, nil
    }
    
    // 3. Configuration file
    if val, exists := r.configFile[key]; exists {
        return val, nil
    }
    
    // 4. Built-in defaults (lowest priority)
    if val, exists := r.defaults[key]; exists {
        return val, nil
    }
    
    return nil, ErrConfigNotFound
}
```

### Configuration File Format

```yaml
# ~/.atlassian-cli/config.yaml
default_jira_project: "DEMO"
default_confluence_space: "DEV"
output: "table"
cache_ttl: "5m"
cache_enabled: true
jira_timeout: "30s"
confluence_timeout: "30s"
debug: false
```

### Environment Variables

```bash
# Naming convention: ATLASSIAN_<CONFIG_KEY>
ATLASSIAN_DEFAULT_JIRA_PROJECT=DEMO
ATLASSIAN_DEFAULT_CONFLUENCE_SPACE=DEV
ATLASSIAN_OUTPUT=json
ATLASSIAN_CACHE_TTL=10m
ATLASSIAN_DEBUG=true
```

## Authentication System

### Token Management

```go
type TokenManager interface {
    Store(server, email, token string) error
    Retrieve(server string) (*Credentials, error)
    Delete(server string) error
    List() ([]string, error)
}

type MemoryTokenManager struct {
    credentials map[string]*Credentials
    mutex       sync.RWMutex
}
```

### Security Features
- **Memory-only storage**: No credentials persisted to disk
- **Session-based**: Credentials cleared on CLI exit
- **Validation**: Token validation during login
- **Multi-instance**: Support for multiple Atlassian instances

## API Client Architecture

### Client Interface

```go
type JiraClient interface {
    CreateIssue(ctx context.Context, req *CreateIssueRequest) (*Issue, error)
    GetIssue(ctx context.Context, key string) (*Issue, error)
    UpdateIssue(ctx context.Context, key string, req *UpdateIssueRequest) (*Issue, error)
    SearchIssues(ctx context.Context, jql string, opts *SearchOptions) (*SearchResult, error)
}

type ConfluenceClient interface {
    CreatePage(ctx context.Context, req *CreatePageRequest) (*Page, error)
    GetPage(ctx context.Context, id string) (*Page, error)
    UpdatePage(ctx context.Context, id string, req *UpdatePageRequest) (*Page, error)
    ListPages(ctx context.Context, spaceKey string, opts *ListOptions) (*PageList, error)
}
```

### HTTP Client Configuration

```go
type HTTPClient struct {
    client      *http.Client
    baseURL     string
    credentials *Credentials
    retryConfig *RetryConfig
    cache       *Cache
}

type RetryConfig struct {
    MaxAttempts int
    BaseDelay   time.Duration
    MaxDelay    time.Duration
    Multiplier  float64
    Jitter      bool
}
```

## Caching System

### Cache Implementation

```go
type Cache struct {
    store map[string]*CacheEntry
    mutex sync.RWMutex
    ttl   time.Duration
}

type CacheEntry struct {
    Data      interface{}
    ExpiresAt time.Time
}

func (c *Cache) Get(key string) (interface{}, bool) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    
    entry, exists := c.store[key]
    if !exists || time.Now().After(entry.ExpiresAt) {
        return nil, false
    }
    
    return entry.Data, true
}
```

### Cache Strategy
- **TTL-based expiration**: Default 5-minute TTL
- **Key generation**: URL + parameters hash
- **Memory management**: LRU eviction for large caches
- **Cache warming**: Proactive cache population

## Output Formatting

### Formatter Interface

```go
type Formatter interface {
    Format(data interface{}) (string, error)
}

type OutputFormatter struct {
    format string
}

func (f *OutputFormatter) Format(data interface{}) (string, error) {
    switch f.format {
    case "json":
        return f.formatJSON(data)
    case "yaml":
        return f.formatYAML(data)
    case "table":
        return f.formatTable(data)
    default:
        return "", ErrUnsupportedFormat
    }
}
```

### Format Support
- **JSON**: Machine-readable, perfect for scripting
- **YAML**: Human-readable, configuration-friendly
- **Table**: Terminal-optimized, interactive use

## Error Handling

### Error Types

```go
type APIError struct {
    StatusCode int
    Message    string
    Details    map[string]interface{}
}

type ConfigError struct {
    Key     string
    Message string
}

type AuthError struct {
    Server  string
    Message string
}
```

### Error Recovery

```go
func (c *HTTPClient) executeWithRetry(req *http.Request) (*http.Response, error) {
    var lastErr error
    
    for attempt := 0; attempt < c.retryConfig.MaxAttempts; attempt++ {
        resp, err := c.client.Do(req)
        if err == nil && resp.StatusCode < 500 {
            return resp, nil
        }
        
        lastErr = err
        if attempt < c.retryConfig.MaxAttempts-1 {
            delay := c.calculateBackoff(attempt)
            time.Sleep(delay)
        }
    }
    
    return nil, lastErr
}
```

## Testing Strategy

### Test Structure

```
test/
├── fixtures/           # Test data and mock responses
├── mocks/             # API client mocks
└── integration/       # End-to-end test helpers
```

### Testing Patterns

```go
func TestIssueCreate(t *testing.T) {
    tests := []struct {
        name    string
        input   *CreateIssueRequest
        want    *Issue
        wantErr bool
    }{
        {
            name: "valid issue creation",
            input: &CreateIssueRequest{
                Type:    "Story",
                Summary: "Test issue",
            },
            want: &Issue{
                Key:     "DEMO-123",
                Summary: "Test issue",
            },
            wantErr: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            client := &MockJiraClient{}
            client.On("CreateIssue", mock.Anything, tt.input).Return(tt.want, nil)
            
            got, err := client.CreateIssue(context.Background(), tt.input)
            
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.want, got)
            }
        })
    }
}
```

### Coverage Requirements
- **Unit tests**: 90%+ coverage for business logic
- **Integration tests**: API client behavior validation
- **End-to-end tests**: Complete workflow validation

## Performance Characteristics

### Benchmarks

```go
func BenchmarkConfigResolution(b *testing.B) {
    resolver := NewConfigResolver()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = resolver.Resolve("default_jira_project")
    }
}

func BenchmarkCacheAccess(b *testing.B) {
    cache := NewCache(5 * time.Minute)
    cache.Set("test-key", "test-value")
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = cache.Get("test-key")
    }
}
```

### Performance Targets
- **Startup time**: < 100ms for cached operations
- **API response**: < 500ms for cached responses
- **Memory usage**: < 50MB for typical operations
- **Cache hit ratio**: > 80% for repeated operations

## Security Considerations

### Threat Model
- **Credential exposure**: Mitigated by memory-only storage
- **Man-in-the-middle**: HTTPS enforcement for all API calls
- **Token leakage**: No logging of sensitive data
- **Injection attacks**: Input validation and parameterized queries

### Security Controls
- **Input validation**: All user inputs validated and sanitized
- **HTTPS enforcement**: TLS 1.2+ required for all connections
- **Token rotation**: Support for token refresh workflows
- **Audit logging**: Structured logging for security events

## Deployment Architecture

### Binary Distribution
- **Multi-platform**: Linux (amd64, arm64), macOS (amd64, arm64), Windows (amd64)
- **Static linking**: No external dependencies required
- **Compression**: UPX compression for smaller binaries
- **Checksums**: SHA256 verification for all releases

### Installation Methods
- **Direct download**: GitHub releases with automated installation script
- **Package managers**: Homebrew, Chocolatey, Snap packages
- **Container**: Docker images for CI/CD environments
- **Source build**: Go toolchain compilation

### Configuration Management
- **File locations**: XDG Base Directory specification compliance
- **Permissions**: Secure file permissions (600) for config files
- **Migration**: Automatic config format migration
- **Validation**: Schema validation for configuration files

This technical architecture provides a solid foundation for reliable, maintainable, and extensible CLI operations while maintaining high performance and security standards.