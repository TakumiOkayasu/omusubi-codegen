#!/bin/bash
# Test UI script to verify multi-select display

cd "$(dirname "$0")"

echo "Testing multi-select UI..."
echo ""
echo "Instructions:"
echo "  ↑/↓  : Move cursor"
echo "  Space: Select/Deselect"
echo "  a    : Select ALL"
echo "  i    : Invert selection"
echo "  Enter: Confirm"
echo ""
echo "Press Enter to start..."
read

docker exec -it platform-builder-devcontainer /workspace/omusubi-codegen generate --repo /workspace/testdata/sample
