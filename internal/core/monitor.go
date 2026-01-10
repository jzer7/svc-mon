package core

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

// CheckService performs an HTTP health check on the given URL
func CheckService(ctx context.Context, url string, timeout time.Duration) MonitoringResult {
	result := MonitoringResult{
		URL:       url,
		Status:    StatusUp,
		Timestamp: time.Now(),
	}

	client := &http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		result.Status = StatusDown
		result.Reason = ReasonDNSFailure
		result.Error = err
		return result
	}

	resp, err := client.Do(req)
	if err != nil {
		// Classify error type
		result.Status = StatusDown
		result.Error = err

		// Check for timeout
		if errors.Is(err, context.DeadlineExceeded) || isTimeoutError(err) {
			result.Reason = ReasonTimeout
			return result
		}

		// Check for DNS failure
		var dnsErr *net.DNSError
		if errors.As(err, &dnsErr) {
			result.Reason = ReasonDNSFailure
			return result
		}

		// Default to timeout for network errors
		result.Reason = ReasonTimeout
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode

	// Check for 5xx status codes
	if resp.StatusCode >= 500 && resp.StatusCode < 600 {
		result.Status = StatusDown
		result.Reason = ReasonHTTP5xx
		result.Error = fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return result
}

// isTimeoutError checks if an error is a timeout error
func isTimeoutError(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}
	return false
}
