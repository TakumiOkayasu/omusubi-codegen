#!/bin/bash
# Run omusubi-codegen in Docker container from host

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Build Docker image if not exists
IMAGE_NAME="omusubi-codegen:latest"

if ! docker images -q ${IMAGE_NAME} >/dev/null 2>&1; then
    echo "Building Docker image..."
    docker build -t ${IMAGE_NAME} -f "${PROJECT_ROOT}/Dockerfile.builder" "${PROJECT_ROOT}"
fi

# Run with current directory and pre-omusubi mounted
exec docker run --rm -it \
    -v "${PWD}:/workspace" \
    -w /workspace \
    ${IMAGE_NAME} "$@"
