package project

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

const ForemanDir = ".foreman"

// Project represents the project.yaml schema.
type Project struct {
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description"`
	Created      string   `yaml:"created"`
	Updated      string   `yaml:"updated"`
	Requirements string   `yaml:"requirements"`
	Constraints  string   `yaml:"constraints"`
	TechStack    []string `yaml:"tech_stack"`
	Tags         []string `yaml:"tags"`
}

// FindRoot walks up from dir looking for .foreman/.
func FindRoot(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(abs, ForemanDir)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return abs, nil
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			break
		}
		abs = parent
	}
	return "", fmt.Errorf("no .foreman/ directory found (walked up from %s)\nRun 'foreman init' to create a project", dir)
}

// ForemanPath returns the path to .foreman/ for a project root.
func ForemanPath(root string) string {
	return filepath.Join(root, ForemanDir)
}

// ProjectFilePath returns the path to project.yaml.
func ProjectFilePath(root string) string {
	return filepath.Join(ForemanPath(root), "project.yaml")
}

// Load reads and parses project.yaml from the given root.
func Load(root string) (*Project, error) {
	data, err := os.ReadFile(ProjectFilePath(root))
	if err != nil {
		return nil, fmt.Errorf("failed to read project.yaml: %w", err)
	}
	var p Project
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("failed to parse project.yaml: %w", err)
	}
	return &p, nil
}

// Save writes the project back to project.yaml.
func Save(root string, p *Project) error {
	p.Updated = nowString()
	data, err := yaml.Marshal(p)
	if err != nil {
		return fmt.Errorf("failed to marshal project.yaml: %w", err)
	}
	return os.WriteFile(ProjectFilePath(root), data, 0644)
}

// Init creates a new .foreman/ directory with initial files.
func Init(dir, name string) (string, error) {
	foremanDir := filepath.Join(dir, ForemanDir)
	if _, err := os.Stat(foremanDir); err == nil {
		return "", fmt.Errorf(".foreman/ already exists in %s", dir)
	}

	if err := os.MkdirAll(filepath.Join(foremanDir, "lanes"), 0755); err != nil {
		return "", fmt.Errorf("failed to create .foreman/: %w", err)
	}

	now := nowString()

	if name == "" {
		name = filepath.Base(dir)
	}

	p := &Project{
		Name:         name,
		Description:  "",
		Created:      now,
		Updated:      now,
		Requirements: "",
		Constraints:  "",
		TechStack:    []string{},
		Tags:         []string{},
	}

	data, err := yaml.Marshal(p)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(foremanDir, "project.yaml"), data, 0644); err != nil {
		return "", err
	}

	// Create empty plan.md
	if err := os.WriteFile(filepath.Join(foremanDir, "plan.md"), []byte("# Plan\n\n_No plan defined yet. Use `foreman plan set` to add one._\n"), 0644); err != nil {
		return "", err
	}

	// Create empty design.md
	if err := os.WriteFile(filepath.Join(foremanDir, "design.md"), []byte("# Design\n\n_No design defined yet. Use `foreman design set` to add one._\n"), 0644); err != nil {
		return "", err
	}

	// Create empty progress.yaml
	prog := Progress{
		Updated:    now,
		TotalLanes: 0,
		ByStatus: map[string]int{
			"planned":     0,
			"ready":       0,
			"in-progress": 0,
			"review":      0,
			"done":        0,
		},
		Lanes: []ProgressLane{},
	}
	progData, err := yaml.Marshal(&prog)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(foremanDir, "progress.yaml"), progData, 0644); err != nil {
		return "", err
	}

	return dir, nil
}

// Progress represents progress.yaml schema.
type Progress struct {
	Updated    string         `yaml:"updated"`
	TotalLanes int            `yaml:"total_lanes"`
	ByStatus   map[string]int `yaml:"by_status"`
	Lanes      []ProgressLane `yaml:"lanes"`
}

// ProgressLane is a lane entry in progress.yaml.
type ProgressLane struct {
	Name   string `yaml:"name"`
	Status string `yaml:"status"`
}

// LoadProgress reads progress.yaml.
func LoadProgress(root string) (*Progress, error) {
	data, err := os.ReadFile(filepath.Join(ForemanPath(root), "progress.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to read progress.yaml: %w", err)
	}
	var p Progress
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("failed to parse progress.yaml: %w", err)
	}
	return &p, nil
}

// SaveProgress writes progress.yaml.
func SaveProgress(root string, p *Progress) error {
	p.Updated = nowString()
	data, err := yaml.Marshal(p)
	if err != nil {
		return fmt.Errorf("failed to marshal progress.yaml: %w", err)
	}
	return os.WriteFile(filepath.Join(ForemanPath(root), "progress.yaml"), data, 0644)
}

// PlanPath returns path to plan.md.
func PlanPath(root string) string {
	return filepath.Join(ForemanPath(root), "plan.md")
}

// DesignPath returns path to design.md.
func DesignPath(root string) string {
	return filepath.Join(ForemanPath(root), "design.md")
}

func nowString() string {
	return time.Now().Format(time.RFC3339)
}
