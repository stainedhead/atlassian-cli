#!/bin/bash
# Build script for Atlassian CLI

set -e

VERSION=${VERSION:-"1.0.0"}
BUILD_DIR="bin"
DIST_DIR="dist"

echo "Building Atlassian CLI v${VERSION}..."

# Clean previous builds
rm -rf ${BUILD_DIR} ${DIST_DIR}
mkdir -p ${BUILD_DIR} ${DIST_DIR}

# Build for current platform
echo "Building for current platform..."
go build -ldflags "-X main.version=${VERSION}" -o ${BUILD_DIR}/atlassian-cli .

# Build for all platforms
echo "Building for all platforms..."
platforms=("linux/amd64" "darwin/amd64" "darwin/arm64" "windows/amd64")

for platform in "${platforms[@]}"; do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    
    output_name="atlassian-cli-${GOOS}-${GOARCH}"
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi
    
    echo "Building for ${GOOS}/${GOARCH}..."
    env GOOS=$GOOS GOARCH=$GOARCH go build -ldflags "-X main.version=${VERSION}" -o ${DIST_DIR}/${output_name} .
done

echo "Build complete! Binaries available in ${DIST_DIR}/"
ls -la ${DIST_DIR}/