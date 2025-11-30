package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// RequestInfo holds details about a captured HTTP request
type RequestInfo struct {
	ID         int               `json:"id"`
	Method     string            `json:"method"`
	URL        string            `json:"url"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Timestamp  time.Time         `json:"timestamp"`
	RemoteAddr string            `json:"remote_addr"`
}

var (
	requests []RequestInfo
	mu       sync.RWMutex
	nextID   = 1
)

func main() {
	// Serve static files for the UI
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/ui/", http.StripPrefix("/ui/", fs))

	// API endpoint to get requests
	http.HandleFunc("/api/requests", getRequestsHandler)

	// API endpoint to clear requests
	http.HandleFunc("/api/clear", clearRequestsHandler)

	// Catch-all handler for webhooks
	http.HandleFunc("/", webhookHandler)

	port := ":8080"
	fmt.Printf("Server started on http://localhost%s\n", port)
	fmt.Printf("UI available at http://localhost%s/ui/\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	// Skip if it matches our internal paths (just in case, though ServeMux handles specificity)
	// But since "/" matches everything, we don't strictly need this if we trust ServeMux.
	// However, let's be safe.

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	headers := make(map[string]string)
	for k, v := range r.Header {
		headers[k] = v[0] // Just taking the first value for simplicity
	}

	info := RequestInfo{
		Method:     r.Method,
		URL:        r.URL.String(),
		Headers:    headers,
		Body:       string(bodyBytes),
		Timestamp:  time.Now(),
		RemoteAddr: r.RemoteAddr,
	}

	mu.Lock()
	info.ID = nextID
	nextID++
	// Prepend to show newest first
	requests = append([]RequestInfo{info}, requests...)
	// Keep only last 100 requests to avoid memory issues
	if len(requests) > 100 {
		requests = requests[:100]
	}
	mu.Unlock()

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Webhook received")
}

func getRequestsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	mu.RLock()
	defer mu.RUnlock()
	json.NewEncoder(w).Encode(requests)
}

func clearRequestsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	mu.Lock()
	requests = []RequestInfo{}
	mu.Unlock()
	w.WriteHeader(http.StatusOK)
}
