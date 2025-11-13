package client

import (
	"encoding/json"
	"time"
)

// Response represents an HTTP response with timing information
type Response struct {
	StatusCode int
	Headers    map[string][]string
	Body       json.RawMessage
	Duration   time.Duration
	Error      error
}

// Request represents an HTTP request to be made
type Request struct {
	URL         string
	Method      string
	Headers     map[string]string
	QueryParams map[string]string
	Body        json.RawMessage
}
