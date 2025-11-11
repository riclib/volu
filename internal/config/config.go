package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// RadioSeries defines a radio series configuration for the radio command.
type RadioSeries struct {
	Name        string `yaml:"name"`         // Display name (e.g., "A State of Trance")
	SearchQuery string `yaml:"search_query"` // Search query for Volumio API
	Pattern     string `yaml:"pattern"`      // Regex pattern to match album names
}

// Config represents the volu configuration file structure.
type Config struct {
	Host  string                 `yaml:"host"`  // Volumio host (hostname or IP)
	Radio map[string]RadioSeries `yaml:"radio"` // Radio series configurations
}

// DefaultConfig returns a Config with default values.
func DefaultConfig() *Config {
	return &Config{
		Host:  "volumio.local",
		Radio: make(map[string]RadioSeries),
	}
}

// GetConfigPath returns the path to the config file.
// On Linux, this is typically ~/.config/volu/config.yaml.
func GetConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory: %w", err)
	}
	return filepath.Join(configDir, "volu", "config.yaml"), nil
}

// Load reads the config file and returns a Config.
// If the config file doesn't exist, returns default config.
// Returns error only for actual read/parse failures.
func Load() (*Config, error) {
	cfg := DefaultConfig()

	configPath, err := GetConfigPath()
	if err != nil {
		// Can't determine config path, return defaults
		return cfg, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Config file doesn't exist, return defaults
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Ensure Radio map is initialized even if not in YAML
	if cfg.Radio == nil {
		cfg.Radio = make(map[string]RadioSeries)
	}

	return cfg, nil
}

// Save writes the config to the config file.
func Save(cfg *Config) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	// Create directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
