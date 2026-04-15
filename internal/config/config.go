package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Service represents a single log source to tail.
type Service struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
	Color   string `yaml:"color"`
}

// Config holds the top-level logdrift configuration.
type Config struct {
	Services []Service `yaml:"services"`
	DiffMode string    `yaml:"diff_mode"` // "line" or "word"
}

// Load reads and parses a YAML config file at the given path.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("config: open %q: %w", path, err)
	}
	defer f.Close()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("config: decode %q: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// validate checks required fields and sets defaults.
func (c *Config) validate() error {
	if len(c.Services) == 0 {
		return fmt.Errorf("config: at least one service must be defined")
	}

	seen := make(map[string]bool)
	for i, svc := range c.Services {
		if svc.Name == "" {
			return fmt.Errorf("config: service[%d] missing name", i)
		}
		if svc.Command == "" {
			return fmt.Errorf("config: service %q missing command", svc.Name)
		}
		if seen[svc.Name] {
			return fmt.Errorf("config: duplicate service name %q", svc.Name)
		}
		seen[svc.Name] = true
	}

	if c.DiffMode == "" {
		c.DiffMode = "line"
	}
	if c.DiffMode != "line" && c.DiffMode != "word" {
		return fmt.Errorf("config: diff_mode must be \"line\" or \"word\", got %q", c.DiffMode)
	}

	return nil
}
