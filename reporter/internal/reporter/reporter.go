package reporter

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jonathanleahy/prroxy/reporter/internal/client"
	"github.com/jonathanleahy/prroxy/reporter/internal/comparer"
	"github.com/jonathanleahy/prroxy/reporter/internal/config"
)

// Reporter runs endpoint tests and generates reports
type Reporter struct {
	config      *config.Config
	client      *client.Client
	comparer    *comparer.Comparer
	outputDir   string
	maxFailures int
}

// NewReporter creates a new Reporter
func NewReporter(cfg *config.Config, outputDir string, maxFailures int) *Reporter {
	return &Reporter{
		config:      cfg,
		client:      client.NewClient(30 * time.Second),
		comparer:    comparer.NewComparer(cfg.IgnoreFields),
		outputDir:   outputDir,
		maxFailures: maxFailures,
	}
}

// Run executes all endpoint tests
func (r *Reporter) Run() Report {
	report := Report{
		TotalEndpoints: len(r.config.Endpoints),
		Endpoints:      make([]EndpointReport, 0, len(r.config.Endpoints)),
	}

	// Create output directories if outputDir is specified
	if r.outputDir != "" {
		matchesDir := r.outputDir + "/matches"
		mismatchesDir := r.outputDir + "/mismatches"

		if err := os.MkdirAll(matchesDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating matches directory: %v\n", err)
		}
		if err := os.MkdirAll(mismatchesDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating mismatches directory: %v\n", err)
		}
	}

	start := time.Now()

	for i, endpoint := range r.config.Endpoints {
		epReport := r.testEndpoint(endpoint)
		report.Endpoints = append(report.Endpoints, epReport)

		if epReport.Match {
			report.MatchedEndpoints++
		} else {
			report.FailedEndpoints++
		}

		// Write individual files if outputDir is specified
		if r.outputDir != "" {
			r.writeIndividualReports(i+1, epReport)
		}

		// Print progress after each test
		status := "✓"
		if !epReport.Match || epReport.Error != "" {
			status = "✗"
		}

		fmt.Fprintf(os.Stderr, "[%d/%d] %s %s %s (V1: %v, V2: %v) [%d failures so far]\n",
			i+1,
			len(r.config.Endpoints),
			status,
			epReport.Method,
			epReport.Path,
			epReport.V1AvgTime.Round(time.Millisecond),
			epReport.V2AvgTime.Round(time.Millisecond),
			report.FailedEndpoints)

		// Stop if we've reached max failures
		if r.maxFailures > 0 && report.FailedEndpoints >= r.maxFailures {
			fmt.Fprintf(os.Stderr, "\nStopping after %d failures\n", report.FailedEndpoints)
			break
		}
	}

	report.TotalDuration = time.Since(start)

	return report
}

// testEndpoint tests a single endpoint with retries on mismatch
func (r *Reporter) testEndpoint(endpoint config.Endpoint) EndpointReport {
	const maxRetries = 5

	for attempt := 1; attempt <= maxRetries; attempt++ {
		epReport := r.runSingleTest(endpoint)

		// If test passed or had an error (not just mismatch), return
		if epReport.Match || epReport.Error != "" {
			if attempt > 1 {
				fmt.Fprintf(os.Stderr, "   ✓ Matched on attempt %d/%d\n", attempt, maxRetries)
			}
			epReport.Retries = attempt - 1
			return epReport
		}

		// Test failed (mismatch) - retry if we haven't reached max
		if attempt < maxRetries {
			fmt.Fprintf(os.Stderr, "   ⟳ Mismatch on attempt %d/%d, retrying...\n", attempt, maxRetries)
			time.Sleep(100 * time.Millisecond) // Small delay between retries
		} else {
			fmt.Fprintf(os.Stderr, "   ✗ Still mismatched after %d attempts\n", maxRetries)
			epReport.Retries = maxRetries - 1
			return epReport
		}
	}

	// Should never reach here
	return EndpointReport{Path: endpoint.Path, Method: endpoint.Method, Error: "Max retries exceeded"}
}

