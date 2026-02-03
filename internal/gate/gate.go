package gate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/thinkshake/foreman/internal/project"
	"github.com/thinkshake/foreman/internal/state"
)

// ValidationResult represents the result of a gate validation check.
type ValidationResult struct {
	Passed  bool
	Message string
	Details []string
}

// ValidateRequirements checks if the requirements stage is ready to pass.
func ValidateRequirements(root string) *ValidationResult {
	reqPath := filepath.Join(project.ForemanPath(root), "requirements.md")
	
	// Check if file exists
	if _, err := os.Stat(reqPath); os.IsNotExist(err) {
		return &ValidationResult{
			Passed:  false,
			Message: "Requirements document missing",
			Details: []string{
				"Create requirements.md in .foreman/ directory",
				"Define what the project should accomplish",
			},
		}
	}
	
	// Check if file has content (more than just placeholder)
	data, err := os.ReadFile(reqPath)
	if err != nil {
		return &ValidationResult{
			Passed:  false,
			Message: "Cannot read requirements.md",
			Details: []string{err.Error()},
		}
	}
	
	content := strings.TrimSpace(string(data))
	if len(content) < 50 { // Arbitrary minimum length
		return &ValidationResult{
			Passed:  false,
			Message: "Requirements document too short",
			Details: []string{
				"requirements.md should contain detailed project requirements",
				"Current length: " + fmt.Sprintf("%d characters", len(content)),
				"Minimum expected: 50 characters",
			},
		}
	}
	
	// Check for placeholder text
	if strings.Contains(content, "_Define what this project should accomplish") || 
	   strings.Contains(content, "Replace this placeholder") ||
	   strings.Contains(content, "_No requirements") {
		return &ValidationResult{
			Passed:  false,
			Message: "Requirements document contains placeholder text",
			Details: []string{
				"Replace placeholder content with actual requirements",
				"Define project goals, features, and constraints",
			},
		}
	}
	
	return &ValidationResult{
		Passed:  true,
		Message: "Requirements stage is ready",
		Details: []string{
			"requirements.md exists and has content",
			fmt.Sprintf("Content length: %d characters", len(content)),
		},
	}
}

// ValidateDesign checks if the design stage is ready to pass.
func ValidateDesign(root string) *ValidationResult {
	designsDir := filepath.Join(project.ForemanPath(root), "designs")
	
	// Check if designs directory exists
	if _, err := os.Stat(designsDir); os.IsNotExist(err) {
		return &ValidationResult{
			Passed:  false,
			Message: "Designs directory missing",
			Details: []string{
				"Create designs/ directory in .foreman/",
				"Add design documents (.md files) describing the architecture",
			},
		}
	}
	
	// Check for .md files in designs directory
	entries, err := os.ReadDir(designsDir)
	if err != nil {
		return &ValidationResult{
			Passed:  false,
			Message: "Cannot read designs directory",
			Details: []string{err.Error()},
		}
	}
	
	var mdFiles []string
	totalContent := 0
	
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		if !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		
		// Check file content
		path := filepath.Join(designsDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		
		content := strings.TrimSpace(string(data))
		if len(content) < 20 { // Minimal content check
			continue
		}
		
		mdFiles = append(mdFiles, entry.Name())
		totalContent += len(content)
	}
	
	if len(mdFiles) == 0 {
		return &ValidationResult{
			Passed:  false,
			Message: "No substantial design documents found",
			Details: []string{
				"Create .md files in designs/ directory",
				"Document architecture, data models, APIs, etc.",
				"Each file should have meaningful content (>20 characters)",
			},
		}
	}
	
	return &ValidationResult{
		Passed:  true,
		Message: "Design stage is ready",
		Details: []string{
			fmt.Sprintf("Found %d design documents: %s", len(mdFiles), strings.Join(mdFiles, ", ")),
			fmt.Sprintf("Total content: %d characters", totalContent),
		},
	}
}

