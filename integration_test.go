package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thinkshake/foreman/internal/brief"
	"github.com/thinkshake/foreman/internal/config"
	"github.com/thinkshake/foreman/internal/gate"
	"github.com/thinkshake/foreman/internal/project"
	"github.com/thinkshake/foreman/internal/state"
)

// TestFullWorkflow tests the complete project workflow from init to completion.
func TestFullWorkflow(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "foreman-integration-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Step 1: Initialize project
	root, err := project.Init(tempDir, "test-project")
	if err != nil {
		t.Fatalf("failed to initialize project: %v", err)
	}

	// Verify project structure was created
	expectedDirs := []string{
		".foreman",
		".foreman/designs",
		".foreman/phases", 
		".foreman/briefs",
	}

	for _, dir := range expectedDirs {
		path := filepath.Join(root, dir)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected directory %s to exist", dir)
		}
	}

	// Verify initial files
	configPath := filepath.Join(root, ".foreman", "config.yaml")
	statePath := filepath.Join(root, ".foreman", "state.yaml")
	reqPath := filepath.Join(root, ".foreman", "requirements.md")

	for _, file := range []string{configPath, statePath, reqPath} {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", file)
		}
	}

	// Step 2: Load and verify initial state
	st, err := state.Load(root)
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}

	if st.CurrentStage != "requirements" {
		t.Errorf("expected current stage 'requirements', got '%s'", st.CurrentStage)
	}

	if st.Gates["requirements"].Status != "open" {
		t.Errorf("expected requirements gate to be open, got '%s'", st.Gates["requirements"].Status)
	}

	// Step 3: Test requirements stage
	// Initially should fail validation (placeholder content)
	result, err := gate.ValidateStage(root, "requirements", st)
	if err != nil {
		t.Fatalf("failed to validate requirements: %v", err)
	}

	if result.Passed {
		t.Error("expected requirements validation to fail with placeholder content")
	}

	// Add real requirements
	requirementsContent := `# Requirements

## Goal
Build a comprehensive CLI tool for managing AI-native software projects.

## Features
- Stage-based workflow (requirements → design → phases → implementation)
- Gate validation system to ensure quality at each stage
- Phase-based implementation tracking
- Comprehensive brief compilation for coding agents
- Human and automatic review capabilities

## Constraints
- Must be written in Go using Cobra framework
- Must maintain backward compatibility where possible
- Must be thoroughly tested
- Configuration stored in YAML files

## Success Criteria
- All CLI commands work as specified
- Full test coverage including integration tests
- Documentation is complete and accurate
- Tool can manage complete project lifecycle from start to finish`

	if err := os.WriteFile(reqPath, []byte(requirementsContent), 0644); err != nil {
		t.Fatalf("failed to write requirements: %v", err)
	}

	// Now validation should pass
	result, err = gate.ValidateStage(root, "requirements", st)
	if err != nil {
		t.Fatalf("failed to validate requirements: %v", err)
	}

	if !result.Passed {
		t.Errorf("expected requirements validation to pass: %s", result.Message)
	}

	// Approve requirements gate (auto reviewer)
	cfg, err := config.Load(root)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Reviewers.Default != "auto" {
		t.Errorf("expected default reviewer to be 'auto', got '%s'", cfg.Reviewers.Default)
	}

	// Approve gate
	if err := st.ApproveGate("requirements", "auto"); err != nil {
		t.Fatalf("failed to approve requirements gate: %v", err)
	}

	if err := state.Save(root, st); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	// Verify advancement
	if st.CurrentStage != "design" {
		t.Errorf("expected current stage 'design', got '%s'", st.CurrentStage)
	}

	if st.Gates["design"].Status != "open" {
		t.Errorf("expected design gate to be open, got '%s'", st.Gates["design"].Status)
	}

	// Step 4: Test design stage
	designsDir := filepath.Join(root, ".foreman", "designs")

	// Create design documents
	archContent := `# Architecture Design

## Overview
Foreman v2 uses a modular architecture with distinct layers:

## Core Modules
- **Config**: Project configuration management (config.yaml)
- **State**: Project state tracking (state.yaml) 
- **Gate**: Stage validation and advancement logic
- **Project**: File system utilities and project structure
- **Brief**: Context compilation for coding agents

## CLI Structure
- Built on Cobra framework for consistent command interface
- Each command handles specific workflow aspect
- Error handling with helpful user guidance

## Data Flow
1. Commands load config and state from .foreman/ directory
2. Validation ensures data integrity and workflow rules
3. State transitions controlled by gate system
4. Briefs compile all context for handoff to coding agents`

	apiContent := `# API Design

## Command Structure

### Core Commands
- **init**: Initialize new project with v2 structure
- **status**: Show current stage and gate status  
- **gate**: Validate stages and control advancement
- **phase**: Manage phase status in implementation
- **brief**: Generate context briefs for phases
- **config**: Manage project configuration

### Gate System
Each stage has a gate that validates completion:
- requirements: Check requirements.md exists and has content
- design: Check designs/ directory has meaningful .md files
- phases: Check phases/overview.md and individual phase plans exist  
- implementation: Check all phases marked as done

### Reviewer Configuration
- auto: Gate automatically approves when validation passes
- human: Gate waits for manual approval via --approve flag`

	if err := os.WriteFile(filepath.Join(designsDir, "architecture.md"), []byte(archContent), 0644); err != nil {
		t.Fatalf("failed to write architecture design: %v", err)
	}

	if err := os.WriteFile(filepath.Join(designsDir, "api.md"), []byte(apiContent), 0644); err != nil {
		t.Fatalf("failed to write API design: %v", err)
	}

	// Validate and approve design stage
	result, err = gate.ValidateStage(root, "design", st)
	if err != nil {
		t.Fatalf("failed to validate design: %v", err)
	}

	if !result.Passed {
		t.Errorf("expected design validation to pass: %s", result.Message)
	}

	if err := st.ApproveGate("design", "auto"); err != nil {
		t.Fatalf("failed to approve design gate: %v", err)
	}

	if err := state.Save(root, st); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	// Step 5: Test phases stage
	phasesDir := filepath.Join(root, ".foreman", "phases")

	// Create phase documents
	overviewContent := `# Implementation Phases

## Strategy
Break the implementation into independent phases that can be handed off to coding agents.
Each phase should be self-contained with clear objectives and deliverables.

## Phase Breakdown
1. **Setup**: Basic project structure, dependencies, core types
2. **Logic**: Core business logic - state management, config, validation
3. **CLI**: Command implementation and user interface
4. **Testing**: Comprehensive testing and documentation

## Dependencies
- Phase 1 must complete before Phase 2
- Phases 2 and 3 can partially overlap
- Phase 4 requires all implementation phases complete

## Handoff Process
Each phase gets a comprehensive brief with:
- Project context and requirements
- Design documentation  
- Specific phase objectives
- Dependencies and their current status
- Implementation guidelines`

	phase1Content := `# Phase 1: Project Setup

## Objectives
- Set up Go module with proper dependencies
- Create core type definitions for config, state, gates
- Implement basic file I/O and project structure utilities
- Create foundation for CLI with Cobra

## Deliverables
- go.mod with all required dependencies (cobra, yaml.v3)
- Core package structure (internal/config, internal/state, internal/gate)
- Basic project initialization (project.Init function)
- Foundation CLI structure

## Acceptance Criteria
- go build succeeds without errors
- Basic project initialization creates correct directory structure
- Core types marshal/unmarshal to YAML correctly
- All code has appropriate documentation`

	phase2Content := `# Phase 2: Core Logic Implementation 

## Objectives
- Implement complete config and state management
- Build gate validation system for all stages
- Create project utilities (FindRoot, file reading, etc.)
- Implement state transitions and workflow logic

## Deliverables
- Full config.go and state.go with all methods
- Complete gate validation for each stage
- Project utility functions
- State transition logic with proper error handling

## Acceptance Criteria
- Config can be loaded, modified, and saved
- State transitions work correctly (stage advancement)
- Gate validation accurately checks stage completion
- All core logic has unit tests with good coverage`

	phase3Content := `# Phase 3: CLI Commands

## Objectives
- Implement all CLI commands per specification
- Create proper error handling and user feedback
- Add colored output and helpful messages
- Ensure commands integrate properly with core logic

## Deliverables
- All commands: init, status, gate, phase, brief, config
- Proper flag handling and argument validation
- User-friendly output with colors and formatting
- Integration between commands and core modules

## Acceptance Criteria
- All commands work as specified in design
- Error messages are helpful and actionable
- Output is well-formatted and informative
- Commands properly validate inputs and handle edge cases`

	phase4Content := `# Phase 4: Testing and Documentation

## Objectives
- Write comprehensive unit tests for all modules
- Create integration tests for full workflow
- Update documentation (README.md)
- Ensure code quality and maintainability

## Deliverables
- Unit tests for config, state, gate, project modules
- Integration tests for complete workflow scenarios
- Updated README.md with v2 documentation
- Code cleanup and final review

## Acceptance Criteria
- go test ./... passes with good coverage
- Integration tests verify complete workflow works
- Documentation is accurate and complete
- Code follows Go best practices and is well-documented`

	files := map[string]string{
		"overview.md": overviewContent,
		"1-setup.md":  phase1Content,
		"2-logic.md":  phase2Content,
		"3-cli.md":    phase3Content,
		"4-testing.md": phase4Content,
	}

	for filename, content := range files {
		path := filepath.Join(phasesDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", filename, err)
		}
	}

	// Validate and approve phases stage
	result, err = gate.ValidateStage(root, "phases", st)
	if err != nil {
		t.Fatalf("failed to validate phases: %v", err)
	}

	if !result.Passed {
		t.Errorf("expected phases validation to pass: %s", result.Message)
	}

	if err := st.ApproveGate("phases", "auto"); err != nil {
		t.Fatalf("failed to approve phases gate: %v", err)
	}

	// Step 6: Test implementation stage
	// First sync phases to state
	if err := project.SyncPhasesToState(root, st); err != nil {
		t.Fatalf("failed to sync phases: %v", err)
	}

	if err := state.Save(root, st); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	// Verify phases were loaded
	expectedPhases := []string{"1-setup", "2-logic", "3-cli", "4-testing"}
	if len(st.Phases) != len(expectedPhases) {
		t.Errorf("expected %d phases, got %d", len(expectedPhases), len(st.Phases))
	}

	for _, expectedPhase := range expectedPhases {
		phase := st.GetPhase(expectedPhase)
		if phase == nil {
			t.Errorf("expected phase %s to exist", expectedPhase)
		} else if phase.Status != "planned" {
			t.Errorf("expected phase %s to be planned, got %s", expectedPhase, phase.Status)
		}
	}

	// Test implementation validation (should fail - phases not done)
	result, err = gate.ValidateStage(root, "implementation", st)
	if err != nil {
		t.Fatalf("failed to validate implementation: %v", err)
	}

	if result.Passed {
		t.Error("expected implementation validation to fail with incomplete phases")
	}

	// Mark all phases as done
	for _, phaseName := range expectedPhases {
		if err := st.SetPhaseStatus(phaseName, "done"); err != nil {
			t.Fatalf("failed to set phase %s status: %v", phaseName, err)
		}
	}

	// Now implementation should validate
	result, err = gate.ValidateStage(root, "implementation", st)
	if err != nil {
		t.Fatalf("failed to validate implementation: %v", err)
	}

	if !result.Passed {
		t.Errorf("expected implementation validation to pass: %s", result.Message)
	}

	// Approve implementation gate
	if err := st.ApproveGate("implementation", "auto"); err != nil {
		t.Fatalf("failed to approve implementation gate: %v", err)
	}

	if err := state.Save(root, st); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	// Step 7: Verify final state
	// All gates should be approved
	for _, stageName := range state.Stages {
		gate := st.Gates[stageName]
		if gate.Status != "approved" {
			t.Errorf("expected gate %s to be approved, got %s", stageName, gate.Status)
		}
	}

	// Should still be in implementation stage (no stage after it)
	if st.CurrentStage != "implementation" {
		t.Errorf("expected to remain in implementation stage, got %s", st.CurrentStage)
	}
}