// runSingleTest executes a single test attempt for an endpoint
func (r *Reporter) runSingleTest(endpoint config.Endpoint) EndpointReport {
	epReport := EndpointReport{
		Path:       endpoint.Path,
		Method:     endpoint.Method,
		Iterations: r.config.Iterations,
		V1Timings:  make([]time.Duration, 0, r.config.Iterations),
		V2Timings:  make([]time.Duration, 0, r.config.Iterations),
	}

	var v1RespBody, v2RespBody []byte

	// Run iterations
	for i := 0; i < r.config.Iterations; i++ {
		// Call V1
		v1Req := r.buildRequest(r.config.BaseURLV1, endpoint)
		v1Resp := r.client.Do(v1Req)
		if v1Resp.Error != nil {
			epReport.Error = fmt.Sprintf("V1 error: %v", v1Resp.Error)
			return epReport
		}
		epReport.V1Timings = append(epReport.V1Timings, v1Resp.Duration)
		epReport.StatusCodeV1 = v1Resp.StatusCode
		if i == 0 {
			v1RespBody = v1Resp.Body
			epReport.V1Request = v1Req
		}

		// Call V2
		v2Req := r.buildRequest(r.config.BaseURLV2, endpoint)
		v2Resp := r.client.Do(v2Req)
		if v2Resp.Error != nil {
			epReport.Error = fmt.Sprintf("V2 error: %v", v2Resp.Error)
			return epReport
		}
		epReport.V2Timings = append(epReport.V2Timings, v2Resp.Duration)
		epReport.StatusCodeV2 = v2Resp.StatusCode
		if i == 0 {
			v2RespBody = v2Resp.Body
			epReport.V2Request = v2Req
		}

		// Compare responses (only on first iteration to avoid noise)
		if i == 0 {
			comparison := r.comparer.Compare(comparer.ResponsePair{
				V1: v1Resp.Body,
				V2: v2Resp.Body,
			})
			epReport.Match = comparison.Match
			epReport.Differences = comparison.Differences
			epReport.V1ResponseBody = v1RespBody
			epReport.V2ResponseBody = v2RespBody
		}
	}

	// Calculate averages
	epReport.V1AvgTime = average(epReport.V1Timings)
	epReport.V2AvgTime = average(epReport.V2Timings)

	return epReport
}

// buildRequest constructs an HTTP request from endpoint config
func (r *Reporter) buildRequest(baseURL string, endpoint config.Endpoint) client.Request {
	return client.Request{
		URL:         baseURL + endpoint.Path,
		Method:      endpoint.Method,
		Headers:     endpoint.Headers,
		QueryParams: endpoint.QueryParams,
		Body:        endpoint.Body,
	}
}

// average calculates the average duration
func average(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}

	var total time.Duration
	for _, d := range durations {
		total += d
	}

	return total / time.Duration(len(durations))
}

