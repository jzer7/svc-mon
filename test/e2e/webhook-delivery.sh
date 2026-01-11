#!/usr/bin/env bash
set -euo pipefail

# E2E test for webhook alert delivery
# This test starts a mock webhook server, runs svc-mon against a failing service,
# and verifies the webhook receives the alert

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
BINARY="$PROJECT_ROOT/bin/svc-mon"
MOCK_SERVER="$PROJECT_ROOT/test/e2e/mock-webhook-server"

echo "==> E2E Test: Webhook Alert Delivery"

# Ensure binaries exist
if [ ! -f "$BINARY" ]; then
    echo "ERROR: svc-mon binary not found at $BINARY"
    echo "Run 'task build' first"
    exit 1
fi

if [ ! -f "$MOCK_SERVER" ]; then
    echo "Building mock webhook server..."
    cd "$SCRIPT_DIR"
    go build -o mock-webhook-server mock_webhook_server.go
fi

# Start mock webhook server in background
echo "Starting mock webhook server on :8888..."
"$MOCK_SERVER" 8888 &
WEBHOOK_PID=$!
trap "kill $WEBHOOK_PID 2>/dev/null || true" EXIT

# Give server time to start
sleep 1

# Verify webhook server is running
if ! curl -s http://localhost:8888/health >/dev/null 2>&1; then
    echo "ERROR: Mock webhook server failed to start"
    exit 1
fi
echo "  ✓ Webhook server ready"

# TODO: Once monitor is implemented, this should:
# 1. Create a config pointing to a known-bad service URL
# 2. Start svc-mon with webhook URL http://localhost:8888/webhook
# 3. Wait for svc-mon to detect failure and send webhook
# 4. Query mock server for received webhooks
# 5. Verify webhook payload matches expected format

echo ""
echo "Test: Webhook delivery (SKIPPED - not yet implemented)"
echo "  ⊘ Monitor and webhook logic not yet implemented"
echo ""
echo "When implemented, this test will:"
echo "  - Start svc-mon monitoring a failing service"
echo "  - Verify webhook POST to http://localhost:8888/webhook"
echo "  - Validate JSON payload structure"
echo "  - Check alert contains: service, url, status=down, reason"

echo ""
echo "==> Mock infrastructure test PASSED"
exit 0
