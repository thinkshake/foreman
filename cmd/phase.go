package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/thinkshake/foreman/internal/project"
	"github.com/thinkshake/foreman/internal/state"
)

var phaseCmd = &cobra.Command{
	Use:   "phase <name> <status>",
	Short: "Update phase status",
	Long: `Update the status of a phase during the implementation stage.

Valid statuses: planned | in-progress | done

Example:
  foreman phase 1-setup in-progress
  foreman phase 2-backend done`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		root, err := project.FindRoot(wd)
		if err != nil {
			return err
		}

		st, err := state.Load(root)
		if err != nil {
			return err
		}

		phaseName := args[0]
		phaseStatus := args[1]

		// Validate we're in implementation stage
		if st.CurrentStage != "implementation" {
			yellow := color.New(color.FgYellow)
			yellow.Printf("⚠️  Warning: Not in implementation stage (currently: %s)\n", st.CurrentStage)
			fmt.Println("Phase status can be updated at any time, but it's typically used during implementation.")
			fmt.Println()
		}

		// Sync phases from directory first
		if err := project.SyncPhasesToState(root, st); err != nil {
			return fmt.Errorf("failed to sync phases: %w", err)
		}

		// Update phase status
		if err := st.SetPhaseStatus(phaseName, phaseStatus); err != nil {
			return err
		}

		if err := state.Save(root, st); err != nil {
			return err
		}

		green := color.New(color.FgGreen)
		green.Printf("✓ ")
		fmt.Printf("Updated phase %s to: %s\n", phaseName, phaseStatus)

		// Show current phase status summary
		fmt.Println()
		fmt.Println("Phase Status:")
		for _, phase := range st.Phases {
			indicator := getPhaseStatusIndicator(phase.Status)
			highlight := ""
			if phase.Name == phaseName {
				highlight = " ← updated"
			}
			fmt.Printf("  %s %-15s %s%s\n", indicator, phase.Name, phase.Status, highlight)
		}

		// Check if all phases are done for implementation gate
		if st.AllPhasesDone() {
			fmt.Println()
			cyan := color.New(color.FgCyan, color.Bold)
			cyan.Println("🎉 All phases completed!")
			fmt.Println("Run 'foreman gate implementation' to complete the project")
		}

		return nil
	},
}

func getPhaseStatusIndicator(status string) string {
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
	rootCmd.AddCommand(phaseCmd)
}