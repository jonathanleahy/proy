package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/pismo/testing-proxy/internal/config"
	"github.com/pismo/testing-proxy/internal/handler"
	"github.com/pismo/testing-proxy/internal/storage"
)

func main() {
	// ASCII Art Banner
	fmt.Println(`
╔══════════════════════════════════════════════╗
║     HTTP Testing Proxy - Simple & Elegant     ║
║         Record • Replay • Test • Win           ║
╚══════════════════════════════════════════════╝`)

	// Load configuration
	cfg := config.GetInstance()
	if err := cfg.Load(); err != nil {
		log.Printf("Warning: Failed to load config file: %v", err)
	}

	// Display configuration
	fmt.Printf("📍 Starting proxy server on %s\n", cfg.GetAddress())
	fmt.Printf("📁 Recordings directory: %s\n", cfg.Storage.Path)
	fmt.Printf("🎯 Default mode: %s\n", cfg.Mode.Default)
	fmt.Printf("🔒 TLS verification: %v\n", !cfg.TLS.SkipVerify)
	fmt.Println()

	// Initialize storage repository
	repository, err := storage.NewFileSystemRepository(cfg.Storage.Path)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Display initial statistics
	count, _ := repository.Count()
	fmt.Printf("📊 Existing recordings: %d\n", count)

	// Create handlers
	proxyHandler := handler.NewProxyHandler(repository)
	managementHandler := handler.NewManagementHandler(repository, proxyHandler)

	// Setup HTTP routes
	mux := http.NewServeMux()

	// Management endpoints (must be registered first)
	mux.HandleFunc("/admin/status", managementHandler.HandleStatus)
	mux.HandleFunc("/admin/mode", managementHandler.HandleMode)
	mux.HandleFunc("/admin/history", managementHandler.HandleHistory)
	mux.HandleFunc("/admin/recordings", managementHandler.HandleRecordings)
	mux.HandleFunc("/admin/recording", managementHandler.HandleRecording)
	mux.HandleFunc("/admin/ui", managementHandler.HandleDashboard)
	mux.HandleFunc("/health", managementHandler.HandleHealth)

	// Proxy handles all other paths (catch-all)
	mux.Handle("/", proxyHandler)

	// Setup graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	server := &http.Server{
		Addr:    cfg.GetAddress(),
		Handler: loggingMiddleware(mux),
	}

	go func() {
		fmt.Println("\n✅ Proxy server is ready!")
		fmt.Println("📖 Documentation:")
		fmt.Printf("   • Proxy endpoint: http://%s/<any-path>?target=<target-host>\n", cfg.GetAddress())
		fmt.Printf("   • Dashboard UI:   http://%s/admin/ui\n", cfg.GetAddress())
		fmt.Printf("   • Health check:   http://%s/health\n", cfg.GetAddress())
		fmt.Println("\n🎮 Management API:")
		fmt.Printf("   • GET    /admin/status     - View status and statistics\n")
		fmt.Printf("   • POST   /admin/mode       - Switch between record/playback\n")
		fmt.Printf("   • GET    /admin/recordings - List all recordings\n")
		fmt.Printf("   • GET    /admin/recording?id=<id> - Get recording details\n")
		fmt.Printf("   • DELETE /admin/recordings - Clear all recordings\n")
		fmt.Println("\n⌨️  Press Ctrl+C to stop the server")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-stop
	fmt.Println("\n🛑 Shutting down proxy server...")

	// Final statistics
	finalCount, _ := repository.Count()
	fmt.Printf("📊 Total recordings saved: %d\n", finalCount)
	fmt.Println("👋 Goodbye!")
}

// loggingMiddleware logs all incoming requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip logging for dashboard assets
		if r.URL.Path != "/" && r.URL.Path != "/health" && r.URL.Path != "/admin/ui" {
			log.Printf("[%s] %s %s", r.RemoteAddr, r.Method, r.URL.String())
		}
		next.ServeHTTP(w, r)
	})
}