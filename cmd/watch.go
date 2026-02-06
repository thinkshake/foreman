package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/thinkshake/foreman/internal/config"
	"github.com/thinkshake/foreman/internal/project"
	"github.com/thinkshake/foreman/internal/state"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch project progress in real-time",
	Long: `Watch the foreman project for state changes and display progress.

This is useful when a coding agent is working on phases and you want
to see progress without manually running 'foreman status'.

Press Ctrl+C to stop watching.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		interval, _ := cmd.Flags().GetInt("interval")

		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		root, err := project.FindRoot(wd)
		if err != nil {
			return err
		}

		fmt.Println("👀 Watching project progress... (Ctrl+C to stop)")
		fmt.Println()

		// Track last state for change detection
		var lastStage string
		var lastPhaseStates map[string]string

		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		defer ticker.Stop()

		// Initial display
		displayProgress(root, &lastStage, &lastPhaseStates, true)

		for range ticker.C {
			displayProgress(root, &lastStage, &lastPhaseStates, false)
		}

		return nil
	},
}

func displayProgress(root string, lastStage *string, lastPhaseStates *map[string]string, initial bool) {
	st, err := state.Load(root)
	if err != nil {
		fmt.Printf("Error loading state: %v\n", err)
		return
	}

	cfg, err := config.Load(root)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	// Sync phases
	_ = project.SyncPhasesToState(root, st)
	_ = state.Save(root, st)

	// Check for stage change
	stageChanged := *lastStage != st.CurrentStage
	if stageChanged && !initial {
		green := color.New(color.FgGreen, color.Bold)
		green.Printf("\n🎉 Stage advanced: %s → %s\n", *lastStage, st.CurrentStage)
	}
	*lastStage = st.CurrentStage

	// Check for phase changes
	if *lastPhaseStates == nil {
		*lastPhaseStates = make(map[string]string)
	}

	for _, phase := range st.Phases {
		oldStatus, exists := (*lastPhaseStates)[phase.Name]
		if exists && oldStatus != phase.Status && !initial {
			cyan := color.New(color.FgCyan)
			cyan.Printf("   📝 Phase %s: %s → %s\n", phase.Name, oldStatus, phase.Status)
		}
		(*lastPhaseStates)[phase.Name] = phase.Status
	}

	// Only show full status on initial or stage change
	if initial || stageChanged {
		fmt.Printf("\n")
		dim := color.New(color.Faint)
		dim.Printf("[%s] ", time.Now().Format("15:04:05"))
		fmt.Printf("Project: %s\n", cfg.Name)

		// Show mode
		if st.QuickMode {
			dim.Printf("Mode: quick\n")
		}

		// Show stages progress
		stages := st.GetActiveStages()
		fmt.Printf("Stages: ")
		for i, stage := range stages {
			gate := st.Gates[stage]
			if gate == nil {
				gate = &state.Gate{Status: "blocked"}
			}

			indicator := "⬜"
			if stage == st.CurrentStage {
				indicator = "🔵"
			} else if gate.Status == "approved" {
				indicator = "✅"
			}

			fmt.Printf("%s %s", indicator, stage)
			if i < len(stages)-1 {
				fmt.Print(" → ")
			}
		}
		fmt.Println()

		// Show phases if in implementation
		if st.CurrentStage == "implementation" && len(st.Phases) > 0 {
			fmt.Print("Phases: ")
			done := 0
			for _, phase := range st.Phases {
				if phase.Status == "done" {
					done++
				}
			}
			fmt.Printf("%d/%d complete\n", done, len(st.Phases))
			for _, phase := range st.Phases {
				indicator := "⬜"
				switch phase.Status {
				case "done":
					indicator = "✅"
				case "in-progress":
					indicator = "🔵"
				}
				fmt.Printf("  %s %s\n", indicator, phase.Name)
			}
		}
	}
}

func init() {
	watchCmd.Flags().Int("interval", 5, "Check interval in seconds")
	rootCmd.AddCommand(watchCmd)
}
