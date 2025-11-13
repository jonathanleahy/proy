package models

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Interaction represents a recorded HTTP request/response pair
type Interaction struct {
	ID        string            `json:"id"`
	Timestamp time.Time         `json:"timestamp"`
	Request   RecordedRequest   `json:"request"`
	Response  RecordedResponse  `json:"response"`
	Metadata  InteractionMetadata `json:"metadata"`
}

// RecordedRequest contains the recorded request details
type RecordedRequest struct {
	Method  string              `json:"method"`
	URL     string              `json:"url"`
	Headers map[string][]string `json:"headers"`
	Body    []byte              `json:"body,omitempty"`
}

// RecordedResponse contains the recorded response details
type RecordedResponse struct {
	StatusCode int                 `json:"status_code"`
	Headers    map[string][]string `json:"headers"`
	Body       []byte              `json:"body,omitempty"`
}

// InteractionMetadata contains additional information about the interaction
type InteractionMetadata struct {
	Target     string `json:"target"`
	DurationMS int64  `json:"duration_ms"`
}

// GenerateHash creates a unique hash for request matching
// Uses simplified match strategy: URL + Method + Body (excludes headers)
// This allows different HTTP clients to match the same recordings
func (r *RecordedRequest) GenerateHash() string {
	h := sha256.New()

	// Add method and URL
	h.Write([]byte(r.Method))
	h.Write([]byte(r.URL))

	// NOTE: Headers are intentionally excluded from hashing
	// Different HTTP clients send different auto-generated headers
	// (Accept, User-Agent, Accept-Encoding, etc.), so we only match
	// on method, URL, and body for maximum compatibility

	// Add body if present
	if r.Body != nil {
		h.Write(r.Body)
	}

	return hex.EncodeToString(h.Sum(nil))
}

// ToHTTPRequest converts RecordedRequest to http.Request for forwarding
func (r *RecordedRequest) ToHTTPRequest(targetURL string) (*http.Request, error) {
	// Build the full URL
	fullURL := targetURL + r.URL

	// Create request with body if present
	var bodyReader strings.Reader
	if r.Body != nil {
		bodyReader = *strings.NewReader(string(r.Body))
	}

	req, err := http.NewRequest(r.Method, fullURL, &bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Copy headers
	for k, values := range r.Headers {
		for _, v := range values {
			req.Header.Add(k, v)
		}
	}

	return req, nil
}

// FromHTTPRequest creates a RecordedRequest from an http.Request
func FromHTTPRequest(req *http.Request, body []byte, target string) *RecordedRequest {
	recorded := &RecordedRequest{
		Method:  req.Method,
		URL:     target, // Store the full target URL
		Headers: make(map[string][]string),
	}

	// Copy headers
	for k, v := range req.Header {
		recorded.Headers[k] = v
	}

	// Set body if present
	if body != nil && len(body) > 0 {
		recorded.Body = body
	}

	return recorded
}

// FromHTTPResponse creates a RecordedResponse from an http.Response
func FromHTTPResponse(resp *http.Response, body []byte) *RecordedResponse {
	recorded := &RecordedResponse{
		StatusCode: resp.StatusCode,
		Headers:    make(map[string][]string),
	}

	// Copy headers
	for k, v := range resp.Header {
		recorded.Headers[k] = v
	}

	// Set body if present
	if body != nil && len(body) > 0 {
		recorded.Body = body
	}

	return recorded
}