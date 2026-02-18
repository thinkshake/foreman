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

	// TDD Instructions (if enabled)
	if cfg.IsTDDEnabled() {
		b.WriteString("## Test-Driven Development\n\n")
		b.WriteString("⚠️ **TDD is enabled for this project.** Follow this workflow:\n\n")
		b.WriteString("1. **Write tests first** — Define expected behavior before implementation\n")
		b.WriteString("2. **Run tests (they should fail)** — Confirm the test is valid\n")
		b.WriteString("3. **Implement the feature** — Write minimal code to pass the test\n")
		b.WriteString("4. **Refactor** — Clean up while keeping tests green\n")
		b.WriteString("5. **Repeat** — For each feature/function\n\n")
		if cfg.Testing.Framework != "" {
			b.WriteString(fmt.Sprintf("**Testing framework:** %s\n\n", cfg.Testing.Framework))
		}
		if cfg.Testing.Required {
			b.WriteString("**⚠️ Tests are required** — Phase cannot be marked complete without passing tests.\n\n")
		}
	}

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
	if cfg.IsTDDEnabled() {
		b.WriteString("- **Write tests first** (TDD enabled)\n")
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

// GenerateQuickBrief creates a streamlined brief for quick mode.
func GenerateQuickBrief(root, task string) (string, error) {
	// Load project config
	cfg, err := config.Load(root)
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	// Read requirements (which contains the task details)
	requirements := project.ReadRequirements(root)

	// Build the brief
	var b strings.Builder

	// Header
	b.WriteString("# Implementation Brief\n\n")
	b.WriteString(fmt.Sprintf("**Project:** %s\n", cfg.Name))
	b.WriteString(fmt.Sprintf("**Generated:** %s\n", getCurrentTimestamp()))

	// Mode indicator
	presetName := config.NormalizePreset(cfg.Preset)
	switch presetName {
	case config.PresetMinimal:
		b.WriteString("**Mode:** Minimal (no gates)\n")
	case config.PresetLight:
		b.WriteString("**Mode:** Light (requirements gate only)\n")
	default:
		b.WriteString("**Mode:** Quick (no design/phases)\n")
	}

	// TDD indicator
	if cfg.IsTDDEnabled() {
		b.WriteString("**Testing:** TDD enabled")
		if cfg.Testing.Framework != "" {
			b.WriteString(fmt.Sprintf(" (%s)", cfg.Testing.Framework))
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Task
	b.WriteString("## Task\n\n")
	b.WriteString(task)
	b.WriteString("\n\n")

	// Requirements
	b.WriteString("## Requirements\n\n")
	b.WriteString(requirements)
	b.WriteString("\n\n")

	// Tech Stack (if specified)
	if len(cfg.TechStack) > 0 {
		b.WriteString("## Tech Stack\n\n")
		for _, tech := range cfg.TechStack {
			b.WriteString(fmt.Sprintf("- %s\n", tech))
		}
		b.WriteString("\n")
	}

	// TDD Instructions (if enabled)
	if cfg.IsTDDEnabled() {
		b.WriteString("## Test-Driven Development\n\n")
		b.WriteString("⚠️ **TDD is enabled for this project.** Follow this workflow:\n\n")
		b.WriteString("1. **Write tests first** — Define expected behavior before implementation\n")
		b.WriteString("2. **Run tests (they should fail)** — Confirm the test is valid\n")
		b.WriteString("3. **Implement the feature** — Write minimal code to pass the test\n")
		b.WriteString("4. **Refactor** — Clean up while keeping tests green\n")
		b.WriteString("5. **Repeat** — For each feature/function\n\n")
		if cfg.Testing.Framework != "" {
			b.WriteString(fmt.Sprintf("**Testing framework:** %s\n\n", cfg.Testing.Framework))
		}
		if cfg.Testing.Required {
			b.WriteString("**⚠️ Tests are required** — Phase cannot be marked complete without passing tests.\n\n")
		}
	}

	// Implementation Guidelines
	b.WriteString("## Implementation Guidelines\n\n")
	switch presetName {
	case config.PresetMinimal:
		b.WriteString("This is a **minimal build** — move fast, ship it.\n\n")
	case config.PresetLight:
		b.WriteString("This is a **light build** — balance speed with quality.\n\n")
	default:
		b.WriteString("This is a **quick build** — focus on getting a working solution.\n\n")
	}
	b.WriteString("- Keep it simple and functional\n")
	b.WriteString("- Write clean, readable code\n")
	if cfg.IsTDDEnabled() {
		b.WriteString("- **Write tests first** (TDD enabled)\n")
	} else {
		b.WriteString("- Include basic tests for core functionality\n")
	}
	b.WriteString("- Add a README with usage instructions\n")
	b.WriteString("\n")

	// Completion
	b.WriteString("## Completion\n\n")
	b.WriteString("When done:\n")
	b.WriteString("- Run `foreman gate implementation` to mark as complete\n")
	b.WriteString("- Ensure the build compiles/runs successfully\n")
	b.WriteString("- All tests pass\n")

	return b.String(), nil
}

// GenerateQuickBriefAndSave creates a quick brief and saves it.
func GenerateQuickBriefAndSave(root, task string) (string, error) {
	brief, err := GenerateQuickBrief(root, task)
	if err != nil {
		return "", err
	}

	briefPath := project.BriefPath(root, "impl")
	if err := os.WriteFile(briefPath, []byte(brief), 0644); err != nil {
		return "", fmt.Errorf("failed to write brief: %w", err)
	}

	return brief, nil
}