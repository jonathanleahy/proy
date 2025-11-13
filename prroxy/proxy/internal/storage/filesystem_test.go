package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pismo/testing-proxy/internal/models"
)

func TestFileSystemRepository(t *testing.T) {
	// Create temporary directory for tests
	tempDir, err := os.MkdirTemp("", "proxy-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("NewFileSystemRepository", func(t *testing.T) {
		repo, err := NewFileSystemRepository(tempDir)
		if err != nil {
			t.Fatalf("Failed to create repository: %v", err)
		}
		if repo == nil {
			t.Fatal("Repository is nil")
		}
	})

	t.Run("Save and Find", func(t *testing.T) {
		repo, _ := NewFileSystemRepository(filepath.Join(tempDir, "save-find"))

		interaction := &models.Interaction{
			ID:        "test-123",
			Timestamp: time.Now(),
			Request: models.RecordedRequest{
				Method: "GET",
				URL:    "/api/test",
				Headers: map[string][]string{
					"Content-Type": {"application/json"},
				},
				Body: json.RawMessage(`{"test":"data"}`),
			},
			Response: models.RecordedResponse{
				StatusCode: 200,
				Headers: map[string][]string{
					"Content-Type": {"application/json"},
				},
				Body: json.RawMessage(`{"result":"success"}`),
			},
			Metadata: models.InteractionMetadata{
				Target:     "api.example.com",
				DurationMS: 150,
			},
		}

		// Save interaction
		err := repo.Save(interaction)
		if err != nil {
			t.Fatalf("Failed to save interaction: %v", err)
		}

		// Find by hash
		hash := interaction.Request.GenerateHash()
		found, err := repo.Find(hash)
		if err != nil {
			t.Fatalf("Failed to find interaction: %v", err)
		}

		// Verify the found interaction
		if found.ID != interaction.ID {
			t.Errorf("ID mismatch. Expected: %s, Got: %s", interaction.ID, found.ID)
		}
		if found.Request.Method != interaction.Request.Method {
			t.Errorf("Method mismatch. Expected: %s, Got: %s", interaction.Request.Method, found.Request.Method)
		}
		if found.Response.StatusCode != interaction.Response.StatusCode {
			t.Errorf("StatusCode mismatch. Expected: %d, Got: %d", interaction.Response.StatusCode, found.Response.StatusCode)
		}
	})

	t.Run("Find non-existent", func(t *testing.T) {
		repo, _ := NewFileSystemRepository(filepath.Join(tempDir, "find-missing"))

		_, err := repo.Find("non-existent-hash")
		if err == nil {
			t.Fatal("Expected error for non-existent hash")
		}

		if _, ok := err.(ErrNotFound); !ok {
			t.Errorf("Expected ErrNotFound, got %T", err)
		}
	})

	t.Run("FindAll", func(t *testing.T) {
		repo, _ := NewFileSystemRepository(filepath.Join(tempDir, "find-all"))

		// Save multiple interactions
		interactions := []*models.Interaction{
			{
				ID:        "test-1",
				Timestamp: time.Now(),
				Request: models.RecordedRequest{
					Method: "GET",
					URL:    "/api/users",
				},
				Metadata: models.InteractionMetadata{
					Target: "api.example.com",
				},
			},
			{
				ID:        "test-2",
				Timestamp: time.Now(),
				Request: models.RecordedRequest{
					Method: "POST",
					URL:    "/api/users",
					Body:   json.RawMessage(`{"name":"test"}`),
				},
				Metadata: models.InteractionMetadata{
					Target: "api.example.com",
				},
			},
		}

		for _, interaction := range interactions {
			if err := repo.Save(interaction); err != nil {
				t.Fatalf("Failed to save interaction: %v", err)
			}
		}

		// Find all
		found, err := repo.FindAll()
		if err != nil {
			t.Fatalf("Failed to find all: %v", err)
		}

		if len(found) != len(interactions) {
			t.Errorf("Expected %d interactions, got %d", len(interactions), len(found))
		}

		// Verify all interactions are found
		foundIDs := make(map[string]bool)
		for _, f := range found {
			foundIDs[f.ID] = true
		}

		for _, i := range interactions {
			if !foundIDs[i.ID] {
				t.Errorf("Interaction %s not found", i.ID)
			}
		}
	})

	t.Run("Count", func(t *testing.T) {
		repo, _ := NewFileSystemRepository(filepath.Join(tempDir, "count"))

		// Initially should be 0
		count, err := repo.Count()
		if err != nil {
			t.Fatalf("Failed to count: %v", err)
		}
		if count != 0 {
			t.Errorf("Expected count 0, got %d", count)
		}

		// Save some interactions
		for i := 0; i < 3; i++ {
			interaction := &models.Interaction{
				ID:        string(rune('a' + i)),
				Timestamp: time.Now(),
				Request: models.RecordedRequest{
					Method: "GET",
					URL:    "/api/test/" + string(rune('a'+i)),
				},
				Metadata: models.InteractionMetadata{
					Target: "api.example.com",
				},
			}
			repo.Save(interaction)
		}

		// Count should be 3
		count, err = repo.Count()
		if err != nil {
			t.Fatalf("Failed to count: %v", err)
		}
		if count != 3 {
			t.Errorf("Expected count 3, got %d", count)
		}
	})

	t.Run("Clear", func(t *testing.T) {
		repo, _ := NewFileSystemRepository(filepath.Join(tempDir, "clear"))

		// Save some interactions
		for i := 0; i < 5; i++ {
			interaction := &models.Interaction{
				ID:        string(rune('a' + i)),
				Timestamp: time.Now(),
				Request: models.RecordedRequest{
					Method: "GET",
					URL:    "/api/test/" + string(rune('a'+i)),
				},
				Metadata: models.InteractionMetadata{
					Target: "api.example.com",
				},
			}
			repo.Save(interaction)
		}

		// Verify they exist
		count, _ := repo.Count()
		if count != 5 {
			t.Errorf("Expected 5 interactions before clear, got %d", count)
		}

		// Clear all
		err := repo.Clear()
		if err != nil {
			t.Fatalf("Failed to clear: %v", err)
		}

		// Verify they're gone
		count, _ = repo.Count()
		if count != 0 {
			t.Errorf("Expected 0 interactions after clear, got %d", count)
		}
	})

	t.Run("Service organization", func(t *testing.T) {
		repo, _ := NewFileSystemRepository(filepath.Join(tempDir, "services"))

		// Save interactions for different services
		services := []string{
			"api.users.com",
			"api.accounts.com",
			"api.transactions.com",
		}

		for _, service := range services {
			interaction := &models.Interaction{
				ID:        service,
				Timestamp: time.Now(),
				Request: models.RecordedRequest{
					Method: "GET",
					URL:    "/" + service,
				},
				Metadata: models.InteractionMetadata{
					Target: service,
				},
			}
			repo.Save(interaction)
		}

		// Check that service directories were created
		basePath := filepath.Join(tempDir, "services")
		for _, service := range services {
			serviceName := extractServiceName(service)
			serviceDir := filepath.Join(basePath, serviceName)
			if _, err := os.Stat(serviceDir); os.IsNotExist(err) {
				t.Errorf("Service directory not created for %s", service)
			}
		}
	})
}

func TestExtractServiceName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"api.example.com", "api_example_com"},
		{"http://api.example.com", "api_example_com"},
		{"https://api.example.com", "api_example_com"},
		{"api.example.com/v1/users", "api_example_com"},
		{"0.0.0.0:8080", "0.0.0.0_8080"},
		{"", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := extractServiceName(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}