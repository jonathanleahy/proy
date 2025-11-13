package mode

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pismo/testing-proxy/internal/models"
	"github.com/pismo/testing-proxy/internal/storage"
)

// MockRepository implements storage.Repository for testing
type MockRepository struct {
	interactions map[string]*models.Interaction
	saveError    error
	findError    error
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		interactions: make(map[string]*models.Interaction),
	}
}

func (m *MockRepository) Save(interaction *models.Interaction) error {
	if m.saveError != nil {
		return m.saveError
	}
	hash := interaction.Request.GenerateHash()
	m.interactions[hash] = interaction
	return nil
}

func (m *MockRepository) Find(hash string) (*models.Interaction, error) {
	if m.findError != nil {
		return nil, m.findError
	}
	if interaction, ok := m.interactions[hash]; ok {
		return interaction, nil
	}
	return nil, storage.ErrNotFound{Hash: hash}
}

func (m *MockRepository) FindAll() ([]*models.Interaction, error) {
	var result []*models.Interaction
	for _, interaction := range m.interactions {
		result = append(result, interaction)
	}
	return result, nil
}

func (m *MockRepository) Clear() error {
	m.interactions = make(map[string]*models.Interaction)
	return nil
}

func (m *MockRepository) Count() (int, error) {
	return len(m.interactions), nil
}

func TestRecorder(t *testing.T) {
	// Create test server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Echo back request details
		response := map[string]interface{}{
			"method": r.Method,
			"path":   r.URL.Path,
			"query":  r.URL.RawQuery,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer testServer.Close()

	t.Run("Record GET request", func(t *testing.T) {
		repo := NewMockRepository()
		recorder := NewRecorder(repo)

		// Create test request
		req, _ := http.NewRequest("GET", "/api/test", nil)
		req.Header.Set("Accept", "application/json")

		// Handle request
		interaction, err := recorder.Handle(req, testServer.URL, nil)
		if err != nil {
			t.Fatalf("Failed to handle request: %v", err)
		}

		// Verify interaction was recorded
		if interaction == nil {
			t.Fatal("Interaction is nil")
		}
		if interaction.Request.Method != "GET" {
			t.Errorf("Expected method GET, got %s", interaction.Request.Method)
		}
		if interaction.Response.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", interaction.Response.StatusCode)
		}

		// Verify it was saved
		count, _ := repo.Count()
		if count != 1 {
			t.Errorf("Expected 1 saved interaction, got %d", count)
		}
	})

	t.Run("Record POST request with body", func(t *testing.T) {
		repo := NewMockRepository()
		recorder := NewRecorder(repo)

		// Create test request with body
		body := []byte(`{"name":"test"}`)
		req, _ := http.NewRequest("POST", "/api/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		// Handle request
		interaction, err := recorder.Handle(req, testServer.URL, body)
		if err != nil {
			t.Fatalf("Failed to handle request: %v", err)
		}

		// Verify body was recorded
		if string(interaction.Request.Body) != string(body) {
			t.Errorf("Body mismatch. Expected %s, got %s", body, interaction.Request.Body)
		}
	})

	t.Run("Handle save error gracefully", func(t *testing.T) {
		repo := NewMockRepository()
		repo.saveError = storage.ErrNotFound{Hash: "test"}
		recorder := NewRecorder(repo)

		req, _ := http.NewRequest("GET", "/api/test", nil)

		// Should not fail even if save fails
		interaction, err := recorder.Handle(req, testServer.URL, nil)
		if err != nil {
			t.Fatalf("Should not fail on save error: %v", err)
		}
		if interaction == nil {
			t.Fatal("Interaction should still be returned")
		}
	})
}

func TestPlayer(t *testing.T) {
	t.Run("Playback existing recording", func(t *testing.T) {
		repo := NewMockRepository()
		player := NewPlayer(repo)

		// Pre-save an interaction
		interaction := &models.Interaction{
			ID: "test-123",
			Request: models.RecordedRequest{
				Method: "GET",
				URL:    "/api/test",
				Headers: map[string][]string{
					"Accept": {"application/json"},
				},
			},
			Response: models.RecordedResponse{
				StatusCode: 200,
				Body:       json.RawMessage(`{"result":"success"}`),
			},
		}
		repo.Save(interaction)

		// Create matching request
		req, _ := http.NewRequest("GET", "/api/test", nil)
		req.Header.Set("Accept", "application/json")

		// Handle request
		found, err := player.Handle(req, nil)
		if err != nil {
			t.Fatalf("Failed to handle request: %v", err)
		}

		// Verify correct interaction was returned
		if found.ID != interaction.ID {
			t.Errorf("Wrong interaction returned. Expected ID %s, got %s", interaction.ID, found.ID)
		}
		if found.Response.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", found.Response.StatusCode)
		}
	})

	t.Run("Return error when no recording exists", func(t *testing.T) {
		repo := NewMockRepository()
		player := NewPlayer(repo)

		req, _ := http.NewRequest("GET", "/api/unknown", nil)

		// Handle request
		_, err := player.Handle(req, nil)
		if err == nil {
			t.Fatal("Expected error for non-existent recording")
		}

		// Check it's the right error type
		if _, ok := err.(*ErrNoRecording); !ok {
			t.Errorf("Expected ErrNoRecording, got %T", err)
		}
	})

	t.Run("Match based on full request", func(t *testing.T) {
		repo := NewMockRepository()
		player := NewPlayer(repo)

		// Save interaction with specific headers
		interaction := &models.Interaction{
			ID: "test-456",
			Request: models.RecordedRequest{
				Method: "POST",
				URL:    "/api/users",
				Headers: map[string][]string{
					"Content-Type": {"application/json"},
					"X-Tenant":     {"org-123"},
				},
				Body: json.RawMessage(`{"name":"Alice"}`),
			},
			Response: models.RecordedResponse{
				StatusCode: 201,
			},
		}
		repo.Save(interaction)

		// Request with matching everything
		body := []byte(`{"name":"Alice"}`)
		req1, _ := http.NewRequest("POST", "/api/users", bytes.NewReader(body))
		req1.Header.Set("Content-Type", "application/json")
		req1.Header.Set("X-Tenant", "org-123")

		found, err := player.Handle(req1, body)
		if err != nil {
			t.Errorf("Should find matching request: %v", err)
		}
		if found.ID != interaction.ID {
			t.Error("Wrong interaction returned")
		}

		// Request with different header
		req2, _ := http.NewRequest("POST", "/api/users", bytes.NewReader(body))
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set("X-Tenant", "org-456") // Different tenant

		_, err = player.Handle(req2, body)
		if err == nil {
			t.Error("Should not find request with different header")
		}

		// Request with different body
		body2 := []byte(`{"name":"Bob"}`)
		req3, _ := http.NewRequest("POST", "/api/users", bytes.NewReader(body2))
		req3.Header.Set("Content-Type", "application/json")
		req3.Header.Set("X-Tenant", "org-123")

		_, err = player.Handle(req3, body2)
		if err == nil {
			t.Error("Should not find request with different body")
		}
	})
}

func TestBuildTargetURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"api.example.com", "https://api.example.com"},
		{"http://api.example.com", "http://api.example.com"},
		{"https://api.example.com", "https://api.example.com"},
		{"0.0.0.0:8080", "https://0.0.0.0:8080"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := buildTargetURL(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}