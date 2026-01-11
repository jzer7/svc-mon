---
description: Implementation plan for svc-mon service monitoring tool.
agent: agent
tools: ["execute", "read", "edit", "search", "web", "agent", "todo"]
---

# svc-mon Implementation Plan

- **Version:** 1.0
- **Date:** January 10, 2026
- **Status:** Ready for Implementation

## Overview

Implement the core service monitoring engine for svc-mon: HTTP health checks, webhook alerts, concurrent monitoring with graceful shutdown.

## Architecture

```
┌─────────────┐
│   main.go   │ Parse CLI flags
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  config.go  │ Load & validate YAML → Config struct
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  server.go  │ Orchestrate: spawn monitor goroutines
└──┬───────┬──┘
   │       │
   ▼       ▼
┌───────┐ ┌──────────┐
│monitor│ │ webhook  │ Per-service: check → alert if down
│  .go  │ │   .go    │
└───────┘ └──────────┘
```

## Implementation Order

### Phase 1: Foundation (Core Package)

#### 1.1 Constants & Models

**File:** `internal/core/constants.go`

- Default values: intervals, timeouts, retry delays
- Alert condition constants
- Status constants ("up", "down")

**File:** `internal/core/model.go`

- Complete all structs: `Config`, `ServiceConfig`, `Defaults`, `MonitoringResult`, `WebhookPayload`
- Add `NewConfig()` constructor with defaults merging
- YAML/JSON struct tags

#### 1.2 Configuration Loading

**File:** `internal/core/config.go`

- `LoadConfig(path string) (*Config, error)`
- Parse YAML with `gopkg.in/yaml.v3`
- Validate required fields, parse duration strings
- Merge service config with defaults

#### 1.3 HTTP Monitoring

**File:** `internal/core/monitor.go`

- `CheckService(ctx context.Context, url string, timeout time.Duration) MonitoringResult`
- Create `http.Client` with timeout
- Error classification: timeout, DNS failure, HTTP 5xx
- Use `errors.Is`, `errors.As` for error types

#### 1.4 Webhook Delivery

**File:** `internal/core/webhook.go`

- `SendWebhooks(ctx context.Context, urls []string, payload WebhookPayload) []error`
- JSON marshal payload
- Concurrent delivery with `sync.WaitGroup.Go()`
- 3 retries with 1-second delay on 5xx
- 10-second timeout per webhook POST

### Phase 2: Orchestration

#### 2.1 Monitoring Server

**File:** `internal/core/server.go`

- `Run(configPath string) error`
- Load config via `LoadConfig()`
- Create cancellable context for shutdown
- Spawn goroutine per service:
  - `time.Ticker` for interval-based checks
  - Call `CheckService()`
  - Check if `result.Reason` in `service.AlertIf`
  - If alert needed, call `SendWebhooks()`
  - Structured logging with `slog`
- Signal handling: SIGINT, SIGTERM
- Graceful shutdown: cancel context, 30s wait for WaitGroup

#### 2.2 CLI Integration

**File:** `cmd/svc-mon/main.go`

- Wire `--config` flag to `core.Run()` in monitor command
- Pass through errors and exit codes

### Phase 3: Testing & Cleanup

#### 3.1 Enable Integration Tests

- Remove `t.Skip()` and TODOs from:
  - `internal/core/monitor_integration_test.go`
  - `internal/core/webhook_integration_test.go`
- Verify all test scenarios pass

#### 3.2 Enable E2E Tests

- Update `test/e2e/webhook-delivery.sh` to run (remove skip)
- Verify full webhook flow works

## Constants & Defaults

**File:** `internal/core/constants.go`

```go
const (
    DefaultInterval         = 300 * time.Second  // 5 minutes
    DefaultTimeout          = 5 * time.Second
    WebhookTimeout          = 10 * time.Second
    WebhookRetries          = 3
    WebhookRetryDelay       = 1 * time.Second
    GracefulShutdownTimeout = 30 * time.Second

    StatusUp   = "up"
    StatusDown = "down"

    ReasonTimeout     = "timeout"
    ReasonHTTP5xx     = "http_5xx"
    ReasonDNSFailure  = "dns_failure"
)
```

## Concurrency Patterns

### Service Monitoring (Go 1.25.5 WaitGroup.Go)

```go
var wg sync.WaitGroup
for _, svc := range config.Services {
    svc := svc // capture loop var
    wg.Go(func() {
        monitorService(ctx, svc)
    })
}
wg.Wait()
```

### Graceful Shutdown

```go
ctx, cancel := context.WithCancel(context.Background())
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

go func() {
    <-sigChan
    slog.Info("Shutdown signal received")
    cancel()
}()

// Wait with timeout
done := make(chan struct{})
go func() {
    wg.Wait()
    close(done)
}()

select {
case <-done:
    slog.Info("All monitors stopped gracefully")
case <-time.After(GracefulShutdownTimeout):
    slog.Warn("Graceful shutdown timeout, forcing exit")
}
```

## Error Handling Strategy

- **Config loading errors**: Return immediately, don't start monitoring
- **Service check errors**: Log with structured fields, continue monitoring other services
- **Webhook failures**: Log errors, don't stop monitoring (alerts are best-effort)
- **Context cancellation**: Clean exit from all goroutines

## Validation Checklist

- [ ] Unit tests cover core functions
- [ ] All integration tests pass (`task test:integration`)
- [ ] E2E webhook delivery test passes (`task test:e2e`)
- [ ] `task validate` passes (format, lint, test)
- [ ] Graceful shutdown works (SIGINT/SIGTERM)
- [ ] Multiple services monitored concurrently
- [ ] Webhook retries work on 5xx responses
- [ ] Config validation catches invalid YAML
- [ ] Structured logging includes service names

## Testing Commands

```bash
task setup              # First time setup
task build              # Build binary
task test:unit          # Fast unit tests
task test:integration   # HTTP/webhook integration tests
task test:e2e           # End-to-end with real binary
task validate           # Full validation pipeline
```

## Dependencies

**Standard Library Only:**

- `context` - cancellation & timeouts
- `net/http` - HTTP client
- `time` - tickers, durations, parsing
- `sync` - WaitGroup.Go(), Mutex
- `encoding/json` - webhook payloads
- `os`, `os/signal` - signal handling
- `errors` - error wrapping & inspection
- `log/slog` - structured logging

**Third-Party (minimal):**

- `gopkg.in/yaml.v3` - YAML parsing
- `github.com/spf13/cobra` - CLI framework (already imported)

## Notes

- Keep `internal/core/` as single implementation package
- Use struct validation, not runtime JSON schema validation
- HTTP 4xx responses are NOT failures (only 5xx)
- MVP sends alerts on every check failure (no deduplication)
- MVP only supports GET method for service checks
- MVP only loads config from file at startup (no hot reload)
- MVP only supports basic auth via URL (no custom headers)
- MVP uses a fixed 1-second delay between webhook retries (not exponential backoff)
- MVP checks services at fixed intervals (no jitter/randomization)
- MVP checks all services in the configuration file (no filtering, tags, etc.)