// FormatJSON formats the report as JSON
func FormatJSON(report Report) (string, error) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FormatMarkdown formats the report as Markdown
func FormatMarkdown(report Report) string {
	var md strings.Builder

	// Header
	md.WriteString("# API Comparison Report\n\n")
	md.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	// Summary
	md.WriteString(fmt.Sprintf("- Total: %d endpoints\n", report.TotalEndpoints))
	md.WriteString(fmt.Sprintf("- Passing: %d (%.0f%%)\n",
		report.MatchedEndpoints,
		float64(report.MatchedEndpoints)/float64(report.TotalEndpoints)*100))
	md.WriteString(fmt.Sprintf("- Failing: %d (%.0f%%)\n",
		report.FailedEndpoints,
		float64(report.FailedEndpoints)/float64(report.TotalEndpoints)*100))
	md.WriteString(fmt.Sprintf("- Duration: %v\n\n", report.TotalDuration))

	// Separate passing and failing
	var passing, failing []EndpointReport
	for _, ep := range report.Endpoints {
		if ep.Match {
			passing = append(passing, ep)
		} else {
			failing = append(failing, ep)
		}
	}

	// Failing endpoints
	if len(failing) > 0 {
		md.WriteString(fmt.Sprintf("## Failing Endpoints (%d)\n\n", len(failing)))

		for i, ep := range failing {
			md.WriteString(fmt.Sprintf("### %d. %s %s\n\n", i+1, ep.Method, ep.Path))
			md.WriteString(fmt.Sprintf("- Status: %s\n", ep.Summary()))
			md.WriteString(fmt.Sprintf("- V1: %d (%v)\n", ep.StatusCodeV1, ep.V1AvgTime))
			md.WriteString(fmt.Sprintf("- V2: %d (%v)\n", ep.StatusCodeV2, ep.V2AvgTime))

			if ep.Error != "" {
				md.WriteString(fmt.Sprintf("\nError: %s\n", ep.Error))
			}

			if len(ep.Differences) > 0 {
				md.WriteString("\nDifferences:\n")
				for j, diff := range ep.Differences {
					md.WriteString(fmt.Sprintf("%d. Path: %s (%s)\n", j+1, diff.Path, diff.DiffType))
					md.WriteString(fmt.Sprintf("   V1: %v\n", diff.Value1))
					md.WriteString(fmt.Sprintf("   V2: %v\n", diff.Value2))
				}
			}

			md.WriteString("\n")
		}
	}

	// Passing endpoints
	if len(passing) > 0 {
		md.WriteString(fmt.Sprintf("## Passing Endpoints (%d)\n\n", len(passing)))

		for i, ep := range passing {
			md.WriteString(fmt.Sprintf("%d. %s %s - V1: %v, V2: %v\n",
				i+1, ep.Method, ep.Path, ep.V1AvgTime, ep.V2AvgTime))
		}
		md.WriteString("\n")
	}

	return md.String()
}

