#!/bin/bash

# Atlassian CLI Installation Script
# This script downloads and installs the latest version of atlassian-cli

set -e

# Configuration
REPO="your-org/atlassian-cli"
BINARY_NAME="atlassian-cli"
INSTALL_DIR="/usr/local/bin"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Detect OS and architecture
detect_platform() {
    local os arch
    
    case "$(uname -s)" in
        Linux*)     os="linux" ;;
        Darwin*)    os="darwin" ;;
        CYGWIN*|MINGW*|MSYS*) os="windows" ;;
        *)          log_error "Unsupported operating system: $(uname -s)"; exit 1 ;;
    esac
    
    case "$(uname -m)" in
        x86_64|amd64)   arch="amd64" ;;
        arm64|aarch64)  arch="arm64" ;;
        *)              log_error "Unsupported architecture: $(uname -m)"; exit 1 ;;
    esac
    
    echo "${os}-${arch}"
}

# Get latest release version from GitHub
get_latest_version() {
    local version
    version=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$version" ]; then
        log_error "Failed to get latest version"
        exit 1
    fi
    
    echo "$version"
}

# Download and install binary
install_binary() {
    local platform="$1"
    local version="$2"
    local binary_name="${BINARY_NAME}"
    
    if [[ "$platform" == *"windows"* ]]; then
        binary_name="${BINARY_NAME}.exe"
    fi
    
    local download_url="https://github.com/${REPO}/releases/download/${version}/${BINARY_NAME}-${platform}"
    if [[ "$platform" == *"windows"* ]]; then
        download_url="${download_url}.exe"
    fi
    
    local temp_file="/tmp/${binary_name}"
    
    log_info "Downloading ${BINARY_NAME} ${version} for ${platform}..."
    if ! curl -L -o "$temp_file" "$download_url"; then
        log_error "Failed to download binary"
        exit 1
    fi
    
    # Make executable
    chmod +x "$temp_file"
    
    # Install to system directory
    log_info "Installing to ${INSTALL_DIR}..."
    if [ -w "$INSTALL_DIR" ]; then
        mv "$temp_file" "${INSTALL_DIR}/${BINARY_NAME}"
    else
        sudo mv "$temp_file" "${INSTALL_DIR}/${BINARY_NAME}"
    fi
    
    log_info "Installation completed successfully!"
}

# Verify installation
verify_installation() {
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        local version
        version=$("$BINARY_NAME" --version 2>/dev/null || echo "unknown")
        log_info "✓ ${BINARY_NAME} installed successfully"
        log_info "Version: $version"
        log_info "Location: $(which $BINARY_NAME)"
    else
        log_error "Installation verification failed"
        exit 1
    fi
}

# Setup shell completion
setup_completion() {
    local shell
    shell=$(basename "$SHELL")
    
    case "$shell" in
        bash)
            log_info "Setting up bash completion..."
            if [ -d "/etc/bash_completion.d" ]; then
                sudo "$BINARY_NAME" completion bash > "/etc/bash_completion.d/$BINARY_NAME"
                log_info "✓ System-wide bash completion installed"
            else
                log_warn "Add this to your ~/.bashrc:"
                echo "source <($BINARY_NAME completion bash)"
            fi
            ;;
        zsh)
            log_info "Setting up zsh completion..."
            local comp_dir="${HOME}/.zsh/completions"
            mkdir -p "$comp_dir"
            "$BINARY_NAME" completion zsh > "${comp_dir}/_${BINARY_NAME}"
            log_info "✓ Zsh completion installed to $comp_dir"
            log_warn "Add this to your ~/.zshrc if not already present:"
            echo "fpath=(~/.zsh/completions \$fpath)"
            echo "autoload -U compinit && compinit"
            ;;
        fish)
            log_info "Setting up fish completion..."
            local comp_dir="${HOME}/.config/fish/completions"
            mkdir -p "$comp_dir"
            "$BINARY_NAME" completion fish > "${comp_dir}/${BINARY_NAME}.fish"
            log_info "✓ Fish completion installed"
            ;;
        *)
            log_warn "Unknown shell: $shell. Completion setup skipped."
            log_info "Run '$BINARY_NAME completion --help' for manual setup instructions"
            ;;
    esac
}

# Main installation process
main() {
    log_info "Starting Atlassian CLI installation..."
    
    # Check dependencies
    if ! command -v curl >/dev/null 2>&1; then
        log_error "curl is required but not installed"
        exit 1
    fi
    
    # Detect platform
    local platform
    platform=$(detect_platform)
    log_info "Detected platform: $platform"
    
    # Get latest version
    local version
    version=$(get_latest_version)
    log_info "Latest version: $version"
    
    # Install binary
    install_binary "$platform" "$version"
    
    # Verify installation
    verify_installation
    
    # Setup completion (optional)
    if [ "${SKIP_COMPLETION:-}" != "true" ]; then
        setup_completion
    fi
    
    log_info ""
    log_info "🎉 Installation complete!"
    log_info ""
    log_info "Next steps:"
    log_info "1. Authenticate: $BINARY_NAME auth login --server <url> --email <email> --token <token>"
    log_info "2. Configure defaults: $BINARY_NAME config set default_jira_project <project>"
    log_info "3. Get help: $BINARY_NAME --help"
    log_info ""
    log_info "Documentation: https://github.com/${REPO}#readme"
}

# Handle command line arguments
case "${1:-}" in
    --help|-h)
        echo "Atlassian CLI Installation Script"
        echo ""
        echo "Usage: $0 [options]"
        echo ""
        echo "Options:"
        echo "  --help, -h          Show this help message"
        echo "  --skip-completion   Skip shell completion setup"
        echo ""
        echo "Environment variables:"
        echo "  INSTALL_DIR         Installation directory (default: /usr/local/bin)"
        echo "  SKIP_COMPLETION     Skip completion setup (default: false)"
        exit 0
        ;;
    --skip-completion)
        export SKIP_COMPLETION=true
        ;;
esac

# Run main installation
main