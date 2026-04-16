// Package config loads and validates the logdrift YAML configuration file.
package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Source describes a single log source (process or file tail).
type Source struct {
	Name    string   `yaml:"name"`
	Command string   `yaml:"command"`
	Args    []string `yaml:"args"`
	File    string   `yaml:"file"`
}

// ThrottleConfig holds optional per-run throttle settings.
type ThrottleConfig struct {
	LinesPerSec int `yaml:"lines_per_sec"`
}

// Config is the top-level configuration structure.
type Config struct {
	Sources  []Source       `yaml:"sources"`
	DiffMode string         `yaml:"diff_mode"`
	Throttle ThrottleConfig `yaml:"throttle"`
}

// Load reads and validates a Config from the YAML file at path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read %s: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse: %w", err)
	}
	if err := validate(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func validate(cfg *Config) error {
	if len(cfg.Sources) == 0 {
		return errors.New("config: at least one source is required")
	}
	seen := make(map[string]bool)
	for i, s := range cfg.Sources {
		if s.Name == "" {
			return fmt.Errorf("config: source[%d]: name is required", i)
		}
		if s.Command == "" && s.File == "" {
			return fmt.Errorf("config: source %q: command or file is required", s.Name)
		}
		if seen[s.Name] {
			return fmt.Errorf("config: duplicate source name %q", s.Name)
		}
		seen[s.Name] = true
	}
	if cfg.DiffMode == "" {
		cfg.DiffMode = "uniq"
	}
	allowed := map[string]bool{"none": true, "uniq": true, "fuzzy": true}
	if !allowed[cfg.DiffMode] {
		return fmt.Errorf("config: unknown diff_mode %q", cfg.DiffMode)
	}
	if cfg.Throttle.LinesPerSec < 0 {
		return errors.New("config: throttle.lines_per_sec must be >= 0")
	}
	return nil
}
