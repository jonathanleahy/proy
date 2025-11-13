package reporter

import (
	"time"

	"github.com/jonathanleahy/prroxy/reporter/internal/comparer"
)

// Report represents the complete test report
type Report struct {
	TotalEndpoints  int
	MatchedEndpoints int
	FailedEndpoints  int
	TotalDuration   time.Duration
	Endpoints       []EndpointReport
}

// EndpointReport represents the test results for a single endpoint
type EndpointReport struct {
	Path        string
	Method      string
	Match       bool
	Iterations  int
	V1Timings   []time.Duration
	V2Timings   []time.Duration
	V1AvgTime   time.Duration
	V2AvgTime   time.Duration
	StatusCodeV1 int
	StatusCodeV2 int
	Differences []comparer.Difference
	Error       string
}

// Summary returns a brief summary of the endpoint test
func (er *EndpointReport) Summary() string {
	if er.Error != "" {
		return "ERROR: " + er.Error
	}
	if er.Match {
		return "MATCH"
	}
	return "MISMATCH"
}
