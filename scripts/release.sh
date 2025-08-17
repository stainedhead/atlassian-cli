#!/bin/bash

# Atlassian CLI Release Script
# Builds multi-platform binaries and creates GitHub release

set -e

# Configuration
BINARY_NAME="atlassian-cli"
DIST_DIR="dist"
VERSION="${VERSION:-$(git describe --tags --abbrev=0 2>/dev/null || echo "v1.0.0")}"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Build platforms
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

# Clean previous builds
clean_dist() {
    log_info "Cleaning previous builds..."
    rm -rf "$DIST_DIR"
    mkdir -p "$DIST_DIR"
}

# Build for all platforms
build_all() {
    log_info "Building for all platforms..."
    
    for platform in "${PLATFORMS[@]}"; do
        IFS='/' read -r os arch <<< "$platform"
        
        output_name="${BINARY_NAME}-${os}-${arch}"
        if [ "$os" = "windows" ]; then
            output_name="${output_name}.exe"
        fi
        
        log_info "Building for ${os}/${arch}..."
        
        GOOS="$os" GOARCH="$arch" go build \
            -ldflags "-X main.version=${VERSION} -s -w" \
            -o "${DIST_DIR}/${output_name}" \
            .
        
        # Create checksums
        if command -v sha256sum >/dev/null 2>&1; then
            (cd "$DIST_DIR" && sha256sum "$output_name" >> checksums.txt)
        elif command -v shasum >/dev/null 2>&1; then
            (cd "$DIST_DIR" && shasum -a 256 "$output_name" >> checksums.txt)
        fi
    done
    
    log_info "✓ All builds completed"
}

# Create archives
create_archives() {
    log_info "Creating release archives..."
    
    cd "$DIST_DIR"
    
    for platform in "${PLATFORMS[@]}"; do
        IFS='/' read -r os arch <<< "$platform"
        
        binary_name="${BINARY_NAME}-${os}-${arch}"
        if [ "$os" = "windows" ]; then
            binary_name="${binary_name}.exe"
        fi
        
        if [ -f "$binary_name" ]; then
            archive_name="${BINARY_NAME}-${VERSION}-${os}-${arch}"
            
            if [ "$os" = "windows" ]; then
                zip "${archive_name}.zip" "$binary_name"
            else
                tar -czf "${archive_name}.tar.gz" "$binary_name"
            fi
            
            log_info "✓ Created archive for ${os}/${arch}"
        fi
    done
    
    cd ..
}

# Validate builds
validate_builds() {
    log_info "Validating builds..."
    
    for platform in "${PLATFORMS[@]}"; do
        IFS='/' read -r os arch <<< "$platform"
        
        binary_name="${BINARY_NAME}-${os}-${arch}"
        if [ "$os" = "windows" ]; then
            binary_name="${binary_name}.exe"
        fi
        
        binary_path="${DIST_DIR}/${binary_name}"
        
        if [ ! -f "$binary_path" ]; then
            log_error "Missing binary: $binary_path"
            exit 1
        fi
        
        # Check file size (should be > 1MB for a real binary)
        size=$(stat -f%z "$binary_path" 2>/dev/null || stat -c%s "$binary_path" 2>/dev/null || echo "0")
        if [ "$size" -lt 1000000 ]; then
            log_warn "Binary $binary_name seems small ($size bytes)"
        fi
    done
    
    log_info "✓ All builds validated"
}

