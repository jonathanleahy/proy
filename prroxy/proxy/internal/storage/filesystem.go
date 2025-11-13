package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/pismo/testing-proxy/internal/models"
)

// FileSystemRepository implements Repository using filesystem storage
type FileSystemRepository struct {
	basePath string
	mu       sync.RWMutex // Ensures thread-safe operations
}

// NewFileSystemRepository creates a new filesystem-based repository
func NewFileSystemRepository(basePath string) (*FileSystemRepository, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &FileSystemRepository{
		basePath: basePath,
	}, nil
}

// Save stores an interaction to the filesystem
func (r *FileSystemRepository) Save(interaction *models.Interaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Generate hash for the request
	hash := interaction.Request.GenerateHash()

	// Extract service name from target URL for organization
	serviceName := extractServiceName(interaction.Metadata.Target)

	// Create service directory if it doesn't exist
	serviceDir := filepath.Join(r.basePath, serviceName)
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		return fmt.Errorf("failed to create service directory: %w", err)
	}

	// Save interaction as JSON file
	filename := filepath.Join(serviceDir, hash+".json")

	// Marshal interaction to JSON
	data, err := json.MarshalIndent(interaction, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal interaction: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write interaction file: %w", err)
	}

	return nil
}

// Find retrieves an interaction by request hash
func (r *FileSystemRepository) Find(hash string) (*models.Interaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Search for the file in all service directories
	pattern := filepath.Join(r.basePath, "*", hash+".json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search for interaction: %w", err)
	}

	if len(matches) == 0 {
		return nil, ErrNotFound{Hash: hash}
	}

	// Read the first match (there should only be one)
	data, err := os.ReadFile(matches[0])
	if err != nil {
		return nil, fmt.Errorf("failed to read interaction file: %w", err)
	}

	// Unmarshal JSON
	var interaction models.Interaction
	if err := json.Unmarshal(data, &interaction); err != nil {
		return nil, fmt.Errorf("failed to unmarshal interaction: %w", err)
	}

	return &interaction, nil
}

// FindAll returns all stored interactions
func (r *FileSystemRepository) FindAll() ([]*models.Interaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var interactions []*models.Interaction

	// Walk through all JSON files in the directory
	err := filepath.Walk(r.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-JSON files
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".json") {
			return nil
		}

		// Read and unmarshal the file
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		var interaction models.Interaction
		if err := json.Unmarshal(data, &interaction); err != nil {
			return fmt.Errorf("failed to unmarshal file %s: %w", path, err)
		}

		interactions = append(interactions, &interaction)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	// Sort by timestamp descending (newest first)
	sort.Slice(interactions, func(i, j int) bool {
		return interactions[i].Timestamp.After(interactions[j].Timestamp)
	})

	return interactions, nil
}

// Clear removes all stored interactions
func (r *FileSystemRepository) Clear() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Remove all subdirectories
	entries, err := os.ReadDir(r.basePath)
	if err != nil {
		return fmt.Errorf("failed to read directory entries: %w", err)
	}

	for _, entry := range entries {
		path := filepath.Join(r.basePath, entry.Name())
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("failed to remove %s: %w", path, err)
		}
	}

	return nil
}

// Count returns the number of stored interactions
func (r *FileSystemRepository) Count() (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0

	// Count all JSON files
	err := filepath.Walk(r.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".json") {
			count++
		}

		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("failed to count interactions: %w", err)
	}

	return count, nil
}

// extractServiceName extracts a clean service name from a target URL
func extractServiceName(target string) string {
	// Remove protocol if present
	target = strings.TrimPrefix(target, "http://")
	target = strings.TrimPrefix(target, "https://")

	// Take the first part (domain/service name)
	parts := strings.Split(target, "/")
	if len(parts) > 0 && parts[0] != "" {
		// Replace dots with underscores for filesystem compatibility
		service := strings.ReplaceAll(parts[0], ".", "_")
		// Replace colons (port numbers) with underscores
		service = strings.ReplaceAll(service, ":", "_")
		return service
	}

	return "unknown"
}