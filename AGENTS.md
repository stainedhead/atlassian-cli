# AGENTS.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an Atlassian CLI tool project for JIRA and Confluence automation. The project is currently in early planning/research phase with comprehensive documentation but no Go code implementation yet.

## Project Structure

- `documentation/` - Contains product research and architectural planning
  - `product-research.md` - Comprehensive guide for building Atlassian CLI with Go, including API patterns, architecture decisions, and implementation examples
- `prompts/` - Contains initial project setup prompts
- `bin/` - Binary output directory (excluded from git)
- `tests/` - Test directory (excluded from git)

## Development Commands

Since this is a new Go project without implementation yet, standard Go commands will be used:

- **Build**: `go build -o bin/atlassian-cli`
- **Test**: `go test ./...`
- **Lint**: `golangci-lint run` (assuming golangci-lint will be used)
- **Clean**: `rm -rf bin/`

## Architecture & Implementation Guide

The `documentation/product-research.md` file contains a comprehensive blueprint for implementation:

### Recommended Tech Stack
- **CLI Framework**: Cobra for command structure and argument parsing
- **Atlassian SDK**: go-atlassian (github.com/ctreminiom/go-atlassian/v2) for unified JIRA/Confluence API access
- **Configuration**: Viper for hierarchical configuration management
- **Authentication**: OS keychain integration for secure token storage
- **Output**: Multi-format support (JSON, table, YAML)

### Command Structure
```
atlassian-cli
├── issue (JIRA issue management)
│   ├── create, get, list, update, comment, link, transition
├── sprint (Agile/Sprint management)
│   ├── create, list, start, complete, issues, move
├── project (Project operations)
│   ├── list, get, components, versions
├── page (Confluence pages)
│   ├── create, get, list, update, delete, history
├── space (Confluence spaces)
│   ├── list, get, content, permissions
└── auth (Authentication)
    ├── login, logout, status, switch
```

### Authentication Pattern
- API token-based authentication (email + token)
- Secure credential storage using OS keychain
- Support for multiple Atlassian instances/profiles

### Key Implementation Patterns
- Service-oriented API client architecture
- Interface-based design for testability
- Comprehensive error handling with actionable suggestions
- Progress indicators for long-running operations
- Intelligent output formatting based on terminal capabilities

## Next Steps for Implementation

1. Initialize Go module: `go mod init atlassian-cli`
2. Add core dependencies (Cobra, Viper, go-atlassian)
3. Implement basic command structure following Cobra patterns
4. Set up authentication and configuration management
5. Implement core JIRA issue operations
6. Add Confluence page management
7. Implement comprehensive testing strategy

## Testing Strategy

- Unit tests with mocked API clients
- Integration tests with HTTP mocking
- CLI command testing using Cobra's testing utilities
- Separate test packages for different domains (issue, page, auth)

---

## Feature Specification Workflow

### Specs Directory Structure

All feature development uses the `specs/` directory for planning and tracking. Each feature gets its own subdirectory named after the feature.

**Directory Structure:**
```
specs/
└── <feature-name>/
    ├── spec.md                  # Feature specification and requirements
    ├── status.md                # **CRITICAL**: Phase progress tracking (update after each task)
    ├── plan.md                  # Implementation plan and architecture decisions
    ├── tasks.md                 # Task breakdown and progress tracking
    ├── research.md              # Research findings, API docs, examples
    ├── data-dictionary.md       # Data structures, types, schemas
    ├── architecture.md          # System architecture and component design
    └── implementation-notes.md  # Implementation details, gotchas, decisions
```

### Progressive Documentation Build

Documents are created progressively as the feature develops:

**Phase 0: Initial Research (PRD/Feature Research)**
- Input: Product Requirement Document, RFC, or feature research
- Purpose: Understand the problem, gather requirements, identify constraints
- **Update status.md**: Mark Phase 0 as "In Progress"

**Phase 1: Specification (spec.md)**
- Define what the feature does
- User requirements and acceptance criteria
- Goals and non-goals
- Success criteria
- **Update status.md**: Mark Phase 0 complete, Phase 1 in progress

**Phase 2: Research & Data Modeling (research.md, data-dictionary.md)**
- Gather API documentation
- Explore existing code and implementations
- Define domain entities and data structures
- Document types, interfaces, and schemas
- **Update status.md**: Mark Phase 1 complete, Phase 2 in progress

**Phase 3: Architecture & Planning (architecture.md, plan.md)**
- Design the implementation approach
- Identify affected layers (Domain, Use Case, Infrastructure, Adapter)
- Document component architecture and data flows
- Create implementation plan with phases and deliverables
- List files to create/modify
- **Update status.md**: Mark Phase 2 complete, Phase 3 in progress

**Phase 4: Task Breakdown (tasks.md)**
- Break down work into concrete, testable tasks
- Define dependencies and critical path
- Estimate durations
- Set up quality gates
- **Update status.md**: Mark Phase 3 complete, Phase 4 in progress

**Phase 5: Implementation (code + implementation-notes.md)**
- Follow TDD (Red-Green-Refactor)
- Record decisions made during implementation
- Document edge cases and solutions
- Note performance optimizations
- Track deviations from plan
- **Update status.md**: After EACH task completion - MANDATORY

**Phase 6: Completion & Archival**
- Update product documentation
- Move spec to specs/archive/
- Capture lessons learned
- **Verify status.md**: Must show 100% completion before archiving

**MANDATORY**: Update `status.md` after completing each task or phase. This file is the single source of truth for progress tracking.

### Specs Workflow Rules

- **Create feature directory** before starting any new feature work
- **Update progressively** as understanding evolves - specs are living documents
- **Update status.md ALWAYS** after completing each task, phase, or milestone - this is MANDATORY
- **Reference from commits** - link to spec directory in commit messages
- **Archive completed** - move to `specs/archive/` when feature is fully implemented and stable
- **Version control** - specs are committed to the repository for team collaboration

**Critical Rule**: Every time you complete a task, update `status.md` immediately to reflect:
- Task completion status
- Phase progress percentage
- Any blockers or issues encountered
- Next steps

### Example Feature Development Flow

```bash
# 1. Initialize specs process (one time)
/prd-to-spec init

# 2. Create feature spec from PRD
/prd-to-spec new-spec docs/prd-gmail-send.md

# 3. Work through phases progressively
# - Fill in spec.md (requirements, goals)
# - UPDATE status.md: Mark Phase 1 complete
# - Research and create research.md, data-dictionary.md
# - UPDATE status.md: Mark Phase 2 complete
# - Design architecture.md, plan.md
# - UPDATE status.md: Mark Phase 3 complete
# - Break down tasks.md
# - UPDATE status.md: Mark Phase 4 complete

# 4. Implement following TDD workflow
# - Update implementation-notes.md as you go
# - **CRITICAL**: Update status.md after EACH task completion

# 5. Archive when complete and stable
# - Verify status.md shows 100% completion
/prd-to-spec archive-spec gmail-send-command
```
