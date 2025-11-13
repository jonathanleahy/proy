package reporter

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jonathanleahy/prroxy/reporter/internal/comparer"
	"github.com/jonathanleahy/prroxy/reporter/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewReporter tests the NewReporter constructor
func TestNewReporter(t *testing.T) {
	cfg := &config.Config{
		BaseURLV1:  "http://0.0.0.0:3000",
		BaseURLV2:  "http://0.0.0.0:8080",
		Iterations: 5,
		Endpoints: []config.Endpoint{
			{Path: "/api/test", Method: "GET"},
		},
	}

	r := NewReporter(cfg)

	assert.NotNil(t, r)
	assert.Equal(t, cfg, r.config)
	assert.NotNil(t, r.client)
	assert.NotNil(t, r.comparer)
}

// TestRun_Success tests successful endpoint comparison
func TestRun_Success(t *testing.T) {
	// Create mock servers for v1 and v2
	v1Server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   1,
			"name": "Test User",
		})
	}))
	defer v1Server.Close()

	v2Server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   1,
			"name": "Test User",
		})
	}))
	defer v2Server.Close()

	cfg := &config.Config{
		BaseURLV1:  v1Server.URL,
		BaseURLV2:  v2Server.URL,
		Iterations: 2,
		Endpoints: []config.Endpoint{
			{Path: "/api/user", Method: "GET"},
		},
	}

	r := NewReporter(cfg)
	report := r.Run()

	assert.Equal(t, 1, report.TotalEndpoints)
	assert.Equal(t, 1, report.MatchedEndpoints)
	assert.Equal(t, 0, report.FailedEndpoints)
	assert.Greater(t, report.TotalDuration, time.Duration(0))
	assert.Len(t, report.Endpoints, 1)
	assert.True(t, report.Endpoints[0].Match)
}

// TestRun_Mismatch tests endpoint comparison with differences
func TestRun_Mismatch(t *testing.T) {
	v1Server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":    1,
			"value": "old",
		})
	}))
	defer v1Server.Close()

	v2Server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":    1,
			"value": "new",
		})
	}))
	defer v2Server.Close()

	cfg := &config.Config{
		BaseURLV1:  v1Server.URL,
		BaseURLV2:  v2Server.URL,
		Iterations: 1,
		Endpoints: []config.Endpoint{
			{Path: "/api/data", Method: "GET"},
		},
	}

	r := NewReporter(cfg)
	report := r.Run()

	assert.Equal(t, 1, report.TotalEndpoints)
	assert.Equal(t, 0, report.MatchedEndpoints)
	assert.Equal(t, 1, report.FailedEndpoints)
	assert.Len(t, report.Endpoints, 1)
	assert.False(t, report.Endpoints[0].Match)
	assert.NotEmpty(t, report.Endpoints[0].Differences)
}

// TestRun_V1Error tests error handling when V1 fails
func TestRun_V1Error(t *testing.T) {
	v2Server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"})
	}))
	defer v2Server.Close()

	cfg := &config.Config{
		BaseURLV1:  "http://0.0.0.0:9999", // Non-existent server
		BaseURLV2:  v2Server.URL,
		Iterations: 1,
		Endpoints: []config.Endpoint{
			{Path: "/api/test", Method: "GET"},
		},
	}

	r := NewReporter(cfg)
	report := r.Run()

	assert.Equal(t, 1, report.TotalEndpoints)
	assert.Equal(t, 0, report.MatchedEndpoints)
	assert.Equal(t, 1, report.FailedEndpoints)
	assert.Contains(t, report.Endpoints[0].Error, "V1 error")
}

// TestRun_V2Error tests error handling when V2 fails
func TestRun_V2Error(t *testing.T) {
	v1Server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"})
	}))
	defer v1Server.Close()

	cfg := &config.Config{
		BaseURLV1:  v1Server.URL,
		BaseURLV2:  "http://0.0.0.0:9999", // Non-existent server
		Iterations: 1,
		Endpoints: []config.Endpoint{
			{Path: "/api/test", Method: "GET"},
		},
	}

	r := NewReporter(cfg)
	report := r.Run()

	assert.Equal(t, 1, report.TotalEndpoints)
	assert.Equal(t, 0, report.MatchedEndpoints)
	assert.Equal(t, 1, report.FailedEndpoints)
	assert.Contains(t, report.Endpoints[0].Error, "V2 error")
}

