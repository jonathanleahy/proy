package storage

import (
	"github.com/pismo/testing-proxy/internal/models"
)

// Repository defines the interface for storing and retrieving interactions
type Repository interface {
	// Save stores an interaction
	Save(interaction *models.Interaction) error

	// Find retrieves an interaction by request hash
	Find(hash string) (*models.Interaction, error)

	// FindAll returns all stored interactions
	FindAll() ([]*models.Interaction, error)

	// Clear removes all stored interactions
	Clear() error

	// Count returns the number of stored interactions
	Count() (int, error)
}

// ErrNotFound is returned when an interaction is not found
type ErrNotFound struct {
	Hash string
}

func (e ErrNotFound) Error() string {
	return "interaction not found for hash: " + e.Hash
}