package mode

import (
	"fmt"
	"net/http"

	"github.com/pismo/testing-proxy/internal/models"
	"github.com/pismo/testing-proxy/internal/storage"
)

// Player handles playback of recorded HTTP interactions
type Player struct {
	repository storage.Repository
}

// NewPlayer creates a new Player instance
func NewPlayer(repository storage.Repository) *Player {
	return &Player{
		repository: repository,
	}
}

// Handle processes a request in playback mode
func (r *Player) Handle(req *http.Request, target string, body []byte) (*models.Interaction, error) {
	// Create recorded request from incoming request
	recordedReq := models.FromHTTPRequest(req, body, target)

	// Generate hash for lookup
	hash := recordedReq.GenerateHash()

	// Find matching interaction
	interaction, err := r.repository.Find(hash)
	if err != nil {
		if _, ok := err.(storage.ErrNotFound); ok {
			return nil, &ErrNoRecording{
				Method: recordedReq.Method,
				URL:    recordedReq.URL,
				Hash:   hash,
			}
		}
		return nil, fmt.Errorf("failed to retrieve recording: %w", err)
	}

	return interaction, nil
}

// ErrNoRecording indicates that no recording was found for the request
type ErrNoRecording struct {
	Method string
	URL    string
	Hash   string
}

func (e *ErrNoRecording) Error() string {
	return fmt.Sprintf("no recording found for %s %s (hash: %s)", e.Method, e.URL, e.Hash)
}