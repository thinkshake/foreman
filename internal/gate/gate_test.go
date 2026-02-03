package gate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/thinkshake/foreman/internal/state"
)

func setupTestProject(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "foreman-test")
	if err != nil {
		t.Fatal(err)
	}

	// Create .foreman directory structure
	foremanDir := filepath.Join(tempDir, ".foreman")
	if err := os.MkdirAll(foremanDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create subdirectories
	for _, dir := range []string{"designs", "phases", "briefs"} {
		if err := os.MkdirAll(filepath.Join(foremanDir, dir), 0755); err != nil {
			t.Fatal(err)
		}
	}

	return tempDir
}

func TestValidateRequirements(t *testing.T) {
	root := setupTestProject(t)
	defer os.RemoveAll(root)

	// Test missing requirements file
	result := ValidateRequirements(root)
	if result.Passed {
		t.Error("expected validation to fail with missing requirements.md")
	}

	// Test empty requirements file
	reqPath := filepath.Join(root, ".foreman", "requirements.md")
	if err := os.WriteFile(reqPath, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	result = ValidateRequirements(root)
	if result.Passed {
		t.Error("expected validation to fail with empty requirements.md")
	}

	// Test file with placeholder content
	placeholder := "# Requirements\n\n_No requirements defined yet._"
	if err := os.WriteFile(reqPath, []byte(placeholder), 0644); err != nil {
		t.Fatal(err)
	}

	result = ValidateRequirements(root)
	if result.Passed {
		t.Error("expected validation to fail with placeholder content")
	}

	// Test valid requirements file
	validReq := `# Requirements

## Goal
Build a CLI tool for project management.

## Features
- Stage-based workflow
- Gate validation
- Phase tracking

## Constraints
- Must be written in Go
- Must use Cobra for CLI

## Success Criteria
- All tests pass
- Tool can manage a complete project lifecycle`

	if err := os.WriteFile(reqPath, []byte(validReq), 0644); err != nil {
		t.Fatal(err)
	}

	result = ValidateRequirements(root)
	if !result.Passed {
		t.Errorf("expected validation to pass with valid requirements: %s", result.Message)
	}
}

func TestValidateDesign(t *testing.T) {
	root := setupTestProject(t)
	defer os.RemoveAll(root)

	// Test missing designs directory (already created in setup)
	result := ValidateDesign(root)
	if result.Passed {
		t.Error("expected validation to fail with empty designs directory")
	}

	// Test with non-markdown files
	designsDir := filepath.Join(root, ".foreman", "designs")
	if err := os.WriteFile(filepath.Join(designsDir, "readme.txt"), []byte("not markdown"), 0644); err != nil {
		t.Fatal(err)
	}

	result = ValidateDesign(root)
	if result.Passed {
		t.Error("expected validation to fail with no markdown files")
	}

	// Test with empty markdown file
	if err := os.WriteFile(filepath.Join(designsDir, "architecture.md"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	result = ValidateDesign(root)
	if result.Passed {
		t.Error("expected validation to fail with empty markdown file")
	}

	// Test with valid design files
	archContent := `# Architecture

## Overview
The system uses a CLI-based architecture with multiple stages.

## Components
- Config management
- State tracking  
- Gate validation
- Brief compilation`

	apiContent := `# API Design

## Commands
- init: Initialize project
- status: Show current state
- gate: Validate stages
- phase: Manage phases`

	if err := os.WriteFile(filepath.Join(designsDir, "architecture.md"), []byte(archContent), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(designsDir, "api.md"), []byte(apiContent), 0644); err != nil {
		t.Fatal(err)
	}

	result = ValidateDesign(root)
	if !result.Passed {
		t.Errorf("expected validation to pass with valid design files: %s", result.Message)
	}
}

func TestValidatePhases(t *testing.T) {
	root := setupTestProject(t)
	defer os.RemoveAll(root)

	phasesDir := filepath.Join(root, ".foreman", "phases")

	// Test missing overview.md
	result := ValidatePhases(root)
	if result.Passed {
		t.Error("expected validation to fail with missing overview.md")
	}

	// Test with empty overview.md
	overviewPath := filepath.Join(phasesDir, "overview.md")
	if err := os.WriteFile(overviewPath, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	result = ValidatePhases(root)
	if result.Passed {
		t.Error("expected validation to fail with empty overview.md")
	}

	// Test with valid overview but no phase files
	overviewContent := `# Implementation Phases

## Strategy
Break implementation into manageable phases that can be executed independently.

## Phase Breakdown
1. Setup - Project structure and basic CLI
2. Core - State management and validation
3. Commands - Implement all CLI commands
4. Testing - Comprehensive test coverage`

	if err := os.WriteFile(overviewPath, []byte(overviewContent), 0644); err != nil {
		t.Fatal(err)
	}

	result = ValidatePhases(root)
	if result.Passed {
		t.Error("expected validation to fail with no phase files")
	}

	// Test with valid phase files
	phase1Content := `# Phase 1: Setup

## Objectives
- Set up Go module and dependencies
- Create basic CLI structure with Cobra
- Implement project initialization

## Deliverables
- go.mod with dependencies
- Basic CLI structure
- init command working`

	phase2Content := `# Phase 2: Core Logic

## Objectives  
- Implement config and state management
- Create gate validation system
- Build brief compilation

## Deliverables
- Config/state loading and saving
- Gate validation for all stages
- Brief generation for phases`

	if err := os.WriteFile(filepath.Join(phasesDir, "1-setup.md"), []byte(phase1Content), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(phasesDir, "2-core.md"), []byte(phase2Content), 0644); err != nil {
		t.Fatal(err)
	}

	result = ValidatePhases(root)
	if !result.Passed {
		t.Errorf("expected validation to pass with valid phase files: %s", result.Message)
	}
}

func TestValidateImplementation(t *testing.T) {
	root := setupTestProject(t)
	defer os.RemoveAll(root)

	st := state.NewDefault()

	// Test with no phases
	result := ValidateImplementation(root, st)
	if result.Passed {
		t.Error("expected validation to fail with no phases")
	}

	// Test with incomplete phases
	st.AddPhase("1-setup")
	st.AddPhase("2-core")
	st.SetPhaseStatus("1-setup", "done")
	// Leave 2-core as "planned"

	result = ValidateImplementation(root, st)
	if result.Passed {
		t.Error("expected validation to fail with incomplete phases")
	}

	// Test with all phases done
	st.SetPhaseStatus("2-core", "done")

	result = ValidateImplementation(root, st)
	if !result.Passed {
		t.Errorf("expected validation to pass with all phases done: %s", result.Message)
	}
}

func TestValidateStage(t *testing.T) {
	root := setupTestProject(t)
	defer os.RemoveAll(root)

	st := state.NewDefault()

	// Test invalid stage
	_, err := ValidateStage(root, "invalid", st)
	if err == nil {
		t.Error("expected error for invalid stage")
	}

	// Test valid stages
	validStages := []string{"requirements", "design", "phases", "implementation"}
	for _, stage := range validStages {
		result, err := ValidateStage(root, stage, st)
		if err != nil {
			t.Errorf("unexpected error validating stage %s: %v", stage, err)
		}
		if result == nil {
			t.Errorf("expected result for stage %s", stage)
		}
	}
}