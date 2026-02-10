# CI/CD Integration Guide

This guide shows how to integrate the Atlassian CLI into your CI/CD pipelines for automated workflows.

## GitHub Actions Integration

### Basic Setup

```yaml
# .github/workflows/atlassian-integration.yml
name: Atlassian Integration

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  atlassian-integration:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Install Atlassian CLI
      run: |
        curl -sSL https://raw.githubusercontent.com/your-org/atlassian-cli/main/scripts/install.sh | bash
        echo "/usr/local/bin" >> $GITHUB_PATH
    
    - name: Configure Atlassian CLI
      env:
        ATLASSIAN_SERVER: ${{ secrets.ATLASSIAN_SERVER }}
        ATLASSIAN_EMAIL: ${{ secrets.ATLASSIAN_EMAIL }}
        ATLASSIAN_TOKEN: ${{ secrets.ATLASSIAN_TOKEN }}
      run: |
        atlassian-cli auth login \
          --server "$ATLASSIAN_SERVER" \
          --email "$ATLASSIAN_EMAIL" \
          --token "$ATLASSIAN_TOKEN"
        
        atlassian-cli config set default_jira_project "${{ vars.JIRA_PROJECT }}"
        atlassian-cli config set default_confluence_space "${{ vars.CONFLUENCE_SPACE }}"
        atlassian-cli config set output json
    
    - name: Update JIRA Issues
      run: |
        # Extract issue keys from commit messages
        git log --oneline ${{ github.event.before }}..${{ github.sha }} | \
        grep -o '[A-Z]\+-[0-9]\+' | sort -u | \
        while read issue_key; do
          echo "Updating issue: $issue_key"
          atlassian-cli issue update "$issue_key" \
            --add-comment "Build triggered by commit ${{ github.sha }}" \
            --add-labels "ci-build"
        done
```

### Advanced Workflow with Issue Transitions

```yaml
# .github/workflows/release-workflow.yml
name: Release Workflow

on:
  release:
    types: [published]

jobs:
  update-jira-issues:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
      with:
        fetch-depth: 0  # Fetch full history
    
    - name: Setup Atlassian CLI
      uses: ./.github/actions/setup-atlassian-cli
    
    - name: Transition Issues to Done
      run: |
        # Get all commits since last release
        LAST_TAG=$(git describe --tags --abbrev=0 HEAD~1 2>/dev/null || echo "")
        if [ -n "$LAST_TAG" ]; then
          COMMIT_RANGE="$LAST_TAG..HEAD"
        else
          COMMIT_RANGE="HEAD~10..HEAD"  # Fallback for first release
        fi
        
        # Extract and process issue keys
        git log --oneline $COMMIT_RANGE | \
        grep -o '[A-Z]\+-[0-9]\+' | sort -u | \
        while read issue_key; do
          echo "Processing issue: $issue_key"
          
          # Check current status
          current_status=$(atlassian-cli issue get "$issue_key" --output json | jq -r '.status')
          
          if [ "$current_status" != "Done" ]; then
            atlassian-cli issue update "$issue_key" \
              --status "Done" \
              --add-comment "Resolved in release ${{ github.event.release.tag_name }}"
            echo "✓ Transitioned $issue_key to Done"
          else
            echo "ℹ Issue $issue_key already Done"
          fi
        done
    
    - name: Create Release Notes Page
      run: |
        # Generate release notes content
        RELEASE_NOTES=$(cat << EOF
        <h1>Release ${{ github.event.release.tag_name }}</h1>
        <h2>Release Date</h2>
        <p>$(date '+%Y-%m-%d %H:%M:%S UTC')</p>
        
        <h2>Changes</h2>
        <p>${{ github.event.release.body }}</p>
        
        <h2>Deployment Information</h2>
        <ul>
        <li><strong>Build:</strong> ${{ github.run_number }}</li>
        <li><strong>Commit:</strong> ${{ github.sha }}</li>
        <li><strong>Branch:</strong> ${{ github.ref_name }}</li>
        </ul>
        EOF
        )
        
        # Create Confluence page
        atlassian-cli page create \
          --title "Release ${{ github.event.release.tag_name }}" \
          --content "$RELEASE_NOTES" \
          --confluence-space "RELEASES"
```

## Jenkins Integration

### Pipeline Script