// ValidatePhases checks if the phases stage is ready to pass.
func ValidatePhases(root string) *ValidationResult {
	phasesDir := filepath.Join(project.ForemanPath(root), "phases")
	
	// Check if phases directory exists
	if _, err := os.Stat(phasesDir); os.IsNotExist(err) {
		return &ValidationResult{
			Passed:  false,
			Message: "Phases directory missing",
			Details: []string{
				"Create phases/ directory in .foreman/",
				"Add overview.md and individual phase plans",
			},
		}
	}
	
	// Check for overview.md
	overviewPath := filepath.Join(phasesDir, "overview.md")
	if _, err := os.Stat(overviewPath); os.IsNotExist(err) {
		return &ValidationResult{
			Passed:  false,
			Message: "Phase overview missing",
			Details: []string{
				"Create phases/overview.md",
				"Document the overall phase breakdown strategy",
			},
		}
	}
	
	// Check overview content
	overviewData, err := os.ReadFile(overviewPath)
	if err != nil {
		return &ValidationResult{
			Passed:  false,
			Message: "Cannot read phases/overview.md",
			Details: []string{err.Error()},
		}
	}
	
	overviewContent := strings.TrimSpace(string(overviewData))
	if len(overviewContent) < 50 {
		return &ValidationResult{
			Passed:  false,
			Message: "Phase overview too short",
			Details: []string{
				"phases/overview.md should explain the implementation strategy",
				fmt.Sprintf("Current length: %d characters", len(overviewContent)),
				"Minimum expected: 50 characters",
			},
		}
	}
	
	// Check for individual phase plans
	entries, err := os.ReadDir(phasesDir)
	if err != nil {
		return &ValidationResult{
			Passed:  false,
			Message: "Cannot read phases directory",
			Details: []string{err.Error()},
		}
	}
	
	var phaseFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		name := entry.Name()
		if name == "overview.md" {
			continue
		}
		
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		
		// Check if it looks like a phase file (e.g., "1-setup.md", "2-backend.md")
		if len(name) >= 3 && name[1] == '-' {
			phaseFiles = append(phaseFiles, name)
		}
	}
	
	if len(phaseFiles) == 0 {
		return &ValidationResult{
			Passed:  false,
			Message: "No phase plan files found",
			Details: []string{
				"Create individual phase plans in phases/ directory",
				"Use naming convention: 1-setup.md, 2-backend.md, etc.",
				"Each phase should have its own detailed plan",
			},
		}
	}
	
	return &ValidationResult{
		Passed:  true,
		Message: "Phases stage is ready",
		Details: []string{
			"overview.md exists with content",
			fmt.Sprintf("Found %d phase plans: %s", len(phaseFiles), strings.Join(phaseFiles, ", ")),
		},
	}
}

// ValidateImplementation checks if the implementation stage is ready to pass.
func ValidateImplementation(root string, s *state.State) *ValidationResult {
	if len(s.Phases) == 0 {
		return &ValidationResult{
			Passed:  false,
			Message: "No phases defined yet",
			Details: []string{
				"Phases should be populated after phases stage completion",
				"Run sync to update phases from phases/ directory",
			},
		}
	}
	
	var incompletePhases []string
	for _, phase := range s.Phases {
		if phase.Status != "done" {
			incompletePhases = append(incompletePhases, fmt.Sprintf("%s (%s)", phase.Name, phase.Status))
		}
	}
	
	if len(incompletePhases) > 0 {
		return &ValidationResult{
			Passed:  false,
			Message: fmt.Sprintf("%d phases not completed", len(incompletePhases)),
			Details: append([]string{
				"All phases must be marked as 'done' before implementation can be completed",
				"Use 'foreman phase <name> done' to mark phases as complete",
				"Incomplete phases:",
			}, incompletePhases...),
		}
	}
	
	return &ValidationResult{
		Passed:  true,
		Message: "Implementation stage is ready",
		Details: []string{
			fmt.Sprintf("All %d phases completed", len(s.Phases)),
		},
	}
}

// ValidateStage validates a specific stage's readiness.
func ValidateStage(root, stage string, s *state.State) (*ValidationResult, error) {
	switch stage {
	case "requirements":
		return ValidateRequirements(root), nil
	case "design":
		return ValidateDesign(root), nil
	case "phases":
		return ValidatePhases(root), nil
	case "implementation":
		return ValidateImplementation(root, s), nil
	default:
		return nil, fmt.Errorf("unknown stage: %s", stage)
	}
}