// writeIndividualReports writes individual JSON and MD files for a single endpoint test
func (r *Reporter) writeIndividualReports(index int, ep EndpointReport) {
	// Determine subdirectory based on match status
	subdir := "matches"
	if !ep.Match || ep.Error != "" {
		subdir = "mismatches"
	}

	// Create filename from index, method, and sanitized path
	baseFilename := fmt.Sprintf("%03d-%s-%s", index, ep.Method, sanitizePath(ep.Path))

	// Write JSON file
	jsonFilename := fmt.Sprintf("%s/%s/%s.json", r.outputDir, subdir, baseFilename)
	jsonData, err := json.MarshalIndent(ep, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting JSON for %s: %v\n", ep.Path, err)
		return
	}
	if err := os.WriteFile(jsonFilename, jsonData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing JSON file %s: %v\n", jsonFilename, err)
	}

	// Write Markdown file
	mdFilename := fmt.Sprintf("%s/%s/%s.md", r.outputDir, subdir, baseFilename)
	mdContent := formatIndividualMarkdown(ep)
	if err := os.WriteFile(mdFilename, []byte(mdContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing MD file %s: %v\n", mdFilename, err)
	}

	// Write V1 curl command
	v1CurlFilename := fmt.Sprintf("%s/%s/%s-v1.curl.sh", r.outputDir, subdir, baseFilename)
	v1CurlCmd := generateCurlCommand(ep.V1Request)
	if err := os.WriteFile(v1CurlFilename, []byte(v1CurlCmd), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing V1 curl file %s: %v\n", v1CurlFilename, err)
	}

	// Write V2 curl command
	v2CurlFilename := fmt.Sprintf("%s/%s/%s-v2.curl.sh", r.outputDir, subdir, baseFilename)
	v2CurlCmd := generateCurlCommand(ep.V2Request)
	if err := os.WriteFile(v2CurlFilename, []byte(v2CurlCmd), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing V2 curl file %s: %v\n", v2CurlFilename, err)
	}

	// Write V1 response
	v1RespFilename := fmt.Sprintf("%s/%s/%s-v1.response.json", r.outputDir, subdir, baseFilename)
	if err := os.WriteFile(v1RespFilename, ep.V1ResponseBody, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing V1 response file %s: %v\n", v1RespFilename, err)
	}

	// Write V2 response
	v2RespFilename := fmt.Sprintf("%s/%s/%s-v2.response.json", r.outputDir, subdir, baseFilename)
	if err := os.WriteFile(v2RespFilename, ep.V2ResponseBody, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing V2 response file %s: %v\n", v2RespFilename, err)
	}
}

// sanitizePath converts a URL path to a safe filename
func sanitizePath(path string) string {
	// Remove leading slash
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	// Replace slashes and special characters with dashes
	path = strings.ReplaceAll(path, "/", "-")
	path = strings.ReplaceAll(path, "?", "-")
	path = strings.ReplaceAll(path, "&", "-")
	path = strings.ReplaceAll(path, "=", "-")
	// Limit length
	if len(path) > 80 {
		path = path[:80]
	}
	return path
}

// generateCurlCommand generates a curl command from a request
func generateCurlCommand(req client.Request) string {
	var cmd strings.Builder
	cmd.WriteString("#!/bin/bash\n\n")
	cmd.WriteString("# Auto-generated curl command\n\n")
	cmd.WriteString(fmt.Sprintf("curl -X %s \\\n", req.Method))

	// Add headers
	for key, value := range req.Headers {
		cmd.WriteString(fmt.Sprintf("  -H '%s: %s' \\\n", key, value))
	}

	// Add body if present
	if len(req.Body) > 0 {
		cmd.WriteString(fmt.Sprintf("  -d '%s' \\\n", string(req.Body)))
	}

	// Build full URL with query params
	fullURL := req.URL
	if len(req.QueryParams) > 0 {
		separator := "?"
		if strings.Contains(fullURL, "?") {
			separator = "&"
		}
		params := make([]string, 0, len(req.QueryParams))
		for key, value := range req.QueryParams {
			params = append(params, fmt.Sprintf("%s=%s", key, value))
		}
		fullURL += separator + strings.Join(params, "&")
	}

	cmd.WriteString(fmt.Sprintf("  '%s'\n", fullURL))

	return cmd.String()
}

// formatIndividualMarkdown formats a single endpoint report as Markdown
func formatIndividualMarkdown(ep EndpointReport) string {
	md := fmt.Sprintf("# %s %s\n\n", ep.Method, ep.Path)
	md += fmt.Sprintf("**Status**: %s\n\n", ep.Summary())

	md += "## Performance\n\n"
	md += fmt.Sprintf("- **V1 Average Time**: %v\n", ep.V1AvgTime)
	md += fmt.Sprintf("- **V2 Average Time**: %v\n", ep.V2AvgTime)
	md += fmt.Sprintf("- **Iterations**: %d\n\n", ep.Iterations)

	md += "## Response Codes\n\n"
	md += fmt.Sprintf("- **V1**: %d\n", ep.StatusCodeV1)
	md += fmt.Sprintf("- **V2**: %d\n\n", ep.StatusCodeV2)

	if ep.Error != "" {
		md += fmt.Sprintf("## Error\n\n```\n%s\n```\n\n", ep.Error)
	}

	if len(ep.Differences) > 0 {
		md += "## Differences\n\n"
		for i, diff := range ep.Differences {
			md += fmt.Sprintf("%d. **Path**: `%s`\n", i+1, diff.Path)
			md += fmt.Sprintf("   - **Type**: %s\n", diff.DiffType)
			md += fmt.Sprintf("   - **V1 Value**: %v\n", diff.Value1)
			md += fmt.Sprintf("   - **V2 Value**: %v\n\n", diff.Value2)
		}
	} else if ep.Error == "" {
		md += "## Result\n\n✅ Responses match perfectly!\n"
	}

	return md
}
