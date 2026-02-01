package lane

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/thinkshake/foreman/internal/project"
	"gopkg.in/yaml.v3"
)

// ValidStatuses defines allowed lane statuses.
var ValidStatuses = []string{"planned", "ready", "in-progress", "review", "done"}

// Lane represents a lane YAML file.
type Lane struct {
	Name         string   `yaml:"name"`
	Order        int      `yaml:"order"`
	Status       string   `yaml:"status"`
	Summary      string   `yaml:"summary"`
	Dependencies []string `yaml:"dependencies"`
	Created      string   `yaml:"created"`
	Updated      string   `yaml:"updated"`
}

// IsValidStatus checks if a status string is valid.
func IsValidStatus(s string) bool {
	for _, v := range ValidStatuses {
		if v == s {
			return true
		}
	}
	return false
}

// LanesDir returns the path to the lanes directory.
func LanesDir(root string) string {
	return filepath.Join(project.ForemanPath(root), "lanes")
}

// ListAll reads all lane YAML files and returns sorted lanes.
func ListAll(root string) ([]Lane, error) {
	dir := LanesDir(root)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read lanes directory: %w", err)
	}

	var lanes []Lane
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", e.Name(), err)
		}
		var l Lane
		if err := yaml.Unmarshal(data, &l); err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", e.Name(), err)
		}
		lanes = append(lanes, l)
	}

	sort.Slice(lanes, func(i, j int) bool {
		return lanes[i].Order < lanes[j].Order
	})

	return lanes, nil
}

// FindByName finds a lane by name.
func FindByName(root, name string) (*Lane, error) {
	lanes, err := ListAll(root)
	if err != nil {
		return nil, err
	}
	for _, l := range lanes {
		if l.Name == name {
			return &l, nil
		}
	}
	return nil, fmt.Errorf("lane %q not found", name)
}

// nextOrder determines the next order number.
func nextOrder(root string) (int, error) {
	lanes, err := ListAll(root)
	if err != nil {
		return 0, err
	}
	if len(lanes) == 0 {
		return 1, nil
	}
	max := 0
	for _, l := range lanes {
		if l.Order > max {
			max = l.Order
		}
	}
	return max + 1, nil
}

// laneFilePrefix returns the file prefix like "01-setup".
func laneFilePrefix(order int, name string) string {
	return fmt.Sprintf("%02d-%s", order, name)
}

// YAMLPath returns the YAML file path for a lane.
func YAMLPath(root string, order int, name string) string {
	return filepath.Join(LanesDir(root), laneFilePrefix(order, name)+".yaml")
}

// MDPath returns the Markdown file path for a lane.
func MDPath(root string, order int, name string) string {
	return filepath.Join(LanesDir(root), laneFilePrefix(order, name)+".md")
}

// FindMDPath finds the MD path for a lane by name.
func FindMDPath(root, name string) (string, error) {
	l, err := FindByName(root, name)
	if err != nil {
		return "", err
	}
	return MDPath(root, l.Order, l.Name), nil
}

// Add creates a new lane.
func Add(root, name, summary string, deps []string) (*Lane, error) {
	// Validate name
	if name == "" {
		return nil, fmt.Errorf("lane name cannot be empty")
	}
	if strings.Contains(name, " ") {
		return nil, fmt.Errorf("lane name cannot contain spaces (use hyphens)")
	}

	// Check for duplicates
	existing, err := ListAll(root)
	if err != nil {
		return nil, err
	}
	for _, l := range existing {
		if l.Name == name {
			return nil, fmt.Errorf("lane %q already exists", name)
		}
	}

	// Validate dependencies exist
	if len(deps) > 0 {
		existingNames := make(map[string]bool)
		for _, l := range existing {
			existingNames[l.Name] = true
		}
		for _, d := range deps {
			if !existingNames[d] {
				return nil, fmt.Errorf("dependency %q does not exist", d)
			}
		}
	}

	order, err := nextOrder(root)
	if err != nil {
		return nil, err
	}

	now := time.Now().Format(time.RFC3339)
	l := &Lane{
		Name:         name,
		Order:        order,
		Status:       "planned",
		Summary:      summary,
		Dependencies: deps,
		Created:      now,
		Updated:      now,
	}

	// Write YAML
	data, err := yaml.Marshal(l)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(YAMLPath(root, order, name), data, 0644); err != nil {
		return nil, err
	}

	// Write initial MD
	md := fmt.Sprintf("# Lane: %s\n\n%s\n\n_Use `foreman lane set %s` to add the detailed spec._\n", name, summary, name)
	if err := os.WriteFile(MDPath(root, order, name), []byte(md), 0644); err != nil {
		return nil, err
	}

	// Update progress
	if err := UpdateProgress(root); err != nil {
		return nil, err
	}

	return l, nil
}

// SetStatus updates a lane's status.
func SetStatus(root, name, status string) error {
	if !IsValidStatus(status) {
		return fmt.Errorf("invalid status %q (valid: %s)", status, strings.Join(ValidStatuses, ", "))
	}

	l, err := FindByName(root, name)
	if err != nil {
		return err
	}

	l.Status = status
	l.Updated = time.Now().Format(time.RFC3339)

	data, err := yaml.Marshal(l)
	if err != nil {
		return err
	}
	if err := os.WriteFile(YAMLPath(root, l.Order, l.Name), data, 0644); err != nil {
		return err
	}

	return UpdateProgress(root)
}

// SetSpec writes the lane's markdown spec.
func SetSpec(root, name, content string) error {
	mdPath, err := FindMDPath(root, name)
	if err != nil {
		return err
	}
	return os.WriteFile(mdPath, []byte(content), 0644)
}

// GetSpec reads the lane's markdown spec.
func GetSpec(root, name string) (string, error) {
	mdPath, err := FindMDPath(root, name)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(mdPath)
	if err != nil {
		return "", fmt.Errorf("failed to read lane spec: %w", err)
	}
	return string(data), nil
}

// UpdateProgress recalculates and writes progress.yaml.
func UpdateProgress(root string) error {
	lanes, err := ListAll(root)
	if err != nil {
		return err
	}

	byStatus := map[string]int{
		"planned":     0,
		"ready":       0,
		"in-progress": 0,
		"review":      0,
		"done":        0,
	}

	var progressLanes []project.ProgressLane
	for _, l := range lanes {
		byStatus[l.Status]++
		progressLanes = append(progressLanes, project.ProgressLane{
			Name:   l.Name,
			Status: l.Status,
		})
	}

	p := &project.Progress{
		TotalLanes: len(lanes),
		ByStatus:   byStatus,
		Lanes:      progressLanes,
	}

	return project.SaveProgress(root, p)
}

// ParseOrderFromFilename extracts the order number from a lane filename.
func ParseOrderFromFilename(filename string) (int, error) {
	base := filepath.Base(filename)
	parts := strings.SplitN(base, "-", 2)
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid lane filename: %s", filename)
	}
	n, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid order number in %s: %w", filename, err)
	}
	return n, nil
}
