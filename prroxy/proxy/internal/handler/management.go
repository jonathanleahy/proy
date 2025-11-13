package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pismo/testing-proxy/internal/config"
	"github.com/pismo/testing-proxy/internal/storage"
	"github.com/pismo/testing-proxy/web"
)

// ManagementHandler handles administrative endpoints
type ManagementHandler struct {
	config     *config.Config
	repository storage.Repository
	proxy      *ProxyHandler
	startTime  time.Time
}

// NewManagementHandler creates a new management handler
func NewManagementHandler(repository storage.Repository, proxy *ProxyHandler) *ManagementHandler {
	return &ManagementHandler{
		config:     config.GetInstance(),
		repository: repository,
		proxy:      proxy,
		startTime:  time.Now(),
	}
}

// HandleStatus returns current proxy status
func (h *ManagementHandler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	count, _ := h.repository.Count()
	stats := h.proxy.GetStatistics()

	// Add additional status info with formatted uptime
	stats["uptime"] = formatDuration(time.Since(h.startTime))
	stats["total_recordings"] = count

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// HandleMode handles mode switching
func (h *ManagementHandler) HandleMode(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Check if mode parameter is provided (for switching mode via GET)
		modeParam := r.URL.Query().Get("mode")
		if modeParam != "" {
			// Switch mode via GET
			if err := h.config.SetMode(modeParam); err != nil {
				http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
				return
			}

			response := map[string]string{
				"mode":    modeParam,
				"message": fmt.Sprintf("Switched to %s mode", modeParam),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		// Return current mode if no mode parameter
		response := map[string]string{
			"mode": h.config.GetMode(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	case http.MethodPost:
		// Switch mode via POST (kept for backward compatibility)
		var request struct {
			Mode string `json:"mode"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
			return
		}

		if err := h.config.SetMode(request.Mode); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
			return
		}

		response := map[string]string{
			"mode":    request.Mode,
			"message": fmt.Sprintf("Switched to %s mode", request.Mode),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleRecording handles individual recording retrieval
func (h *ManagementHandler) HandleRecording(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get recording ID from query parameter
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, `{"error":"Missing recording ID"}`, http.StatusBadRequest)
		return
	}

	// Find the recording
	interaction, err := h.repository.Find(id)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"Recording not found: %s"}`, err.Error()), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(interaction)
}

// HandleHistory returns the request history log
func (h *ManagementHandler) HandleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	history := h.proxy.GetHistory()

	response := map[string]interface{}{
		"count":   len(history),
		"history": history,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleRecordings handles recording management
func (h *ManagementHandler) HandleRecordings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// List all recordings
		interactions, err := h.repository.FindAll()
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"Failed to list recordings: %s"}`, err.Error()), http.StatusInternalServerError)
			return
		}

		// Convert to summary format for listing
		var recordings []map[string]interface{}
		for _, interaction := range interactions {
			recordings = append(recordings, map[string]interface{}{
				"id":        interaction.Request.GenerateHash(), // Use hash as ID for retrieval
				"uuid":      interaction.ID,                      // Keep UUID for reference
				"timestamp": interaction.Timestamp,
				"method":    interaction.Request.Method,
				"url":       interaction.Request.URL,
				"target":    interaction.Metadata.Target,
				"status":    interaction.Response.StatusCode,
				"duration":  interaction.Metadata.DurationMS,
			})
		}

		response := map[string]interface{}{
			"count":      len(recordings),
			"recordings": recordings,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	case http.MethodDelete:
		// Clear all recordings
		if err := h.repository.Clear(); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"Failed to clear recordings: %s"}`, err.Error()), http.StatusInternalServerError)
			return
		}

		response := map[string]string{
			"message": "All recordings cleared successfully",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleDashboard serves the web UI
func (h *ManagementHandler) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Serve the embedded dashboard HTML
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(web.DashboardHTML)
}

// HandleHealth provides a simple health check endpoint
func (h *ManagementHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"status": "healthy",
		"time":   time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// formatDuration formats a duration into a clean human-readable string
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%dh%dm%ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm%ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}