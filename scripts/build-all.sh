#!/bin/bash
# Script to build binaries for multiple platforms

set -e

VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

BINARY_NAME="codegen"
DIST_DIR="dist"

# Build flags
LDFLAGS="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}"

echo "Building version: ${VERSION}"
echo "Commit: ${COMMIT}"
echo "Date: ${DATE}"
echo ""

# Clean dist directory
rm -rf ${DIST_DIR}
mkdir -p ${DIST_DIR}

# Platform configurations
# Format: "OS/ARCH"
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
    "windows/arm64"
)

for PLATFORM in "${PLATFORMS[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$PLATFORM"

    OUTPUT_NAME="${BINARY_NAME}-${VERSION}-${GOOS}-${GOARCH}"

    if [ "$GOOS" = "windows" ]; then
        OUTPUT_NAME="${OUTPUT_NAME}.exe"
    fi

    OUTPUT_PATH="${DIST_DIR}/${OUTPUT_NAME}"

    echo "Building for ${GOOS}/${GOARCH}..."

    env GOOS=${GOOS} GOARCH=${GOARCH} go build \
        -ldflags="${LDFLAGS}" \
        -o "${OUTPUT_PATH}" \
        ./cmd/codegen

    if [ $? -ne 0 ]; then
        echo "Error building for ${GOOS}/${GOARCH}"
        exit 1
    fi

    echo "  ✓ Built: ${OUTPUT_PATH}"

    # Create archive
    ARCHIVE_DIR="${DIST_DIR}/${BINARY_NAME}-${VERSION}-${GOOS}-${GOARCH}"
    mkdir -p "${ARCHIVE_DIR}"

    cp "${OUTPUT_PATH}" "${ARCHIVE_DIR}/${BINARY_NAME}$([ "$GOOS" = "windows" ] && echo ".exe" || echo "")"
    cp README.md "${ARCHIVE_DIR}/"
    cp USAGE_EXAMPLES.md "${ARCHIVE_DIR}/" 2>/dev/null || true

    # Create tar.gz or zip
    cd "${DIST_DIR}"
    if [ "$GOOS" = "windows" ]; then
        zip -q -r "${BINARY_NAME}-${VERSION}-${GOOS}-${GOARCH}.zip" \
            "${BINARY_NAME}-${VERSION}-${GOOS}-${GOARCH}"
        echo "  ✓ Created: ${BINARY_NAME}-${VERSION}-${GOOS}-${GOARCH}.zip"
    else
        tar -czf "${BINARY_NAME}-${VERSION}-${GOOS}-${GOARCH}.tar.gz" \
            "${BINARY_NAME}-${VERSION}-${GOOS}-${GOARCH}"
        echo "  ✓ Created: ${BINARY_NAME}-${VERSION}-${GOOS}-${GOARCH}.tar.gz"
    fi
    cd ..

    rm -rf "${ARCHIVE_DIR}"
done

# Generate checksums
echo ""
echo "Generating checksums..."
cd "${DIST_DIR}"
shasum -a 256 *.tar.gz *.zip 2>/dev/null > checksums.txt || \
    sha256sum *.tar.gz *.zip 2>/dev/null > checksums.txt
cd ..

echo ""
echo "✓ Build completed!"
echo ""
echo "Artifacts in ${DIST_DIR}:"
ls -lh ${DIST_DIR}/*.tar.gz ${DIST_DIR}/*.zip 2>/dev/null || true
echo ""
echo "Checksums:"
cat ${DIST_DIR}/checksums.txt
