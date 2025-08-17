
**Creating a modern Go CLI tool that transforms REST API complexity into intuitive developer workflows requires careful architecture decisions, robust authentication patterns, and thoughtful command design. This comprehensive guide provides the blueprint for building production-ready Atlassian tooling.**

The landscape for Atlassian CLI development has matured significantly, with established patterns from tools like GitHub CLI providing proven architectures. For developers seeking to streamline JIRA and Confluence workflows, the combination of Go's performance, Cobra's command structure, and modern authentication patterns creates powerful automation possibilities. This research reveals the essential components, from API endpoint selection through secure credential management, needed to build enterprise-grade developer tooling.

## Table of Contents

1. [Understanding the Atlassian API landscape](#understanding-the-atlassian-api-landscape)
2. [Go ecosystem evaluation for Atlassian APIs](#go-ecosystem-evaluation-for-atlassian-apis)
3. [Architectural patterns for robust CLI design](#architectural-patterns-for-robust-cli-design)
4. [Implementation examples and workflow patterns](#implementation-examples-and-workflow-patterns)
5. [Testing strategies and quality assurance](#testing-strategies-and-quality-assurance)
6. [Conclusion](#conclusion)
7. [Appendix: CLI Command Reference](#appendix-cli-command-reference)

## Understanding the Atlassian API landscape

Both JIRA and Confluence offer comprehensive REST APIs designed for developer integration, though each serves distinct purposes in the development workflow.

### JIRA Cloud REST API endpoints for developer workflows

JIRA provides two primary API versions with **Version 3 recommended for new development** due to its support for Atlassian Document Format (ADF) and improved data structures. The base URL pattern follows `https://your-domain.atlassian.net/rest/api/3/` for core operations.

**Issue management forms the backbone of developer JIRA usage.** The most critical endpoints include issue creation (`POST /rest/api/3/issue`), retrieval with expansion parameters (`GET /rest/api/3/issue/{key}?expand=transitions,changelog`), and bulk operations for efficiency (`POST /rest/api/3/issue/bulk`). Search capabilities through JQL (JIRA Query Language) enable powerful filtering with queries like `project = DEMO AND assignee = currentUser() AND status IN ("In Progress", "Code Review") ORDER BY created DESC`.

**Agile workflow support requires the JIRA Software API** at `/rest/agile/1.0/`. Essential operations include board management (`GET /rest/agile/1.0/board/{boardId}/issue`), sprint operations (`POST /rest/agile/1.0/sprint` for creation), and epic handling (`GET /rest/agile/1.0/epic/{epicId}/issue`). These endpoints enable CLI tools to integrate seamlessly with development team workflows.

**Custom field handling** presents unique challenges, as fields use identifiers like `customfield_10000` rather than human-readable names. The field metadata endpoint (`GET /rest/api/3/field`) provides mapping between IDs and field definitions, essential for robust CLI implementation.

### Confluence Cloud REST API for content operations

Confluence APIs support comprehensive content management through both v1 and v2 endpoints, with **v1 providing broader feature coverage** despite v2's performance improvements.

**Content operations center around pages and blog posts** using the `/wiki/rest/api/content` endpoint family. Page creation requires understanding Confluence's storage format - an XHTML-based structure that supports both basic HTML and Confluence-specific macros. The API supports format conversion (`/wiki/rest/api/contentbody/convert/{to}`) enabling CLI tools to accept wiki markup input while storing proper format internally.

**Search capabilities use CQL (Confluence Query Language)** through `/wiki/rest/api/search` with powerful filtering options like `type=page AND space=DEV AND lastModified >= "2024-01-01" ORDER BY lastModified DESC`. This enables CLI tools to provide sophisticated content discovery features.

**Space management and navigation** operations (`/wiki/rest/api/space`) allow CLI tools to understand organizational structure and provide context-aware content creation. Template operations, though limited in v1, enable standardized page creation workflows essential for documentation automation.

### Authentication patterns and rate limiting considerations

**API token authentication provides the most straightforward approach** for CLI tools targeting Atlassian Cloud. The pattern uses HTTP Basic Auth with email and API token: `Authorization: Basic base64(email:api_token)`. Tokens created at id.atlassian.com provide configurable expiration periods and fine-grained permission control.

**Rate limiting implementation uses cost-based budgeting** rather than fixed limits, with `429 Too Many Requests` responses including `Retry-After` headers. Successful CLI implementations require exponential backoff strategies and intelligent batching of operations. The APIs support expansion parameters (`?expand=names,renderedFields,transitions`) to reduce request counts by fetching related data in single calls.

## Go ecosystem evaluation for Atlassian APIs

The Go ecosystem lacks official Atlassian SDKs, making community library evaluation critical for production CLI development.

### go-atlassian emerges as the comprehensive solution

**The go-atlassian library (github.com/ctreminiom/go-atlassian/v2) stands out as the top recommendation** for unified JIRA and Confluence development. Unlike alternatives that focus on single products, this library provides comprehensive coverage including JIRA v2/v3, Confluence v1/v2, JIRA Agile, Service Management, and Admin APIs.

The library's architecture follows modern Go practices with service-oriented design inspired by google/go-github:

```go
import "github.com/ctreminiom/go-atlassian/v2/jira/v3"

atlassian, err := v3.New(nil, "https://your-domain.atlassian.net")
if err != nil {
    return err
}
atlassian.Auth.SetBasicAuth("email@example.com", "api-token")

// Service-oriented API calls
issue, response, err := atlassian.Issue.Get(ctx, "PROJ-123", nil, []string{"transitions"})
sprint, response, err := atlassian.Sprint.Create(ctx, &models.SprintPayloadScheme{
    Name:            "Sprint 1",
    StartDate:       "2024-01-01T10:00:00.000Z",
    EndDate:         "2024-01-15T10:00:00.000Z",
    OriginBoardID:   123,
})
```

**Authentication flexibility** includes Basic Auth, OAuth 2.0 with auto-renewal, API tokens, and Personal Access Tokens. The library handles complex authentication flows while providing simple configuration interfaces for CLI integration.

### Alternative approaches and their trade-offs

**For JIRA-only requirements, andygrunwald/go-jira offers mature functionality** with 1,400+ GitHub stars and stable v1.16.0 release. However, its limitation to JIRA operations requires additional libraries for Confluence integration, increasing dependency complexity.

**Generic HTTP client approaches** using `net/http` provide maximum control but require significant development overhead. Manual JSON marshaling, authentication handling, and error processing make this approach suitable only for highly specialized requirements or minimal API usage.

The comparison reveals **specialized SDKs provide superior developer experience** through typed structs, built-in error handling, and maintained API compatibility. For production CLI tools, the development velocity and reliability advantages outweigh dependency concerns.

## Architectural patterns for robust CLI design

Modern Go CLI development follows established patterns proven by tools like GitHub CLI, kubectl, and Docker CLI. These patterns emphasize maintainability, testability, and user experience.

### Command structure implementation with Cobra

**Cobra framework dominates Go CLI development** for compelling reasons: nested command support, automatic help generation, and extensive plugin ecosystem. The architecture separates command definitions from business logic, enabling clean testing and maintenance.

```go
// cmd/root.go - Foundation structure
var rootCmd = &cobra.Command{
    Use:   "atlassian-cli",
    Short: "Developer toolkit for JIRA and Confluence",
    Long:  `Streamline development workflows with integrated JIRA and Confluence operations`,
}

func Execute() error {
    return rootCmd.Execute()
}

func init() {
    cobra.OnInitialize(initConfig)
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")
    rootCmd.PersistentFlags().StringVar(&outputFormat, "output", "table", "output format (json, table, yaml)")
    rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}
```

**Command organization follows domain-driven patterns** with separate packages for issue management, project operations, and confluence content. This structure mirrors the API organization while providing intuitive command hierarchies:

```bash
atlassian-cli issue create --project DEMO --type Story --summary "Feature request"
atlassian-cli issue list --assignee currentUser --status "In Progress"
atlassian-cli confluence page create --space DEV --title "API Documentation"
atlassian-cli project boards --project DEMO
```

### Configuration management with hierarchical precedence

**Viper integration provides sophisticated configuration management** following the standard precedence: explicit calls > command flags > environment variables > configuration files > defaults. This approach accommodates diverse deployment scenarios from developer laptops to CI/CD pipelines.

```go
type Config struct {
    APIEndpoint string        `mapstructure:"api_endpoint" validate:"required,url"`
    Email       string        `mapstructure:"email" validate:"required,email"`
    Token       string        `mapstructure:"token" validate:"required"`
    Timeout     time.Duration `mapstructure:"timeout"`
    Output      string        `mapstructure:"output" validate:"oneof=json table yaml"`
    Debug       bool          `mapstructure:"debug"`
}

func LoadConfig() (*Config, error) {
    // Environment variable configuration
    viper.SetEnvPrefix("ATLASSIAN")
    viper.AutomaticEnv()
    viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
    
    // Configuration file locations
    viper.AddConfigPath("$HOME/.atlassian-cli")
    viper.AddConfigPath(".")
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    
    // Defaults
    viper.SetDefault("timeout", 30*time.Second)
    viper.SetDefault("output", "table")
    viper.SetDefault("api_endpoint", "https://your-domain.atlassian.net")
    
    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            return nil, err
        }
    }
    
    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("config unmarshal failed: %w", err)
    }
    
    return &config, validator.New().Struct(&config)
}
```

### Authentication and secure credential storage

**Token-based authentication requires careful security consideration** beyond simple file storage. Production CLI tools implement OS keychain integration with secure fallback mechanisms.

Modern approaches use credential helper libraries for cross-platform compatibility:

```go
type TokenManager struct {
    store credentials.Helper
}

func NewTokenManager() *TokenManager {
    var store credentials.Helper
    
    switch runtime.GOOS {
    case "darwin":
        store = osxkeychain.Osxkeychain{}
    case "linux":
        store = secretservice.Secretservice{}
    case "windows":
        store = wincred.Wincred{}
    default:
        store = &encryptedFileStore{} // Fallback with encryption
    }
    
    return &TokenManager{store: store}
}

func (tm *TokenManager) SaveToken(serverURL, email, token string) error {
    creds := &credentials.Credentials{
        ServerURL: serverURL,
        Username:  email,
        Secret:    token,
    }
    return tm.store.Add(creds)
}
```

**Authentication flow implementation** should guide users through token creation while maintaining security. Interactive prompts combined with secure storage create user-friendly onboarding:

```go
func (a *AuthCommand) runLogin(cmd *cobra.Command, args []string) error {
    fmt.Println("Please create an API token at: https://id.atlassian.com/manage/api-tokens")
    
    email, err := promptForInput("Email address: ")
    if err != nil {
        return err
    }
    
    token, err := promptForPassword("API token: ")
    if err != nil {
        return err
    }
    
    // Verify credentials
    client := api.NewClient(email, token)
    user, err := client.GetCurrentUser()
    if err != nil {
        return fmt.Errorf("authentication failed: %w", err)
    }
    
    // Store securely
    tm := auth.NewTokenManager()
    if err := tm.SaveToken(viper.GetString("api_endpoint"), email, token); err != nil {
        return fmt.Errorf("failed to store credentials: %w", err)
    }
    
    fmt.Printf("✓ Authenticated as %s\n", user.DisplayName)
    return nil
}
```

## Implementation examples and workflow patterns

Successful CLI implementations require translating REST API complexity into intuitive command interfaces that match developer mental models.

### Issue management command implementation

**Issue operations demonstrate the full spectrum of CLI patterns** from simple retrieval to complex creation workflows. The implementation balances API flexibility with command-line simplicity:

```go
// cmd/issue/create.go
func newIssueCreateCmd(apiClient *api.Client) *cobra.Command {
    var opts IssueCreateOptions
    
    cmd := &cobra.Command{
        Use:   "create",
        Short: "Create a new issue",
        Example: `  atlassian-cli issue create --project DEMO --type Story --summary "New feature"
  atlassian-cli issue create --project DEMO --type Bug --summary "Fix login" --description "Login fails on mobile"`,
        RunE: func(cmd *cobra.Command, args []string) error {
            return runIssueCreate(apiClient, &opts)
        },
    }
    
    cmd.Flags().StringVar(&opts.Project, "project", "", "project key (required)")
    cmd.Flags().StringVar(&opts.IssueType, "type", "Task", "issue type")
    cmd.Flags().StringVar(&opts.Summary, "summary", "", "issue summary (required)")
    cmd.Flags().StringVar(&opts.Description, "description", "", "issue description")
    cmd.Flags().StringVar(&opts.Assignee, "assignee", "", "assignee username or 'me'")
    cmd.Flags().StringVar(&opts.Priority, "priority", "Medium", "issue priority")
    cmd.Flags().StringSliceVar(&opts.Labels, "labels", nil, "issue labels")
    
    cmd.MarkFlagRequired("project")
    cmd.MarkFlagRequired("summary")
    
    return cmd
}

func runIssueCreate(client *api.Client, opts *IssueCreateOptions) error {
    // Validate project exists
    project, err := client.GetProject(opts.Project)
    if err != nil {
        return fmt.Errorf("invalid project %q: %w", opts.Project, err)
    }
    
    // Resolve assignee
    var assigneeID string
    if opts.Assignee == "me" {
        user, err := client.GetCurrentUser()
        if err != nil {
            return fmt.Errorf("failed to get current user: %w", err)
        }
        assigneeID = user.AccountID
    }
    
    // Create issue payload
    issuePayload := &api.IssueCreateRequest{
        Fields: api.IssueFields{
            Project:     api.Project{Key: opts.Project},
            IssueType:   api.IssueType{Name: opts.IssueType},
            Summary:     opts.Summary,
            Description: formatDescription(opts.Description),
            Assignee:    api.User{AccountID: assigneeID},
            Priority:    api.Priority{Name: opts.Priority},
            Labels:      opts.Labels,
        },
    }
    
    issue, err := client.CreateIssue(issuePayload)
    if err != nil {
        return fmt.Errorf("issue creation failed: %w", err)
    }
    
    fmt.Printf("✓ Created issue %s: %s\n", issue.Key, issue.Fields.Summary)
    fmt.Printf("  URL: %s/browse/%s\n", client.BaseURL, issue.Key)
    
    return nil
}
```

### Confluence content management workflows

**Content operations require understanding Confluence's storage format** while providing user-friendly input methods. The CLI abstracts format complexity through intelligent content processing:

```go
// cmd/confluence/page.go
func runPageCreate(client *api.Client, opts *PageCreateOptions) error {
    // Process content based on input method
    var content string
    var err error
    
    switch {
    case opts.ContentFile != "":
        content, err = processContentFile(opts.ContentFile)
    case opts.Template != "":
        content, err = processTemplate(opts.Template, opts.TemplateVars)
    default:
        content = convertToStorageFormat(opts.Content)
    }
    
    if err != nil {
        return fmt.Errorf("content processing failed: %w", err)
    }
    
    pagePayload := &api.PageCreateRequest{
        Type:  "page",
        Title: opts.Title,
        Space: api.Space{Key: opts.Space},
        Body: api.Body{
            Storage: api.Content{
                Value:          content,
                Representation: "storage",
            },
        },
    }
    
    // Handle parent page relationship
    if opts.Parent != "" {
        parent, err := client.GetPageByTitle(opts.Space, opts.Parent)
        if err != nil {
            return fmt.Errorf("parent page not found: %w", err)
        }
        pagePayload.Ancestors = []api.Page{{ID: parent.ID}}
    }
    
    page, err := client.CreatePage(pagePayload)
    if err != nil {
        return fmt.Errorf("page creation failed: %w", err)
    }
    
    fmt.Printf("✓ Created page: %s\n", page.Title)
    fmt.Printf("  URL: %s%s\n", client.BaseURL, page.Links.WebUI)
    
    return nil
}

func processContentFile(filename string) (string, error) {
    content, err := os.ReadFile(filename)
    if err != nil {
        return "", err
    }
    
    // Auto-detect format and convert
    switch filepath.Ext(filename) {
    case ".md":
        return convertMarkdownToStorage(string(content)), nil
    case ".wiki":
        return convertWikiToStorage(string(content)), nil
    default:
        return string(content), nil
    }
}
```

### Error handling and user experience patterns

**Production CLI tools require sophisticated error handling** that provides actionable feedback while maintaining technical accuracy for debugging. The implementation follows GitHub CLI patterns for consistent user experience:

```go
type CLIError struct {
    Message    string
    Suggestion string
    ExitCode   int
    Details    error
}

func (e CLIError) Error() string {
    if e.Suggestion != "" {
        return fmt.Sprintf("%s\n\nSuggestion: %s", e.Message, e.Suggestion)
    }
    return e.Message
}

func FormatAPIError(err error) error {
    var apiErr *api.APIError
    if errors.As(err, &apiErr) {
        switch apiErr.StatusCode {
        case 401:
            return CLIError{
                Message:    "Authentication failed",
                Suggestion: "Run 'atlassian-cli auth login' to authenticate",
                ExitCode:   1,
                Details:    err,
            }
        case 403:
            return CLIError{
                Message:    "Access denied",
                Suggestion: "Check that you have permission for this operation",
                ExitCode:   1,
                Details:    err,
            }
        case 404:
            return CLIError{
                Message:    "Resource not found",
                Suggestion: "Verify the project key, issue key, or space key is correct",
                ExitCode:   1,
                Details:    err,
            }
        }
    }
    return err
}

// Progress indication for long operations
func withProgress(message string, fn func() error) error {
    if !isatty.IsTerminal(os.Stdout.Fd()) {
        return fn()
    }
    
    spinner := spin.New()
    done := make(chan bool)
    
    go func() {
        for {
            select {
            case <-done:
                return
            default:
                fmt.Printf("\r%s %s", message, spinner.Next())
                time.Sleep(100 * time.Millisecond)
            }
        }
    }()
    
    err := fn()
    done <- true
    
    if err != nil {
        fmt.Printf("\r%s ✗\n", message)
    } else {
        fmt.Printf("\r%s ✓\n", message)
    }
    
    return err
}
```

### Output formatting and data presentation

**Multi-format output support accommodates diverse usage scenarios** from human consumption to script integration. The implementation provides consistent interfaces while optimizing for each format:

```go
type OutputFormatter interface {
    Format(data interface{}) (string, error)
}

type TableFormatter struct {
    headers []string
}

func (f TableFormatter) Format(data interface{}) (string, error) {
    t := table.NewWriter()
    t.SetStyle(table.StyleLight)
    
    switch v := data.(type) {
    case []api.Issue:
        t.AppendHeader(table.Row{"Key", "Summary", "Status", "Assignee"})
        for _, issue := range v {
            assignee := "Unassigned"
            if issue.Fields.Assignee != nil {
                assignee = issue.Fields.Assignee.DisplayName
            }
            t.AppendRow(table.Row{
                issue.Key,
                truncate(issue.Fields.Summary, 50),
                issue.Fields.Status.Name,
                assignee,
            })
        }
    }
    
    return t.Render(), nil
}

type JSONFormatter struct {
    Pretty bool
}

func (f JSONFormatter) Format(data interface{}) (string, error) {
    if f.Pretty {
        bytes, err := json.MarshalIndent(data, "", "  ")
        return string(bytes), err
    }
    bytes, err := json.Marshal(data)
    return string(bytes), err
}

// Format selection and execution
func formatOutput(data interface{}, format string) error {
    var formatter OutputFormatter
    
    switch format {
    case "json":
        formatter = JSONFormatter{Pretty: true}
    case "table":
        formatter = TableFormatter{}
    case "yaml":
        formatter = YAMLFormatter{}
    default:
        return fmt.Errorf("unsupported format: %s", format)
    }
    
    output, err := formatter.Format(data)
    if err != nil {
        return err
    }
    
    fmt.Print(output)
    return nil
}
```

## Testing strategies and quality assurance

Robust CLI testing requires multiple approaches to ensure reliability across diverse usage scenarios and deployment environments.

### Unit testing with mocked dependencies

**Interface-based design enables comprehensive unit testing** through dependency injection and mock implementations. This approach isolates business logic from external dependencies:

```go
type APIClient interface {
    GetIssue(key string) (*api.Issue, error)
    CreateIssue(payload *api.IssueCreateRequest) (*api.Issue, error)
    SearchIssues(jql string) (*api.SearchResponse, error)
}

type MockAPIClient struct {
    issues map[string]*api.Issue
    fail   bool
}

func (m *MockAPIClient) GetIssue(key string) (*api.Issue, error) {
    if m.fail {
        return nil, errors.New("API error")
    }
    
    if issue, exists := m.issues[key]; exists {
        return issue, nil
    }
    
    return nil, &api.APIError{StatusCode: 404, Message: "Issue not found"}
}

func TestIssueGetCommand(t *testing.T) {
    tests := []struct {
        name           string
        args           []string
        setupMock      func(*MockAPIClient)
        expectedOutput string
        expectedError  string
    }{
        {
            name: "successful issue retrieval",
            args: []string{"DEMO-123"},
            setupMock: func(m *MockAPIClient) {
                m.issues["DEMO-123"] = &api.Issue{
                    Key: "DEMO-123",
                    Fields: api.IssueFields{
                        Summary: "Test issue",
                        Status:  api.Status{Name: "In Progress"},
                    },
                }
            },
            expectedOutput: "DEMO-123: Test issue",
        },
        {
            name:          "missing issue key",
            args:          []string{},
            expectedError: "issue key required",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockClient := &MockAPIClient{issues: make(map[string]*api.Issue)}
            if tt.setupMock != nil {
                tt.setupMock(mockClient)
            }
            
            cmd := newIssueGetCmd(mockClient)
            
            var output, errorOutput bytes.Buffer
            cmd.SetOut(&output)
            cmd.SetErr(&errorOutput)
            cmd.SetArgs(tt.args)
            
            err := cmd.Execute()
            
            if tt.expectedError != "" {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.expectedError)
            } else {
                assert.NoError(t, err)
            }
            
            if tt.expectedOutput != "" {
                assert.Contains(t, output.String(), tt.expectedOutput)
            }
        })
    }
}
```

### Integration testing with API mocking

**HTTP-level testing validates complete request/response handling** including authentication, error processing, and data transformation:

```go
func TestAPIIntegration(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Validate authentication header
        auth := r.Header.Get("Authorization")
        if !strings.HasPrefix(auth, "Basic ") {
            w.WriteHeader(http.StatusUnauthorized)
            json.NewEncoder(w).Encode(map[string]string{"message": "Unauthorized"})
            return
        }
        
        switch r.URL.Path {
        case "/rest/api/3/issue/DEMO-123":
            response := api.Issue{
                Key: "DEMO-123",
                Fields: api.IssueFields{
                    Summary: "Mock issue",
                    Status:  api.Status{Name: "Open"},
                },
            }
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(response)
            
        default:
            w.WriteHeader(http.StatusNotFound)
        }
    }))
    defer server.Close()
    
    // Configure CLI for test server
    viper.Set("api_endpoint", server.URL)
    viper.Set("email", "test@example.com")
    viper.Set("token", "test-token")
    
    cmd := newRootCmd()
    cmd.SetArgs([]string{"issue", "get", "DEMO-123"})
    
    err := cmd.Execute()
    assert.NoError(t, err)
}
```

## Conclusion

Building a production-ready CLI tool for Atlassian JIRA and Confluence requires careful orchestration of multiple components: comprehensive API understanding, robust Go libraries, proven architectural patterns, and meticulous attention to developer experience. The combination of go-atlassian for API integration, Cobra for command structure, and secure credential management creates a foundation for powerful automation tools.

The research reveals that successful implementations prioritize user experience through intuitive command hierarchies, intelligent error handling, and flexible output formatting. Security considerations around token storage, combined with proper testing strategies, ensure enterprise-grade reliability. By following these established patterns and leveraging the mature Go ecosystem, developers can create CLI tools that significantly enhance productivity in Atlassian-centric development workflows.

The architectural decisions outlined here—from authentication flows through output formatting—provide a comprehensive blueprint for building developer tools that rival the sophistication of GitHub CLI while addressing the unique challenges of Atlassian API integration.

## Appendix: CLI Command Reference

This appendix provides a structured reference for implementing commands across JIRA, Confluence, and configuration management. Each section includes the command definition, parameters, and purpose to guide implementation.

### JIRA Commands

#### Issue Management

| Command | Parameters | Description |
|---------|------------|-------------|
| `issue create` | `--project` (required): Project key<br>`--type`: Issue type (default: "Task")<br>`--summary` (required): Issue summary<br>`--description`: Issue description<br>`--assignee`: Username or "me"<br>`--priority`: Priority (default: "Medium")<br>`--labels`: Comma-separated labels<br>`--components`: Comma-separated components<br>`--epic-link`: Epic key for linking | Creates a new issue with specified attributes |
| `issue get` | `<issue-key>` (required): Issue identifier<br>`--expand`: Fields to expand (e.g., "transitions,changelog")<br>`--output`: Output format (json, table, yaml) | Retrieves and displays issue details |
| `issue list` | `--project`: Filter by project<br>`--assignee`: Filter by assignee<br>`--status`: Filter by status<br>`--type`: Filter by issue type<br>`--component`: Filter by component<br>`--label`: Filter by label<br>`--jql`: Custom JQL query<br>`--limit`: Max issues to return<br>`--output`: Output format | Lists issues matching specified criteria |
| `issue update` | `<issue-key>` (required): Issue identifier<br>`--summary`: New summary<br>`--description`: New description<br>`--assignee`: New assignee<br>`--priority`: New priority<br>`--status`: New status<br>`--add-label`: Labels to add<br>`--remove-label`: Labels to remove | Updates specified issue fields |
| `issue comment` | `<issue-key>` (required): Issue identifier<br>`--text` (required): Comment text<br>`--internal`: Mark as internal comment | Adds a comment to the specified issue |
| `issue comments` | `<issue-key>` (required): Issue identifier<br>`--limit`: Max comments to return<br>`--output`: Output format | Lists comments for specified issue |
| `issue link` | `<from-issue>` (required): Source issue key<br>`<to-issue>` (required): Target issue key<br>`--type` (required): Link type (e.g., "blocks", "relates to") | Creates a link between two issues |
| `issue transition` | `<issue-key>` (required): Issue identifier<br>`--to` (required): Target status<br>`--comment`: Transition comment<br>`--resolution`: Resolution for resolving issues | Transitions issue to specified status |
| `issue attachments` | `<issue-key>` (required): Issue identifier<br>`--upload`: File path to upload<br>`--download`: Download attachment by ID<br>`--list`: List all attachments | Manages issue attachments |

#### Agile & Sprint Management

| Command | Parameters | Description |
|---------|------------|-------------|
| `sprint create` | `--board` (required): Board ID<br>`--name` (required): Sprint name<br>`--goal`: Sprint goal<br>`--start-date`: Start date (YYYY-MM-DD)<br>`--end-date`: End date (YYYY-MM-DD) | Creates a new sprint |
| `sprint list` | `--board` (required): Board ID<br>`--state`: Filter by state (active, future, closed)<br>`--output`: Output format | Lists sprints for specified board |
| `sprint start` | `<sprint-id>` (required): Sprint identifier<br>`--start-date`: Start date (YYYY-MM-DD)<br>`--end-date`: End date (YYYY-MM-DD) | Starts the specified sprint |
| `sprint complete` | `<sprint-id>` (required): Sprint identifier<br>`--date`: Completion date (YYYY-MM-DD) | Completes the specified sprint |
| `sprint issues` | `<sprint-id>` (required): Sprint identifier<br>`--status`: Filter by status<br>`--assignee`: Filter by assignee<br>`--output`: Output format | Lists issues in specified sprint |
| `sprint move` | `<issue-key>` (required): Issue identifier<br>`--to` (required): Target sprint ID | Moves issue to specified sprint |
| `board list` | `--project`: Filter by project<br>`--type`: Filter by board type (scrum, kanban)<br>`--output`: Output format | Lists boards with optional filtering |
| `board issues` | `<board-id>` (required): Board identifier<br>`--filter`: JQL filter<br>`--output`: Output format | Lists issues on specified board |

#### Project & Component Management

| Command | Parameters | Description |
|---------|------------|-------------|
| `project list` | `--output`: Output format | Lists all accessible projects |
| `project get` | `<project-key>` (required): Project identifier<br>`--expand`: Fields to expand<br>`--output`: Output format | Shows detailed project information |
| `project components` | `<project-key>` (required): Project identifier<br>`--output`: Output format | Lists components for specified project |
| `project versions` | `<project-key>` (required): Project identifier<br>`--output`: Output format | Lists versions for specified project |
| `component create` | `--project` (required): Project key<br>`--name` (required): Component name<br>`--description`: Component description<br>`--lead`: Component lead username | Creates a project component |
| `version create` | `--project` (required): Project key<br>`--name` (required): Version name<br>`--description`: Version description<br>`--start-date`: Start date (YYYY-MM-DD)<br>`--release-date`: Release date (YYYY-MM-DD) | Creates a project version |

#### Query & Search

| Command | Parameters | Description |
|---------|------------|-------------|
| `search` | `--jql` (required): JQL query string<br>`--fields`: Fields to include in results<br>`--limit`: Max results to return<br>`--output`: Output format | Executes JQL search and displays results |
| `jql` | `<command>`: Mode (validate, execute, explain)<br>`--query`: JQL query string | Advanced JQL operations |

#### Workflow Management

| Command | Parameters | Description |
|---------|------------|-------------|
| `workflow list` | `--project`: Filter by project<br>`--output`: Output format | Lists workflows |
| `workflow get` | `<workflow-id>` (required): Workflow identifier<br>`--output`: Output format | Shows workflow details |
| `workflow transitions` | `<issue-key>` (required): Issue identifier<br>`--output`: Output format | Lists available transitions for issue |

### Confluence Commands

#### Page Management

| Command | Parameters | Description |
|---------|------------|-------------|
| `page create` | `--space` (required): Space key<br>`--title` (required): Page title<br>`--content`: Page content text<br>`--content-file`: File containing content<br>`--parent`: Parent page title<br>`--template`: Template name<br>`--template-vars`: Template variables (JSON)<br>`--labels`: Comma-separated labels | Creates a new page with specified content |
| `page get` | `<page-id>` or `--title` & `--space`: Page identifier<br>`--expand`: Elements to expand<br>`--version`: Specific version to retrieve<br>`--output`: Output format | Retrieves and displays page content |
| `page list` | `--space` (required): Space key<br>`--limit`: Max pages to return<br>`--labels`: Filter by labels<br>`--output`: Output format | Lists pages in specified space |
| `page update` | `<page-id>` (required): Page identifier<br>`--title`: New title<br>`--content`: New content text<br>`--content-file`: File containing new content<br>`--version-comment`: Comment for version history | Updates existing page |
| `page delete` | `<page-id>` (required): Page identifier<br>`--force`: Skip confirmation | Deletes specified page |
| `page history` | `<page-id>` (required): Page identifier<br>`--limit`: Max versions to show<br>`--output`: Output format | Shows version history for page |
| `page compare` | `<page-id>` (required): Page identifier<br>`--from-version`: Source version number<br>`--to-version`: Target version number | Shows differences between versions |
| `page labels` | `<page-id>` (required): Page identifier<br>`--add`: Labels to add<br>`--remove`: Labels to remove<br>`--output`: Output format | Manages page labels |
| `page export` | `<page-id>` (required): Page identifier<br>`--format`: Export format (PDF, Word, HTML, etc.)<br>`--output-file`: Destination file path | Exports page to specified format |

#### Space Management

| Command | Parameters | Description |
|---------|------------|-------------|
| `space list` | `--type`: Filter by space type (global, personal)<br>`--output`: Output format | Lists available spaces |
| `space get` | `<space-key>` (required): Space identifier<br>`--expand`: Elements to expand<br>`--output`: Output format | Shows detailed space information |
| `space content` | `<space-key>` (required): Space identifier<br>`--type`: Content type filter (page, blogpost, etc.)<br>`--limit`: Max items to return<br>`--output`: Output format | Lists content in specified space |
| `space permissions` | `<space-key>` (required): Space identifier<br>`--output`: Output format | Lists permissions for specified space |

#### Content Search & Navigation

| Command | Parameters | Description |
|---------|------------|-------------|
| `search` | `--cql`: Confluence Query Language query<br>`--text`: Full-text search<br>`--space`: Limit to space<br>`--limit`: Max results to return<br>`--output`: Output format | Searches for content matching criteria |
| `tree` | `<page-id>` or `--space` & `--root`: Starting point<br>`--depth`: Maximum depth to display<br>`--output`: Output format | Displays content hierarchy as tree |

#### Attachments & Media

| Command | Parameters | Description |
|---------|------------|-------------|
| `attachment upload` | `<page-id>` (required): Page identifier<br>`--file` (required): File path to upload<br>`--comment`: Attachment comment | Uploads file attachment to page |
| `attachment list` | `<page-id>` (required): Page identifier<br>`--output`: Output format | Lists attachments for specified page |
| `attachment download` | `<attachment-id>` (required): Attachment identifier<br>`--output-file`: Destination file path | Downloads specified attachment |

#### Blogging & Commenting

| Command | Parameters | Description |
|---------|------------|-------------|
| `blog create` | `--space` (required): Space key<br>`--title` (required): Blog post title<br>`--content`: Post content text<br>`--content-file`: File containing content<br>`--labels`: Comma-separated labels | Creates a new blog post |
| `comment add` | `<page-id>` (required): Page identifier<br>`--text` (required): Comment text | Adds comment to specified page |
| `comment list` | `<page-id>` (required): Page identifier<br>`--limit`: Max comments to return<br>`--output`: Output format | Lists comments for specified page |

### Configuration Commands

#### Authentication & Setup

| Command | Parameters | Description |
|---------|------------|-------------|
| `auth login` | `--instance`: Atlassian instance URL<br>`--email`: User email<br>`--token`: API token<br>`--no-store`: Don't store credentials | Authenticates with Atlassian instance |
| `auth logout` | None | Clears stored credentials |
| `auth status` | `--output`: Output format | Shows authentication status |
| `auth switch` | `--instance`: Instance URL to switch to | Switches between configured instances |

#### Configuration Management

| Command | Parameters | Description |
|---------|------------|-------------|
| `config set` | `<key>` (required): Configuration key<br>`<value>` (required): Configuration value | Sets configuration value |
| `config get` | `<key>` (optional): Configuration key to retrieve | Shows current configuration |
| `config list` | `--output`: Output format | Lists all configuration settings |
| `config init` | `--instance`: Atlassian instance URL<br>`--default-project`: Default project<br>`--default-space`: Default Confluence space | Creates initial configuration file |
| `config import` | `--file` (required): Configuration file to import | Imports configuration from file |
| `config export` | `--file` (required): Destination file path<br>`--include-credentials`: Include auth credentials | Exports configuration to file |

#### Profile & Workspace

| Command | Parameters | Description |
|---------|------------|-------------|
| `profile list` | `--output`: Output format | Lists configured profiles |
| `profile create` | `--name` (required): Profile name<br>`--instance`: Atlassian instance URL<br>`--default-project`: Default project<br>`--default-space`: Default Confluence space | Creates new profile |
| `profile switch` | `<name>` (required): Profile to activate | Switches to specified profile |
| `workspace init` | `--dir`: Directory to initialize | Creates local workspace configuration |

#### Utility Commands

| Command | Parameters | Description |
|---------|------------|-------------|
| `version` | None | Shows CLI version information |
| `completion` | `<shell>`: Shell type (bash, zsh, fish, powershell) | Generates shell completion scripts |
| `help` | `<command>`: Command to show help for | Shows help information |
| `update` | `--check-only`: Only check for updates | Updates CLI to latest version |

### Global Flags

The following flags apply to all commands:

| Flag | Description |
|------|-------------|
| `--help, -h` | Shows help for command |
| `--verbose, -v` | Enables verbose output |
| `--quiet, -q` | Suppresses all output except errors |
| `--output, -o` | Output format (json, table, yaml) |
| `--config` | Custom config file path |
| `--profile` | Use specific profile |
| `--no-color` | Disables colored output |
| `--debug` | Enables debug output |
| `--trace` | Enables API request/response tracing |

Each command implementation should follow consistent patterns for parameter handling, error reporting, and output formatting to ensure a cohesive user experience across the entire CLI tool.# Building a Developer-Focused CLI for Atlassian JIRA and Confluence