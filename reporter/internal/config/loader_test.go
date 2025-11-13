package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		configJSON  string
		wantErr     bool
		errContains string
		validate    func(t *testing.T, cfg *Config)
	}{
		{
			name: "valid config with all fields",
			configJSON: `{
				"base_url_v1": "http://0.0.0.0:3000",
				"base_url_v2": "http://0.0.0.0:8080",
				"iterations": 5,
				"endpoints": [
					{
						"path": "/api/user/1",
						"method": "GET",
						"headers": {"Authorization": "Bearer token"}
					}
				]
			}`,
			wantErr: false,
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "http://0.0.0.0:3000", cfg.BaseURLV1)
				assert.Equal(t, "http://0.0.0.0:8080", cfg.BaseURLV2)
				assert.Equal(t, 5, cfg.Iterations)
				assert.Len(t, cfg.Endpoints, 1)
				assert.Equal(t, "/api/user/1", cfg.Endpoints[0].Path)
				assert.Equal(t, "GET", cfg.Endpoints[0].Method)
			},
		},
		{
			name: "defaults iterations to 1",
			configJSON: `{
				"base_url_v1": "http://0.0.0.0:3000",
				"base_url_v2": "http://0.0.0.0:8080",
				"endpoints": [{"path": "/api/test"}]
			}`,
			wantErr: false,
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 1, cfg.Iterations)
			},
		},
		{
			name: "defaults method to GET",
			configJSON: `{
				"base_url_v1": "http://0.0.0.0:3000",
				"base_url_v2": "http://0.0.0.0:8080",
				"endpoints": [{"path": "/api/test"}]
			}`,
			wantErr: false,
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "GET", cfg.Endpoints[0].Method)
			},
		},
		{
			name: "missing base_url_v1",
			configJSON: `{
				"base_url_v2": "http://0.0.0.0:8080",
				"endpoints": [{"path": "/api/test"}]
			}`,
			wantErr:     true,
			errContains: "base_url_v1",
		},
		{
			name: "missing base_url_v2",
			configJSON: `{
				"base_url_v1": "http://0.0.0.0:3000",
				"endpoints": [{"path": "/api/test"}]
			}`,
			wantErr:     true,
			errContains: "base_url_v2",
		},
		{
			name: "no endpoints",
			configJSON: `{
				"base_url_v1": "http://0.0.0.0:3000",
				"base_url_v2": "http://0.0.0.0:8080",
				"endpoints": []
			}`,
			wantErr:     true,
			errContains: "endpoint",
		},
		{
			name: "endpoint missing path",
			configJSON: `{
				"base_url_v1": "http://0.0.0.0:3000",
				"base_url_v2": "http://0.0.0.0:8080",
				"endpoints": [{"method": "GET"}]
			}`,
			wantErr:     true,
			errContains: "path",
		},
		{
			name: "POST with body",
			configJSON: `{
				"base_url_v1": "http://0.0.0.0:3000",
				"base_url_v2": "http://0.0.0.0:8080",
				"endpoints": [
					{
						"path": "/api/user/1/report",
						"method": "POST",
						"body": {"includeCompleted": true}
					}
				]
			}`,
			wantErr: false,
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "POST", cfg.Endpoints[0].Method)
				assert.NotNil(t, cfg.Endpoints[0].Body)
			},
		},
		{
			name: "with ignore fields",
			configJSON: `{
				"base_url_v1": "http://0.0.0.0:3000",
				"base_url_v2": "http://0.0.0.0:8080",
				"endpoints": [{"path": "/api/test"}],
				"ignore_fields": ["timestamp", "generatedAt"]
			}`,
			wantErr: false,
			validate: func(t *testing.T, cfg *Config) {
				assert.Len(t, cfg.IgnoreFields, 2)
				assert.Contains(t, cfg.IgnoreFields, "timestamp")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpFile := filepath.Join(t.TempDir(), "config.json")
			err := os.WriteFile(tmpFile, []byte(tt.configJSON), 0644)
			require.NoError(t, err)

			// Load config
			cfg, err := Load(tmpFile)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cfg)
			if tt.validate != nil {
				tt.validate(t, cfg)
			}
		})
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/file.json")
	assert.Error(t, err)
}

func TestLoad_InvalidJSON(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "invalid.json")
	err := os.WriteFile(tmpFile, []byte("{invalid json}"), 0644)
	require.NoError(t, err)

	_, err = Load(tmpFile)
	assert.Error(t, err)
}
