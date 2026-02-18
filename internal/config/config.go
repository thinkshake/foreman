package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Preset constants
const (
	PresetMinimal = "minimal" // Script, hotfix: 0 gates, 1 phase
	PresetLight   = "light"   // Small tool: 1 gate, 1-2 phases, no design
	PresetFull    = "full"    // Product: 3 gates (requirements, design, phases)
	PresetNightly = "nightly" // Alias for minimal (backward compat)
	PresetProduct = "product" // Alias for full (backward compat)
)

// TestingStyle constants
const (
	TestingStyleTDD      = "tdd"      // Test-driven development
	TestingStyleCoverage = "coverage" // Coverage-focused
	TestingStyleNone     = "none"     // No testing requirements
)

// Testing defines testing configuration for the project.
type Testing struct {
	Style     string `yaml:"style,omitempty"`     // "tdd", "coverage", "none"
	Required  bool   `yaml:"required,omitempty"`  // Block phase completion without tests?
	Framework string `yaml:"framework,omitempty"` // Hint for coding agent (e.g., "vitest", "go test")
	MinCover  int    `yaml:"min_cover,omitempty"` // Minimum coverage percentage (for coverage style)
}

// Config represents the config.yaml schema.
type Config struct {
	Name        string    `yaml:"name"`
	Description string    `yaml:"description"`
	TechStack   []string  `yaml:"tech_stack"`
	Created     time.Time `yaml:"created"`
	Reviewers   Reviewers `yaml:"reviewers"`
	Preset      string    `yaml:"preset,omitempty"`       // minimal, light, full (or aliases: nightly, product)
	AutoAdvance int       `yaml:"auto_advance,omitempty"` // confidence threshold for auto-advance (0-100)
	Testing     *Testing  `yaml:"testing,omitempty"`      // v2.1: testing configuration
	Workflow    []string  `yaml:"workflow,omitempty"`     // v2.1: custom workflow stages (power users)
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
		Preset:      "",
		AutoAdvance: 0,
		Testing:     nil,
		Workflow:    nil,
	}
}

// NewWithPreset creates a config with a specific preset.
func NewWithPreset(name, preset string) *Config {
	cfg := NewDefault(name)
	cfg.Preset = NormalizePreset(preset)

	switch cfg.Preset {
	case PresetMinimal:
		// Minimal: scripts/hotfixes - no gates, straight to implementation
		cfg.AutoAdvance = 100 // Auto-advance everything
		cfg.Reviewers.Default = "auto"
		cfg.Workflow = []string{"requirements", "implementation"}
	case PresetLight:
		// Light: small tools - requirements gate only, no design phase
		cfg.AutoAdvance = 70 // Auto-advance at 70% confidence
		cfg.Reviewers.Default = "auto"
		cfg.Workflow = []string{"requirements", "implementation"}
	case PresetFull:
		// Full: products - all gates, human review for key stages
		cfg.AutoAdvance = 0 // No auto-advance
		cfg.Reviewers.Default = "auto"
		cfg.Reviewers.Overrides = map[string]string{
			"requirements": "human",
			"design":       "human",
		}
		cfg.Workflow = []string{"requirements", "design", "phases", "implementation"}
	}

	return cfg
}

// NewWithTesting creates a config with testing enabled.
func NewWithTesting(name, preset string, tdd bool) *Config {
	cfg := NewWithPreset(name, preset)
	if tdd {
		cfg.Testing = &Testing{
			Style:    TestingStyleTDD,
			Required: false, // Don't block by default
		}
	}
	return cfg
}

// NormalizePreset converts preset aliases to canonical names.
func NormalizePreset(preset string) string {
	switch preset {
	case PresetNightly:
		return PresetMinimal // nightly → minimal
	case PresetProduct:
		return PresetFull // product → full
	case PresetMinimal, PresetLight, PresetFull:
		return preset
	default:
		return preset // allow custom or empty
	}
}

// IsQuickPreset returns true if the preset implies quick mode (no design/phases).
func (c *Config) IsQuickPreset() bool {
	p := NormalizePreset(c.Preset)
	return p == PresetMinimal || p == PresetLight
}

// IsMinimalPreset returns true if using minimal preset (no gates at all).
func (c *Config) IsMinimalPreset() bool {
	return NormalizePreset(c.Preset) == PresetMinimal
}

// GetWorkflow returns the active workflow stages.
func (c *Config) GetWorkflow() []string {
	// Custom workflow takes priority
	if len(c.Workflow) > 0 {
		return c.Workflow
	}

	// Default based on preset
	switch NormalizePreset(c.Preset) {
	case PresetMinimal, PresetLight:
		return []string{"requirements", "implementation"}
	case PresetFull:
		return []string{"requirements", "design", "phases", "implementation"}
	default:
		// Default to full workflow
		return []string{"requirements", "design", "phases", "implementation"}
	}
}

// HasDesignPhase returns true if the workflow includes a design phase.
func (c *Config) HasDesignPhase() bool {
	for _, stage := range c.GetWorkflow() {
		if stage == "design" {
			return true
		}
	}
	return false
}

// HasPhasesPhase returns true if the workflow includes a phases phase.
func (c *Config) HasPhasesPhase() bool {
	for _, stage := range c.GetWorkflow() {
		if stage == "phases" {
			return true
		}
	}
	return false
}

// IsTDDEnabled returns true if TDD style testing is enabled.
func (c *Config) IsTDDEnabled() bool {
	return c.Testing != nil && c.Testing.Style == TestingStyleTDD
}

// ValidateWorkflow checks if a custom workflow is valid.
func ValidateWorkflow(workflow []string) error {
	if len(workflow) == 0 {
		return fmt.Errorf("workflow cannot be empty")
	}

	// Must have at least implementation
	hasImpl := false
	for _, stage := range workflow {
		if stage == "implementation" {
			hasImpl = true
			break
		}
	}
	if !hasImpl {
		return fmt.Errorf("workflow must include 'implementation' stage")
	}

	// Validate all stages
	validStages := map[string]bool{
		"requirements":   true,
		"design":         true,
		"phases":         true,
		"implementation": true,
	}
	for _, stage := range workflow {
		if !validStages[stage] {
			return fmt.Errorf("invalid workflow stage: %s", stage)
		}
	}

	return nil
}