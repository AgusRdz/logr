package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds user preferences for logr.
type Config struct {
	NoColor    bool     `json:"no_color"`
	Compact    bool     `json:"compact"`
	HideFields []string `json:"hide_fields"`
}

// Path returns ~/.config/logr/config.json
func Path() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "logr", "config.json")
}

// Load reads the config file. Returns defaults if file missing or invalid.
func Load() *Config {
	data, err := os.ReadFile(Path())
	if err != nil {
		return Default()
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Default()
	}
	return &cfg
}

// Default returns a Config with default values.
func Default() *Config {
	return &Config{}
}
