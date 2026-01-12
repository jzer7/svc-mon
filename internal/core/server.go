package core

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Run starts the monitoring server with the given configuration file
// If dryRun is true, validates configuration and exits without monitoring
func Run(configPath string, dryRun bool) error {
	// Load configuration
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	slog.Info("loaded configuration", "services", len(cfg.Services))

	if dryRun {
		slog.Info("dry-run mode: configuration is valid")
		for _, svc := range cfg.Services {
			slog.Info("service configured",
				"name", svc.Name,
				"url", svc.URL,
				"interval", svc.Interval,
				"timeout", svc.Timeout,
				"alert_conditions", svc.AlertIf,
				"webhooks", len(svc.Webhooks))
		}
		return nil
	}

	// Create cancellable context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		slog.Info("shutdown signal received", "signal", sig)
		cancel()
	}()

	// Start monitoring each service in a separate goroutine
	var wg sync.WaitGroup
	for _, svc := range cfg.Services {
		svc := svc // capture loop var
		wg.Add(1)
		go func() {
			defer wg.Done()
			monitorService(ctx, svc)
		}()
	}

	// Wait for all goroutines with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		slog.Info("all monitors stopped gracefully")
	case <-time.After(GracefulShutdownTimeout):
		slog.Warn("graceful shutdown timeout, forcing exit")
	}

	return nil
}

// monitorService continuously monitors a single service
func monitorService(ctx context.Context, svc ServiceConfig) {
	slog.Info("starting monitor", "service", svc.Name, "url", svc.URL, "interval", svc.Interval)

	ticker := time.NewTicker(svc.Interval)
	defer ticker.Stop()

	// Check immediately on startup
	checkAndAlert(ctx, svc)

	for {
		select {
		case <-ctx.Done():
			slog.Info("stopping monitor", "service", svc.Name)
			return
		case <-ticker.C:
			checkAndAlert(ctx, svc)
		}
	}
}

// checkAndAlert performs a health check and sends alerts if needed
func checkAndAlert(ctx context.Context, svc ServiceConfig) {
	result := CheckService(ctx, svc.URL, svc.Timeout)
	result.ServiceName = svc.Name

	logAttrs := []any{
		"service", svc.Name,
		"url", svc.URL,
		"status", result.Status,
	}

	if result.Reason != "" {
		logAttrs = append(logAttrs, "reason", result.Reason)
	}

	if result.StatusCode > 0 {
		logAttrs = append(logAttrs, "status_code", result.StatusCode)
	}

	if result.Error != nil {
		logAttrs = append(logAttrs, "error", result.Error)
	}

	slog.Info("health check completed", logAttrs...)

	// Check if we should alert
	if result.Status == StatusDown && shouldAlert(result.Reason, svc.AlertIf) {
		slog.Info("alert condition met", "service", svc.Name, "reason", result.Reason)

		payload := WebhookPayload{
			ServiceName: svc.Name,
			ServiceURL:  svc.URL,
			Status:      result.Status,
			Reason:      result.Reason,
			StatusCode:  result.StatusCode,
			Timestamp:   result.Timestamp,
		}

		if result.Error != nil {
			payload.Message = result.Error.Error()
		}

		errs := SendWebhooks(ctx, svc.Webhooks, payload)
		for i, err := range errs {
			if err != nil {
				slog.Warn("webhook failed", "service", svc.Name, "webhook", svc.Webhooks[i], "error", err)
			}
		}
	}
}

// shouldAlert checks if the given reason matches any alert condition
func shouldAlert(reason string, alertConditions []string) bool {
	if reason == "" {
		return false
	}

	for _, condition := range alertConditions {
		if condition == reason {
			return true
		}
	}

	return false
}
