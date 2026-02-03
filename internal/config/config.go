package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the config.yaml schema.
type Config struct {
	Name        string    `yaml:"name"`
	Description string    `yaml:"description"`
	TechStack   []string  `yaml:"tech_stack"`
	Created     time.Time `yaml:"created"`
	Reviewers   Reviewers `yaml:"reviewers"`
}

// Reviewers defines gate reviewer configuration.
type Reviewers struct {
	Default   string            `yaml:"default"`   // "auto" or "human"
	Overrides map[string]string `yaml:"overrides"` // stage -> reviewer mapping
}

// GetReviewer returns the reviewer for a given stage.
func (r *Reviewers) GetReviewer(stage string) string {
	if override, ok := r.Overrides[stage]; ok {
		return override
	}
	return r.Default
}

// SetReviewer sets the reviewer for a stage.
func (r *Reviewers) SetReviewer(stage, reviewer string) {
	if r.Overrides == nil {
		r.Overrides = make(map[string]string)
	}
	if reviewer == r.Default {
		// Remove override if it matches default
		delete(r.Overrides, stage)
	} else {
		r.Overrides[stage] = reviewer
	}
}

// ConfigPath returns the path to config.yaml.
func ConfigPath(root string) string {
	return filepath.Join(root, ".foreman", "config.yaml")
}

// Load reads and parses config.yaml from the given root.
func Load(root string) (*Config, error) {
	path := ConfigPath(root)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config.yaml: %w", err)
	}
	
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("failed to parse config.yaml: %w", err)
	}
	
	// Ensure reviewers has defaults
	if c.Reviewers.Default == "" {
		c.Reviewers.Default = "auto"
	}
	if c.Reviewers.Overrides == nil {
		c.Reviewers.Overrides = make(map[string]string)
	}
	
	return &c, nil
}

// Save writes the config back to config.yaml.
func Save(root string, c *Config) error {
	path := ConfigPath(root)
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config.yaml: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// NewDefault creates a default config.
func NewDefault(name string) *Config {
	return &Config{
		Name:        name,
		Description: "",
		TechStack:   []string{},
		Created:     time.Now(),
		Reviewers: Reviewers{
			Default:   "auto",
			Overrides: make(map[string]string),
		},
	}
}