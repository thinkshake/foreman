package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewDefault(t *testing.T) {
	cfg := NewDefault("test-project")
	
	if cfg.Name != "test-project" {
		t.Errorf("expected name 'test-project', got '%s'", cfg.Name)
	}
	
	if cfg.Description != "" {
		t.Errorf("expected empty description, got '%s'", cfg.Description)
	}
	
	if len(cfg.TechStack) != 0 {
		t.Errorf("expected empty tech stack, got %v", cfg.TechStack)
	}
	
	if cfg.Reviewers.Default != "auto" {
		t.Errorf("expected default reviewer 'auto', got '%s'", cfg.Reviewers.Default)
	}
	
	if cfg.Reviewers.Overrides == nil {
		t.Error("expected overrides map to be initialized")
	}
}

func TestReviewersMethods(t *testing.T) {
	reviewers := &Reviewers{
		Default:   "auto",
		Overrides: make(map[string]string),
	}
	
	// Test default reviewer
	if got := reviewers.GetReviewer("requirements"); got != "auto" {
		t.Errorf("expected 'auto', got '%s'", got)
	}
	
	// Test setting override
	reviewers.SetReviewer("requirements", "human")
	if got := reviewers.GetReviewer("requirements"); got != "human" {
		t.Errorf("expected 'human', got '%s'", got)
	}
	
	// Test removing override by setting to default
	reviewers.SetReviewer("requirements", "auto")
	if _, exists := reviewers.Overrides["requirements"]; exists {
		t.Error("expected override to be removed when set to default")
	}
	
	if got := reviewers.GetReviewer("requirements"); got != "auto" {
		t.Errorf("expected 'auto', got '%s'", got)
	}
}

func TestNewWithPreset(t *testing.T) {
	tests := []struct {
		name         string
		preset       string
		wantPreset   string
		wantWorkflow []string
		wantAdvance  int
	}{
		{
			name:         "minimal preset",
			preset:       PresetMinimal,
			wantPreset:   PresetMinimal,
			wantWorkflow: []string{"requirements", "implementation"},
			wantAdvance:  100,
		},
		{
			name:         "light preset",
			preset:       PresetLight,
			wantPreset:   PresetLight,
			wantWorkflow: []string{"requirements", "implementation"},
			wantAdvance:  70,
		},
		{
			name:         "full preset",
			preset:       PresetFull,
			wantPreset:   PresetFull,
			wantWorkflow: []string{"requirements", "design", "phases", "implementation"},
			wantAdvance:  0,
		},
		{
			name:         "nightly alias",
			preset:       PresetNightly,
			wantPreset:   PresetMinimal,
			wantWorkflow: []string{"requirements", "implementation"},
			wantAdvance:  100,
		},
		{
			name:         "product alias",
			preset:       PresetProduct,
			wantPreset:   PresetFull,
			wantWorkflow: []string{"requirements", "design", "phases", "implementation"},
			wantAdvance:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewWithPreset("test", tt.preset)

			if cfg.Preset != tt.wantPreset {
				t.Errorf("preset = %q, want %q", cfg.Preset, tt.wantPreset)
			}

			if cfg.AutoAdvance != tt.wantAdvance {
				t.Errorf("auto_advance = %d, want %d", cfg.AutoAdvance, tt.wantAdvance)
			}

			if len(cfg.Workflow) != len(tt.wantWorkflow) {
				t.Errorf("workflow length = %d, want %d", len(cfg.Workflow), len(tt.wantWorkflow))
			}

			for i, stage := range cfg.Workflow {
				if stage != tt.wantWorkflow[i] {
					t.Errorf("workflow[%d] = %q, want %q", i, stage, tt.wantWorkflow[i])
				}
			}
		})
	}
}

func TestNewWithTesting(t *testing.T) {
	cfg := NewWithTesting("test", PresetLight, true)

	if cfg.Testing == nil {
		t.Fatal("expected testing config to be set")
	}

	if cfg.Testing.Style != TestingStyleTDD {
		t.Errorf("testing style = %q, want %q", cfg.Testing.Style, TestingStyleTDD)
	}

	if cfg.Testing.Required {
		t.Error("expected required = false by default")
	}
}

