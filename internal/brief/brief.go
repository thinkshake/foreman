package brief

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/thinkshake/foreman/internal/config"
	"github.com/thinkshake/foreman/internal/project"
	"github.com/thinkshake/foreman/internal/state"
)

// Generate creates a self-contained brief for a phase.
func Generate(root, phaseName string) (string, error) {
	// Load project config
	cfg, err := config.Load(root)
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	// Load state
	st, err := state.Load(root)
	if err != nil {
		return "", fmt.Errorf("failed to load state: %w", err)
	}

	// Find the target phase
	targetPhase := st.GetPhase(phaseName)
	if targetPhase == nil {
		return "", fmt.Errorf("phase %q not found", phaseName)
	}

	// Read content files
	requirements := project.ReadRequirements(root)
	designs := project.ReadDesigns(root)
	phaseOverview := project.ReadPhaseOverview(root)
	phasePlan := project.ReadPhasePlan(root, phaseName)

	// Build the brief
	var b strings.Builder

	// Header
	b.WriteString(fmt.Sprintf("# Phase Brief: %s\n\n", phaseName))
	b.WriteString(fmt.Sprintf("**Generated:** %s\n", getCurrentTimestamp()))
	b.WriteString(fmt.Sprintf("**Status:** %s\n\n", targetPhase.Status))

	// Project Context
	b.WriteString("## Project Context\n\n")
	b.WriteString(fmt.Sprintf("**Name:** %s\n", cfg.Name))
	if cfg.Description != "" {
		b.WriteString(fmt.Sprintf("**Description:** %s\n", cfg.Description))
	}
	if len(cfg.TechStack) > 0 {
		b.WriteString(fmt.Sprintf("**Tech Stack:** %s\n", strings.Join(cfg.TechStack, ", ")))
	}
	b.WriteString("\n")

	// Requirements
	b.WriteString("## Requirements\n\n")
	b.WriteString(requirements)
	b.WriteString("\n\n")

	// Design Context
	b.WriteString("## Design Context\n\n")
	b.WriteString(designs)
	b.WriteString("\n\n")

	// Phase Overview
	b.WriteString("## Phase Overview\n\n")
	b.WriteString(phaseOverview)
	b.WriteString("\n\n")

	// Dependencies - show status of preceding phases
	b.WriteString("## Dependencies\n\n")
	precedingPhases := getPrecedingPhases(st.Phases, phaseName)
	if len(precedingPhases) > 0 {
		for _, phase := range precedingPhases {
			status := phase.Status
			indicator := getStatusIndicator(status)
			b.WriteString(fmt.Sprintf("- **%s**: %s `%s`\n", phase.Name, indicator, status))
		}
		
		// Check for blockers
		hasBlockers := false
		for _, phase := range precedingPhases {
			if phase.Status != "done" {
				if !hasBlockers {
					b.WriteString("\n### ⚠️  Dependency Warnings\n\n")
					hasBlockers = true
				}
				b.WriteString(fmt.Sprintf("- Phase **%s** is `%s` (not done yet)\n", phase.Name, phase.Status))
			}
		}
	} else {
		b.WriteString("_No dependencies - this is the first phase._\n")
	}
	b.WriteString("\n")

	// This Phase Spec
	b.WriteString(fmt.Sprintf("## Phase Spec: %s\n\n", phaseName))
	b.WriteString(phasePlan)
	b.WriteString("\n\n")

	// Implementation Guidelines
	b.WriteString("## Implementation Guidelines\n\n")
	b.WriteString(fmt.Sprintf("- This phase is currently: **%s**\n", targetPhase.Status))
	if targetPhase.Status == "planned" {
		b.WriteString("- Ready to start implementation\n")
	} else if targetPhase.Status == "in-progress" {
		b.WriteString("- Implementation is ongoing\n")
	} else if targetPhase.Status == "done" {
		b.WriteString("- This phase is marked as completed\n")
	}

	// Show context about other phases for awareness
	b.WriteString("\n### Related Phases\n\n")
	for _, phase := range st.Phases {
		if phase.Name == phaseName {
			continue
		}
		indicator := getStatusIndicator(phase.Status)
		b.WriteString(fmt.Sprintf("- %s %s (`%s`)\n", indicator, phase.Name, phase.Status))
	}

	// Completion criteria
	b.WriteString("\n### Completion\n\n")
	b.WriteString("When this phase is complete:\n")
	b.WriteString(fmt.Sprintf("- Run `foreman phase %s done` to mark it as finished\n", phaseName))
	b.WriteString("- Ensure all deliverables are implemented and tested\n")
	b.WriteString("- Document any changes or decisions made during implementation\n")

	return b.String(), nil
}

// GenerateAndSave creates a brief and saves it to the briefs directory.
func GenerateAndSave(root, phaseName string) (string, error) {
	brief, err := Generate(root, phaseName)
	if err != nil {
		return "", err
	}

	briefPath := project.BriefPath(root, phaseName)
	if err := os.WriteFile(briefPath, []byte(brief), 0644); err != nil {
		return "", fmt.Errorf("failed to write brief: %w", err)
	}

	return brief, nil
}

// getPrecedingPhases returns all phases that come before the target phase.
func getPrecedingPhases(phases []state.Phase, targetName string) []state.Phase {
	var preceding []state.Phase
	
	for _, phase := range phases {
		if phase.Name == targetName {
			break
		}
		preceding = append(preceding, phase)
	}
	
	return preceding
}

// getStatusIndicator returns a visual indicator for phase status.
func getStatusIndicator(status string) string {
	switch status {
	case "done":
		return "✅"
	case "in-progress":
		return "🔵"
	case "planned":
		return "⬜"
	default:
		return "❓"
	}
}

// getCurrentTimestamp returns the current timestamp for brief generation.
func getCurrentTimestamp() string {
	return time.Now().Format(time.RFC3339)
}