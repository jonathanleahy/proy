package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jonathanleahy/prroxy/rest-v2/internal/person"
	"github.com/jonathanleahy/prroxy/rest-v2/internal/user"
	"go.uber.org/zap"
)

const version = "2.0.0"

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Get configuration from environment
	config := loadConfig()

	// Initialize dependencies
	// User domain
	userClient := user.NewClient(config.JSONPlaceholderTarget)
	userService := user.NewService(userClient)
	userHandler := user.NewHandler(userService)

	// Person domain
	personClient := person.NewClient(config.ExternalUserTarget)
	personService := person.NewService(personClient)
	personHandler := person.NewHandler(personService)

	// Setup HTTP router
	mux := http.NewServeMux()

	// Health endpoint
	mux.HandleFunc("/health", handleHealth)

	// User endpoints
	mux.HandleFunc("/api/user/", func(w http.ResponseWriter, r *http.Request) {
		// Route to appropriate handler based on path
		if r.URL.Path == "/api/user/"+extractUserID(r.URL.Path)+"/summary" {
			userHandler.GetUserSummary(w, r)
		} else if r.URL.Path == "/api/user/"+extractUserID(r.URL.Path)+"/report" {
			userHandler.GetUserReport(w, r)
		} else {
			userHandler.GetUser(w, r)
		}
	})

	// Person endpoints
	mux.HandleFunc("/api/person", personHandler.FindPerson)
	mux.HandleFunc("/api/people", personHandler.FindPeople)

	// Create server
	server := &http.Server{
		Addr:    ":" + config.Port,
		Handler: mux,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting server",
			zap.String("version", version),
			zap.String("port", config.Port),
		)
		logger.Info("Available endpoints",
			zap.Strings("user", []string{
				"GET  /api/user/:id",
				"GET  /api/user/:id/summary",
				"POST /api/user/:id/report",
			}),
			zap.Strings("person", []string{
				"GET /api/person?surname=X&dob=YYYY-MM-DD",
				"GET /api/people?surname=X",
				"GET /api/people?dob=YYYY-MM-DD",
			}),
		)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Server shutting down...")
}

// Config holds application configuration
type Config struct {
	Port                  string
	JSONPlaceholderTarget string
	ExternalUserTarget    string
}

// loadConfig loads configuration from environment variables
func loadConfig() Config {
	return Config{
		Port:                  getEnv("PORT", "3004"),
		JSONPlaceholderTarget: getEnv("JSONPLACEHOLDER_TARGET", "http://localhost:8099/proxy?target=https://jsonplaceholder.typicode.com"),
		ExternalUserTarget:    getEnv("EXTERNAL_USER_TARGET", "http://localhost:8099/proxy?target=http://0.0.0.0:3006"),
	}
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// handleHealth handles the health check endpoint
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","service":"rest-v2","version":"%s"}`, version)
}

// extractUserID extracts user ID from path /api/user/:id/*
func extractUserID(path string) string {
	// Simple extraction - just get the part after /api/user/
	const prefix = "/api/user/"
	if len(path) <= len(prefix) {
		return ""
	}
	remaining := path[len(prefix):]
	// Find first /
	for i, ch := range remaining {
		if ch == '/' {
			return remaining[:i]
		}
	}
	return remaining
}
