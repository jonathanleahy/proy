package mode

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/pismo/testing-proxy/internal/models"
	"github.com/pismo/testing-proxy/internal/storage"
)

// Recorder handles recording of HTTP interactions
type Recorder struct {
	repository storage.Repository
	httpClient *http.Client
}

// NewRecorder creates a new Recorder instance
func NewRecorder(repository storage.Repository) *Recorder {
	// Create HTTP client that accepts any certificate (for testing)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return &Recorder{
		repository: repository,
		httpClient: &http.Client{
			Transport: tr,
			Timeout:   30 * time.Second, // Generous timeout for external services
		},
	}
}

// Handle processes a request in record mode
func (r *Recorder) Handle(req *http.Request, target string, body []byte) (*models.Interaction, error) {
	startTime := time.Now()

	// Create recorded request
	recordedReq := models.FromHTTPRequest(req, body, target)

	// Build the full target URL (adds https:// if needed)
	targetURL := buildTargetURL(target)

	// The target from Query().Get() is URL-decoded, which corrupts URLs with
	// special characters (spaces, etc). We need to parse and manually re-encode
	// the query parameters to preserve proper URL encoding.
	parsedTarget, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse target URL: %w", err)
	}

	// If there are query parameters, we need to re-encode them properly
	// because Query().Get() decoded them
	if parsedTarget.RawQuery != "" {
		// Parse the query parameters
		queryValues, err := url.ParseQuery(parsedTarget.RawQuery)
		if err != nil {
			return nil, fmt.Errorf("failed to parse query: %w", err)
		}
		// Re-encode them properly - this will convert spaces to %20
		parsedTarget.RawQuery = queryValues.Encode()
	}

	properlyEncodedURL := parsedTarget.String()

	// Create forward request using the properly encoded target URL
	forwardReq, err := http.NewRequest(recordedReq.Method, properlyEncodedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create forward request: %w", err)
	}

	// Copy headers from original request
	for k, values := range recordedReq.Headers {
		for _, v := range values {
			forwardReq.Header.Add(k, v)
		}
	}

	// Add body if present
	if body != nil && len(body) > 0 {
		forwardReq.Body = io.NopCloser(bytes.NewReader(body))
		forwardReq.ContentLength = int64(len(body))
	}

	// Execute the request
	resp, err := r.httpClient.Do(forwardReq)
	if err != nil {
		return nil, fmt.Errorf("failed to forward request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Create recorded response
	recordedResp := models.FromHTTPResponse(resp, respBody)

	// Calculate duration
	duration := time.Since(startTime).Milliseconds()

	// Create interaction
	interaction := &models.Interaction{
		ID:        uuid.New().String(),
		Timestamp: startTime,
		Request:   *recordedReq,
		Response:  *recordedResp,
		Metadata: models.InteractionMetadata{
			Target:     target,
			DurationMS: duration,
		},
	}

	// Save to repository
	if err := r.repository.Save(interaction); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: Failed to save interaction: %v\n", err)
	}

	return interaction, nil
}

// buildTargetURL constructs the full target URL
func buildTargetURL(target string) string {
	// Check if target already has protocol
	if !hasProtocol(target) {
		// Default to https for security
		return "https://" + target
	}
	return target
}

// hasProtocol checks if URL has http:// or https:// prefix
func hasProtocol(url string) bool {
	return len(url) >= 7 && (url[:7] == "http://" || (len(url) >= 8 && url[:8] == "https://"))
}
