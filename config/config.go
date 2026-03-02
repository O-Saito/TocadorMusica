package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const FRAME_RATE = 44100
const DEFAULT_SEARCH_RESULTS = 5

var configFile = ".config"

type Config struct {
	Volume        float64
	SearchResults int
}

func Load() (*Config, error) {
	return LoadFrom(configFile)
}

func LoadFrom(path string) (*Config, error) {
	cfg := &Config{
		Volume:        1.0,
		SearchResults: DEFAULT_SEARCH_RESULTS,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, nil
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "volume" {
			if v, err := strconv.ParseFloat(value, 64); err == nil {
				cfg.Volume = v
			}
		}
		if key == "search_results" {
			if v, err := strconv.Atoi(value); err == nil && v > 0 {
				cfg.SearchResults = v
			}
		}
	}

	return cfg, nil
}

func (c *Config) Save() error {
	return c.SaveTo(configFile)
}

func (c *Config) SaveTo(path string) error {
	dir, _ := filepath.Split(path)
	if dir != "" {
		os.MkdirAll(dir, 0755)
	}

	content := fmt.Sprintf("volume=%.2f\nsearch_results=%d\n", c.Volume, c.SearchResults)
	return os.WriteFile(path, []byte(content), 0644)
}