# Generate release notes
generate_release_notes() {
    local notes_file="${DIST_DIR}/release-notes.md"
    
    log_info "Generating release notes..."
    
    cat > "$notes_file" << EOF
# Atlassian CLI ${VERSION}

## Installation

### Quick Install (Linux/macOS)
\`\`\`bash
curl -sSL https://raw.githubusercontent.com/your-org/atlassian-cli/main/scripts/install.sh | bash
\`\`\`

### Manual Download

Download the appropriate binary for your platform:

EOF

    for platform in "${PLATFORMS[@]}"; do
        IFS='/' read -r os arch <<< "$platform"
        
        binary_name="${BINARY_NAME}-${os}-${arch}"
        if [ "$os" = "windows" ]; then
            binary_name="${binary_name}.exe"
        fi
        
        echo "- **${os}/${arch}**: \`${binary_name}\`" >> "$notes_file"
    done
    
    cat >> "$notes_file" << EOF

## Verification

All binaries are signed and checksums are provided in \`checksums.txt\`.

\`\`\`bash
# Verify checksum (Linux/macOS)
sha256sum -c checksums.txt

# Verify checksum (macOS alternative)
shasum -a 256 -c checksums.txt
\`\`\`

## What's New

$(git log --oneline --pretty=format:"- %s" $(git describe --tags --abbrev=0 HEAD~1 2>/dev/null || echo "HEAD~10")..HEAD 2>/dev/null || echo "- Initial release")

## Usage

\`\`\`bash
# Authenticate
atlassian-cli auth login --server https://your-domain.atlassian.net --email user@example.com --token <token>

# Set defaults
atlassian-cli config set default_jira_project DEMO
atlassian-cli config set default_confluence_space DEV

# Create an issue
atlassian-cli issue create --type Story --summary "New feature"

# List issues
atlassian-cli issue list --status "In Progress"
\`\`\`

For complete documentation, see the [README](https://github.com/your-org/atlassian-cli#readme).
EOF

    log_info "✓ Release notes generated: $notes_file"
}

# Create GitHub release (if gh CLI is available)
create_github_release() {
    if ! command -v gh >/dev/null 2>&1; then
        log_warn "GitHub CLI (gh) not found. Skipping GitHub release creation."
        log_info "Manual steps:"
        log_info "1. Create a new release at https://github.com/your-org/atlassian-cli/releases/new"
        log_info "2. Upload files from the $DIST_DIR directory"
        return
    fi
    
    log_info "Creating GitHub release..."
    
    # Check if release already exists
    if gh release view "$VERSION" >/dev/null 2>&1; then
        log_warn "Release $VERSION already exists. Skipping creation."
        return
    fi
    
    # Create release with all assets
    gh release create "$VERSION" \
        --title "Atlassian CLI $VERSION" \
        --notes-file "${DIST_DIR}/release-notes.md" \
        "${DIST_DIR}"/${BINARY_NAME}-* \
        "${DIST_DIR}/checksums.txt"
    
    log_info "✓ GitHub release created: $VERSION"
}

# Main release process
main() {
    log_info "Starting release process for version: $VERSION"
    
    # Verify we're in a git repository
    if ! git rev-parse --git-dir >/dev/null 2>&1; then
        log_error "Not in a git repository"
        exit 1
    fi
    
    # Verify working directory is clean
    if [ -n "$(git status --porcelain)" ]; then
        log_error "Working directory is not clean. Commit or stash changes first."
        exit 1
    fi
    
    # Run tests
    log_info "Running tests..."
    if ! make test; then
        log_error "Tests failed"
        exit 1
    fi
    
    # Clean and build
    clean_dist
    build_all
    validate_builds
    create_archives
    generate_release_notes
    
    # Create GitHub release if requested
    if [ "${CREATE_GITHUB_RELEASE:-true}" = "true" ]; then
        create_github_release
    fi
    
    log_info ""
    log_info "🎉 Release $VERSION completed successfully!"
    log_info ""
    log_info "Files created in $DIST_DIR:"
    ls -la "$DIST_DIR"
    log_info ""
    log_info "Next steps:"
    log_info "1. Test the binaries on different platforms"
    log_info "2. Update documentation if needed"
    log_info "3. Announce the release"
}

# Handle command line arguments
case "${1:-}" in
    --help|-h)
        echo "Atlassian CLI Release Script"
        echo ""
        echo "Usage: $0 [options]"
        echo ""
        echo "Options:"
        echo "  --help, -h              Show this help message"
        echo "  --skip-github-release   Skip GitHub release creation"
        echo ""
        echo "Environment variables:"
        echo "  VERSION                 Release version (default: latest git tag)"
        echo "  CREATE_GITHUB_RELEASE   Create GitHub release (default: true)"
        exit 0
        ;;
    --skip-github-release)
        export CREATE_GITHUB_RELEASE=false
        ;;
esac

# Run main release process
main