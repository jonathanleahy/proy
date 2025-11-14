package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	httpAdapter "github.com/jonathanleahy/prroxy/rest-v2/internal/adapters/inbound/http"
	"github.com/jonathanleahy/prroxy/rest-v2/internal/domain/health"
)

const version = "2.0.0"

func main() {
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Setup router with dependency injection
	router := setupRouter()

	// Start server
	log.Printf("Starting REST v2 server (Hexagonal Architecture) on port %s...", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// setupRouter configures and returns the Gin router
// This is where we wire up the hexagonal architecture (dependency injection)
func setupRouter() *gin.Engine {
	router := gin.Default()

	// Domain layer: Create services (business logic)
	healthService := health.NewService(version)

	// Adapter layer: Create HTTP handlers (depend on services through ports)
	healthHandler := httpAdapter.NewHealthHandler(healthService)

	// Route configuration
	router.GET("/health", healthHandler.GetHealth)

	return router
}
