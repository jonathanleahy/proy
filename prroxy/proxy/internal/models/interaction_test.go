package models

import (
	"net/http"
	"strings"
	"testing"
)

func TestGenerateHash(t *testing.T) {
	tests := []struct {
		name      string
		request1  RecordedRequest
		request2  RecordedRequest
		shouldMatch bool
	}{
		{
			name: "identical requests should have same hash",
			request1: RecordedRequest{
				Method: "GET",
				URL:    "/api/users",
				Headers: map[string][]string{
					"Content-Type": {"application/json"},
					"X-Tenant":     {"org-123"},
				},
				Body: []byte(`{"id":1}`),
			},
			request2: RecordedRequest{
				Method: "GET",
				URL:    "/api/users",
				Headers: map[string][]string{
					"Content-Type": {"application/json"},
					"X-Tenant":     {"org-123"},
				},
				Body: []byte(`{"id":1}`),
			},
			shouldMatch: true,
		},
		{
			name: "different methods should have different hash",
			request1: RecordedRequest{
				Method: "GET",
				URL:    "/api/users",
			},
			request2: RecordedRequest{
				Method: "POST",
				URL:    "/api/users",
			},
			shouldMatch: false,
		},
		{
			name: "different URLs should have different hash",
			request1: RecordedRequest{
				Method: "GET",
				URL:    "/api/users",
			},
			request2: RecordedRequest{
				Method: "GET",
				URL:    "/api/accounts",
			},
			shouldMatch: false,
		},
		{
			name: "different headers should have different hash",
			request1: RecordedRequest{
				Method: "GET",
				URL:    "/api/users",
				Headers: map[string][]string{
					"X-Tenant": {"org-123"},
				},
			},
			request2: RecordedRequest{
				Method: "GET",
				URL:    "/api/users",
				Headers: map[string][]string{
					"X-Tenant": {"org-456"},
				},
			},
			shouldMatch: false,
		},
		{
			name: "different bodies should have different hash",
			request1: RecordedRequest{
				Method: "POST",
				URL:    "/api/users",
				Body:   []byte(`{"name":"Alice"}`),
			},
			request2: RecordedRequest{
				Method: "POST",
				URL:    "/api/users",
				Body:   []byte(`{"name":"Bob"}`),
			},
			shouldMatch: false,
		},
		{
			name: "headers in different order should have same hash",
			request1: RecordedRequest{
				Method: "GET",
				URL:    "/api/users",
				Headers: map[string][]string{
					"X-Tenant":     {"org-123"},
					"Content-Type": {"application/json"},
				},
			},
			request2: RecordedRequest{
				Method: "GET",
				URL:    "/api/users",
				Headers: map[string][]string{
					"Content-Type": {"application/json"},
					"X-Tenant":     {"org-123"},
				},
			},
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := tt.request1.GenerateHash()
			hash2 := tt.request2.GenerateHash()

			if tt.shouldMatch && hash1 != hash2 {
				t.Errorf("Expected hashes to match, but they didn't.\nHash1: %s\nHash2: %s", hash1, hash2)
			}
			if !tt.shouldMatch && hash1 == hash2 {
				t.Errorf("Expected hashes to differ, but they matched.\nHash: %s", hash1)
			}
		})
	}
}

func TestToHTTPRequest(t *testing.T) {
	tests := []struct {
		name          string
		recorded      RecordedRequest
		targetURL     string
		expectedURL   string
		expectedError bool
	}{
		{
			name: "simple GET request",
			recorded: RecordedRequest{
				Method: "GET",
				URL:    "/api/users",
				Headers: map[string][]string{
					"Accept": {"application/json"},
				},
			},
			targetURL:   "https://api.example.com",
			expectedURL: "https://api.example.com/api/users",
		},
		{
			name: "POST request with body",
			recorded: RecordedRequest{
				Method: "POST",
				URL:    "/api/users",
				Headers: map[string][]string{
					"Content-Type": {"application/json"},
				},
				Body: []byte(`{"name":"Test"}`),
			},
			targetURL:   "https://api.example.com",
			expectedURL: "https://api.example.com/api/users",
		},
		{
			name: "request with query parameters",
			recorded: RecordedRequest{
				Method: "GET",
				URL:    "/api/users?page=1&limit=10",
			},
			targetURL:   "https://api.example.com",
			expectedURL: "https://api.example.com/api/users?page=1&limit=10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := tt.recorded.ToHTTPRequest(tt.targetURL)

			if tt.expectedError && err == nil {
				t.Error("Expected error but got none")
				return
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if err != nil {
				return
			}

			if req.URL.String() != tt.expectedURL {
				t.Errorf("URL mismatch.\nExpected: %s\nGot: %s", tt.expectedURL, req.URL.String())
			}

			if req.Method != tt.recorded.Method {
				t.Errorf("Method mismatch. Expected: %s, Got: %s", tt.recorded.Method, req.Method)
			}

			// Check headers
			for k, values := range tt.recorded.Headers {
				for _, v := range values {
					if !contains(req.Header.Values(k), v) {
						t.Errorf("Header %s missing value %s", k, v)
					}
				}
			}
		})
	}
}

