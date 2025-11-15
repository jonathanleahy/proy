// Package httpclient provides a simple HTTP client wrapper.
// Proxy configuration is handled externally via source code modification.
package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"
)

const (
	// DefaultTimeout is the default HTTP request timeout
	DefaultTimeout = 10 * time.Second
)

// Client is a simple HTTP client wrapper.
type Client struct {
	httpClient *http.Client
}

// New creates a new HTTP client.
// If timeout is 0, DefaultTimeout is used.
func New(timeout time.Duration) *Client {
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Get performs a GET request to the target URL.
// Sets Accept-Encoding: identity to disable compression for recording compatibility.
func (c *Client) Get(ctx context.Context, target string) (*http.Response, error) {
	// Create request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers for proxy compatibility
	req.Header.Set("Accept-Encoding", "identity")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// Post performs a POST request to the target URL with JSON body.
// Sets Content-Type: application/json and Accept-Encoding: identity headers.
func (c *Client) Post(ctx context.Context, target string, body []byte) (*http.Response, error) {
	// Create request with context and body
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", "identity")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}
