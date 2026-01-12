package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
)

type WebhookServer struct {
	trackingKey string
	counts      map[string]int
	mu          sync.RWMutex
}

func NewWebhookServer(trackingKey string) *WebhookServer {
	return &WebhookServer{
		trackingKey: trackingKey,
		counts:      make(map[string]int),
	}
}

func (m *WebhookServer) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Failed to read request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse JSON payload
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("Received payload: not JSON")
		w.WriteHeader(http.StatusOK)
		return
	}

	// Log the valid JSON payload
	log.Printf("Received payload: %s", string(body))

	// Extract the tracking key value
	keyValue := "unknown"
	if val, ok := payload[m.trackingKey]; ok {
		keyValue = fmt.Sprintf("%v", val)
	}

	// Update counts
	m.mu.Lock()
	m.counts[keyValue]++
	m.mu.Unlock()

	w.WriteHeader(http.StatusOK)
}

func (m *WebhookServer) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status": "ok"}`))
}

func (m *WebhookServer) HandleStats(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Calculate total
	total := 0
	for _, count := range m.counts {
		total += count
	}

	// Build stats response
	stats := map[string]interface{}{
		"key":    m.trackingKey,
		"total":  total,
		"counts": m.counts,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		log.Printf("Error encoding stats: %v", err)
		http.Error(w, "Failed to encode stats", http.StatusInternalServerError)
		return
	}
}

func main() {
	port := flag.String("p", "8080", "port where the server listens")
	key := flag.String("k", "", "key used by the JSON payload tracker")
	flag.Parse()

	// Validate required flags
	if *key == "" {
		fmt.Fprintf(os.Stderr, "Error: -k flag is required\n")
		flag.Usage()
		os.Exit(1)
	}

	server := NewWebhookServer(*key)

	http.HandleFunc("/webhook", server.HandleWebhook)
	http.HandleFunc("/health", server.HandleHealth)
	http.HandleFunc("/stats", server.HandleStats)

	addr := ":" + *port
	log.Printf("Webhook logger server listening on %s", addr)
	log.Printf("  POST %s/webhook - receive alerts", addr)
	log.Printf("  GET  %s/health  - health check", addr)
	log.Printf("  GET  %s/stats   - view received webhooks", addr)
	log.Printf("Using key: %s", *key)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
