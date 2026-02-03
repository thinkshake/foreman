package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/thinkshake/foreman/internal/config"
	"github.com/thinkshake/foreman/internal/state"
)

const ForemanDir = ".foreman"

// FindRoot walks up from dir looking for .foreman/.
func FindRoot(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(abs, ForemanDir)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return abs, nil
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			break
		}
		abs = parent
	}
	return "", fmt.Errorf("no .foreman/ directory found (walked up from %s)\nRun 'foreman init' to create a project", dir)
}

// ForemanPath returns the path to .foreman/ for a project root.
func ForemanPath(root string) string {
	return filepath.Join(root, ForemanDir)
}

// RequirementsPath returns the path to requirements.md.
func RequirementsPath(root string) string {
	return filepath.Join(ForemanPath(root), "requirements.md")
}

// DesignsPath returns the path to the designs directory.
func DesignsPath(root string) string {
	return filepath.Join(ForemanPath(root), "designs")
}

// PhasesPath returns the path to the phases directory.
func PhasesPath(root string) string {
	return filepath.Join(ForemanPath(root), "phases")
}

// BriefsPath returns the path to the briefs directory.
func BriefsPath(root string) string {
	return filepath.Join(ForemanPath(root), "briefs")
}

// PhaseOverviewPath returns the path to phases/overview.md.
func PhaseOverviewPath(root string) string {
	return filepath.Join(PhasesPath(root), "overview.md")
}

// PhasePlanPath returns the path to a specific phase plan.
func PhasePlanPath(root, phaseName string) string {
	return filepath.Join(PhasesPath(root), phaseName+".md")
}

// BriefPath returns the path to a specific brief.
func BriefPath(root, phaseName string) string {
	return filepath.Join(BriefsPath(root), phaseName+".md")
}

// Init creates a new .foreman/ directory with v2 structure.
func Init(dir, name string) (string, error) {
	foremanDir := filepath.Join(dir, ForemanDir)
	if _, err := os.Stat(foremanDir); err == nil {
		return "", fmt.Errorf(".foreman/ already exists in %s", dir)
	}

	// Create directory structure
	dirs := []string{
		foremanDir,
		DesignsPath(dir),
		PhasesPath(dir),
		BriefsPath(dir),
	}
	
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return "", fmt.Errorf("failed to create %s: %w", d, err)
		}
	}

	// Determine project name
	if name == "" {
		name = filepath.Base(dir)
	}

	// Create config.yaml
	cfg := config.NewDefault(name)
	if err := config.Save(dir, cfg); err != nil {
		return "", fmt.Errorf("failed to create config.yaml: %w", err)
	}

	// Create state.yaml
	st := state.NewDefault()
	if err := state.Save(dir, st); err != nil {
		return "", fmt.Errorf("failed to create state.yaml: %w", err)
	}

	// Create initial requirements.md placeholder
	reqContent := `# Requirements

_Define what this project should accomplish. Include:_

- **Goal**: What problem does this solve?
- **Features**: What functionality should it have?
- **Constraints**: Any technical or business limitations?
- **Success criteria**: How will you know it's working?

Replace this placeholder with actual requirements.
`
	if err := os.WriteFile(RequirementsPath(dir), []byte(reqContent), 0644); err != nil {
		return "", fmt.Errorf("failed to create requirements.md: %w", err)
	}

	return dir, nil
}

// SyncPhasesToState reads phase files and updates state with phase list.
func SyncPhasesToState(root string, st *state.State) error {
	phasesDir := PhasesPath(root)
	
	// Check if phases directory exists
	if _, err := os.Stat(phasesDir); os.IsNotExist(err) {
		return nil // No phases yet
	}
	
	entries, err := os.ReadDir(phasesDir)
	if err != nil {
		return fmt.Errorf("failed to read phases directory: %w", err)
	}
	
	// Collect phase names from files (excluding overview.md)
	var phaseNames []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		name := entry.Name()
		if name == "overview.md" || !strings.HasSuffix(name, ".md") {
			continue
		}
		
		// Remove .md extension to get phase name
		phaseName := strings.TrimSuffix(name, ".md")
		phaseNames = append(phaseNames, phaseName)
	}
	
	// Update state with found phases, preserving existing status
	existingPhases := make(map[string]string)
	for _, phase := range st.Phases {
		existingPhases[phase.Name] = phase.Status
	}
	
	// Rebuild phases list
	st.Phases = []state.Phase{}
	for _, name := range phaseNames {
		status := "planned" // default
		if existingStatus, exists := existingPhases[name]; exists {
			status = existingStatus
		}
		
		st.Phases = append(st.Phases, state.Phase{
			Name:   name,
			Status: status,
		})
	}
	
	return nil
}

// ReadFileContent safely reads a file and returns its content, or a placeholder if not found.
func ReadFileContent(path, placeholder string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return placeholder
	}
	return strings.TrimSpace(string(data))
}

// ReadRequirements reads the requirements.md file.
func ReadRequirements(root string) string {
	return ReadFileContent(RequirementsPath(root), "_No requirements defined yet._")
}

// ReadDesigns reads all design documents and concatenates them.
func ReadDesigns(root string) string {
	designsDir := DesignsPath(root)
	
	entries, err := os.ReadDir(designsDir)
	if err != nil {
		return "_No design documents found._"
	}
	
	var designs []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		
		path := filepath.Join(designsDir, entry.Name())
		content := ReadFileContent(path, "")
		if content != "" {
			designs = append(designs, fmt.Sprintf("## %s\n\n%s", entry.Name(), content))
		}
	}
	
	if len(designs) == 0 {
		return "_No design documents with content found._"
	}
	
	return strings.Join(designs, "\n\n---\n\n")
}

// ReadPhaseOverview reads the phases/overview.md file.
func ReadPhaseOverview(root string) string {
	return ReadFileContent(PhaseOverviewPath(root), "_No phase overview defined yet._")
}

// ReadPhasePlan reads a specific phase plan file.
func ReadPhasePlan(root, phaseName string) string {
	return ReadFileContent(PhasePlanPath(root, phaseName), fmt.Sprintf("_No plan defined for phase %s._", phaseName))
}