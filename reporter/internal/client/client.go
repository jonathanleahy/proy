package client

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client is an HTTP client that records timing information
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new Client with the specified timeout
func NewClient(timeout time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Do executes an HTTP request and records timing information
func (c *Client) Do(req Request) Response {
	start := time.Now()

	resp := Response{}

	// Build URL with query parameters
	fullURL, err := buildURL(req.URL, req.QueryParams)
	if err != nil {
		resp.Error = fmt.Errorf("invalid URL: %w", err)
		resp.Duration = time.Since(start)
		return resp
	}

	// Create HTTP request
	var bodyReader io.Reader
	if len(req.Body) > 0 {
		bodyReader = bytes.NewReader(req.Body)
	}

	httpReq, err := http.NewRequest(req.Method, fullURL, bodyReader)
	if err != nil {
		resp.Error = fmt.Errorf("failed to create request: %w", err)
		resp.Duration = time.Since(start)
		return resp
	}

	// Add headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Execute request
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		resp.Error = err
		resp.Duration = time.Since(start)
		return resp
	}
	defer httpResp.Body.Close()

	// Read response body
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		resp.Error = fmt.Errorf("failed to read response body: %w", err)
		resp.Duration = time.Since(start)
		return resp
	}

	// Record response
	resp.StatusCode = httpResp.StatusCode
	resp.Headers = httpResp.Header
	resp.Body = body
	resp.Duration = time.Since(start)

	return resp
}

// buildURL constructs a full URL with query parameters
func buildURL(baseURL string, queryParams map[string]string) (string, error) {
	if len(queryParams) == 0 {
		return baseURL, nil
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	q := parsedURL.Query()
	for key, value := range queryParams {
		q.Set(key, value)
	}
	parsedURL.RawQuery = q.Encode()

	return parsedURL.String(), nil
}
