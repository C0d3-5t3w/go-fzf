package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	RipgrepPath string   `yaml:"ripgrep_path"`
	SearchDirs  []string `yaml:"search_dirs"`
}

func LoadConfig(configPath string) (*Config, error) {
	cfg := &Config{
		RipgrepPath: "rg",
		SearchDirs:  []string{"."},
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
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
