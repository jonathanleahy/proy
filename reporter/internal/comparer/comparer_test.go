package comparer

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComparer_Compare(t *testing.T) {
	tests := []struct {
		name         string
		v1JSON       string
		v2JSON       string
		ignoreFields []string
		wantMatch    bool
		wantDiffs    int
		validate     func(t *testing.T, result ComparisonResult)
	}{
		{
			name:      "identical simple objects",
			v1JSON:    `{"id": 1, "name": "test"}`,
			v2JSON:    `{"id": 1, "name": "test"}`,
			wantMatch: true,
			wantDiffs: 0,
		},
		{
			name:      "identical arrays",
			v1JSON:    `[1, 2, 3]`,
			v2JSON:    `[1, 2, 3]`,
			wantMatch: true,
			wantDiffs: 0,
		},
		{
			name:      "identical nested objects",
			v1JSON:    `{"user": {"id": 1, "name": "John"}, "active": true}`,
			v2JSON:    `{"user": {"id": 1, "name": "John"}, "active": true}`,
			wantMatch: true,
			wantDiffs: 0,
		},
		{
			name:      "simple value mismatch",
			v1JSON:    `{"id": 1, "name": "test"}`,
			v2JSON:    `{"id": 1, "name": "different"}`,
			wantMatch: false,
			wantDiffs: 1,
			validate: func(t *testing.T, result ComparisonResult) {
				assert.Equal(t, "name", result.Differences[0].Path)
				assert.Equal(t, "test", result.Differences[0].Value1)
				assert.Equal(t, "different", result.Differences[0].Value2)
				assert.Equal(t, DiffTypeValueMismatch, result.Differences[0].DiffType)
			},
		},
		{
			name:      "number mismatch",
			v1JSON:    `{"id": 1, "count": 5}`,
			v2JSON:    `{"id": 1, "count": 10}`,
			wantMatch: false,
			wantDiffs: 1,
			validate: func(t *testing.T, result ComparisonResult) {
				assert.Equal(t, "count", result.Differences[0].Path)
				assert.Equal(t, float64(5), result.Differences[0].Value1)
				assert.Equal(t, float64(10), result.Differences[0].Value2)
			},
		},
		{
			name:      "missing field in v2",
			v1JSON:    `{"id": 1, "name": "test", "extra": "field"}`,
			v2JSON:    `{"id": 1, "name": "test"}`,
			wantMatch: false,
			wantDiffs: 1,
			validate: func(t *testing.T, result ComparisonResult) {
				assert.Equal(t, "extra", result.Differences[0].Path)
				assert.Equal(t, DiffTypeMissing, result.Differences[0].DiffType)
			},
		},
		{
			name:      "extra field in v2",
			v1JSON:    `{"id": 1, "name": "test"}`,
			v2JSON:    `{"id": 1, "name": "test", "new": "field"}`,
			wantMatch: false,
			wantDiffs: 1,
			validate: func(t *testing.T, result ComparisonResult) {
				assert.Equal(t, "new", result.Differences[0].Path)
				assert.Equal(t, DiffTypeExtra, result.Differences[0].DiffType)
			},
		},
		{
			name:      "type mismatch",
			v1JSON:    `{"value": "123"}`,
			v2JSON:    `{"value": 123}`,
			wantMatch: false,
			wantDiffs: 1,
			validate: func(t *testing.T, result ComparisonResult) {
				assert.Equal(t, "value", result.Differences[0].Path)
				assert.Equal(t, DiffTypeTypeMismatch, result.Differences[0].DiffType)
			},
		},
		{
			name:      "nested object mismatch",
			v1JSON:    `{"user": {"id": 1, "name": "John"}}`,
			v2JSON:    `{"user": {"id": 1, "name": "Jane"}}`,
			wantMatch: false,
			wantDiffs: 1,
			validate: func(t *testing.T, result ComparisonResult) {
				assert.Equal(t, "user.name", result.Differences[0].Path)
				assert.Equal(t, "John", result.Differences[0].Value1)
				assert.Equal(t, "Jane", result.Differences[0].Value2)
			},
		},
		{
			name:      "array element mismatch",
			v1JSON:    `{"items": [1, 2, 3]}`,
			v2JSON:    `{"items": [1, 2, 4]}`,
			wantMatch: false,
			wantDiffs: 1,
			validate: func(t *testing.T, result ComparisonResult) {
				assert.Contains(t, result.Differences[0].Path, "items")
			},
		},
		{
			name:         "ignore specified fields",
			v1JSON:       `{"id": 1, "timestamp": "2023-01-01", "name": "test"}`,
			v2JSON:       `{"id": 1, "timestamp": "2023-12-31", "name": "test"}`,
			ignoreFields: []string{"timestamp"},
			wantMatch:    true,
			wantDiffs:    0,
		},
		{
			name:         "ignore nested fields",
			v1JSON:       `{"user": {"id": 1, "createdAt": "2023-01-01"}}`,
			v2JSON:       `{"user": {"id": 1, "createdAt": "2023-12-31"}}`,
			ignoreFields: []string{"user.createdAt"},
			wantMatch:    true,
			wantDiffs:    0,
		},
		{
			name:         "ignore field but other differences exist",
			v1JSON:       `{"id": 1, "timestamp": "2023-01-01", "name": "test1"}`,
			v2JSON:       `{"id": 1, "timestamp": "2023-12-31", "name": "test2"}`,
			ignoreFields: []string{"timestamp"},
			wantMatch:    false,
			wantDiffs:    1,
			validate: func(t *testing.T, result ComparisonResult) {
				assert.Equal(t, "name", result.Differences[0].Path)
			},
		},
		{
			name:      "multiple differences",
			v1JSON:    `{"id": 1, "name": "test", "count": 5}`,
			v2JSON:    `{"id": 2, "name": "different", "count": 5}`,
			wantMatch: false,
			wantDiffs: 2,
		},
		{
			name:      "boolean mismatch",
			v1JSON:    `{"active": true}`,
			v2JSON:    `{"active": false}`,
			wantMatch: false,
			wantDiffs: 1,
		},
		{
			name:      "null vs value",
			v1JSON:    `{"value": null}`,
			v2JSON:    `{"value": "something"}`,
			wantMatch: false,
			wantDiffs: 1,
		},
		{
			name:      "empty object vs populated",
			v1JSON:    `{"data": {}}`,
			v2JSON:    `{"data": {"key": "value"}}`,
			wantMatch: false,
			wantDiffs: 1,
		},
		{
			name:      "empty array vs populated",
			v1JSON:    `{"items": []}`,
			v2JSON:    `{"items": [1, 2, 3]}`,
			wantMatch: false,
			wantDiffs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comparer := NewComparer(tt.ignoreFields)

			pair := ResponsePair{
				V1: json.RawMessage(tt.v1JSON),
				V2: json.RawMessage(tt.v2JSON),
			}

			result := comparer.Compare(pair)

			assert.Equal(t, tt.wantMatch, result.Match, "Match status incorrect")
			assert.Len(t, result.Differences, tt.wantDiffs, "Number of differences incorrect")

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestComparer_Compare_InvalidJSON(t *testing.T) {
	comparer := NewComparer(nil)

	tests := []struct {
		name   string
		v1JSON string
		v2JSON string
	}{
		{
			name:   "invalid v1 JSON",
			v1JSON: `{invalid}`,
			v2JSON: `{"valid": true}`,
		},
		{
			name:   "invalid v2 JSON",
			v1JSON: `{"valid": true}`,
			v2JSON: `{invalid}`,
		},
		{
			name:   "both invalid",
			v1JSON: `{invalid1}`,
			v2JSON: `{invalid2}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pair := ResponsePair{
				V1: json.RawMessage(tt.v1JSON),
				V2: json.RawMessage(tt.v2JSON),
			}

			result := comparer.Compare(pair)

			// Should not match and have at least one difference
			assert.False(t, result.Match)
			assert.NotEmpty(t, result.Differences)
		})
	}
}

func TestComparer_Compare_ComplexRealWorld(t *testing.T) {
	// Real-world example from the REST API
	v1 := `{
		"userId": 1,
		"userName": "Leanne Graham",
		"email": "Sincere@april.biz",
		"stats": {
			"totalPosts": 3,
			"totalTodos": 5,
			"completedTodos": 2,
			"pendingTodos": 3,
			"completionRate": "40.0%"
		},
		"generatedAt": "2025-11-01T15:00:00.000Z"
	}`

	v2 := `{
		"userId": 1,
		"userName": "Leanne Graham",
		"email": "Sincere@april.biz",
		"stats": {
			"totalPosts": 3,
			"totalTodos": 5,
			"completedTodos": 2,
			"pendingTodos": 3,
			"completionRate": "40.0%"
		},
		"generatedAt": "2025-11-01T16:00:00.000Z"
	}`

	comparer := NewComparer([]string{"generatedAt"})

	result := comparer.Compare(ResponsePair{
		V1: json.RawMessage(v1),
		V2: json.RawMessage(v2),
	})

	require.True(t, result.Match, "Should match when ignoring generatedAt")
	assert.Empty(t, result.Differences)
}
