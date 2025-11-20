#!/bin/bash
# Multi-platform build using Docker buildx
# This script builds for multiple platforms using Docker's cross-compilation support

set -e

VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

BINARY_NAME="codegen"
DIST_DIR="dist"
IMAGE_NAME="codegen-builder"

echo "Building version: ${VERSION}"
echo "Commit: ${COMMIT}"
echo "Date: ${DATE}"
echo ""

# Clean dist directory
rm -rf ${DIST_DIR}
mkdir -p ${DIST_DIR}

# Check if buildx is available
if ! docker buildx version >/dev/null 2>&1; then
    echo "Error: Docker buildx is not available"
    echo "Please install Docker Desktop or enable buildx"
    exit 1
fi

# Create buildx builder if not exists
BUILDER_NAME="multiplatform-builder"
if ! docker buildx inspect ${BUILDER_NAME} >/dev/null 2>&1; then
    echo "Creating buildx builder: ${BUILDER_NAME}"
    docker buildx create --name ${BUILDER_NAME} --use
fi

docker buildx use ${BUILDER_NAME}

# Platforms to build
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
)

for PLATFORM in "${PLATFORMS[@]}"; do
    echo ""
    echo "Building for ${PLATFORM}..."

    IFS='/' read -r OS ARCH <<< "$PLATFORM"

    # Build Docker image for this platform
    docker buildx build \
        --platform "${PLATFORM}" \
        --build-arg VERSION="${VERSION}" \
        --build-arg COMMIT="${COMMIT}" \
        --build-arg DATE="${DATE}" \
        --build-arg TARGETOS="${OS}" \
        --build-arg TARGETARCH="${ARCH}" \
        --output type=local,dest="${DIST_DIR}/tmp-${OS}-${ARCH}" \
        -f Dockerfile.builder \
        .

    if [ $? -ne 0 ]; then
        echo "Error building for ${PLATFORM}"
        exit 1
    fi

    # Extract binary
    BINARY_PATH="${DIST_DIR}/tmp-${OS}-${ARCH}/usr/local/bin/codegen"
    OUTPUT_NAME="${BINARY_NAME}-${VERSION}-${OS}-${ARCH}"

    if [ -f "${BINARY_PATH}" ]; then
        mv "${BINARY_PATH}" "${DIST_DIR}/${OUTPUT_NAME}"
        echo "  ✓ Extracted: ${OUTPUT_NAME}"

        # Create archive
        ARCHIVE_DIR="${DIST_DIR}/${BINARY_NAME}-${VERSION}-${OS}-${ARCH}"
        mkdir -p "${ARCHIVE_DIR}"

        cp "${DIST_DIR}/${OUTPUT_NAME}" "${ARCHIVE_DIR}/${BINARY_NAME}"
        chmod +x "${ARCHIVE_DIR}/${BINARY_NAME}"
        cp README.md "${ARCHIVE_DIR}/"
        cp USAGE_EXAMPLES.md "${ARCHIVE_DIR}/" 2>/dev/null || true

        # Create tar.gz
        cd "${DIST_DIR}"
        tar -czf "${BINARY_NAME}-${VERSION}-${OS}-${ARCH}.tar.gz" \
            "${BINARY_NAME}-${VERSION}-${OS}-${ARCH}"
        echo "  ✓ Created: ${BINARY_NAME}-${VERSION}-${OS}-${ARCH}.tar.gz"
        cd ..

        rm -rf "${ARCHIVE_DIR}"
    else
        echo "  ✗ Binary not found: ${BINARY_PATH}"
    fi

    # Cleanup temp directory
    rm -rf "${DIST_DIR}/tmp-${OS}-${ARCH}"
done

# Generate checksums
echo ""
echo "Generating checksums..."
cd "${DIST_DIR}"
shasum -a 256 *.tar.gz 2>/dev/null > checksums.txt || \
    sha256sum *.tar.gz 2>/dev/null > checksums.txt
cd ..

echo ""
echo "✓ Build completed!"
echo ""
echo "Artifacts in ${DIST_DIR}:"
ls -lh ${DIST_DIR}/*.tar.gz 2>/dev/null || true
echo ""
echo "Checksums:"
cat ${DIST_DIR}/checksums.txt
