package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/pismo/testing-proxy/internal/config"
	"github.com/pismo/testing-proxy/internal/mode"
	"github.com/pismo/testing-proxy/internal/models"
	"github.com/pismo/testing-proxy/internal/storage"
)

// ProxyHandler handles incoming proxy requests
type ProxyHandler struct {
	config    *config.Config
	recorder  *mode.Recorder
	player    *mode.Player
	stats     *Statistics
	history   *RequestHistory
	mu        sync.Mutex // Sequential processing
}

// Statistics tracks proxy metrics
type Statistics struct {
	RecordCount   int64 `json:"record_count"`
	PlaybackHits  int64 `json:"playback_hits"`
	PlaybackMisses int64 `json:"playback_misses"`
	mu            sync.RWMutex
}

// RequestHistoryEntry tracks a single request in the session
type RequestHistoryEntry struct {
	ID        string                `json:"id"`
	Timestamp string                `json:"timestamp"`
	Method    string                `json:"method"`
	URL       string                `json:"url"`
	Target    string                `json:"target"`
	Status    int                   `json:"status"`
	Duration  int64                 `json:"duration"`
	Saved     bool                  `json:"saved"` // Whether this was saved (not a duplicate)
}

// RequestHistory tracks all requests this session
type RequestHistory struct {
	entries []RequestHistoryEntry
	mu      sync.RWMutex
}

// NewProxyHandler creates a new proxy handler
func NewProxyHandler(repository storage.Repository) *ProxyHandler {
	return &ProxyHandler{
		config:   config.GetInstance(),
		recorder: mode.NewRecorder(repository),
		player:   mode.NewPlayer(repository),
		stats:    &Statistics{},
		history:  &RequestHistory{entries: make([]RequestHistoryEntry, 0, 100)},
	}
}

// AddToHistory adds a request to the history log
func (h *ProxyHandler) AddToHistory(entry RequestHistoryEntry) {
	h.history.mu.Lock()
	defer h.history.mu.Unlock()

	// Prepend to show newest first
	h.history.entries = append([]RequestHistoryEntry{entry}, h.history.entries...)

	// Keep last 1000 entries to prevent memory bloat
	if len(h.history.entries) > 1000 {
		h.history.entries = h.history.entries[:1000]
	}
}

// GetHistory returns the request history
func (h *ProxyHandler) GetHistory() []RequestHistoryEntry {
	h.history.mu.RLock()
	defer h.history.mu.RUnlock()

	// Return a copy to prevent race conditions
	result := make([]RequestHistoryEntry, len(h.history.entries))
	copy(result, h.history.entries)
	return result
}

// ServeHTTP handles incoming HTTP requests
func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Sequential processing with mutex
	h.mu.Lock()
	defer h.mu.Unlock()

	// Extract target from query parameter
	target := r.URL.Query().Get("target")
	if target == "" {
		http.Error(w, `{"error":"Missing 'target' query parameter"}`, http.StatusBadRequest)
		return
	}

	// Validate target URL
	if _, err := url.Parse(target); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"Invalid target URL: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error":"Failed to read request body"}`, http.StatusBadRequest)
		return
	}

	// Handle based on current mode
	currentMode := h.config.GetMode()
	var interaction *models.Interaction
	startTime := time.Now()

	if currentMode == "record" {
		interaction, err = h.handleRecord(r, target, body)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"Record failed: %s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
		h.stats.incrementRecord()
	} else {
		interaction, err = h.handlePlayback(r, target, body)
		if err != nil {
			if _, ok := err.(*mode.ErrNoRecording); ok {
				h.stats.incrementMiss()
				http.Error(w, fmt.Sprintf(`{"error":"No recording found: %s"}`, err.Error()), http.StatusNotFound)
				return
			}
			http.Error(w, fmt.Sprintf(`{"error":"Playback failed: %s"}`, err.Error()), http.StatusInternalServerError)
			return
		}
		h.stats.incrementHit()
	}

	// Add to history log
	h.AddToHistory(RequestHistoryEntry{
		ID:        interaction.Request.GenerateHash(),
		Timestamp: time.Now().Format(time.RFC3339),
		Method:    interaction.Request.Method,
		URL:       interaction.Request.URL,
		Target:    interaction.Metadata.Target,
		Status:    interaction.Response.StatusCode,
		Duration:  time.Since(startTime).Milliseconds(),
		Saved:     currentMode == "record", // In record mode, assume saved (could check if exists)
	})

	// Write response
	h.writeResponse(w, interaction.Response)
}

// handleRecord processes request in record mode
func (h *ProxyHandler) handleRecord(r *http.Request, target string, body []byte) (*models.Interaction, error) {
	return h.recorder.Handle(r, target, body)
}

// handlePlayback processes request in playback mode
func (h *ProxyHandler) handlePlayback(r *http.Request, target string, body []byte) (*models.Interaction, error) {
	return h.player.Handle(r, target, body)
}

// writeResponse writes the recorded response to the client
func (h *ProxyHandler) writeResponse(w http.ResponseWriter, resp models.RecordedResponse) {
	// Copy headers
	for key, values := range resp.Headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Write status code
	w.WriteHeader(resp.StatusCode)

	// Write body if present
	if resp.Body != nil {
		w.Write(resp.Body)
	}
}

// GetStatistics returns current statistics
func (h *ProxyHandler) GetStatistics() map[string]interface{} {
	h.stats.mu.RLock()
	defer h.stats.mu.RUnlock()

	return map[string]interface{}{
		"mode":            h.config.GetMode(),
		"record_count":    h.stats.RecordCount,
		"playback_hits":   h.stats.PlaybackHits,
		"playback_misses": h.stats.PlaybackMisses,
	}
}

// Statistics increment methods
func (s *Statistics) incrementRecord() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.RecordCount++
}

func (s *Statistics) incrementHit() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PlaybackHits++
}

func (s *Statistics) incrementMiss() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PlaybackMisses++
}