```groovy
// Jenkinsfile
pipeline {
    agent any
    
    environment {
        ATLASSIAN_SERVER = credentials('atlassian-server')
        ATLASSIAN_EMAIL = credentials('atlassian-email')
        ATLASSIAN_TOKEN = credentials('atlassian-token')
        JIRA_PROJECT = 'DEMO'
        CONFLUENCE_SPACE = 'DEV'
    }
    
    stages {
        stage('Setup') {
            steps {
                script {
                    // Install Atlassian CLI
                    sh '''
                        curl -sSL https://raw.githubusercontent.com/your-org/atlassian-cli/main/scripts/install.sh | bash
                        export PATH="/usr/local/bin:$PATH"
                        
                        # Authenticate and configure
                        atlassian-cli auth login \
                            --server "$ATLASSIAN_SERVER" \
                            --email "$ATLASSIAN_EMAIL" \
                            --token "$ATLASSIAN_TOKEN"
                        
                        atlassian-cli config set default_jira_project "$JIRA_PROJECT"
                        atlassian-cli config set default_confluence_space "$CONFLUENCE_SPACE"
                    '''
                }
            }
        }
        
        stage('Build') {
            steps {
                sh 'make build'
            }
            post {
                success {
                    script {
                        // Update issues on successful build
                        sh '''
                            export PATH="/usr/local/bin:$PATH"
                            git log --oneline ${GIT_PREVIOUS_COMMIT}..${GIT_COMMIT} | \
                            grep -o '[A-Z]\\+-[0-9]\\+' | sort -u | \
                            while read issue_key; do
                                atlassian-cli issue update "$issue_key" \
                                    --add-comment "Build #${BUILD_NUMBER} successful" \
                                    --add-labels "build-success"
                            done
                        '''
                    }
                }
                failure {
                    script {
                        // Update issues on build failure
                        sh '''
                            export PATH="/usr/local/bin:$PATH"
                            git log --oneline ${GIT_PREVIOUS_COMMIT}..${GIT_COMMIT} | \
                            grep -o '[A-Z]\\+-[0-9]\\+' | sort -u | \
                            while read issue_key; do
                                atlassian-cli issue update "$issue_key" \
                                    --add-comment "Build #${BUILD_NUMBER} failed - ${BUILD_URL}" \
                                    --add-labels "build-failure" \
                                    --priority "High"
                            done
                        '''
                    }
                }
            }
        }
        
        stage('Test') {
            steps {
                sh 'make test'
            }
            post {
                always {
                    // Publish test results to Confluence
                    script {
                        sh '''
                            export PATH="/usr/local/bin:$PATH"
                            
                            # Generate test report
                            TEST_RESULTS=$(make test 2>&1 || true)
                            
                            # Create Confluence page with results
                            atlassian-cli page create \
                                --title "Test Results - Build #${BUILD_NUMBER}" \
                                --content "<h2>Test Results</h2><pre>$TEST_RESULTS</pre>" \
                                --confluence-space "QA"
                        '''
                    }
                }
            }
        }
        
        stage('Deploy') {
            when {
                branch 'main'
            }
            steps {
                sh 'make deploy'
            }
            post {
                success {
                    script {
                        // Create deployment notification
                        sh '''
                            export PATH="/usr/local/bin:$PATH"
                            
                            # Create deployment issue
                            atlassian-cli issue create \
                                --type "Deployment" \
                                --summary "Production Deployment - Build #${BUILD_NUMBER}" \
                                --description "Deployed commit ${GIT_COMMIT} to production" \
                                --assignee "devops-team" \
                                --priority "Medium"
                        '''
                    }
                }
            }
        }
    }
}
```

## GitLab CI Integration

### GitLab CI Configuration

