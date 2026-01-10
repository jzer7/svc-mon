package core

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// rawConfig is used to parse YAML with duration strings
type rawConfig struct {
	Services []rawServiceConfig `yaml:"services"`
	Defaults rawDefaults        `yaml:"defaults"`
}

type rawServiceConfig struct {
	Name     string   `yaml:"name"`
	URL      string   `yaml:"url"`
	Interval string   `yaml:"interval"`
	Timeout  string   `yaml:"timeout"`
	AlertIf  []string `yaml:"alert_if"`
	Webhooks []string `yaml:"webhooks"`
}

type rawDefaults struct {
	Interval string `yaml:"interval"`
	Timeout  string `yaml:"timeout"`
}

// LoadConfig reads and parses a YAML configuration file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var raw rawConfig
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	cfg := NewConfig()

	// Parse defaults
	if raw.Defaults.Interval != "" {
		interval, err := time.ParseDuration(raw.Defaults.Interval)
		if err != nil {
			return nil, fmt.Errorf("invalid default interval: %w", err)
		}
		cfg.Defaults.Interval = interval
	}

	if raw.Defaults.Timeout != "" {
		timeout, err := time.ParseDuration(raw.Defaults.Timeout)
		if err != nil {
			return nil, fmt.Errorf("invalid default timeout: %w", err)
		}
		cfg.Defaults.Timeout = timeout
	}

	// Parse services
	if len(raw.Services) == 0 {
		return nil, fmt.Errorf("no services defined in configuration")
	}

	for i, rawSvc := range raw.Services {
		svc := ServiceConfig{
			Name:     rawSvc.Name,
			URL:      rawSvc.URL,
			AlertIf:  rawSvc.AlertIf,
			Webhooks: rawSvc.Webhooks,
			Interval: cfg.Defaults.Interval,
			Timeout:  cfg.Defaults.Timeout,
		}

		// Validate required fields
		if svc.Name == "" {
			return nil, fmt.Errorf("service at index %d missing name", i)
		}
		if svc.URL == "" {
			return nil, fmt.Errorf("service %q missing URL", svc.Name)
		}

		// Parse service-specific interval if provided
		if rawSvc.Interval != "" {
			interval, err := time.ParseDuration(rawSvc.Interval)
			if err != nil {
				return nil, fmt.Errorf("service %q has invalid interval: %w", svc.Name, err)
			}
			svc.Interval = interval
		}

		// Parse service-specific timeout if provided
		if rawSvc.Timeout != "" {
			timeout, err := time.ParseDuration(rawSvc.Timeout)
			if err != nil {
				return nil, fmt.Errorf("service %q has invalid timeout: %w", svc.Name, err)
			}
			svc.Timeout = timeout
		}

		// Validate alert conditions
		for _, condition := range svc.AlertIf {
			if condition != ReasonTimeout && condition != ReasonHTTP5xx && condition != ReasonDNSFailure {
				return nil, fmt.Errorf("service %q has invalid alert condition: %q", svc.Name, condition)
			}
		}

		cfg.Services = append(cfg.Services, svc)
	}

	return cfg, nil
}
