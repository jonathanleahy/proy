package health

import "time"

// Health represents the health status of the application
type Health struct {
	Status    string
	Timestamp time.Time
	Version   string
}

// NewHealth creates a new Health instance
func NewHealth(version string) *Health {
	return &Health{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   version,
	}
}

// IsHealthy checks if the application is healthy
func (h *Health) IsHealthy() bool {
	return h.Status == "healthy"
}