```yaml
# .gitlab-ci.yml
stages:
  - setup
  - build
  - test
  - deploy
  - notify

variables:
  ATLASSIAN_CLI_VERSION: "latest"
  JIRA_PROJECT: "DEMO"
  CONFLUENCE_SPACE: "DEV"

.atlassian_setup: &atlassian_setup
  - curl -sSL https://raw.githubusercontent.com/your-org/atlassian-cli/main/scripts/install.sh | bash
  - export PATH="/usr/local/bin:$PATH"
  - atlassian-cli auth login --server "$ATLASSIAN_SERVER" --email "$ATLASSIAN_EMAIL" --token "$ATLASSIAN_TOKEN"
  - atlassian-cli config set default_jira_project "$JIRA_PROJECT"
  - atlassian-cli config set default_confluence_space "$CONFLUENCE_SPACE"
  - atlassian-cli config set output json

setup:
  stage: setup
  script:
    - *atlassian_setup
    - atlassian-cli auth status
  only:
    - main
    - develop

build:
  stage: build
  script:
    - make build
  after_script:
    - *atlassian_setup
    - |
      # Update JIRA issues with build status
      git log --oneline $CI_COMMIT_BEFORE_SHA..$CI_COMMIT_SHA | \
      grep -o '[A-Z]\+-[0-9]\+' | sort -u | \
      while read issue_key; do
        if [ "$CI_JOB_STATUS" = "success" ]; then
          atlassian-cli issue update "$issue_key" \
            --add-comment "Build pipeline #$CI_PIPELINE_ID completed successfully" \
            --add-labels "build-success"
        else
          atlassian-cli issue update "$issue_key" \
            --add-comment "Build pipeline #$CI_PIPELINE_ID failed - $CI_PIPELINE_URL" \
            --add-labels "build-failure"
        fi
      done

test:
  stage: test
  script:
    - make test
    - make coverage
  coverage: '/Total coverage: (\d+\.\d+)%/'
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage.xml
  after_script:
    - *atlassian_setup
    - |
      # Create test report page
      COVERAGE=$(make coverage 2>&1 | grep "Total coverage" | awk '{print $3}')
      
      atlassian-cli page create \
        --title "Test Report - Pipeline #$CI_PIPELINE_ID" \
        --content "<h2>Test Coverage</h2><p>Coverage: $COVERAGE</p><h2>Pipeline</h2><p><a href='$CI_PIPELINE_URL'>View Pipeline</a></p>"

deploy_staging:
  stage: deploy
  script:
    - make deploy-staging
  environment:
    name: staging
    url: https://staging.example.com
  only:
    - develop
  after_script:
    - *atlassian_setup
    - |
      atlassian-cli issue create \
        --type "Deployment" \
        --summary "Staging Deployment - Pipeline #$CI_PIPELINE_ID" \
        --description "Deployed $CI_COMMIT_SHA to staging environment" \
        --assignee "qa-team"

deploy_production:
  stage: deploy
  script:
    - make deploy-production
  environment:
    name: production
    url: https://example.com
  when: manual
  only:
    - main
  after_script:
    - *atlassian_setup
    - |
      # Create production deployment issue
      atlassian-cli issue create \
        --type "Deployment" \
        --summary "Production Deployment - Pipeline #$CI_PIPELINE_ID" \
        --description "Deployed $CI_COMMIT_SHA to production environment" \
        --assignee "devops-team" \
        --priority "High"
      
      # Update related issues
      git log --oneline $CI_COMMIT_BEFORE_SHA..$CI_COMMIT_SHA | \
      grep -o '[A-Z]\+-[0-9]\+' | sort -u | \
      while read issue_key; do
        atlassian-cli issue update "$issue_key" \
          --status "Done" \
          --add-comment "Deployed to production in pipeline #$CI_PIPELINE_ID"
      done

notify_release:
  stage: notify
  script:
    - *atlassian_setup
    - |
      # Create release announcement
      atlassian-cli page create \
        --title "Release Announcement - $(date +%Y-%m-%d)" \
        --content "<h1>New Release Deployed</h1><p>Pipeline: $CI_PIPELINE_URL</p><p>Commit: $CI_COMMIT_SHA</p>" \
        --confluence-space "ANNOUNCEMENTS"
  only:
    - main
  when: manual
```

## Azure DevOps Integration

### Azure Pipelines YAML

```yaml
# azure-pipelines.yml
trigger:
  branches:
    include:
    - main
    - develop

pool:
  vmImage: 'ubuntu-latest'

variables:
  jiraProject: 'DEMO'
  confluenceSpace: 'DEV'

stages:
- stage: Setup
  jobs:
  - job: InstallCLI
    steps:
    - script: |
        curl -sSL https://raw.githubusercontent.com/your-org/atlassian-cli/main/scripts/install.sh | bash
        echo "##vso[task.prependpath]/usr/local/bin"
      displayName: 'Install Atlassian CLI'
    
    - script: |
        atlassian-cli auth login \
          --server "$(ATLASSIAN_SERVER)" \
          --email "$(ATLASSIAN_EMAIL)" \
          --token "$(ATLASSIAN_TOKEN)"
        
        atlassian-cli config set default_jira_project "$(jiraProject)"
        atlassian-cli config set default_confluence_space "$(confluenceSpace)"
      displayName: 'Configure Atlassian CLI'
      env:
        ATLASSIAN_SERVER: $(atlassianServer)
        ATLASSIAN_EMAIL: $(atlassianEmail)
        ATLASSIAN_TOKEN: $(atlassianToken)

- stage: Build
  dependsOn: Setup
  jobs:
  - job: BuildAndTest
    steps:
    - script: make build
      displayName: 'Build Application'
    
    - script: make test
      displayName: 'Run Tests'
    
    - script: |
        # Update JIRA issues with build results
        git log --oneline $(Build.SourceVersion)~5..$(Build.SourceVersion) | \
        grep -o '[A-Z]\+-[0-9]\+' | sort -u | \
        while read issue_key; do
          atlassian-cli issue update "$issue_key" \
            --add-comment "Azure DevOps build $(Build.BuildNumber) completed" \
            --add-labels "azure-build"
        done
      displayName: 'Update JIRA Issues'
      condition: always()

- stage: Deploy
  dependsOn: Build
  condition: and(succeeded(), eq(variables['Build.SourceBranch'], 'refs/heads/main'))
  jobs:
  - deployment: DeployProduction
    environment: 'production'
    strategy:
      runOnce:
        deploy:
          steps:
          - script: make deploy
            displayName: 'Deploy to Production'
          
          - script: |
              atlassian-cli page create \
                --title "Deployment Report - $(Build.BuildNumber)" \
                --content "<h2>Production Deployment</h2><p>Build: $(Build.BuildNumber)</p><p>Commit: $(Build.SourceVersion)</p>" \
                --confluence-space "DEPLOYMENTS"
            displayName: 'Create Deployment Report'
```

