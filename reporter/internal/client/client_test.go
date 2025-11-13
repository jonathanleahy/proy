package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Do(t *testing.T) {
	tests := []struct {
		name       string
		request    Request
		handler    http.HandlerFunc
		wantStatus int
		wantBody   string
		wantErr    bool
	}{
		{
			name: "successful GET request",
			request: Request{
				Method: "GET",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			},
			wantStatus: http.StatusOK,
			wantBody:   `{"status":"ok"}`,
			wantErr:    false,
		},
		{
			name: "POST request with body",
			request: Request{
				Method: "POST",
				Body:   json.RawMessage(`{"test":"data"}`),
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				var body map[string]string
				err := json.NewDecoder(r.Body).Decode(&body)
				require.NoError(t, err)
				assert.Equal(t, "data", body["test"])

				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(map[string]string{"created": "true"})
			},
			wantStatus: http.StatusCreated,
			wantBody:   `{"created":"true"}`,
			wantErr:    false,
		},
		{
			name: "request with custom headers",
			request: Request{
				Method: "GET",
				Headers: map[string]string{
					"Authorization": "Bearer token123",
					"X-Custom":      "header",
				},
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "Bearer token123", r.Header.Get("Authorization"))
				assert.Equal(t, "header", r.Header.Get("X-Custom"))
				w.WriteHeader(http.StatusOK)
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "request with query parameters",
			request: Request{
				Method: "GET",
				QueryParams: map[string]string{
					"foo": "bar",
					"baz": "qux",
				},
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "bar", r.URL.Query().Get("foo"))
				assert.Equal(t, "qux", r.URL.Query().Get("baz"))
				w.WriteHeader(http.StatusOK)
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "handles 404 response",
			request: Request{
				Method: "GET",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
			},
			wantStatus: http.StatusNotFound,
			wantBody:   `{"error":"not found"}`,
			wantErr:    false,
		},
		{
			name: "handles 500 response",
			request: Request{
				Method: "GET",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "server error"})
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":"server error"}`,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			// Set URL on request
			tt.request.URL = server.URL

			// Create client and execute request
			client := NewClient(5 * time.Second)
			resp := client.Do(tt.request)

			// Assertions
			if tt.wantErr {
				assert.Error(t, resp.Error)
				return
			}

			require.NoError(t, resp.Error)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)

			if tt.wantBody != "" {
				assert.JSONEq(t, tt.wantBody, string(resp.Body))
			}

			// Check timing is recorded
			assert.Greater(t, resp.Duration, time.Duration(0))
			assert.Less(t, resp.Duration, 5*time.Second)
		})
	}
}

func TestClient_Do_Timeout(t *testing.T) {
	// Create server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create client with very short timeout
	client := NewClient(10 * time.Millisecond)
	resp := client.Do(Request{
		URL:    server.URL,
		Method: "GET",
	})

	// Should have timeout error
	assert.Error(t, resp.Error)
	assert.Contains(t, resp.Error.Error(), "deadline exceeded")
}

func TestClient_Do_InvalidURL(t *testing.T) {
	client := NewClient(5 * time.Second)
	resp := client.Do(Request{
		URL:    "://invalid-url",
		Method: "GET",
	})

	assert.Error(t, resp.Error)
}

func TestClient_Do_ServerDown(t *testing.T) {
	client := NewClient(5 * time.Second)
	resp := client.Do(Request{
		URL:    "http://0.0.0.0:99999",
		Method: "GET",
	})

	assert.Error(t, resp.Error)
}

func TestClient_Do_RecordsTiming(t *testing.T) {
	// Create server with known delay
	delay := 50 * time.Millisecond
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(5 * time.Second)
	resp := client.Do(Request{
		URL:    server.URL,
		Method: "GET",
	})

	require.NoError(t, resp.Error)

	// Duration should be at least the delay time
	assert.GreaterOrEqual(t, resp.Duration, delay)

	// But not too much more (accounting for overhead)
	assert.Less(t, resp.Duration, delay+100*time.Millisecond)
}
