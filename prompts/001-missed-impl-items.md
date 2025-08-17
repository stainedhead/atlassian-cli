# Implementation Plan: Global Flag Refactoring and Search Feature Addition

## 📋 **Analysis Summary**

Based on the review of documentation and development rules, I need to implement the following changes:

### **Required Changes:**
1. **Remove global flags** `--jira-project` and `--confluence-space` from root command
2. **Add command-specific flags** `--project` for JIRA commands and `--space` for Confluence commands
3. **Add missing search commands** for both issues and pages
4. **Maintain configuration hierarchy** for default values and environment variables

### **Missing Features Identified:**
- `issue search` command with JQL support
- `page search` command with CQL support
- Command-specific project/space override flags

## 🎯 **Implementation Plan**

### **Phase 1: Global Flag Refactoring**
Following TDD principles and Definition of Done requirements:

#### **Step 1.1: Remove Global Flags**
- [ ] **Red Phase**: Write failing tests for removed global flags
- [ ] **Green Phase**: Remove `--jira-project` and `--confluence-space` from root command
- [ ] **Refactor Phase**: Update viper bindings and configuration handling
- [ ] **Test Coverage**: Ensure 80%+ coverage for configuration changes

#### **Step 1.2: Add Command-Specific Flags**
- [ ] **Red Phase**: Write tests for new command-specific flags
- [ ] **Green Phase**: Add `--project` flag to issue/project commands
- [ ] **Green Phase**: Add `--space` flag to page/space commands
- [ ] **Refactor Phase**: Implement flag resolution hierarchy
- [ ] **Test Coverage**: Unit tests for flag precedence logic

### **Phase 2: Search Feature Implementation**

#### **Step 2.1: Issue Search Command**
- [ ] **Red Phase**: Write failing tests for `issue search` command
- [ ] **Green Phase**: Implement basic JQL search functionality
- [ ] **Refactor Phase**: Add advanced filtering options
- [ ] **Test Coverage**: Integration tests with mocked API responses

#### **Step 2.2: Page Search Command**
- [ ] **Red Phase**: Write failing tests for `page search` command
- [ ] **Green Phase**: Implement basic CQL search functionality
- [ ] **Refactor Phase**: Add content type filtering
- [ ] **Test Coverage**: End-to-end search workflow tests

### **Phase 3: Documentation and Quality Assurance**

#### **Step 3.1: Documentation Updates**
- [ ] **Documentation Sync**: Update command reference documentation
- [ ] **API Documentation**: Document new search parameters
- [ ] **Migration Guide**: Document breaking changes for global flags
- [ ] **Examples**: Add search command examples to usage guides

#### **Step 3.2: Quality Gates**
- [ ] **Static Analysis**: Pass golangci-lint with no new issues
- [ ] **Security Scanning**: Run gosec on new code
- [ ] **Performance Testing**: Benchmark search operations
- [ ] **Integration Testing**: Test with real API endpoints

## 📝 **Definition of Done Checklist**

### **Planning Phase**
- [x] **Requirements Analysis**: Reviewed product documentation and identified missing features
- [x] **Technical Planning**: Analyzed current architecture and flag handling
- [x] **Test Planning**: Identified test cases for flag removal and search features
- [x] **Architecture Review**: Confirmed modular design consistency

### **Implementation Standards**
- [ ] **Modularity**: Search commands organized in logical modules
- [ ] **Error Handling**: Comprehensive error handling for search operations
- [ ] **Input Validation**: JQL and CQL query validation
- [ ] **Resource Management**: Proper cleanup of search resources
- [ ] **Concurrency Safety**: Thread-safe search operations

### **Code Quality & Testing**
- [ ] **TDD Requirements**: Red-Green-Refactor cycle for all changes
- [ ] **Test Coverage**: Minimum 80% coverage for new functionality
- [ ] **Static Analysis**: Pass all configured linters
- [ ] **Security Review**: No hardcoded credentials or injection vulnerabilities

## 🔧 **Technical Implementation Details**

### **Flag Resolution Hierarchy**
```go
// New hierarchy for command-specific flags
func resolveProject(cmd *cobra.Command) (string, error) {
    // 1. Command-specific --project flag (highest priority)
    if project, _ := cmd.Flags().GetString("project"); project != "" {
        return project, nil
    }
    
    // 2. Environment variable
    if project := os.Getenv("ATLASSIAN_DEFAULT_JIRA_PROJECT"); project != "" {
        return project, nil
    }
    
    // 3. Configuration file
    if project := viper.GetString("default_jira_project"); project != "" {
        return project, nil
    }
    
    // 4. Interactive prompt (lowest priority)
    return "", ErrNoProjectConfigured
}
```

### **Search Command Structure**
```go
// Issue search command
func NewIssueSearchCmd(tokenManager auth.TokenManager) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "search",
        Short: "Search issues using JQL",
        Example: `  atlassian-cli issue search --jql "assignee = currentUser()"
  atlassian-cli issue search --project DEMO --status "In Progress"`,
        RunE: func(cmd *cobra.Command, args []string) error {
            return runIssueSearch(cmd, tokenManager)
        },
    }
    
    cmd.Flags().String("project", "", "project key (overrides default)")
    cmd.Flags().String("jql", "", "JQL query string")
    cmd.Flags().String("assignee", "", "filter by assignee")
    cmd.Flags().String("status", "", "filter by status")
    cmd.Flags().String("type", "", "filter by issue type")
    cmd.Flags().Int("limit", 50, "maximum results to return")
    
    return cmd
}
```

### **Breaking Changes Documentation**
```markdown
## Breaking Changes in v2.0.0

### Removed Global Flags
- `--jira-project` and `--confluence-space` global flags have been removed
- Use command-specific flags instead:
  - JIRA commands: `--project`
  - Confluence commands: `--space`

### Migration Guide
```bash
# Old (v1.x)
atlassian-cli --jira-project DEMO issue list

# New (v2.x)
atlassian-cli issue list --project DEMO
```

### New Search Commands
- `atlassian-cli issue search` - JQL-based issue search
- `atlassian-cli page search` - CQL-based page search
```

## 🚀 **Implementation Priority**

### **High Priority (Week 1)**
1. Remove global flags and add command-specific flags
2. Implement basic search functionality
3. Update core documentation

### **Medium Priority (Week 2)**
1. Advanced search filtering options
2. Performance optimization for search
3. Integration testing with real APIs

### **Low Priority (Week 3)**
1. Search result caching
2. Interactive search builders
3. Search history and saved queries

This plan follows the Definition of Done requirements, implements TDD methodology, and maintains the modular architecture while adding the missing search functionality and fixing the global flag structure.