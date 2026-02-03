package state

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStageValidation(t *testing.T) {
	validStages := []string{"requirements", "design", "phases", "implementation"}
	for _, stage := range validStages {
		if !IsValidStage(stage) {
			t.Errorf("expected %s to be valid stage", stage)
		}
	}
	
	if IsValidStage("invalid") {
		t.Error("expected 'invalid' to be invalid stage")
	}
}

func TestGateStatusValidation(t *testing.T) {
	validStatuses := []string{"open", "pending-review", "approved", "blocked"}
	for _, status := range validStatuses {
		if !IsValidGateStatus(status) {
			t.Errorf("expected %s to be valid gate status", status)
		}
	}
	
	if IsValidGateStatus("invalid") {
		t.Error("expected 'invalid' to be invalid gate status")
	}
}

func TestPhaseStatusValidation(t *testing.T) {
	validStatuses := []string{"planned", "in-progress", "done"}
	for _, status := range validStatuses {
		if !IsValidPhaseStatus(status) {
			t.Errorf("expected %s to be valid phase status", status)
		}
	}
	
	if IsValidPhaseStatus("invalid") {
		t.Error("expected 'invalid' to be invalid phase status")
	}
}

func TestStageOrdering(t *testing.T) {
	tests := []struct {
		stage    string
		expected int
	}{
		{"requirements", 0},
		{"design", 1},
		{"phases", 2},
		{"implementation", 3},
	}
	
	for _, test := range tests {
		if got := GetStageIndex(test.stage); got != test.expected {
			t.Errorf("expected index %d for stage %s, got %d", test.expected, test.stage, got)
		}
	}
	
	if GetStageIndex("invalid") != -1 {
		t.Error("expected -1 for invalid stage")
	}
}

func TestGetNextStage(t *testing.T) {
	tests := []struct {
		stage    string
		expected string
	}{
		{"requirements", "design"},
		{"design", "phases"},
		{"phases", "implementation"},
		{"implementation", ""},
	}
	
	for _, test := range tests {
		if got := GetNextStage(test.stage); got != test.expected {
			t.Errorf("expected next stage '%s' for stage %s, got '%s'", test.expected, test.stage, got)
		}
	}
}

func TestNewDefault(t *testing.T) {
	state := NewDefault()
	
	if state.CurrentStage != "requirements" {
		t.Errorf("expected current stage 'requirements', got '%s'", state.CurrentStage)
	}
	
	if state.Gates["requirements"].Status != "open" {
		t.Errorf("expected requirements gate to be open, got '%s'", state.Gates["requirements"].Status)
	}
	
	if state.Gates["design"].Status != "blocked" {
		t.Errorf("expected design gate to be blocked, got '%s'", state.Gates["design"].Status)
	}
	
	if len(state.Phases) != 0 {
		t.Errorf("expected empty phases list, got %d phases", len(state.Phases))
	}
}

func TestGateApproval(t *testing.T) {
	state := NewDefault()
	
	// Cannot approve blocked gate
	err := state.ApproveGate("design", "auto")
	if err == nil {
		t.Error("expected error when approving blocked gate")
	}
	
	// Can approve open gate
	err = state.ApproveGate("requirements", "auto")
	if err != nil {
		t.Errorf("unexpected error approving open gate: %v", err)
	}
	
	// Check gate status
	gate := state.Gates["requirements"]
	if gate.Status != "approved" {
		t.Errorf("expected gate status 'approved', got '%s'", gate.Status)
	}
	
	if gate.ApprovedBy != "auto" {
		t.Errorf("expected approved by 'auto', got '%s'", gate.ApprovedBy)
	}
	
	if gate.ApprovedAt == nil {
		t.Error("expected approved at timestamp to be set")
	}
	
	// Should advance to next stage
	if state.CurrentStage != "design" {
		t.Errorf("expected current stage 'design', got '%s'", state.CurrentStage)
	}
	
	if state.Gates["design"].Status != "open" {
		t.Errorf("expected design gate to be open, got '%s'", state.Gates["design"].Status)
	}
}

func TestGateRejection(t *testing.T) {
	state := NewDefault()
	
	// Set gate to pending review
	state.SetGateStatus("requirements", "pending-review")
	
	// Reject the gate
	err := state.RejectGate("requirements", "needs more detail")
	if err != nil {
		t.Errorf("unexpected error rejecting gate: %v", err)
	}
	
	gate := state.Gates["requirements"]
	if gate.Status != "open" {
		t.Errorf("expected gate status 'open' after rejection, got '%s'", gate.Status)
	}
	
	if gate.Reason != "needs more detail" {
		t.Errorf("expected reason 'needs more detail', got '%s'", gate.Reason)
	}
}

func TestPhaseManagement(t *testing.T) {
	state := NewDefault()
	
	// Add phases
	state.AddPhase("1-setup")
	state.AddPhase("2-backend")
	
	if len(state.Phases) != 2 {
		t.Errorf("expected 2 phases, got %d", len(state.Phases))
	}
	
	// Get phase
	phase := state.GetPhase("1-setup")
	if phase == nil {
		t.Error("expected to find phase '1-setup'")
	}
	
	if phase.Status != "planned" {
		t.Errorf("expected phase status 'planned', got '%s'", phase.Status)
	}
	
	// Update phase status
	err := state.SetPhaseStatus("1-setup", "in-progress")
	if err != nil {
		t.Errorf("unexpected error updating phase status: %v", err)
	}
	
	phase = state.GetPhase("1-setup")
	if phase.Status != "in-progress" {
		t.Errorf("expected phase status 'in-progress', got '%s'", phase.Status)
	}
	
	// Test AllPhasesDone
	if state.AllPhasesDone() {
		t.Error("expected AllPhasesDone to return false")
	}
	
	// Mark all phases done
	state.SetPhaseStatus("1-setup", "done")
	state.SetPhaseStatus("2-backend", "done")
	
	if !state.AllPhasesDone() {
		t.Error("expected AllPhasesDone to return true")
	}
}

func TestLoadSave(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "foreman-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create .foreman directory
	foremanDir := filepath.Join(tempDir, ".foreman")
	if err := os.MkdirAll(foremanDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	// Create and save state
	state := NewDefault()
	state.CurrentStage = "design"
	state.AddPhase("1-setup")
	state.SetPhaseStatus("1-setup", "in-progress")
	
	if err := Save(tempDir, state); err != nil {
		t.Fatal(err)
	}
	
	// Load state back
	loaded, err := Load(tempDir)
	if err != nil {
		t.Fatal(err)
	}
	
	// Verify loaded state
	if loaded.CurrentStage != state.CurrentStage {
		t.Errorf("expected current stage '%s', got '%s'", state.CurrentStage, loaded.CurrentStage)
	}
	
	if len(loaded.Phases) != len(state.Phases) {
		t.Errorf("expected %d phases, got %d", len(state.Phases), len(loaded.Phases))
	}
	
	if loaded.GetPhase("1-setup").Status != "in-progress" {
		t.Errorf("expected phase status 'in-progress', got '%s'", loaded.GetPhase("1-setup").Status)
	}
}