// TestRun_MultipleIterations tests multiple iterations
func TestRun_MultipleIterations(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"count": callCount})
	}))
	defer server.Close()

	cfg := &config.Config{
		BaseURLV1:    server.URL,
		BaseURLV2:    server.URL,
		Iterations:   3,
		IgnoreFields: []string{"count"}, // Ignore count to avoid mismatch
		Endpoints: []config.Endpoint{
			{Path: "/api/test", Method: "GET"},
		},
	}

	r := NewReporter(cfg)
	report := r.Run()

	assert.Equal(t, 1, report.TotalEndpoints)
	assert.Len(t, report.Endpoints[0].V1Timings, 3)
	assert.Len(t, report.Endpoints[0].V2Timings, 3)
	assert.Greater(t, report.Endpoints[0].V1AvgTime, time.Duration(0))
	assert.Greater(t, report.Endpoints[0].V2AvgTime, time.Duration(0))
}

// TestRun_POSTWithBody tests POST requests with body
func TestRun_POSTWithBody(t *testing.T) {
	receivedBody := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			receivedBody = fmt.Sprintf("%v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}))
	defer server.Close()

	bodyData := map[string]interface{}{
		"name": "test",
		"age":  30,
	}
	bodyJSON, _ := json.Marshal(bodyData)

	cfg := &config.Config{
		BaseURLV1:  server.URL,
		BaseURLV2:  server.URL,
		Iterations: 1,
		Endpoints: []config.Endpoint{
			{
				Path:   "/api/create",
				Method: "POST",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: bodyJSON,
			},
		},
	}

	r := NewReporter(cfg)
	report := r.Run()

	assert.Equal(t, 1, report.TotalEndpoints)
	assert.Equal(t, 1, report.MatchedEndpoints)
	assert.NotEmpty(t, receivedBody)
}

// TestRun_WithHeaders tests requests with custom headers
func TestRun_WithHeaders(t *testing.T) {
	receivedHeaders := make(map[string]string)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders["Authorization"] = r.Header.Get("Authorization")
		receivedHeaders["X-Custom"] = r.Header.Get("X-Custom")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"})
	}))
	defer server.Close()

	cfg := &config.Config{
		BaseURLV1:  server.URL,
		BaseURLV2:  server.URL,
		Iterations: 1,
		Endpoints: []config.Endpoint{
			{
				Path:   "/api/secure",
				Method: "GET",
				Headers: map[string]string{
					"Authorization": "Bearer token123",
					"X-Custom":      "custom-value",
				},
			},
		},
	}

	r := NewReporter(cfg)
	report := r.Run()

	assert.Equal(t, 1, report.TotalEndpoints)
	assert.Equal(t, "Bearer token123", receivedHeaders["Authorization"])
	assert.Equal(t, "custom-value", receivedHeaders["X-Custom"])
}

// TestRun_WithQueryParams tests requests with query parameters
func TestRun_WithQueryParams(t *testing.T) {
	receivedParams := make(map[string]string)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedParams["filter"] = r.URL.Query().Get("filter")
		receivedParams["sort"] = r.URL.Query().Get("sort")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"})
	}))
	defer server.Close()

	cfg := &config.Config{
		BaseURLV1:  server.URL,
		BaseURLV2:  server.URL,
		Iterations: 1,
		Endpoints: []config.Endpoint{
			{
				Path:   "/api/items",
				Method: "GET",
				QueryParams: map[string]string{
					"filter": "active",
					"sort":   "name",
				},
			},
		},
	}

	r := NewReporter(cfg)
	report := r.Run()

	assert.Equal(t, 1, report.TotalEndpoints)
	assert.Equal(t, "active", receivedParams["filter"])
	assert.Equal(t, "name", receivedParams["sort"])
}

