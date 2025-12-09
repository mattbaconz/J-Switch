package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/user/jswitch/pkg/models"
)

const (
	configDirName  = ".jswitch"
	configFileName = "config.json"
)

// Config holds the persistent state of the application.
type Config struct {
	CurrentVersion string                    `json:"current_version"`
	Installations  []models.JavaInstallation `json:"installations"`
}

// getConfigPath returns the full path to the config file (e.g. ~/.jswitch/config.json).
func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not find user home directory: %w", err)
	}
	return filepath.Join(home, configDirName, configFileName), nil
}

// LoadConfig reads the config file from disk.
// Returns an empty config if the file does not exist.
func LoadConfig() (*Config, error) {
	path, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		// Return valid empty config if file doesn't exist
		return &Config{Installations: []models.JavaInstallation{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read config file at %s: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// SaveConfig writes the config to disk.
func SaveConfig(cfg *Config) error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file to %s: %w", path, err)
	}

	return nil
}

// CurrentVersionPath returns the path for the given version, or empty string.
func (c *Config) CurrentVersionPath(version string) string {
	for _, inst := range c.Installations {
		if inst.Version == version {
			return inst.Path
		}
	}
	return ""
}
