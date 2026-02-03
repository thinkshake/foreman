package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/thinkshake/foreman/internal/brief"
	"github.com/thinkshake/foreman/internal/project"
	"github.com/thinkshake/foreman/internal/state"
)

var briefCmd = &cobra.Command{
	Use:   "brief <phase-name>",
	Short: "Generate a self-contained brief for a phase",
	Long: `Generates a comprehensive brief for a coding agent working on a specific phase.
The brief includes project context, requirements, design, phase dependencies,
and the specific phase plan.

The brief is saved to .foreman/briefs/<phase-name>.md and also printed to stdout.

Example:
  foreman brief 1-setup
  foreman brief 2-backend`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		root, err := project.FindRoot(wd)
		if err != nil {
			return err
		}

		phaseName := args[0]

		// Load state and sync phases
		st, err := state.Load(root)
		if err != nil {
			return err
		}

		if err := project.SyncPhasesToState(root, st); err != nil {
			return fmt.Errorf("failed to sync phases: %w", err)
		}

		if err := state.Save(root, st); err != nil {
			return err
		}

		// Validate phase exists
		phase := st.GetPhase(phaseName)
		if phase == nil {
			return fmt.Errorf("phase %q not found\n\nAvailable phases:", phaseName)
		}

		// Generate and save brief
		briefContent, err := brief.GenerateAndSave(root, phaseName)
		if err != nil {
			return err
		}

		briefPath := project.BriefPath(root, phaseName)
		
		green := color.New(color.FgGreen, color.Bold)
		green.Printf("✓ ")
		fmt.Printf("Generated brief for phase: %s\n", phaseName)
		
		dim := color.New(color.Faint)
		dim.Printf("Saved to: %s\n", briefPath)
		fmt.Println()

		cyan := color.New(color.FgCyan)
		cyan.Println("Brief content:")
		fmt.Println("=" + string(make([]byte, 50)) + "=")
		fmt.Print(briefContent)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(briefCmd)
}