// TestAverage tests the average function
func TestAverage(t *testing.T) {
	tests := []struct {
		name     string
		input    []time.Duration
		expected time.Duration
	}{
		{
			name:     "empty slice",
			input:    []time.Duration{},
			expected: 0,
		},
		{
			name:     "single value",
			input:    []time.Duration{100 * time.Millisecond},
			expected: 100 * time.Millisecond,
		},
		{
			name:     "multiple values",
			input:    []time.Duration{100 * time.Millisecond, 200 * time.Millisecond, 300 * time.Millisecond},
			expected: 200 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := average(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFormatJSON tests JSON formatting
func TestFormatJSON(t *testing.T) {
	report := Report{
		TotalEndpoints:   2,
		MatchedEndpoints: 1,
		FailedEndpoints:  1,
		TotalDuration:    1500 * time.Millisecond,
		Endpoints: []EndpointReport{
			{
				Path:         "/api/user/1",
				Method:       "GET",
				Match:        true,
				V1AvgTime:    100 * time.Millisecond,
				V2AvgTime:    90 * time.Millisecond,
				StatusCodeV1: 200,
				StatusCodeV2: 200,
			},
		},
	}

	output, err := FormatJSON(report)
	require.NoError(t, err)

	// Verify it's valid JSON
	var parsed map[string]interface{}
	err = json.Unmarshal([]byte(output), &parsed)
	require.NoError(t, err)

	assert.Equal(t, float64(2), parsed["TotalEndpoints"])
	assert.Equal(t, float64(1), parsed["MatchedEndpoints"])
	assert.Equal(t, float64(1), parsed["FailedEndpoints"])
}

// TestFormatMarkdown tests Markdown formatting
func TestFormatMarkdown(t *testing.T) {
	report := Report{
		TotalEndpoints:   2,
		MatchedEndpoints: 1,
		FailedEndpoints:  1,
		TotalDuration:    1500 * time.Millisecond,
		Endpoints: []EndpointReport{
			{
				Path:         "/api/user/1",
				Method:       "GET",
				Match:        true,
				V1AvgTime:    100 * time.Millisecond,
				V2AvgTime:    90 * time.Millisecond,
				StatusCodeV1: 200,
				StatusCodeV2: 200,
			},
			{
				Path:         "/api/user/2",
				Method:       "POST",
				Match:        false,
				V1AvgTime:    150 * time.Millisecond,
				V2AvgTime:    140 * time.Millisecond,
				StatusCodeV1: 200,
				StatusCodeV2: 200,
				Differences: []comparer.Difference{
					{
						Path:     "name",
						Value1:   "old",
						Value2:   "new",
						DiffType: comparer.DiffTypeValueMismatch,
					},
				},
			},
		},
	}

	output := FormatMarkdown(report)

	// Verify markdown structure
	assert.Contains(t, output, "# API Comparison Report")
	assert.Contains(t, output, "**Total Endpoints**: 2")
	assert.Contains(t, output, "**Matched**: 1")
	assert.Contains(t, output, "**Failed**: 1")
	assert.Contains(t, output, "### GET /api/user/1")
	assert.Contains(t, output, "### POST /api/user/2")
	assert.Contains(t, output, "MATCH")
	assert.Contains(t, output, "MISMATCH")
	assert.Contains(t, output, "**Differences**:")
}

// TestEndpointReport_Summary tests the Summary method
func TestEndpointReport_Summary(t *testing.T) {
	tests := []struct {
		name     string
		report   EndpointReport
		expected string
	}{
		{
			name: "match",
			report: EndpointReport{
				Match: true,
			},
			expected: "MATCH",
		},
		{
			name: "mismatch",
			report: EndpointReport{
				Match: false,
			},
			expected: "MISMATCH",
		},
		{
			name: "error",
			report: EndpointReport{
				Error: "connection failed",
			},
			expected: "ERROR: connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.report.Summary()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestRun_MultipleEndpoints tests running multiple endpoints
func TestRun_MultipleEndpoints(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Return different data based on path
		if strings.Contains(r.URL.Path, "user") {
			json.NewEncoder(w).Encode(map[string]interface{}{"type": "user"})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{"type": "post"})
		}
	}))
	defer server.Close()

	cfg := &config.Config{
		BaseURLV1:  server.URL,
		BaseURLV2:  server.URL,
		Iterations: 1,
		Endpoints: []config.Endpoint{
			{Path: "/api/user/1", Method: "GET"},
			{Path: "/api/post/1", Method: "GET"},
		},
	}

	r := NewReporter(cfg)
	report := r.Run()

	assert.Equal(t, 2, report.TotalEndpoints)
	assert.Equal(t, 2, report.MatchedEndpoints)
	assert.Equal(t, 0, report.FailedEndpoints)
	assert.Len(t, report.Endpoints, 2)
}
