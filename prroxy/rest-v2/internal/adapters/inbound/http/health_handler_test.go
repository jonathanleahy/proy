package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jonathanleahy/prroxy/rest-v2/internal/domain/health"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockHealthService is a mock implementation of the HealthService port
type MockHealthService struct {
	health *health.Health
	err    error
}

func (m *MockHealthService) GetHealth() (*health.Health, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.health, nil
}

func TestNewHealthHandler(t *testing.T) {
	mockService := &MockHealthService{}
	handler := NewHealthHandler(mockService)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.healthService)
}

func TestHealthHandler_GetHealth_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := &MockHealthService{
		health: &health.Health{
			Status:    "healthy",
			Timestamp: time.Now(),
			Version:   "2.0.0",
		},
	}
	handler := NewHealthHandler(mockService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/health", nil)

	// Execute
	handler.GetHealth(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response.Status)
	assert.Equal(t, "2.0.0", response.Version)
	assert.NotEmpty(t, response.Timestamp)
}

func TestHealthHandler_GetHealth_ServiceError(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := &MockHealthService{
		err: errors.New("service unavailable"),
	}
	handler := NewHealthHandler(mockService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/health", nil)

	// Execute
	handler.GetHealth(c)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "Failed to get health status")
}

func TestHealthHandler_GetHealth_ReturnsJSON(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := &MockHealthService{
		health: health.NewHealth("2.0.0"),
	}
	handler := NewHealthHandler(mockService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/health", nil)

	// Execute
	handler.GetHealth(c)

	// Assert
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
}

func TestHealthHandler_GetHealth_ConsistentResponse(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockService := &MockHealthService{
		health: health.NewHealth("2.0.0"),
	}
	handler := NewHealthHandler(mockService)

	// Execute multiple times
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/health", nil)

		handler.GetHealth(c)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		var response HealthResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "healthy", response.Status)
		assert.Equal(t, "2.0.0", response.Version)
	}
}
