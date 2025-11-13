package comparer

import "encoding/json"

// ComparisonResult represents the result of comparing two responses
type ComparisonResult struct {
	Match       bool
	Differences []Difference
}

// Difference represents a single difference between two responses
type Difference struct {
	Path     string      // JSON path where difference was found (e.g., "user.name")
	Value1   interface{} // Value from first response
	Value2   interface{} // Value from second response
	DiffType DiffType
}

// DiffType indicates the type of difference found
type DiffType string

const (
	DiffTypeValueMismatch DiffType = "value_mismatch"
	DiffTypeMissing       DiffType = "missing_in_v2"
	DiffTypeExtra         DiffType = "extra_in_v2"
	DiffTypeTypeMismatch  DiffType = "type_mismatch"
)

// ResponsePair holds two responses to compare
type ResponsePair struct {
	V1 json.RawMessage
	V2 json.RawMessage
}
