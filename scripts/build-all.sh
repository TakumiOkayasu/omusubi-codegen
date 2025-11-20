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

# Detect host platform
HOST_OS=$(go env GOOS)
HOST_ARCH=$(go env GOARCH)

echo "Note: tree-sitter requires CGO, so cross-compilation is limited."
echo "Building for host platform only: ${HOST_OS}/${HOST_ARCH}"
echo "For multi-platform builds, use GoReleaser with Docker or GitHub Actions."
echo ""

# Build for host platform only
OUTPUT_NAME="${BINARY_NAME}-${VERSION}-${HOST_OS}-${HOST_ARCH}"

if [ "$HOST_OS" = "windows" ]; then
    OUTPUT_NAME="${OUTPUT_NAME}.exe"
fi

OUTPUT_PATH="${DIST_DIR}/${OUTPUT_NAME}"

echo "Building for ${HOST_OS}/${HOST_ARCH}..."

go build \
    -ldflags="${LDFLAGS}" \
    -o "${OUTPUT_PATH}" \
    ./cmd/codegen

if [ $? -ne 0 ]; then
    echo "Error building for ${HOST_OS}/${HOST_ARCH}"
    exit 1
fi

echo "  ✓ Built: ${OUTPUT_PATH}"

# Create archive
ARCHIVE_DIR="${DIST_DIR}/${BINARY_NAME}-${VERSION}-${HOST_OS}-${HOST_ARCH}"
mkdir -p "${ARCHIVE_DIR}"

cp "${OUTPUT_PATH}" "${ARCHIVE_DIR}/${BINARY_NAME}$([ "$HOST_OS" = "windows" ] && echo ".exe" || echo "")"
cp README.md "${ARCHIVE_DIR}/"
cp USAGE_EXAMPLES.md "${ARCHIVE_DIR}/" 2>/dev/null || true

# Create tar.gz or zip
cd "${DIST_DIR}"
if [ "$HOST_OS" = "windows" ]; then
    zip -q -r "${BINARY_NAME}-${VERSION}-${HOST_OS}-${HOST_ARCH}.zip" \
        "${BINARY_NAME}-${VERSION}-${HOST_OS}-${HOST_ARCH}"
    echo "  ✓ Created: ${BINARY_NAME}-${VERSION}-${HOST_OS}-${HOST_ARCH}.zip"
else
    tar -czf "${BINARY_NAME}-${VERSION}-${HOST_OS}-${HOST_ARCH}.tar.gz" \
        "${BINARY_NAME}-${VERSION}-${HOST_OS}-${HOST_ARCH}"
    echo "  ✓ Created: ${BINARY_NAME}-${VERSION}-${HOST_OS}-${HOST_ARCH}.tar.gz"
fi
cd ..

rm -rf "${ARCHIVE_DIR}"

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
