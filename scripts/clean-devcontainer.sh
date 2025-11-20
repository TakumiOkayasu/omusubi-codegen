#!/bin/bash
set -e

# Script to completely clean devcontainer volumes and containers
# Usage: ./scripts/clean-devcontainer.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

cd "$PROJECT_ROOT"

# Load .env if exists (excluding readonly bash variables)
if [ -f .env ]; then
  export $(cat .env | grep -v '^#' | grep -v '^UID=' | grep -v '^GID=' | xargs)
fi

PROJECT_NAME=${PROJECT_NAME:-platform-builder}

echo -e "${YELLOW}================================${NC}"
echo -e "${YELLOW}DevContainer Cleanup${NC}"
echo -e "${YELLOW}================================${NC}"
echo -e "Project: ${BLUE}${PROJECT_NAME}${NC}\n"

# 1. Stop and remove containers
echo -e "${YELLOW}[1/5] Stopping containers...${NC}"
if docker compose -f .devcontainer/compose.yaml down 2>/dev/null; then
  echo -e "${GREEN}✓ Containers stopped${NC}"
else
  echo -e "${YELLOW}⚠ No containers to stop${NC}"
fi

# 2. Find and stop any related containers
echo -e "\n${YELLOW}[2/5] Checking for running containers...${NC}"
RUNNING_CONTAINERS=$(docker ps -q -f name="${PROJECT_NAME}")
if [ -n "$RUNNING_CONTAINERS" ]; then
  echo -e "${YELLOW}Stopping running containers...${NC}"
  docker stop $RUNNING_CONTAINERS
  echo -e "${GREEN}✓ Stopped running containers${NC}"
else
  echo -e "${GREEN}✓ No running containers found${NC}"
fi

# 3. Remove all related containers (including stopped ones)
echo -e "\n${YELLOW}[3/5] Removing containers...${NC}"
ALL_CONTAINERS=$(docker ps -aq -f name="${PROJECT_NAME}")
if [ -n "$ALL_CONTAINERS" ]; then
  docker rm -f $ALL_CONTAINERS
  echo -e "${GREEN}✓ Removed all containers${NC}"
else
  echo -e "${GREEN}✓ No containers to remove${NC}"
fi

# 4. Remove volumes
echo -e "\n${YELLOW}[4/5] Removing volumes...${NC}"
VOLUMES=(
  "${PROJECT_NAME}-go-modules"
  "${PROJECT_NAME}-vscode-extensions"
)

for volume in "${VOLUMES[@]}"; do
  if docker volume inspect "$volume" >/dev/null 2>&1; then
    if docker volume rm "$volume" 2>/dev/null; then
      echo -e "${GREEN}✓ Removed volume: ${volume}${NC}"
    else
      echo -e "${RED}✗ Failed to remove volume: ${volume}${NC}"
      echo -e "${YELLOW}  Try closing VSCode and run this script again${NC}"
    fi
  else
    echo -e "${BLUE}ℹ Volume not found: ${volume}${NC}"
  fi
done

# 5. Remove network
echo -e "\n${YELLOW}[5/5] Removing network...${NC}"
NETWORK="${PROJECT_NAME}-network"
if docker network inspect "$NETWORK" >/dev/null 2>&1; then
  if docker network rm "$NETWORK" 2>/dev/null; then
    echo -e "${GREEN}✓ Removed network: ${NETWORK}${NC}"
  else
    echo -e "${YELLOW}⚠ Network still in use (will be removed automatically later)${NC}"
  fi
else
  echo -e "${BLUE}ℹ Network not found: ${NETWORK}${NC}"
fi

echo -e "\n${GREEN}================================${NC}"
echo -e "${GREEN}Cleanup completed!${NC}"
echo -e "${GREEN}================================${NC}"
echo -e "\n${BLUE}Next steps:${NC}"
echo -e "1. Close VSCode completely"
echo -e "2. Reopen the project in VSCode"
echo -e "3. Select 'Reopen in Container'"
echo -e "\nOr rebuild manually:"
echo -e "  ${BLUE}make devcontainer-build${NC}"
