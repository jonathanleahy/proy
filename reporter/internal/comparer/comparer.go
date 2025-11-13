package comparer

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// Comparer compares two JSON responses
type Comparer struct {
	ignoreFields map[string]bool
}

// NewComparer creates a new Comparer with optional fields to ignore
func NewComparer(ignoreFields []string) *Comparer {
	ignore := make(map[string]bool)
	for _, field := range ignoreFields {
		ignore[field] = true
	}
	return &Comparer{
		ignoreFields: ignore,
	}
}

// Compare compares two responses and returns the result
func (c *Comparer) Compare(pair ResponsePair) ComparisonResult {
	result := ComparisonResult{
		Match:       true,
		Differences: []Difference{},
	}

	// Parse V1
	var v1Data interface{}
	if err := json.Unmarshal(pair.V1, &v1Data); err != nil {
		result.Match = false
		result.Differences = append(result.Differences, Difference{
			Path:     "v1",
			DiffType: DiffTypeValueMismatch,
			Value1:   string(pair.V1),
			Value2:   fmt.Sprintf("JSON parse error: %v", err),
		})
		return result
	}

	// Parse V2
	var v2Data interface{}
	if err := json.Unmarshal(pair.V2, &v2Data); err != nil {
		result.Match = false
		result.Differences = append(result.Differences, Difference{
			Path:     "v2",
			DiffType: DiffTypeValueMismatch,
			Value1:   fmt.Sprintf("JSON parse error: %v", err),
			Value2:   string(pair.V2),
		})
		return result
	}

	// Compare the data
	diffs := c.compareValues("", v1Data, v2Data)
	if len(diffs) > 0 {
		result.Match = false
		result.Differences = diffs
	}

	return result
}

// compareValues recursively compares two values
func (c *Comparer) compareValues(path string, v1, v2 interface{}) []Difference {
	var diffs []Difference

	// Check if this field should be ignored
	if c.shouldIgnore(path) {
		return diffs
	}

	// Get types
	t1 := reflect.TypeOf(v1)
	t2 := reflect.TypeOf(v2)

	// Type mismatch
	if t1 != t2 {
		diffs = append(diffs, Difference{
			Path:     path,
			Value1:   v1,
			Value2:   v2,
			DiffType: DiffTypeTypeMismatch,
		})
		return diffs
	}

	// Compare based on type
	switch v1 := v1.(type) {
	case map[string]interface{}:
		v2Map := v2.(map[string]interface{})
		diffs = append(diffs, c.compareMaps(path, v1, v2Map)...)

	case []interface{}:
		v2Slice := v2.([]interface{})
		diffs = append(diffs, c.compareSlices(path, v1, v2Slice)...)

	default:
		// Simple value comparison
		if !reflect.DeepEqual(v1, v2) {
			diffs = append(diffs, Difference{
				Path:     path,
				Value1:   v1,
				Value2:   v2,
				DiffType: DiffTypeValueMismatch,
			})
		}
	}

	return diffs
}

// compareMaps compares two maps
func (c *Comparer) compareMaps(path string, m1, m2 map[string]interface{}) []Difference {
	var diffs []Difference

	// Check all keys in m1
	for key, v1Val := range m1 {
		keyPath := buildPath(path, key)

		if c.shouldIgnore(keyPath) {
			continue
		}

		v2Val, exists := m2[key]
		if !exists {
			diffs = append(diffs, Difference{
				Path:     keyPath,
				Value1:   v1Val,
				Value2:   nil,
				DiffType: DiffTypeMissing,
			})
			continue
		}

		// Recursively compare values
		diffs = append(diffs, c.compareValues(keyPath, v1Val, v2Val)...)
	}

	// Check for extra keys in m2
	for key, v2Val := range m2 {
		keyPath := buildPath(path, key)

		if c.shouldIgnore(keyPath) {
			continue
		}

		if _, exists := m1[key]; !exists {
			diffs = append(diffs, Difference{
				Path:     keyPath,
				Value1:   nil,
				Value2:   v2Val,
				DiffType: DiffTypeExtra,
			})
		}
	}

	return diffs
}

// compareSlices compares two slices
func (c *Comparer) compareSlices(path string, s1, s2 []interface{}) []Difference {
	var diffs []Difference

	// Check length
	if len(s1) != len(s2) {
		diffs = append(diffs, Difference{
			Path:     path,
			Value1:   fmt.Sprintf("length: %d", len(s1)),
			Value2:   fmt.Sprintf("length: %d", len(s2)),
			DiffType: DiffTypeValueMismatch,
		})
		return diffs
	}

	// Compare elements
	for i := 0; i < len(s1); i++ {
		indexPath := fmt.Sprintf("%s[%d]", path, i)
		diffs = append(diffs, c.compareValues(indexPath, s1[i], s2[i])...)
	}

	return diffs
}

// shouldIgnore checks if a field path should be ignored
func (c *Comparer) shouldIgnore(path string) bool {
	if path == "" {
		return false
	}
	return c.ignoreFields[path]
}

// buildPath constructs a field path
func buildPath(parent, child string) string {
	if parent == "" {
		return child
	}
	return parent + "." + child
}

// FormatDifferences formats differences into a human-readable string
func FormatDifferences(diffs []Difference) string {
	if len(diffs) == 0 {
		return "No differences found"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d difference(s):\n", len(diffs)))

	for i, diff := range diffs {
		sb.WriteString(fmt.Sprintf("\n%d. Path: %s\n", i+1, diff.Path))
		sb.WriteString(fmt.Sprintf("   Type: %s\n", diff.DiffType))
		sb.WriteString(fmt.Sprintf("   V1: %v\n", diff.Value1))
		sb.WriteString(fmt.Sprintf("   V2: %v\n", diff.Value2))
	}

	return sb.String()
}
