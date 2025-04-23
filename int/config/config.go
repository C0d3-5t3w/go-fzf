package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration.
type Config struct {
	RipgrepPath string   `yaml:"ripgrep_path"`
	SearchDirs  []string `yaml:"search_dirs"`
}

// LoadConfig reads the configuration from the specified YAML file.
func LoadConfig(configPath string) (*Config, error) {
	cfg := &Config{
		RipgrepPath: "rg", // Default value
		SearchDirs:  []string{"."},
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		// If the file doesn't exist, return default config without error
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