// TestBriefGeneration tests brief compilation for phases.
func TestBriefGeneration(t *testing.T) {
	// Set up a project with phases
	tempDir, err := os.MkdirTemp("", "foreman-brief-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize project
	root, err := project.Init(tempDir, "brief-test")
	if err != nil {
		t.Fatalf("failed to initialize project: %v", err)
	}

	// Set up basic content
	reqContent := "# Requirements\n\nBuild a test project for brief generation."
	if err := os.WriteFile(project.RequirementsPath(root), []byte(reqContent), 0644); err != nil {
		t.Fatal(err)
	}

	designContent := "# Design\n\nSimple test design."
	designPath := filepath.Join(project.DesignsPath(root), "test.md")
	if err := os.WriteFile(designPath, []byte(designContent), 0644); err != nil {
		t.Fatal(err)
	}

	overviewContent := "# Phase Overview\n\nTest phases for brief generation."
	if err := os.WriteFile(project.PhaseOverviewPath(root), []byte(overviewContent), 0644); err != nil {
		t.Fatal(err)
	}

	phaseContent := "# Phase 1\n\nFirst test phase."
	if err := os.WriteFile(project.PhasePlanPath(root, "1-test"), []byte(phaseContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Set up state with phases
	st, err := state.Load(root)
	if err != nil {
		t.Fatal(err)
	}

	if err := project.SyncPhasesToState(root, st); err != nil {
		t.Fatal(err)
	}

	st.AddPhase("1-test")
	if err := state.Save(root, st); err != nil {
		t.Fatal(err)
	}

	// Generate brief
	briefContent, err := brief.GenerateAndSave(root, "1-test")
	if err != nil {
		t.Fatalf("failed to generate brief: %v", err)
	}

	// Verify brief content
	expectedSections := []string{
		"# Phase Brief: 1-test",
		"## Project Context",
		"## Requirements",  
		"## Design Context",
		"## Phase Overview",
		"## Dependencies",
		"## Phase Spec: 1-test",
		"## Implementation Guidelines",
	}

	for _, section := range expectedSections {
		if !strings.Contains(briefContent, section) {
			t.Errorf("expected brief to contain section: %s", section)
		}
	}

	// Verify brief file was saved
	briefPath := project.BriefPath(root, "1-test")
	if _, err := os.Stat(briefPath); os.IsNotExist(err) {
		t.Error("expected brief file to be saved")
	}
}