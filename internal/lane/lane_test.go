package lane

import (
	"testing"

	"github.com/thinkshake/foreman/internal/project"
)

func setupProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	_, err := project.Init(dir, "test-project")
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	return dir
}

func TestAddLane(t *testing.T) {
	root := setupProject(t)

	l, err := Add(root, "setup", "Project scaffolding", nil)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if l.Name != "setup" {
		t.Errorf("expected name 'setup', got %q", l.Name)
	}
	if l.Order != 1 {
		t.Errorf("expected order 1, got %d", l.Order)
	}
	if l.Status != "planned" {
		t.Errorf("expected status 'planned', got %q", l.Status)
	}
	if l.Summary != "Project scaffolding" {
		t.Errorf("expected summary 'Project scaffolding', got %q", l.Summary)
	}
}

func TestAutoNumbering(t *testing.T) {
	root := setupProject(t)

	l1, err := Add(root, "first", "First lane", nil)
	if err != nil {
		t.Fatal(err)
	}
	if l1.Order != 1 {
		t.Errorf("expected order 1, got %d", l1.Order)
	}

	l2, err := Add(root, "second", "Second lane", nil)
	if err != nil {
		t.Fatal(err)
	}
	if l2.Order != 2 {
		t.Errorf("expected order 2, got %d", l2.Order)
	}

	l3, err := Add(root, "third", "Third lane", nil)
	if err != nil {
		t.Fatal(err)
	}
	if l3.Order != 3 {
		t.Errorf("expected order 3, got %d", l3.Order)
	}
}

func TestDuplicateLaneName(t *testing.T) {
	root := setupProject(t)

	_, err := Add(root, "setup", "First", nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Add(root, "setup", "Duplicate", nil)
	if err == nil {
		t.Fatal("Expected error on duplicate lane name")
	}
}

func TestListAll(t *testing.T) {
	root := setupProject(t)

	Add(root, "alpha", "Alpha lane", nil)
	Add(root, "beta", "Beta lane", nil)
	Add(root, "gamma", "Gamma lane", nil)

	lanes, err := ListAll(root)
	if err != nil {
		t.Fatal(err)
	}

	if len(lanes) != 3 {
		t.Fatalf("expected 3 lanes, got %d", len(lanes))
	}

	// Should be sorted by order
	if lanes[0].Name != "alpha" || lanes[1].Name != "beta" || lanes[2].Name != "gamma" {
		t.Errorf("lanes not in correct order: %v", lanes)
	}
}

func TestSetStatus(t *testing.T) {
	root := setupProject(t)
	Add(root, "setup", "Setup", nil)

	if err := SetStatus(root, "setup", "in-progress"); err != nil {
		t.Fatal(err)
	}

	l, err := FindByName(root, "setup")
	if err != nil {
		t.Fatal(err)
	}
	if l.Status != "in-progress" {
		t.Errorf("expected 'in-progress', got %q", l.Status)
	}
}

func TestInvalidStatus(t *testing.T) {
	root := setupProject(t)
	Add(root, "setup", "Setup", nil)

	if err := SetStatus(root, "setup", "invalid"); err == nil {
		t.Fatal("Expected error on invalid status")
	}
}

func TestSetAndGetSpec(t *testing.T) {
	root := setupProject(t)
	Add(root, "setup", "Setup", nil)

	content := "# Setup Lane\n\nDetailed spec here.\n"
	if err := SetSpec(root, "setup", content); err != nil {
		t.Fatal(err)
	}

	spec, err := GetSpec(root, "setup")
	if err != nil {
		t.Fatal(err)
	}
	if spec != content {
		t.Errorf("spec mismatch:\nexpected: %q\ngot: %q", content, spec)
	}
}

func TestDependencies(t *testing.T) {
	root := setupProject(t)

	Add(root, "setup", "Setup", nil)
	l, err := Add(root, "core", "Core", []string{"setup"})
	if err != nil {
		t.Fatal(err)
	}

	if len(l.Dependencies) != 1 || l.Dependencies[0] != "setup" {
		t.Errorf("expected deps [setup], got %v", l.Dependencies)
	}
}

func TestDependencyValidation(t *testing.T) {
	root := setupProject(t)

	_, err := Add(root, "core", "Core", []string{"nonexistent"})
	if err == nil {
		t.Fatal("Expected error for nonexistent dependency")
	}
}

func TestUpdateProgress(t *testing.T) {
	root := setupProject(t)

	Add(root, "setup", "Setup", nil)
	Add(root, "core", "Core", nil)
	SetStatus(root, "setup", "done")
	SetStatus(root, "core", "in-progress")

	prog, err := project.LoadProgress(root)
	if err != nil {
		t.Fatal(err)
	}

	if prog.TotalLanes != 2 {
		t.Errorf("expected 2 total lanes, got %d", prog.TotalLanes)
	}
	if prog.ByStatus["done"] != 1 {
		t.Errorf("expected 1 done, got %d", prog.ByStatus["done"])
	}
	if prog.ByStatus["in-progress"] != 1 {
		t.Errorf("expected 1 in-progress, got %d", prog.ByStatus["in-progress"])
	}
}

func TestFindByNameNotFound(t *testing.T) {
	root := setupProject(t)
	_, err := FindByName(root, "nonexistent")
	if err == nil {
		t.Fatal("Expected error for nonexistent lane")
	}
}

func TestEmptyLaneName(t *testing.T) {
	root := setupProject(t)
	_, err := Add(root, "", "Empty", nil)
	if err == nil {
		t.Fatal("Expected error for empty lane name")
	}
}

func TestSpaceInLaneName(t *testing.T) {
	root := setupProject(t)
	_, err := Add(root, "has space", "Bad name", nil)
	if err == nil {
		t.Fatal("Expected error for lane name with spaces")
	}
}

func TestIsValidStatus(t *testing.T) {
	tests := []struct {
		status string
		valid  bool
	}{
		{"planned", true},
		{"ready", true},
		{"in-progress", true},
		{"review", true},
		{"done", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := IsValidStatus(tt.status); got != tt.valid {
			t.Errorf("IsValidStatus(%q) = %v, want %v", tt.status, got, tt.valid)
		}
	}
}
