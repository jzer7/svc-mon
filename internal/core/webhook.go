package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// SendWebhooks sends alert payloads to multiple webhook endpoints concurrently
func SendWebhooks(ctx context.Context, urls []string, payload WebhookPayload) []error {
	if len(urls) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	errors := make([]error, len(urls))

	for i, url := range urls {
		i, url := i, url // capture loop vars
		wg.Add(1)
		go func() {
			defer wg.Done()
			errors[i] = sendWebhook(ctx, url, payload)
		}()
	}

	wg.Wait()
	return errors
}

// sendWebhook sends a single webhook with retry logic
func sendWebhook(ctx context.Context, url string, payload WebhookPayload) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	client := &http.Client{
		Timeout: WebhookTimeout,
	}

	var lastErr error
	for attempt := 0; attempt <= WebhookRetries; attempt++ {
		if attempt > 0 {
			slog.Info("retrying webhook", "url", url, "attempt", attempt)
			time.Sleep(WebhookRetryDelay)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payloadBytes))
		if err != nil {
			lastErr = fmt.Errorf("failed to create request: %w", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		resp.Body.Close()

		// Success on 2xx status
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			slog.Info("webhook sent successfully", "url", url, "status", resp.StatusCode)
			return nil
		}

		// Retry on 5xx status
		if resp.StatusCode >= 500 && resp.StatusCode < 600 {
			lastErr = fmt.Errorf("webhook returned %d", resp.StatusCode)
			continue
		}

		// Don't retry on 4xx status
		return fmt.Errorf("webhook returned %d", resp.StatusCode)
	}

	slog.Warn("webhook failed after retries", "url", url, "error", lastErr)
	return lastErr
}
