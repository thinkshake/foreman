package state

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Stages defines the ordered list of stages for full workflow.
var Stages = []string{"requirements", "design", "phases", "implementation"}

// QuickStages defines the stages for quick mode (no design/phases).
var QuickStages = []string{"requirements", "implementation"}

// ValidStages contains all possible workflow stages.
var ValidStages = map[string]bool{
	"requirements":   true,
	"design":         true,
	"phases":         true,
	"implementation": true,
}

// Gate represents a stage gate with its status and review info.
type Gate struct {
	Status     string     `yaml:"status"`      // "open", "pending-review", "approved", "blocked"
	ApprovedAt *time.Time `yaml:"approved_at"` // when approved (nil if not approved)
	ApprovedBy string     `yaml:"approved_by"` // "auto" or "human"
	Reason     string     `yaml:"reason"`      // rejection reason (if any)
}

// Phase represents a phase within the implementation stage.
type Phase struct {
	Name   string `yaml:"name"`   // e.g. "1-setup", "2-backend"
	Status string `yaml:"status"` // "planned", "in-progress", "done"
}

// State represents the state.yaml schema.
type State struct {
	CurrentStage string             `yaml:"current_stage"`
	Gates        map[string]*Gate   `yaml:"gates"`
	Phases       []Phase            `yaml:"phases"`
	QuickMode    bool               `yaml:"quick_mode,omitempty"`   // v3: skip design/phases
	QuickTask    string             `yaml:"quick_task,omitempty"`   // v3: task description for quick mode
	Confidence   int                `yaml:"confidence,omitempty"`   // v3: auto-advance threshold (0-100)
	Workflow     []string           `yaml:"workflow,omitempty"`     // v2.1: custom workflow stages
	MinimalMode  bool               `yaml:"minimal_mode,omitempty"` // v2.1: no gates at all
}

// StatePath returns the path to state.yaml.
func StatePath(root string) string {
	return filepath.Join(root, ".foreman", "state.yaml")
}

// IsValidStage checks if a stage name is valid.
func IsValidStage(stage string) bool {
	for _, s := range Stages {
		if s == stage {
			return true
		}
	}
	return false
}

// IsValidGateStatus checks if a gate status is valid.
func IsValidGateStatus(status string) bool {
	validStatuses := []string{"open", "pending-review", "approved", "blocked"}
	for _, s := range validStatuses {
		if s == status {
			return true
		}
	}
	return false
}

// IsValidPhaseStatus checks if a phase status is valid.
func IsValidPhaseStatus(status string) bool {
	validStatuses := []string{"planned", "in-progress", "done"}
	for _, s := range validStatuses {
		if s == status {
			return true
		}
	}
	return false
}

// GetStageIndex returns the index of a stage in the ordered list.
func GetStageIndex(stage string) int {
	for i, s := range Stages {
		if s == stage {
			return i
		}
	}
	return -1
}

// GetStageIndexForMode returns stage index respecting quick mode.
func GetStageIndexForMode(stage string, quickMode bool) int {
	stages := Stages
	if quickMode {
		stages = QuickStages
	}
	for i, s := range stages {
		if s == stage {
			return i
		}
	}
	return -1
}

// GetNextStage returns the next stage after the given one, or empty if last.
func GetNextStage(stage string) string {
	idx := GetStageIndex(stage)
	if idx == -1 || idx >= len(Stages)-1 {
		return ""
	}
	return Stages[idx+1]
}

// GetNextStageForMode returns next stage respecting quick mode.
func GetNextStageForMode(stage string, quickMode bool) string {
	stages := Stages
	if quickMode {
		stages = QuickStages
	}
	idx := GetStageIndexForMode(stage, quickMode)
	if idx == -1 || idx >= len(stages)-1 {
		return ""
	}
	return stages[idx+1]
}

// GetActiveStages returns stages for current mode.
func (s *State) GetActiveStages() []string {
	// Custom workflow takes priority
	if len(s.Workflow) > 0 {
		return s.Workflow
	}
	if s.QuickMode {
		return QuickStages
	}
	return Stages
}

// GetStageIndexInWorkflow returns stage index within current workflow.
func (s *State) GetStageIndexInWorkflow(stage string) int {
	stages := s.GetActiveStages()
	for i, st := range stages {
		if st == stage {
			return i
		}
	}
	return -1
}

