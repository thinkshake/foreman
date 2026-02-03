package state

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Stages defines the ordered list of stages.
var Stages = []string{"requirements", "design", "phases", "implementation"}

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

// GetNextStage returns the next stage after the given one, or empty if last.
func GetNextStage(stage string) string {
	idx := GetStageIndex(stage)
	if idx == -1 || idx >= len(Stages)-1 {
		return ""
	}
	return Stages[idx+1]
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
	
	nextStage := GetNextStage(s.CurrentStage)
	if nextStage == "" {
		return fmt.Errorf("already at final stage")
	}
	
	s.CurrentStage = nextStage
	s.Gates[nextStage].Status = "open"
	s.Gates[nextStage].Reason = "" // Clear any previous rejection reason
	
	return nil
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
		nextStage := GetNextStage(stage)
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