#!/bin/bash
set -e

# Script to test devcontainer build without opening in VSCode
# Usage: ./scripts/test-devcontainer-build.sh [--no-cache] [--platform linux/amd64|linux/arm64]

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
COMPOSE_FILE="$PROJECT_ROOT/.devcontainer/compose.yaml"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
NO_CACHE=""
PLATFORM=""

# Enable BuildKit for faster builds
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1

# Parse arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --no-cache)
      NO_CACHE="--no-cache"
      shift
      ;;
    --platform)
      PLATFORM="$2"
      shift 2
      ;;
    *)
      echo -e "${RED}Unknown option: $1${NC}"
      echo "Usage: $0 [--no-cache] [--platform linux/amd64|linux/arm64]"
      exit 1
      ;;
  esac
done

cd "$PROJECT_ROOT"

# Check if .env exists, create from example if not
if [ ! -f .env ]; then
  echo -e "${YELLOW}Creating .env from .env.example...${NC}"
  cp .env.example .env
fi

# Override platform if specified
if [ -n "$PLATFORM" ]; then
  echo -e "${YELLOW}Setting DOCKER_PLATFORM=$PLATFORM${NC}"
  export DOCKER_PLATFORM="$PLATFORM"
fi

# Show build info
echo -e "${GREEN}================================${NC}"
echo -e "${GREEN}Devcontainer Build Test${NC}"
echo -e "${GREEN}================================${NC}"
echo "Compose file: $COMPOSE_FILE"
echo "Platform: ${DOCKER_PLATFORM:-linux/arm64 (default)}"
echo "No cache: ${NO_CACHE:-false}"
echo -e "${GREEN}================================${NC}\n"

# Build the container
echo -e "${YELLOW}Building devcontainer...${NC}"
if docker compose --progress=plain -f "$COMPOSE_FILE" build $NO_CACHE; then
  echo -e "\n${GREEN}✓ Build successful!${NC}\n"

  # Show image info
  echo -e "${GREEN}Image information:${NC}"
  docker compose -f "$COMPOSE_FILE" images

  echo -e "\n${GREEN}================================${NC}"
  echo -e "${GREEN}Build test completed successfully${NC}"
  echo -e "${GREEN}================================${NC}"
  echo -e "\nYou can now open the project in VSCode devcontainer."
else
  echo -e "\n${RED}✗ Build failed!${NC}\n"
  exit 1
fi
