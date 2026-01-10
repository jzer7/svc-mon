package core

import "time"

const (
	// Default monitoring intervals and timeouts
	DefaultInterval         = 300 * time.Second // 5 minutes
	DefaultTimeout          = 5 * time.Second
	WebhookTimeout          = 10 * time.Second
	WebhookRetries          = 3
	WebhookRetryDelay       = 1 * time.Second
	GracefulShutdownTimeout = 30 * time.Second

	// Service status constants
	StatusUp   = "up"
	StatusDown = "down"

	// Alert condition reasons
	ReasonTimeout    = "timeout"
	ReasonHTTP5xx    = "http_5xx"
	ReasonDNSFailure = "dns_failure"
)
