#!/bin/bash
# Build binary for host macOS from within devcontainer using Docker

set -e

VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

BINARY_NAME="omusubi-codegen"

echo "Building macOS binary from devcontainer..."
echo "Version: ${VERSION}"

# Detect host OS from Docker socket or environment
HOST_OS=${HOST_OS:-darwin}
HOST_ARCH=${HOST_ARCH:-arm64}

echo "Target: ${HOST_OS}/${HOST_ARCH}"
echo ""

# Check if osxcross or similar is available
if command -v o64-clang >/dev/null 2>&1 || command -v x86_64-apple-darwin20.4-clang >/dev/null 2>&1; then
    echo "Cross-compilation toolchain found"

    # Build with cross-compiler
    CGO_ENABLED=1 \
    GOOS=${HOST_OS} \
    GOARCH=${HOST_ARCH} \
    go build \
        -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
        -o ${BINARY_NAME}-${HOST_OS}-${HOST_ARCH} \
        ./cmd/codegen

    echo "Built: ${BINARY_NAME}-${HOST_OS}-${HOST_ARCH}"
else
    echo "Warning: Cross-compilation toolchain not available"
    echo "Building for current platform (Linux) instead"
    echo ""
    echo "To use on macOS, please:"
    echo "  1. Install Go on macOS: brew install go"
    echo "  2. Build on macOS: make build"
    echo "  Or use Docker buildx for multi-platform builds"

    # Build for current platform
    make build
fi
