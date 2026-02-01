package brief

import (
	"os"
	"strings"
	"testing"

	"github.com/thinkshake/foreman/internal/lane"
	"github.com/thinkshake/foreman/internal/project"
)

func setupProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	_, err := project.Init(dir, "test-project")
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Set up project metadata
	proj, _ := project.Load(dir)
	proj.Description = "A test project"
	proj.Requirements = "Build an awesome thing\nWith multiple features\n"
	proj.Constraints = "Must be fast\nMust be secure\n"
	proj.TechStack = []string{"Go", "PostgreSQL"}
	project.Save(dir, proj)

	// Set design
	os.WriteFile(project.DesignPath(dir), []byte("# Architecture\n\nMicroservices with event sourcing.\n"), 0644)

	return dir
}

func TestGenerate(t *testing.T) {
	root := setupProject(t)

	lane.Add(root, "setup", "Project scaffolding and CI", nil)
	lane.SetStatus(root, "setup", "done")
	lane.Add(root, "core-api", "Core REST endpoints", []string{"setup"})
	lane.SetStatus(root, "core-api", "in-progress")
	lane.SetSpec(root, "core-api", "# Core API\n\nBuild REST endpoints for users and items.\n\n## Endpoints\n- GET /users\n- POST /users\n- GET /items\n")

	output, err := Generate(root, "core-api")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check all expected sections are present
	expectedSections := []string{
		"# Lane Brief: core-api",
		"## Project Context",
		"**Name:** test-project",
		"**Description:** A test project",
		"**Tech Stack:** Go, PostgreSQL",
		"**Constraints:**",
		"Must be fast",
		"## Requirements",
		"Build an awesome thing",
		"## Design Context",
		"Microservices with event sourcing",
		"## Dependencies",
		"**setup**: `done`",
		"## Lane Spec",
		"# Core API",
		"GET /users",
		"## Guidelines",
		"**in-progress**",
		"### Related Lanes",
		"setup (`done`)",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("brief missing expected content: %q", section)
		}
	}
}

func TestGenerateNoDeps(t *testing.T) {
	root := setupProject(t)

	lane.Add(root, "standalone", "A standalone lane", nil)

	output, err := Generate(root, "standalone")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if strings.Contains(output, "## Dependencies") {
		t.Error("brief should not have Dependencies section for lanes with no deps")
	}
}

func TestGenerateNonexistent(t *testing.T) {
	root := setupProject(t)

	_, err := Generate(root, "nonexistent")
	if err == nil {
		t.Fatal("Expected error for nonexistent lane")
	}
}

func TestGenerateWithBlockedDeps(t *testing.T) {
	root := setupProject(t)

	lane.Add(root, "setup", "Setup", nil)
	lane.Add(root, "blocked", "Blocked lane", []string{"setup"})

	output, err := Generate(root, "blocked")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output, "Dependency Warnings") {
		t.Error("brief should warn about non-done dependencies")
	}
	if !strings.Contains(output, "not done yet") {
		t.Error("brief should mention dependency is not done")
	}
}

func TestGenerateWithDoneDeps(t *testing.T) {
	root := setupProject(t)

	lane.Add(root, "setup", "Setup", nil)
	lane.SetStatus(root, "setup", "done")
	lane.Add(root, "next", "Next lane", []string{"setup"})

	output, err := Generate(root, "next")
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(output, "Dependency Warnings") {
		t.Error("brief should not warn when all deps are done")
	}
}
