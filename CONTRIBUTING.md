# Contributing to Atlassian CLI

Thank you for your interest in contributing to Atlassian CLI! This document provides guidelines and information for contributors.

## Table of Contents

- [Development Setup](#development-setup)
- [Development Workflow](#development-workflow)
- [Testing Requirements](#testing-requirements)
- [Code Quality Standards](#code-quality-standards)
- [Definition of Done](#definition-of-done)
- [Submitting Changes](#submitting-changes)

## Development Setup

### Prerequisites

- **Go 1.21+**: Required for building and testing
- **Make**: For build automation
- **golangci-lint**: For code quality checks
- **Git**: For version control

### Initial Setup

```bash
# Clone the repository
git clone https://github.com/your-org/atlassian-cli.git
cd atlassian-cli

# Install dependencies
make deps

# Verify setup
make test
make build
```

### Install golangci-lint

```bash
# Install golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.0

# Verify installation
golangci-lint --version
```

## Development Workflow

### 1. Branch Strategy

- **Feature branches**: `feature/description-of-feature`
- **Bug fixes**: `fix/description-of-fix`
- **Documentation**: `docs/description-of-change`

```bash
# Create feature branch
git checkout -b feature/add-issue-list-command
```

### 2. Test-Driven Development (TDD)

This project follows strict TDD practices:

1. **Red**: Write failing tests first
2. **Green**: Write minimal code to pass tests
3. **Refactor**: Improve code while maintaining tests

```bash
# Write tests first
go test ./internal/package -v

# Implement code to pass tests
# ...

# Verify all tests pass
make test
```

### 3. Code Implementation

Follow these patterns:

- **Interface-based design**: Use interfaces for testability
- **Error handling**: Return meaningful errors with context
- **Documentation**: Comment exported functions and complex logic
- **Validation**: Validate all inputs

Example:

```go
// APIClient defines the interface for interacting with Atlassian APIs
type APIClient interface {
    GetIssue(ctx context.Context, key string) (*Issue, error)
    CreateIssue(ctx context.Context, req *CreateIssueRequest) (*Issue, error)
}

// GetIssue retrieves an issue by key with proper error handling
func (c *Client) GetIssue(ctx context.Context, key string) (*Issue, error) {
    if key == "" {
        return nil, fmt.Errorf("issue key cannot be empty")
    }
    
    // Implementation...
}
```

## Testing Requirements

### Test Coverage

- **Minimum 90% coverage** for new code
- **Unit tests** for all business logic
- **Integration tests** for API interactions
- **End-to-end tests** for critical workflows

### Test Structure

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name        string
        input       InputType
        expected    ExpectedType
        expectError bool
    }{
        {
            name:        "valid input",
            input:       validInput,
            expected:    expectedOutput,
            expectError: false,
        },
        {
            name:        "invalid input",
            input:       invalidInput,
            expectError: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := FunctionName(tt.input)
            
            if tt.expectError {
                assert.Error(t, err)
                return
            }
            
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests with race detection
make test-race

# Run specific package tests
go test ./internal/config -v
```

## Code Quality Standards

### Linting

All code must pass linting checks:

```bash
# Run linter
make lint

# Auto-fix formatting issues
make fmt
```

### Common Linting Rules

- **Function length**: Maximum 100 lines
- **Complexity**: Maximum cyclomatic complexity of 10
- **Naming**: Use clear, descriptive names
- **Comments**: Document exported functions
- **Error handling**: Always handle errors appropriately

### Security

- **No hardcoded secrets**: Use environment variables or secure storage
- **Input validation**: Validate all user inputs
- **Dependency scanning**: Keep dependencies up to date
- **Error messages**: Don't leak sensitive information

## Definition of Done

Before submitting any code, ensure all items in the [Definition of Done](.github/copilot-instructions.md) are completed:

### Planning Phase
- [ ] Requirements analysis completed
- [ ] Technical design reviewed
- [ ] Test cases identified

### Implementation
- [ ] Code follows TDD approach (Red → Green → Refactor)
- [ ] 90%+ test coverage achieved
- [ ] All tests pass
- [ ] Code passes linting checks
- [ ] Security review completed

### Documentation
- [ ] Code comments added for complex logic
- [ ] Public APIs documented
- [ ] README updated if needed
- [ ] Examples provided where appropriate

### Quality Assurance
- [ ] Peer review completed
- [ ] Integration tests pass
- [ ] Performance impact assessed
- [ ] Security implications reviewed

## Submitting Changes

### 1. Pre-submission Checklist

```bash
# Ensure all tests pass
make test

# Verify code quality
make lint

# Check test coverage
make test-coverage

# Build successfully
make build
```

### 2. Commit Message Format

Use conventional commit format:

```
type(scope): description

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Test additions/changes
- `refactor`: Code refactoring
- `perf`: Performance improvements

Examples:
```
feat(auth): add support for multiple Atlassian instances

Add authentication management for multiple instances with
profile switching capabilities.

Closes #123
```

### 3. Pull Request Process

1. **Create descriptive PR title and description**
2. **Link related issues**
3. **Include testing evidence**
4. **Request review from maintainers**
5. **Address review feedback**
6. **Ensure CI passes**

### 4. Pull Request Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing completed

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex code
- [ ] Documentation updated
- [ ] Tests added/updated
- [ ] All tests pass
```

## Architecture Guidelines

### Package Organization

- `cmd/`: Command implementations
- `internal/`: Private application code
- `pkg/`: Public packages (if any)
- `test/`: Test utilities and fixtures

### Design Principles

1. **Single Responsibility**: Each package/function has one clear purpose
2. **Dependency Injection**: Use interfaces to enable testing
3. **Error Handling**: Wrap errors with context
4. **Configuration**: Support multiple configuration sources
5. **Testing**: Write tests first, aim for high coverage

### API Design

- **Consistent interfaces**: Similar operations should have similar APIs
- **Context support**: Pass context for cancellation and timeouts
- **Error wrapping**: Provide meaningful error messages
- **Validation**: Validate inputs at boundaries

## Getting Help

- **Documentation**: Check [docs/](./docs/) directory
- **Issues**: Search existing issues or create new ones
- **Discussions**: Use GitHub Discussions for questions
- **Code Review**: Request review from maintainers

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and grow
- Follow the project's technical standards

Thank you for contributing to Atlassian CLI!