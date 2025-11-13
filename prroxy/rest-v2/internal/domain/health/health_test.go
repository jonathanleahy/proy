package health

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewHealth(t *testing.T) {
	version := "2.0.0"
	h := NewHealth(version)

	assert.NotNil(t, h)
	assert.Equal(t, "healthy", h.Status)
	assert.Equal(t, version, h.Version)
	assert.WithinDuration(t, time.Now(), h.Timestamp, 2*time.Second)
}

func TestHealth_IsHealthy(t *testing.T) {
	h := NewHealth("2.0.0")
	assert.True(t, h.IsHealthy())
}

func TestHealth_IsHealthy_WhenUnhealthy(t *testing.T) {
	h := &Health{
		Status:    "unhealthy",
		Timestamp: time.Now(),
		Version:   "2.0.0",
	}
	assert.False(t, h.IsHealthy())
}
