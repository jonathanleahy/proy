package config

import "fmt"

var (
	// ErrMissingBaseURLV1 indicates base_url_v1 is missing
	ErrMissingBaseURLV1 = fmt.Errorf("base_url_v1 is required")

	// ErrMissingBaseURLV2 indicates base_url_v2 is missing
	ErrMissingBaseURLV2 = fmt.Errorf("base_url_v2 is required")

	// ErrNoEndpoints indicates no endpoints were configured
	ErrNoEndpoints = fmt.Errorf("at least one endpoint is required")
)

// ValidationError represents a validation error for a specific field
type ValidationError struct {
	Field string
	Index int
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s at endpoint index %d", e.Field, e.Index)
}
