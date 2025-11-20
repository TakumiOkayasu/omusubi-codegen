# syntax=docker/dockerfile:1.4

FROM ubuntu:22.04

# Avoid prompts from apt
ENV DEBIAN_FRONTEND=noninteractive

# Install base dependencies (combined layer)
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    apt-get update && apt-get install -y \
    build-essential \
    cmake \
    git \
    wget \
    curl \
    gnupg \
    software-properties-common \
    pkg-config \
    ca-certificates \
    libssl-dev \
    sudo

# Install Go 1.23 (architecture-aware)
ARG TARGETARCH
RUN --mount=type=cache,target=/tmp/downloads \
    GOARCH=${TARGETARCH:-amd64} && \
    if [ "$GOARCH" = "arm64" ]; then GOARCH="arm64"; else GOARCH="amd64"; fi && \
    wget -O /tmp/downloads/go.tar.gz https://go.dev/dl/go1.23.0.linux-${GOARCH}.tar.gz && \
    tar -C /usr/local -xzf /tmp/downloads/go.tar.gz

ENV PATH="/usr/local/go/bin:/go/bin:${PATH}"
ENV GOPATH="/go"

# Install C++ toolchain (Clang 18 - latest)
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    wget -O - https://apt.llvm.org/llvm-snapshot.gpg.key | apt-key add - && \
    add-apt-repository "deb http://apt.llvm.org/jammy/ llvm-toolchain-jammy-18 main" && \
    apt-get update && apt-get install -y \
        clang-18 \
        clang++-18 \
        clang-format-18 \
        clang-tidy-18 \
        libc++-18-dev \
        libc++abi-18-dev \
        lld-18 && \
    update-alternatives --install /usr/bin/clang clang /usr/bin/clang-18 100 && \
    update-alternatives --install /usr/bin/clang++ clang++ /usr/bin/clang++-18 100 && \
    update-alternatives --install /usr/bin/clang-format clang-format /usr/bin/clang-format-18 100 && \
    update-alternatives --install /usr/bin/clang-tidy clang-tidy /usr/bin/clang-tidy-18 100 && \
    update-alternatives --install /usr/bin/ld.lld ld.lld /usr/bin/ld.lld-18 100

# Install Google Test
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    apt-get update && apt-get install -y \
    libgtest-dev \
    libgmock-dev \
    googletest

# Install Node.js and tree-sitter CLI
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    --mount=type=cache,target=/root/.npm \
    curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && \
    apt-get install -y nodejs && \
    npm install -g tree-sitter-cli

# Create non-root user
ARG USERNAME
ARG USER_UID
ARG USER_GID

# Create /go directory with proper permissions
RUN mkdir -p /go/pkg/mod/cache /go/bin && chmod -R 777 /go

# Create user and group with conditional handling
RUN if ! getent group $USER_GID >/dev/null; then \
        groupadd --gid $USER_GID $USERNAME; \
    fi \
    && if ! getent passwd $USER_UID >/dev/null; then \
        useradd --uid $USER_UID --gid $USER_GID -m -s /bin/bash $USERNAME; \
    fi \
    && echo $USERNAME ALL=\(root\) NOPASSWD:ALL > /etc/sudoers.d/$USERNAME \
    && chmod 0440 /etc/sudoers.d/$USERNAME

# Create directories for VSCode Server with proper permissions
RUN mkdir -p /home/$USERNAME/.vscode-server/bin \
    && mkdir -p /home/$USERNAME/.vscode-server/extensions \
    && chown -R $USER_UID:$USER_GID /home/$USERNAME/.vscode-server

# Install Go tools (as non-root user) with cache mount
USER $USERNAME

RUN --mount=type=cache,target=/go/pkg/mod,uid=${USER_UID},gid=${USER_GID} \
    go install golang.org/x/tools/gopls@latest && \
    go install github.com/go-delve/delve/cmd/dlv@latest && \
    go install honnef.co/go/tools/cmd/staticcheck@latest && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Set working directory
WORKDIR /workspace

