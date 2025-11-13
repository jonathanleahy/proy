package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	httpAdapter "github.com/jonathanleahy/prroxy/rest-v2/internal/adapters/inbound/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupRouter_HexArchitecture(t *testing.T) {
	router := setupRouter()
	assert.NotNil(t, router)

	// Test health endpoint is configured
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response httpAdapter.HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "healthy", response.Status)
	assert.Equal(t, "2.0.0", response.Version)
}

func TestSetupRouter_NotFoundRoute(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/nonexistent", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSetupRouter_DependencyInjection(t *testing.T) {
	// Verify hexagonal architecture wiring
	router := setupRouter()
	assert.NotNil(t, router)

	// The router should have the health endpoint wired through proper DI
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)

	router.ServeHTTP(w, req)

	// Verify the response comes from properly wired components
	var response httpAdapter.HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// These values confirm domain -> port -> adapter wiring
	assert.Equal(t, "healthy", response.Status)
	assert.Equal(t, version, response.Version)
	assert.NotEmpty(t, response.Timestamp)
}