// GetNextStageInWorkflow returns the next stage in current workflow.
func (s *State) GetNextStageInWorkflow(stage string) string {
	stages := s.GetActiveStages()
	idx := s.GetStageIndexInWorkflow(stage)
	if idx == -1 || idx >= len(stages)-1 {
		return ""
	}
	return stages[idx+1]
}

// IsStageInWorkflow checks if a stage is part of the current workflow.
func (s *State) IsStageInWorkflow(stage string) bool {
	return s.GetStageIndexInWorkflow(stage) >= 0
}

// Load reads and parses state.yaml from the given root.
func Load(root string) (*State, error) {
	path := StatePath(root)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read state.yaml: %w", err)
	}
	
	var s State
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("failed to parse state.yaml: %w", err)
	}
	
	// Ensure gates exists
	if s.Gates == nil {
		s.Gates = make(map[string]*Gate)
	}
	
	// Ensure all stages have gates
	for _, stage := range Stages {
		if s.Gates[stage] == nil {
			s.Gates[stage] = &Gate{Status: "blocked"}
		}
	}
	
	return &s, nil
}

// Save writes the state back to state.yaml.
func Save(root string, s *State) error {
	path := StatePath(root)
	data, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("failed to marshal state.yaml: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// NewDefault creates a default initial state.
func NewDefault() *State {
	gates := make(map[string]*Gate)
	
	// First stage is open, rest are blocked
	gates["requirements"] = &Gate{Status: "open"}
	gates["design"] = &Gate{Status: "blocked"}
	gates["phases"] = &Gate{Status: "blocked"}
	gates["implementation"] = &Gate{Status: "blocked"}
	
	return &State{
		CurrentStage: "requirements",
		Gates:        gates,
		Phases:       []Phase{},
		QuickMode:    false,
		Confidence:   0,
	}
}

// NewQuickMode creates a quick mode state (skips design/phases).
func NewQuickMode(task string, confidence int) *State {
	gates := make(map[string]*Gate)

	// Quick mode only has requirements and implementation
	gates["requirements"] = &Gate{Status: "open"}
	gates["implementation"] = &Gate{Status: "blocked"}

	return &State{
		CurrentStage: "requirements",
		Gates:        gates,
		Phases:       []Phase{},
		QuickMode:    true,
		QuickTask:    task,
		Confidence:   confidence,
		Workflow:     QuickStages,
		MinimalMode:  false,
	}
}

// NewMinimalMode creates a minimal mode state (no gates, straight to implementation).
func NewMinimalMode(task string) *State {
	now := time.Now()
	gates := make(map[string]*Gate)

	// Minimal mode: requirements is auto-approved, implementation is open
	gates["requirements"] = &Gate{
		Status:     "approved",
		ApprovedAt: &now,
		ApprovedBy: "auto",
	}
	gates["implementation"] = &Gate{Status: "open"}

	return &State{
		CurrentStage: "implementation", // Skip straight to implementation
		Gates:        gates,
		Phases:       []Phase{},
		QuickMode:    true,
		QuickTask:    task,
		Confidence:   100,
		Workflow:     QuickStages,
		MinimalMode:  true,
	}
}

// NewWithWorkflow creates a state with a custom workflow.
func NewWithWorkflow(workflow []string, confidence int, minimal bool) *State {
	gates := make(map[string]*Gate)
	now := time.Now()

	if len(workflow) == 0 {
		workflow = Stages
	}

	startStage := workflow[0]

	if minimal {
		// Minimal: auto-approve everything except implementation
		for i, stage := range workflow {
			if i == len(workflow)-1 {
				// Last stage (implementation) is open
				gates[stage] = &Gate{Status: "open"}
				startStage = stage
			} else {
				// All others are auto-approved
				gates[stage] = &Gate{
					Status:     "approved",
					ApprovedAt: &now,
					ApprovedBy: "auto",
				}
			}
		}
	} else {
		// Normal: first stage is open, rest are blocked
		for i, stage := range workflow {
			if i == 0 {
				gates[stage] = &Gate{Status: "open"}
			} else {
				gates[stage] = &Gate{Status: "blocked"}
			}
		}
	}

	return &State{
		CurrentStage: startStage,
		Gates:        gates,
		Phases:       []Phase{},
		QuickMode:    len(workflow) <= 2, // Quick if 2 or fewer stages
		Confidence:   confidence,
		Workflow:     workflow,
		MinimalMode:  minimal,
	}
}

// CanAdvanceStage checks if we can advance from the current stage.
func (s *State) CanAdvanceStage() bool {
	currentGate := s.Gates[s.CurrentStage]
	return currentGate != nil && currentGate.Status == "approved"
}

// AdvanceToNextStage moves to the next stage and opens its gate.
func (s *State) AdvanceToNextStage() error {
	if !s.CanAdvanceStage() {
		return fmt.Errorf("cannot advance: current stage gate not approved")
	}

	// Use workflow-aware next stage
	nextStage := s.GetNextStageInWorkflow(s.CurrentStage)
	if nextStage == "" {
		// Fallback to old behavior for backward compat
		nextStage = GetNextStageForMode(s.CurrentStage, s.QuickMode)
	}
	if nextStage == "" {
		return fmt.Errorf("already at final stage")
	}

	s.CurrentStage = nextStage
	if s.Gates[nextStage] == nil {
		s.Gates[nextStage] = &Gate{Status: "open"}
	} else {
		s.Gates[nextStage].Status = "open"
		s.Gates[nextStage].Reason = "" // Clear any previous rejection reason
	}

	return nil
}

// ShouldAutoAdvance checks if the gate should auto-advance based on confidence.
func (s *State) ShouldAutoAdvance(confidence int) bool {
	if s.Confidence <= 0 {
		return false // No auto-advance configured
	}
	return confidence >= s.Confidence
}

// ApproveGate approves a gate and potentially advances the stage.
func (s *State) ApproveGate(stage, approvedBy string) error {
	gate := s.Gates[stage]
	if gate == nil {
		return fmt.Errorf("stage %s not found", stage)
	}
	
	// Can only approve gates that are open or pending-review
	if gate.Status != "open" && gate.Status != "pending-review" {
		return fmt.Errorf("gate %s is %s, cannot approve", stage, gate.Status)
	}
	
	now := time.Now()
	gate.Status = "approved"
	gate.ApprovedAt = &now
	gate.ApprovedBy = approvedBy
	gate.Reason = ""
	
	// If this is the current stage, advance (unless it's the final stage)
	if stage == s.CurrentStage {
		nextStage := GetNextStageForMode(stage, s.QuickMode)
		if nextStage != "" {
			return s.AdvanceToNextStage()
		}
		// Final stage approved, no advancement needed
	}
	
	return nil
}

// RejectGate rejects a gate and sets it back to open.
func (s *State) RejectGate(stage, reason string) error {
	gate := s.Gates[stage]
	if gate == nil {
		return fmt.Errorf("stage %s not found", stage)
	}
	
	if gate.Status != "pending-review" {
		return fmt.Errorf("gate %s is %s, can only reject pending-review gates", stage, gate.Status)
	}
	
	gate.Status = "open"
	gate.ApprovedAt = nil
	gate.ApprovedBy = ""
	gate.Reason = reason
	
	return nil
}

// SetGateStatus sets a gate to pending review.
func (s *State) SetGateStatus(stage, status string) error {
	if !IsValidGateStatus(status) {
		return fmt.Errorf("invalid gate status: %s", status)
	}
	
	gate := s.Gates[stage]
	if gate == nil {
		return fmt.Errorf("stage %s not found", stage)
	}
	
	gate.Status = status
	if status != "approved" {
		gate.ApprovedAt = nil
		gate.ApprovedBy = ""
	}
	
	return nil
}

// GetPhase returns a phase by name.
func (s *State) GetPhase(name string) *Phase {
	for i := range s.Phases {
		if s.Phases[i].Name == name {
			return &s.Phases[i]
		}
	}
	return nil
}

// SetPhaseStatus updates a phase's status.
func (s *State) SetPhaseStatus(name, status string) error {
	if !IsValidPhaseStatus(status) {
		return fmt.Errorf("invalid phase status: %s", status)
	}
	
	phase := s.GetPhase(name)
	if phase == nil {
		return fmt.Errorf("phase %s not found", name)
	}
	
	phase.Status = status
	return nil
}

// AddPhase adds a new phase.
func (s *State) AddPhase(name string) {
	// Check if phase already exists
	if s.GetPhase(name) != nil {
		return
	}
	
	s.Phases = append(s.Phases, Phase{
		Name:   name,
		Status: "planned",
	})
}

// AllPhasesDone checks if all phases are marked as done.
func (s *State) AllPhasesDone() bool {
	if len(s.Phases) == 0 {
		return false
	}
	
	for _, phase := range s.Phases {
		if phase.Status != "done" {
			return false
		}
	}
	
	return true
}