func TestIsQuickPreset(t *testing.T) {
	tests := []struct {
		preset    string
		wantQuick bool
	}{
		{PresetMinimal, true},
		{PresetLight, true},
		{PresetFull, false},
		{PresetNightly, true},
		{PresetProduct, false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.preset, func(t *testing.T) {
			cfg := &Config{Preset: tt.preset}
			if got := cfg.IsQuickPreset(); got != tt.wantQuick {
				t.Errorf("IsQuickPreset() = %v, want %v", got, tt.wantQuick)
			}
		})
	}
}

func TestIsTDDEnabled(t *testing.T) {
	tests := []struct {
		name    string
		testing *Testing
		want    bool
	}{
		{"nil testing", nil, false},
		{"tdd style", &Testing{Style: TestingStyleTDD}, true},
		{"coverage style", &Testing{Style: TestingStyleCoverage}, false},
		{"none style", &Testing{Style: TestingStyleNone}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Testing: tt.testing}
			if got := cfg.IsTDDEnabled(); got != tt.want {
				t.Errorf("IsTDDEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateWorkflow(t *testing.T) {
	tests := []struct {
		name     string
		workflow []string
		wantErr  bool
	}{
		{"valid full", []string{"requirements", "design", "phases", "implementation"}, false},
		{"valid minimal", []string{"requirements", "implementation"}, false},
		{"valid impl only", []string{"implementation"}, false},
		{"empty workflow", []string{}, true},
		{"missing impl", []string{"requirements", "design"}, true},
		{"invalid stage", []string{"requirements", "testing", "implementation"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWorkflow(tt.workflow)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateWorkflow() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetWorkflow(t *testing.T) {
	// Config with custom workflow
	cfg := &Config{
		Workflow: []string{"requirements", "implementation"},
	}
	got := cfg.GetWorkflow()
	if len(got) != 2 || got[0] != "requirements" || got[1] != "implementation" {
		t.Errorf("GetWorkflow() with custom = %v", got)
	}

	// Config with preset, no custom workflow
	cfg2 := &Config{
		Preset:   PresetFull,
		Workflow: nil,
	}
	got2 := cfg2.GetWorkflow()
	if len(got2) != 4 {
		t.Errorf("GetWorkflow() with full preset = %v, want 4 stages", got2)
	}
}

func TestLoadSave(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "foreman-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create .foreman directory
	foremanDir := filepath.Join(tempDir, ".foreman")
	if err := os.MkdirAll(foremanDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	// Create and save config
	cfg := &Config{
		Name:        "test-project",
		Description: "A test project",
		TechStack:   []string{"Go", "YAML"},
		Created:     time.Now(),
		Reviewers: Reviewers{
			Default: "human",
			Overrides: map[string]string{
				"requirements": "auto",
			},
		},
	}
	
	if err := Save(tempDir, cfg); err != nil {
		t.Fatal(err)
	}
	
	// Load config back
	loaded, err := Load(tempDir)
	if err != nil {
		t.Fatal(err)
	}
	
	// Verify loaded config
	if loaded.Name != cfg.Name {
		t.Errorf("expected name '%s', got '%s'", cfg.Name, loaded.Name)
	}
	
	if loaded.Description != cfg.Description {
		t.Errorf("expected description '%s', got '%s'", cfg.Description, loaded.Description)
	}
	
	if len(loaded.TechStack) != len(cfg.TechStack) {
		t.Errorf("expected tech stack length %d, got %d", len(cfg.TechStack), len(loaded.TechStack))
	}
	
	if loaded.Reviewers.Default != cfg.Reviewers.Default {
		t.Errorf("expected default reviewer '%s', got '%s'", cfg.Reviewers.Default, loaded.Reviewers.Default)
	}
	
	if loaded.Reviewers.GetReviewer("requirements") != "auto" {
		t.Errorf("expected requirements reviewer 'auto', got '%s'", loaded.Reviewers.GetReviewer("requirements"))
	}
}