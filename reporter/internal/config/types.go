package config

import "encoding/json"

// Config represents the complete reporter configuration
type Config struct {
	BaseURLV1  string     `json:"base_url_v1"`
	BaseURLV2  string     `json:"base_url_v2"`
	Iterations int        `json:"iterations"`
	Endpoints  []Endpoint `json:"endpoints"`
	IgnoreFields []string `json:"ignore_fields,omitempty"`
}

// Endpoint represents a single API endpoint to test
type Endpoint struct {
	Path        string            `json:"path"`
	Method      string            `json:"method"`
	Headers     map[string]string `json:"headers,omitempty"`
	QueryParams map[string]string `json:"query_params,omitempty"`
	Body        json.RawMessage   `json:"body,omitempty"`
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.BaseURLV1 == "" {
		return ErrMissingBaseURLV1
	}
	if c.BaseURLV2 == "" {
		return ErrMissingBaseURLV2
	}
	if c.Iterations < 1 {
		c.Iterations = 1 // Default to 1 iteration
	}
	if len(c.Endpoints) == 0 {
		return ErrNoEndpoints
	}
	for i, ep := range c.Endpoints {
		if ep.Path == "" {
			return &ValidationError{Field: "path", Index: i}
		}
		if ep.Method == "" {
			c.Endpoints[i].Method = "GET" // Default to GET
		}
	}
	return nil
}
