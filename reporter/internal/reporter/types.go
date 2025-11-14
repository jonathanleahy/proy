package reporter

import (
	"time"

	"github.com/jonathanleahy/prroxy/reporter/internal/client"
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
	Path          string
	Method        string
	Match         bool
	Iterations    int
	Retries       int // Number of retry attempts before success/failure
	V1Timings     []time.Duration
	V2Timings     []time.Duration
	V1AvgTime     time.Duration
	V2AvgTime     time.Duration
	StatusCodeV1  int
	StatusCodeV2  int
	Differences   []comparer.Difference
	Error         string
	V1Request     client.Request     `json:"-"` // Exclude from JSON marshaling
	V2Request     client.Request     `json:"-"`
	V1ResponseBody []byte            `json:"-"`
	V2ResponseBody []byte            `json:"-"`
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
