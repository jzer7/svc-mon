# Service Monitor - AI Agent Instructions

## Project Overview

**svc-mon** is a lightweight service monitoring tool written in Go that checks HTTP/HTTPS endpoints and sends webhook alerts when services fail. The project emphasizes simplicity and single-binary deployment.

### Architecture

- **Binary**: `svc-mon` (service monitor) under `cmd/svc-mon` calls code
- **Core package**: `internal/core` - contains monitoring logic and data models
- **Configuration-driven**: YAML config with JSON schema validation (`schemas/config.schema.json`)
- **CLI framework**: Uses Cobra for command structure

### Key Components

```
cmd/svc-mon/main.go      → Main CLI entry point with Cobra commands (minimal implementation)
internal/core/           → Core monitoring logic (minimal implementation)
  ├── server.go          → Monitoring orchestration
  └── model.go           → Data models and types
examples/config.yaml     → Reference configuration
schemas/config.schema.json → Config validation schema
```

## Development Workflow

### Build System (Task)

This project uses [Task](https://taskfile.dev/) as its build tool, NOT Make or go commands directly. Always use Task commands:

```bash
task                    # Show help and available tasks
task setup              # Install tools and dependencies (run first)
task build              # Build all binaries
task validate           # Run full validation pipeline
task test               # Run all tests (unit, integration, e2e, benchmarks)
task test:unit          # Run unit tests only
task test:coverage      # Generate HTML coverage report
task check              # Run linters and static analysis
task format             # Format code and documentation files
task clean              # Remove most common artifacts
```

### Development Tools

Tools are installed locally in `tools/` directory (not globally) via `task setup:tools`:

- `golangci-lint` - comprehensive linter
- `staticcheck` - static analysis
- `goimports` - code formatting
- `gopls` - language server
- `dlv` - debugger
- `cobra-cli` - CLI scaffolding

### Testing Strategy

- **Unit tests**: Standard Go tests in `*_test.go` files
- **Coverage**: Generated as `coverage.html` via `task test:coverage`
- **Integration tests**: Use `httptest` for isolated testing (see below)
- **E2E tests**: Shell scripts with mock servers via Task (see below)
- **Benchmarks**: Run with 5-second duration via `task test:benchmarks`

#### Integration Testing Pattern

Integration tests use Go's `httptest` package to create mock HTTP servers without external dependencies:

```go
// Example from internal/core/monitor_integration_test.go
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusInternalServerError)
}))
defer server.Close()

// Test monitor against server.URL
// result := CheckService(server.URL, timeout)
```

**Integration test files**:

- `internal/core/monitor_integration_test.go` - HTTP monitoring scenarios (200/5xx/timeouts/DNS failures)
- `internal/core/webhook_integration_test.go` - Webhook delivery, retries, multiple endpoints

**Run**: `task test:integration` (filters for tests with "Integration" in name)

#### E2E Testing Pattern

E2E tests use bash scripts that start real binaries and verify end-to-end behavior:

**Test infrastructure** (`test/e2e/`):

- `basic-monitoring.sh` - Smoke tests (version, config validation, binary execution)
- `webhook-delivery.sh` - Full webhook alert flow with mock server
- `mock_webhook_server.go` - Simple HTTP server that logs received webhooks
- `test-config.yaml` - Sample config for E2E scenarios

**Run**:

- `task test:e2e` - Run all E2E tests
- `task test:e2e:basic` - Run basic smoke tests only
- `task test:e2e:webhook` - Run webhook delivery tests only

**Rationale**: This approach maintains simplicity (shell + standard library), integrates with Task workflow, and avoids heavy test frameworks while providing solid coverage.

## Code Conventions

### Configuration Schema

YAML configs MUST match `schemas/config.schema.json`. Key patterns:

1. **Duration strings**: Use `"60s"` format with regex `^\d+s$`
2. **Alert conditions**: Enum of `["dns_failure", "timeout", "http_5xx"]`
3. **Webhooks**: Array of URIs (unique)
4. **Schema annotation**: Add YAML language server directive:
   ```yaml
   # yaml-language-server: $schema=../schemas/config.schema.json
   ```

### Go Code Patterns

1. **Package structure**:
   - `cmd/` packages are `package main`
   - Business logic lives in `internal/core`
   - Follow standard Go project layout

2. **Logging**: Use `slog` (standard library), initialized in `main.go`:

   ```go
   _appLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
       Level: slog.LevelInfo,
   }))
   slog.SetDefault(_appLogger)
   ```

3. **CLI commands**: Cobra pattern with persistent pre-run hooks for initialization
   - Root command: `svc-mon`
   - Subcommands: `monitor`, `version`
   - Flags: Use Cobra's flag system (e.g., `--config`)

4. **Data models**: Defined in `internal/core/model.go`:
   - `Service`, `MonitorConfig`, `AlertConfig`
   - `MonitoringState`, `MonitoringResult`
   - Constructor pattern: `NewSvcMonModel()`

### File Naming & Organization

- **Multiple Taskfiles**: Domain-specific Taskfiles included from main:
  - `Taskfile-test.yml` → test tasks
  - `Taskfile-static.yml` → linting/static analysis
  - `Taskfile-style.yml` → formatting
  - `Taskfile-clean.yml` → cleanup tasks
  - `Taskfile-act.yml` → GitHub Actions local testing

- **Build outputs**: All binaries go to `bin/` directory
- **Example configs**: Place in `examples/` with schema references

## Important Notes

1. **Early stage**: Core monitoring logic in `internal/core/server.go` is a stub. Implementation is pending.

2. **JSON Schema integration**: The config schema provides autocomplete in editors. Always update schema when changing config structure.

3. **Single binary project**: Binary is defined in `Taskfile.yml` var `CLI_BINS`. Build system handles build.

4. **No GitHub workflows yet**: CI/CD not configured (`.github/workflows/` empty)

5. **Sandbox directory**: Contains experimental/template code, old code, reference code. It is not part of the main application, and MUST be ignored.

## When Adding Features

- Update `schemas/config.schema.json` if adding config options
- Add tests to `*_test.go` files in appropriate packages
- Run `task validate` before committing
- Update `examples/config.yaml` if config schema changes
- Consider webhook payload format (documented in README) when implementing alerts
- Use structured logging with `slog` for operational visibility
