// Package core provides components to implement svc-mon
package core

import "time"

// Config represents the complete monitoring configuration
type Config struct {
	Services []ServiceConfig `yaml:"services" json:"services"`
	Defaults Defaults        `yaml:"defaults" json:"defaults"`
}

// ServiceConfig represents a single service to monitor
type ServiceConfig struct {
	Name     string        `yaml:"name" json:"name"`
	URL      string        `yaml:"url" json:"url"`
	Interval time.Duration `yaml:"interval" json:"interval"`
	Timeout  time.Duration `yaml:"timeout" json:"timeout"`
	AlertIf  []string      `yaml:"alert_if" json:"alert_if"`
	Webhooks []string      `yaml:"webhooks" json:"webhooks"`
}

// Defaults contains default values for service configurations
type Defaults struct {
	Interval time.Duration `yaml:"interval" json:"interval"`
	Timeout  time.Duration `yaml:"timeout" json:"timeout"`
}

// MonitoringResult represents the result of a service health check
type MonitoringResult struct {
	ServiceName string
	URL         string
	Status      string // "up" or "down"
	Reason      string // "timeout", "http_5xx", "dns_failure", or empty
	StatusCode  int
	Timestamp   time.Time
	Error       error
}

// WebhookPayload represents the JSON payload sent to webhook endpoints
type WebhookPayload struct {
	ServiceName string    `json:"service_name"`
	ServiceURL  string    `json:"service_url"`
	Status      string    `json:"status"`
	Reason      string    `json:"reason"`
	StatusCode  int       `json:"status_code,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
	Message     string    `json:"message,omitempty"`
}

// NewConfig creates a new Config with default values applied
func NewConfig() *Config {
	return &Config{
		Defaults: Defaults{
			Interval: DefaultInterval,
			Timeout:  DefaultTimeout,
		},
	}
}