## Docker Integration

### Dockerfile for CI/CD

```dockerfile
# Dockerfile.ci
FROM golang:1.24-alpine AS builder

# Install Atlassian CLI
RUN apk add --no-cache curl bash
RUN curl -sSL https://raw.githubusercontent.com/your-org/atlassian-cli/main/scripts/install.sh | bash

# Copy source code
WORKDIR /app
COPY . .

# Build application
RUN make build

# Runtime image
FROM alpine:latest
RUN apk add --no-cache ca-certificates git
COPY --from=builder /usr/local/bin/atlassian-cli /usr/local/bin/
COPY --from=builder /app/bin/your-app /usr/local/bin/

# CI/CD script
COPY scripts/ci-integration.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/ci-integration.sh

ENTRYPOINT ["/usr/local/bin/ci-integration.sh"]
```

### CI Integration Script

```bash
#!/bin/bash
# scripts/ci-integration.sh

set -e

# Configuration
JIRA_PROJECT="${JIRA_PROJECT:-DEMO}"
CONFLUENCE_SPACE="${CONFLUENCE_SPACE:-DEV}"

# Authenticate
atlassian-cli auth login \
  --server "$ATLASSIAN_SERVER" \
  --email "$ATLASSIAN_EMAIL" \
  --token "$ATLASSIAN_TOKEN"

# Configure defaults
atlassian-cli config set default_jira_project "$JIRA_PROJECT"
atlassian-cli config set default_confluence_space "$CONFLUENCE_SPACE"
atlassian-cli config set output json

# Process git commits for JIRA integration
if [ -n "$CI_COMMIT_RANGE" ]; then
  echo "Processing commits in range: $CI_COMMIT_RANGE"
  
  git log --oneline "$CI_COMMIT_RANGE" | \
  grep -o '[A-Z]\+-[0-9]\+' | sort -u | \
  while read -r issue_key; do
    echo "Processing issue: $issue_key"
    
    case "$CI_EVENT_TYPE" in
      "push")
        atlassian-cli issue update "$issue_key" \
          --add-comment "Code pushed in build $CI_BUILD_NUMBER" \
          --add-labels "ci-push"
        ;;
      "pull_request")
        atlassian-cli issue update "$issue_key" \
          --add-comment "Pull request created: $CI_PULL_REQUEST_URL" \
          --add-labels "pr-created"
        ;;
      "merge")
        atlassian-cli issue update "$issue_key" \
          --add-comment "Code merged in build $CI_BUILD_NUMBER" \
          --add-labels "merged"
        ;;
    esac
  done
fi

# Execute the main CI command
exec "$@"
```

## Best Practices for CI/CD Integration

### Security Considerations

1. **Store credentials securely**:
   ```bash
   # Use CI/CD secret management
   ATLASSIAN_TOKEN: ${{ secrets.ATLASSIAN_TOKEN }}
   ```

2. **Limit token permissions**:
   - Create dedicated service accounts
   - Use project-specific tokens
   - Regularly rotate credentials

3. **Validate inputs**:
   ```bash
   # Validate issue keys before processing
   validate_issue_key() {
     if [[ ! "$1" =~ ^[A-Z]+-[0-9]+$ ]]; then
       echo "Invalid issue key: $1"
       return 1
     fi
   }
   ```

### Performance Optimization

1. **Cache CLI installation**:
   ```yaml
   - name: Cache Atlassian CLI
     uses: actions/cache@v3
     with:
       path: /usr/local/bin/atlassian-cli
       key: atlassian-cli-${{ runner.os }}-${{ env.CLI_VERSION }}
   ```

2. **Batch operations**:
   ```bash
   # Process multiple issues in a single call when possible
   issue_keys=$(git log --oneline $RANGE | grep -o '[A-Z]\+-[0-9]\+' | sort -u | tr '\n' ' ')
   atlassian-cli issue bulk-update $issue_keys --add-labels "ci-build"
   ```

3. **Use appropriate timeouts**:
   ```bash
   atlassian-cli config set timeout 60s
   ```

This integration guide provides comprehensive examples for incorporating the Atlassian CLI into various CI/CD platforms, enabling automated JIRA and Confluence workflows as part of your development process.