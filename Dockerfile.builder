# Multi-platform builder for CGO-enabled Go projects
# This Dockerfile is used for cross-compilation with tree-sitter

ARG GO_VERSION=1.23
ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH

FROM golang:${GO_VERSION} AS builder

WORKDIR /build

# Install build dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    gcc \
    g++ \
    && rm -rf /var/lib/apt/lists/*

# Copy source code
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build arguments
ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown

# Build binary
RUN CGO_ENABLED=1 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
    -o /omusubi-codegen \
    ./cmd/codegen

# Runtime stage
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /omusubi-codegen /usr/local/bin/omusubi-codegen

ENTRYPOINT ["omusubi-codegen"]
CMD ["--help"]
