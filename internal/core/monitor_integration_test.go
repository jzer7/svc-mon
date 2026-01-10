package core

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestMonitorHTTPEndpoint demonstrates integration testing using httptest
// to verify HTTP monitoring logic without external dependencies.
func TestMonitorHTTPEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		serverHandler  http.HandlerFunc
		timeout        time.Duration
		expectedStatus string
		expectedReason string
	}{
		{
			name: "healthy service returns 200",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "healthy"}`))
			},
			timeout:        5 * time.Second,
			expectedStatus: "up",
			expectedReason: "",
		},
		{
			name: "service returns 500",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error": "internal error"}`))
			},
			timeout:        5 * time.Second,
			expectedStatus: "down",
			expectedReason: "http_5xx",
		},
		{
			name: "service returns 503",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusServiceUnavailable)
			},
			timeout:        5 * time.Second,
			expectedStatus: "down",
			expectedReason: "http_5xx",
		},
		{
			name: "slow service causes timeout",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(2 * time.Second)
				w.WriteHeader(http.StatusOK)
			},
			timeout:        500 * time.Millisecond,
			expectedStatus: "down",
			expectedReason: "timeout",
		},
		{
			name: "service returns 404 (not an alert condition)",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			timeout:        5 * time.Second,
			expectedStatus: "up",
			expectedReason: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test HTTP server with the handler
			server := httptest.NewServer(tt.serverHandler)
			defer server.Close()

			ctx := context.Background()
			result := CheckService(ctx, server.URL, tt.timeout)

			if result.Status != tt.expectedStatus {
				t.Errorf("expected status %q, got %q", tt.expectedStatus, result.Status)
			}

			if result.Reason != tt.expectedReason {
				t.Errorf("expected reason %q, got %q", tt.expectedReason, result.Reason)
			}
		})
	}
}

// TestMonitorWithAuth demonstrates testing authentication header handling
func TestMonitorWithAuth(t *testing.T) {
	expectedToken := "Bearer test-token-123"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != expectedToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// TODO: Implement CheckServiceWithAuth function
	// result := CheckServiceWithAuth(server.URL, expectedToken, 5*time.Second)
	//
	// if result.Status != "up" {
	//     t.Errorf("expected authenticated request to succeed, got status: %q", result.Status)
	// }

	t.Skip("CheckServiceWithAuth not yet implemented")
}

// TestMonitorDNSFailure demonstrates handling DNS resolution failures
func TestMonitorDNSFailure(t *testing.T) {
	// Use an invalid hostname that will fail DNS resolution
	invalidURL := "http://this-domain-does-not-exist-12345.invalid"

	ctx := context.Background()
	result := CheckService(ctx, invalidURL, 5*time.Second)

	if result.Status != "down" {
		t.Errorf("expected DNS failure to return down status, got: %q", result.Status)
	}

	if result.Reason != "dns_failure" {
		t.Errorf("expected reason 'dns_failure', got: %q", result.Reason)
	}
}
