package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/thinkshake/foreman/internal/config"
	"github.com/thinkshake/foreman/internal/project"
	"github.com/thinkshake/foreman/internal/state"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show project stage and gate status",
	Long:  "Displays the current project stage, gate statuses, and phase progress.",
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		root, err := project.FindRoot(wd)
		if err != nil {
			return err
		}

		cfg, err := config.Load(root)
		if err != nil {
			return err
		}

		st, err := state.Load(root)
		if err != nil {
			return err
		}

		// Project header
		fmt.Printf("Project: %s\n", cfg.Name)
		if cfg.Description != "" {
			dim := color.New(color.Faint)
			dim.Printf("%s\n", cfg.Description)
		}
		
		// Mode indicator
		if st.QuickMode {
			yellow := color.New(color.FgYellow)
			yellow.Printf("Mode: quick")
			if st.Confidence > 0 {
				fmt.Printf(" (auto-advance at %d%%)", st.Confidence)
			}
			fmt.Println()
		} else if cfg.Preset != "" {
			dim := color.New(color.Faint)
			dim.Printf("Preset: %s\n", cfg.Preset)
		}
		
		// Current stage
		stages := st.GetActiveStages()
		stageIndex := state.GetStageIndexForMode(st.CurrentStage, st.QuickMode) + 1
		totalStages := len(stages)
		fmt.Printf("Stage: %s (%d/%d)\n\n", st.CurrentStage, stageIndex, totalStages)

		// Gates status
		fmt.Println("Gates:")
		for _, stage := range stages {
			gate := st.Gates[stage]
			if gate == nil {
				continue
			}

			indicator := getGateIndicator(gate.Status)
			statusText := gate.Status
			
			extra := ""
			if gate.Status == "approved" && gate.ApprovedAt != nil && gate.ApprovedBy != "" {
				timeStr := gate.ApprovedAt.Format("2006-01-02 15:04")
				extra = fmt.Sprintf(" (%s, %s)", gate.ApprovedBy, timeStr)
			}
			if gate.Reason != "" {
				extra = fmt.Sprintf(" (reason: %s)", gate.Reason)
			}

			fmt.Printf("  %s %-15s %s%s\n", indicator, stage, statusText, extra)
		}

		// Phases (if any)
		if len(st.Phases) > 0 {
			fmt.Println("\nPhases:")
			for _, phase := range st.Phases {
				indicator := getPhaseIndicator(phase.Status)
				fmt.Printf("  %s %-15s %s\n", indicator, phase.Name, phase.Status)
			}
		} else if st.CurrentStage == "implementation" || state.GetStageIndex(st.CurrentStage) > state.GetStageIndex("phases") {
			fmt.Println("\nPhases: (not yet defined)")
			dim := color.New(color.Faint)
			dim.Println("  Run sync to load phases from phases/ directory")
		}

		return nil
	},
}

func getGateIndicator(status string) string {
	switch status {
	case "approved":
		return "✅"
	case "pending-review":
		return "⏳"
	case "open":
		return "🔵"
	case "blocked":
		return "🔒"
	default:
		return "❓"
	}
}

func getPhaseIndicator(status string) string {
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

func init() {
	rootCmd.AddCommand(statusCmd)
}