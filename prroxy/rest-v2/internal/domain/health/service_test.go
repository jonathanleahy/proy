package health

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	version := "2.0.0"
	service := NewService(version)

	assert.NotNil(t, service)
	assert.Equal(t, version, service.version)
}

func TestService_GetHealth(t *testing.T) {
	version := "2.0.0"
	service := NewService(version)

	health, err := service.GetHealth()

	require.NoError(t, err)
	assert.NotNil(t, health)
	assert.Equal(t, "healthy", health.Status)
	assert.Equal(t, version, health.Version)
	assert.WithinDuration(t, time.Now(), health.Timestamp, 2*time.Second)
}

func TestService_GetHealth_MultipleCalls(t *testing.T) {
	service := NewService("2.0.0")

	for i := 0; i < 5; i++ {
		health, err := service.GetHealth()
		require.NoError(t, err)
		assert.True(t, health.IsHealthy())
	}
}
