package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInit(t *testing.T) {
	dir := t.TempDir()

	root, err := Init(dir, "test-project")
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	if root != dir {
		t.Errorf("expected root %s, got %s", dir, root)
	}

	// Check .foreman/ exists
	foremanDir := filepath.Join(dir, ForemanDir)
	if _, err := os.Stat(foremanDir); os.IsNotExist(err) {
		t.Fatal(".foreman/ not created")
	}

	// Check project.yaml
	if _, err := os.Stat(filepath.Join(foremanDir, "project.yaml")); os.IsNotExist(err) {
		t.Fatal("project.yaml not created")
	}

	// Check plan.md
	if _, err := os.Stat(filepath.Join(foremanDir, "plan.md")); os.IsNotExist(err) {
		t.Fatal("plan.md not created")
	}

	// Check design.md
	if _, err := os.Stat(filepath.Join(foremanDir, "design.md")); os.IsNotExist(err) {
		t.Fatal("design.md not created")
	}

	// Check progress.yaml
	if _, err := os.Stat(filepath.Join(foremanDir, "progress.yaml")); os.IsNotExist(err) {
		t.Fatal("progress.yaml not created")
	}

	// Check lanes dir
	if _, err := os.Stat(filepath.Join(foremanDir, "lanes")); os.IsNotExist(err) {
		t.Fatal("lanes/ not created")
	}
}

func TestInitDuplicate(t *testing.T) {
	dir := t.TempDir()

	_, err := Init(dir, "test")
	if err != nil {
		t.Fatalf("First init failed: %v", err)
	}

	_, err = Init(dir, "test")
	if err == nil {
		t.Fatal("Expected error on duplicate init")
	}
}

func TestLoadSave(t *testing.T) {
	dir := t.TempDir()
	_, err := Init(dir, "test-project")
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	proj, err := Load(dir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if proj.Name != "test-project" {
		t.Errorf("expected name 'test-project', got %q", proj.Name)
	}

	proj.Description = "A test project"
	proj.Requirements = "Build something cool"
	proj.TechStack = []string{"Go", "Redis"}

	if err := Save(dir, proj); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	proj2, err := Load(dir)
	if err != nil {
		t.Fatalf("Load after save failed: %v", err)
	}

	if proj2.Description != "A test project" {
		t.Errorf("description not saved")
	}
	if proj2.Requirements != "Build something cool" {
		t.Errorf("requirements not saved")
	}
	if len(proj2.TechStack) != 2 {
		t.Errorf("tech stack not saved")
	}
}

func TestFindRoot(t *testing.T) {
	dir := t.TempDir()
	_, err := Init(dir, "test")
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// From project root
	root, err := FindRoot(dir)
	if err != nil {
		t.Fatalf("FindRoot failed: %v", err)
	}
	if root != dir {
		t.Errorf("expected %s, got %s", dir, root)
	}

	// From subdirectory
	subdir := filepath.Join(dir, "sub", "deep")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatal(err)
	}
	root, err = FindRoot(subdir)
	if err != nil {
		t.Fatalf("FindRoot from subdir failed: %v", err)
	}
	if root != dir {
		t.Errorf("expected %s, got %s", dir, root)
	}
}

func TestFindRootNotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := FindRoot(dir)
	if err == nil {
		t.Fatal("Expected error when no .foreman/ exists")
	}
}

func TestProgressLoadSave(t *testing.T) {
	dir := t.TempDir()
	_, err := Init(dir, "test")
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	prog, err := LoadProgress(dir)
	if err != nil {
		t.Fatalf("LoadProgress failed: %v", err)
	}

	if prog.TotalLanes != 0 {
		t.Errorf("expected 0 total lanes, got %d", prog.TotalLanes)
	}

	prog.TotalLanes = 3
	prog.ByStatus["done"] = 1
	prog.Lanes = append(prog.Lanes, ProgressLane{Name: "test-lane", Status: "done"})

	if err := SaveProgress(dir, prog); err != nil {
		t.Fatalf("SaveProgress failed: %v", err)
	}

	prog2, err := LoadProgress(dir)
	if err != nil {
		t.Fatalf("LoadProgress after save failed: %v", err)
	}

	if prog2.TotalLanes != 3 {
		t.Errorf("expected 3 total lanes, got %d", prog2.TotalLanes)
	}
	if len(prog2.Lanes) != 1 {
		t.Errorf("expected 1 lane, got %d", len(prog2.Lanes))
	}
}
