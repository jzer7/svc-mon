package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
)

// WebhookPayload matches the expected alert structure
type WebhookPayload struct {
	Service string `json:"service"`
	URL     string `json:"url"`
	Status  string `json:"status"`
	Reason  string `json:"reason"`
}

type MockWebhookServer struct {
	mu       sync.Mutex
	webhooks []WebhookPayload
}

func (m *MockWebhookServer) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	m.mu.Lock()
	m.webhooks = append(m.webhooks, payload)
	count := len(m.webhooks)
	m.mu.Unlock()

	log.Printf("Received webhook #%d: service=%s, status=%s, reason=%s",
		count, payload.Service, payload.Status, payload.Reason)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"received": true}`))
}

func (m *MockWebhookServer) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status": "ok"}`))
}

func (m *MockWebhookServer) HandleStats(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	response := struct {
		WebhooksReceived int              `json:"webhooks_received"`
		Webhooks         []WebhookPayload `json:"webhooks"`
	}{
		WebhooksReceived: len(m.webhooks),
		Webhooks:         m.webhooks,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <port>\n", os.Args[0])
		os.Exit(1)
	}

	port := os.Args[1]
	server := &MockWebhookServer{
		webhooks: make([]WebhookPayload, 0),
	}

	http.HandleFunc("/webhook", server.HandleWebhook)
	http.HandleFunc("/health", server.HandleHealth)
	http.HandleFunc("/stats", server.HandleStats)

	addr := ":" + port
	log.Printf("Mock webhook server listening on %s", addr)
	log.Printf("  POST %s/webhook - receive alerts", addr)
	log.Printf("  GET  %s/health  - health check", addr)
	log.Printf("  GET  %s/stats   - view received webhooks", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
