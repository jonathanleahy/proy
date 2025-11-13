package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sync"

	"gopkg.in/yaml.v2"
)

// Config holds the application configuration
type Config struct {
	Server   ServerConfig   `json:"server" yaml:"server"`
	Storage  StorageConfig  `json:"storage" yaml:"storage"`
	Mode     ModeConfig     `json:"mode" yaml:"mode"`
	TLS      TLSConfig      `json:"tls" yaml:"tls"`
	mu       sync.RWMutex   // For thread-safe mode changes
}

// ServerConfig contains server settings
type ServerConfig struct {
	Port string `json:"port" yaml:"port"`
	Host string `json:"host" yaml:"host"`
}

// StorageConfig contains storage settings
type StorageConfig struct {
	Type string `json:"type" yaml:"type"`
	Path string `json:"path" yaml:"path"`
}

// ModeConfig contains mode settings
type ModeConfig struct {
	Default string `json:"default" yaml:"default"`
	current string // Internal field for runtime mode
}

// TLSConfig contains TLS settings
type TLSConfig struct {
	SkipVerify bool `json:"skip_verify" yaml:"skip_verify"`
}

// singleton instance
var (
	instance *Config
	once     sync.Once
)

// GetInstance returns the singleton configuration instance
func GetInstance() *Config {
	once.Do(func() {
		instance = &Config{
			// Default values
			Server: ServerConfig{
				Port: "8099",
				Host: "0.0.0.0",
			},
			Storage: StorageConfig{
				Type: "filesystem",
				Path: "./recordings",
			},
			Mode: ModeConfig{
				Default: "playback",
				current: "playback",
			},
			TLS: TLSConfig{
				SkipVerify: true,
			},
		}
	})
	return instance
}

// Load loads configuration from various sources
func (c *Config) Load() error {
	// 1. Load from config file if exists
	if err := c.loadFromFile(); err == nil {
		fmt.Println("Loaded configuration from file")
	}

	// 2. Override with environment variables
	c.loadFromEnv()

	// 3. Override with command line flags
	c.loadFromFlags()

	// Set initial current mode
	c.Mode.current = c.Mode.Default

	return nil
}

// loadFromFile attempts to load configuration from a file
func (c *Config) loadFromFile() error {
	// Try different config file names
	configFiles := []string{"proxy.yaml", "proxy.yml", "proxy.json", "config.yaml", "config.json"}

	for _, filename := range configFiles {
		file, err := os.Open(filename)
		if err != nil {
			continue
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			continue
		}

		// Try to parse as YAML first (works for both YAML and JSON)
		if err := yaml.Unmarshal(data, c); err == nil {
			return nil
		}

		// Try JSON if YAML fails
		if err := json.Unmarshal(data, c); err == nil {
			return nil
		}
	}

	return fmt.Errorf("no valid configuration file found")
}

// loadFromEnv loads configuration from environment variables
func (c *Config) loadFromEnv() {
	if port := os.Getenv("PROXY_PORT"); port != "" {
		c.Server.Port = port
	}
	if host := os.Getenv("PROXY_HOST"); host != "" {
		c.Server.Host = host
	}
	if path := os.Getenv("PROXY_RECORDINGS_DIR"); path != "" {
		c.Storage.Path = path
	}
	if mode := os.Getenv("PROXY_MODE"); mode != "" {
		c.Mode.Default = mode
	}
	if skipVerify := os.Getenv("PROXY_TLS_SKIP_VERIFY"); skipVerify == "false" {
		c.TLS.SkipVerify = false
	}
}

// loadFromFlags loads configuration from command line flags
func (c *Config) loadFromFlags() {
	port := flag.String("port", c.Server.Port, "Server port")
	host := flag.String("host", c.Server.Host, "Server host")
	recordingsDir := flag.String("recordings-dir", c.Storage.Path, "Recordings directory")
	mode := flag.String("mode", c.Mode.Default, "Default mode (record/playback)")
	skipVerify := flag.Bool("skip-verify", c.TLS.SkipVerify, "Skip TLS verification")

	flag.Parse()

	c.Server.Port = *port
	c.Server.Host = *host
	c.Storage.Path = *recordingsDir
	c.Mode.Default = *mode
	c.TLS.SkipVerify = *skipVerify
}

// GetMode returns the current mode
func (c *Config) GetMode() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Mode.current
}

// SetMode sets the current mode
func (c *Config) SetMode(mode string) error {
	if mode != "record" && mode != "playback" {
		return fmt.Errorf("invalid mode: %s (must be 'record' or 'playback')", mode)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.Mode.current = mode
	return nil
}

// GetAddress returns the server address
func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
}