#!/usr/bin/env bash
set -euo pipefail

# Basic smoke test for svc-mon binary
# This script tests the most basic functionality: start monitor, verify it runs

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
BINARY="$PROJECT_ROOT/bin/svc-mon"
CONFIG="$SCRIPT_DIR/test-config.yaml"

echo "==> E2E Test: Basic Monitoring"

# Ensure binary exists
if [ ! -f "$BINARY" ]; then
    echo "ERROR: Binary not found at $BINARY"
    echo "Run 'task build' first"
    exit 1
fi

# Ensure test config exists
if [ ! -f "$CONFIG" ]; then
    echo "ERROR: Test config not found at $CONFIG"
    exit 1
fi

# Test 1: Verify binary runs with --version
echo "Test 1: Version command"
if "$BINARY" version | grep -q "svc-mon version"; then
    echo "  ✓ Version command works"
else
    echo "  ✗ Version command failed"
    exit 1
fi

# Test 2: Verify binary accepts config file
echo "Test 2: Config file validation"
if "$BINARY" monitor --config "$CONFIG" --help >/dev/null 2>&1; then
    echo "  ✓ Config flag accepted"
else
    echo "  ✗ Config flag not accepted"
    exit 1
fi

# Test 3: Start monitor in background and verify it runs
# TODO: Once monitor logic is implemented, this should:
# - Start the monitor
# - Wait for it to check services
# - Verify logs or webhook calls
# - Cleanup gracefully
echo "Test 3: Monitor execution (SKIPPED - not yet implemented)"
echo "  ⊘ Monitor logic not yet implemented"

echo ""
echo "==> Basic smoke tests PASSED"
exit 0
