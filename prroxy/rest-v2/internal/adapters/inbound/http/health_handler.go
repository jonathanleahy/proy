package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jonathanleahy/prroxy/rest-v2/internal/ports/inbound"
)

// HealthHandler is the HTTP adapter for health check requests
// It depends on the HealthService port (dependency inversion)
type HealthHandler struct {
	healthService inbound.HealthService
}

// NewHealthHandler creates a new HTTP health handler
func NewHealthHandler(healthService inbound.HealthService) *HealthHandler {
	return &HealthHandler{
		healthService: healthService,
	}
}

// HealthResponse represents the HTTP response for health checks
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
}

// GetHealth handles GET /health requests
func (h *HealthHandler) GetHealth(c *gin.Context) {
	health, err := h.healthService.GetHealth()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get health status",
		})
		return
	}

	response := HealthResponse{
		Status:    health.Status,
		Timestamp: health.Timestamp.Format("2006-01-02T15:04:05.999999Z07:00"),
		Version:   health.Version,
	}

	c.JSON(http.StatusOK, response)
}
