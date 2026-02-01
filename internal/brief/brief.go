package brief

import (
	"fmt"
	"os"
	"strings"

	"github.com/thinkshake/foreman/internal/lane"
	"github.com/thinkshake/foreman/internal/project"
)

// Generate creates a self-contained brief for a coding agent.
func Generate(root, laneName string) (string, error) {
	// Load project
	proj, err := project.Load(root)
	if err != nil {
		return "", fmt.Errorf("failed to load project: %w", err)
	}

	// Load the target lane
	targetLane, err := lane.FindByName(root, laneName)
	if err != nil {
		return "", err
	}

	// Load lane spec
	spec, err := lane.GetSpec(root, laneName)
	if err != nil {
		return "", err
	}

	// Load design
	design, err := os.ReadFile(project.DesignPath(root))
	if err != nil {
		design = []byte("_No design document found._")
	}

	// Load dependency info
	allLanes, err := lane.ListAll(root)
	if err != nil {
		return "", err
	}

	laneMap := make(map[string]*lane.Lane)
	for i := range allLanes {
		laneMap[allLanes[i].Name] = &allLanes[i]
	}

	// Build the brief
	var b strings.Builder

	// Header
	b.WriteString(fmt.Sprintf("# Lane Brief: %s\n\n", laneName))

	// Project Context
	b.WriteString("## Project Context\n\n")
	b.WriteString(fmt.Sprintf("**Name:** %s\n", proj.Name))
	if proj.Description != "" {
		b.WriteString(fmt.Sprintf("**Description:** %s\n", proj.Description))
	}
	if len(proj.TechStack) > 0 {
		b.WriteString(fmt.Sprintf("**Tech Stack:** %s\n", strings.Join(proj.TechStack, ", ")))
	}
	if proj.Constraints != "" {
		b.WriteString(fmt.Sprintf("\n**Constraints:**\n%s\n", strings.TrimSpace(proj.Constraints)))
	}
	b.WriteString("\n")

	// Requirements Context
	if proj.Requirements != "" {
		b.WriteString("## Requirements\n\n")
		b.WriteString(strings.TrimSpace(proj.Requirements))
		b.WriteString("\n\n")
	}

	// Design Context
	designStr := strings.TrimSpace(string(design))
	if designStr != "" && !strings.Contains(designStr, "No design defined yet") {
		b.WriteString("## Design Context\n\n")
		b.WriteString(designStr)
		b.WriteString("\n\n")
	}

	// Dependencies
	if len(targetLane.Dependencies) > 0 {
		b.WriteString("## Dependencies\n\n")
		for _, dep := range targetLane.Dependencies {
			if depLane, ok := laneMap[dep]; ok {
				b.WriteString(fmt.Sprintf("- **%s**: `%s` — %s\n", dep, depLane.Status, depLane.Summary))
			} else {
				b.WriteString(fmt.Sprintf("- **%s**: _unknown_\n", dep))
			}
		}
		b.WriteString("\n")
	}

	// Lane Spec
	b.WriteString("## Lane Spec\n\n")
	b.WriteString(strings.TrimSpace(spec))
	b.WriteString("\n\n")

	// Guidelines
	b.WriteString("## Guidelines\n\n")
	b.WriteString(fmt.Sprintf("- This lane is currently: **%s**\n", targetLane.Status))
	b.WriteString(fmt.Sprintf("- Lane summary: %s\n", targetLane.Summary))

	// Show dependency status warnings
	hasBlockers := false
	for _, dep := range targetLane.Dependencies {
		if depLane, ok := laneMap[dep]; ok {
			if depLane.Status != "done" {
				if !hasBlockers {
					b.WriteString("\n### ⚠️  Dependency Warnings\n\n")
					hasBlockers = true
				}
				b.WriteString(fmt.Sprintf("- Lane **%s** is `%s` (not done yet)\n", dep, depLane.Status))
			}
		}
	}

	// Context about other lanes for awareness
	b.WriteString("\n### Related Lanes\n\n")
	for _, l := range allLanes {
		if l.Name == laneName {
			continue
		}
		b.WriteString(fmt.Sprintf("- %s (`%s`): %s\n", l.Name, l.Status, l.Summary))
	}

	return b.String(), nil
}