func TestFromHTTPRequest(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *http.Request
		body     []byte
		expected RecordedRequest
	}{
		{
			name: "GET request without body",
			setup: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com/api/users", nil)
				req.Header.Set("Accept", "application/json")
				return req
			},
			body: nil,
			expected: RecordedRequest{
				Method: "GET",
				URL:    "http://example.com/api/users",
				Headers: map[string][]string{
					"Accept": {"application/json"},
				},
			},
		},
		{
			name: "POST request with body",
			setup: func() *http.Request {
				body := strings.NewReader(`{"name":"Test"}`)
				req, _ := http.NewRequest("POST", "http://example.com/api/users", body)
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			body: []byte(`{"name":"Test"}`),
			expected: RecordedRequest{
				Method: "POST",
				URL:    "http://example.com/api/users",
				Headers: map[string][]string{
					"Content-Type": {"application/json"},
				},
				Body: []byte(`{"name":"Test"}`),
			},
		},
		{
			name: "request with query parameters",
			setup: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com/api/users?page=1&limit=10", nil)
				return req
			},
			body: nil,
			expected: RecordedRequest{
				Method:  "GET",
				URL:     "http://example.com/api/users?page=1&limit=10",
				Headers: map[string][]string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setup()
			// Use the full URL from the request as the target for testing
			target := req.URL.String()
			recorded := FromHTTPRequest(req, tt.body, target)

			if recorded.Method != tt.expected.Method {
				t.Errorf("Method mismatch. Expected: %s, Got: %s", tt.expected.Method, recorded.Method)
			}

			if recorded.URL != tt.expected.URL {
				t.Errorf("URL mismatch. Expected: %s, Got: %s", tt.expected.URL, recorded.URL)
			}

			// Check body
			if string(recorded.Body) != string(tt.expected.Body) {
				t.Errorf("Body mismatch.\nExpected: %s\nGot: %s", tt.expected.Body, recorded.Body)
			}
		})
	}
}

func TestFromHTTPResponse(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *http.Response
		body     []byte
		expected RecordedResponse
	}{
		{
			name: "successful JSON response",
			setup: func() *http.Response {
				resp := &http.Response{
					StatusCode: 200,
					Header:     http.Header{},
				}
				resp.Header.Set("Content-Type", "application/json")
				return resp
			},
			body: []byte(`{"status":"ok"}`),
			expected: RecordedResponse{
				StatusCode: 200,
				Headers: map[string][]string{
					"Content-Type": {"application/json"},
				},
				Body: []byte(`{"status":"ok"}`),
			},
		},
		{
			name: "error response without body",
			setup: func() *http.Response {
				return &http.Response{
					StatusCode: 404,
					Header:     http.Header{},
				}
			},
			body: nil,
			expected: RecordedResponse{
				StatusCode: 404,
				Headers:    map[string][]string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := tt.setup()
			recorded := FromHTTPResponse(resp, tt.body)

			if recorded.StatusCode != tt.expected.StatusCode {
				t.Errorf("StatusCode mismatch. Expected: %d, Got: %d", tt.expected.StatusCode, recorded.StatusCode)
			}

			// Check body
			if string(recorded.Body) != string(tt.expected.Body) {
				t.Errorf("Body mismatch.\nExpected: %s\nGot: %s", tt.expected.Body, recorded.Body)
			}
		})
	}
}

// Helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}