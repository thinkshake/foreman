package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/thinkshake/foreman/internal/config"
	"github.com/thinkshake/foreman/internal/gate"
	"github.com/thinkshake/foreman/internal/project"
	"github.com/thinkshake/foreman/internal/state"
)

var gateCmd = &cobra.Command{
	Use:   "gate [stage]",
	Short: "Validate and control stage gates",
	Long: `Validate stage completion and control gate advancement.

Without arguments: checks the current stage's gate.
With stage name: checks that specific stage's gate.

Gates control advancement between stages. Each gate validates that
the stage work is complete before allowing progression.`,
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

		// Handle subcommands
		approve, _ := cmd.Flags().GetBool("approve")
		reject, _ := cmd.Flags().GetBool("reject")
		reason, _ := cmd.Flags().GetString("reason")
		reviewer, _ := cmd.Flags().GetString("reviewer")

		// Determine target stage
		targetStage := st.CurrentStage
		if len(args) > 0 {
			targetStage = args[0]
		}

		// Validate stage against active stages (respects quick mode)
		activeStages := st.GetActiveStages()
		isValid := false
		for _, s := range activeStages {
			if s == targetStage {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("invalid stage: %s (valid for current mode: %s)", targetStage, strings.Join(activeStages, ", "))
		}

		// Handle reviewer change
		if reviewer != "" {
			if reviewer != "auto" && reviewer != "human" {
				return fmt.Errorf("invalid reviewer: %s (must be 'auto' or 'human')", reviewer)
			}
			cfg.Reviewers.SetReviewer(targetStage, reviewer)
			if err := config.Save(root, cfg); err != nil {
				return err
			}
			fmt.Printf("✓ Set %s gate reviewer to: %s\n", targetStage, reviewer)
			return nil
		}

		// Handle approve
		if approve {
			return handleApprove(root, targetStage, st)
		}

		// Handle reject
		if reject {
			return handleReject(root, targetStage, reason, st)
		}

		// Default: validate gate
		return handleValidate(root, targetStage, cfg, st)
	},
}

func handleValidate(root, stage string, cfg *config.Config, st *state.State) error {
	// First sync phases if we're checking phases or implementation
	if stage == "phases" || stage == "implementation" {
		if err := project.SyncPhasesToState(root, st); err != nil {
			return fmt.Errorf("failed to sync phases: %w", err)
		}
		if err := state.Save(root, st); err != nil {
			return err
		}
	}

	result, err := gate.ValidateStage(root, stage, st)
	if err != nil {
		return err
	}

	gate := st.Gates[stage]
	reviewer := cfg.Reviewers.GetReviewer(stage)

	fmt.Printf("Gate: %s\n", stage)
	fmt.Printf("Status: %s\n", gate.Status)
	fmt.Printf("Reviewer: %s\n\n", reviewer)

	// Show validation result
	if result.Passed {
		green := color.New(color.FgGreen, color.Bold)
		green.Printf("✓ %s\n", result.Message)
	} else {
		red := color.New(color.FgRed, color.Bold)
		red.Printf("✗ %s\n", result.Message)
	}

	fmt.Println()
	if len(result.Details) > 0 {
		for _, detail := range result.Details {
			fmt.Printf("  %s\n", detail)
		}
		fmt.Println()
	}

	// If validation passes and gate is open, advance based on reviewer
	if result.Passed && gate.Status == "open" {
		if reviewer == "auto" {
			// Auto-approve
			if err := st.ApproveGate(stage, "auto"); err != nil {
				return err
			}
			if err := state.Save(root, st); err != nil {
				return err
			}

			cyan := color.New(color.FgCyan, color.Bold)
			cyan.Printf("🎉 Gate approved automatically!\n")
			
			if stage == st.CurrentStage {
				nextStage := state.GetNextStageForMode(stage, st.QuickMode)
				if nextStage != "" {
					fmt.Printf("Advanced to stage: %s\n", nextStage)
				} else {
					fmt.Printf("🏁 All stages completed!\n")
				}
			}
		} else {
			// Set to pending review for human approval
			if err := st.SetGateStatus(stage, "pending-review"); err != nil {
				return err
			}
			if err := state.Save(root, st); err != nil {
				return err
			}

			yellow := color.New(color.FgYellow, color.Bold)
			yellow.Printf("⏳ Gate validation passed - awaiting human review\n")
			fmt.Printf("Run 'foreman gate %s --approve' to approve manually\n", stage)
		}
	} else if !result.Passed {
		fmt.Printf("Complete the requirements above, then run 'foreman gate %s' again\n", stage)
	}

	return nil
}

func handleApprove(root, stage string, st *state.State) error {
	gate := st.Gates[stage]
	if gate == nil {
		return fmt.Errorf("stage %s not found", stage)
	}

	if gate.Status != "pending-review" {
		return fmt.Errorf("gate %s is %s, can only approve pending-review gates", stage, gate.Status)
	}

	if err := st.ApproveGate(stage, "human"); err != nil {
		return err
	}

	if err := state.Save(root, st); err != nil {
		return err
	}

	green := color.New(color.FgGreen, color.Bold)
	green.Printf("✓ Gate %s approved!\n", stage)

	if stage == st.CurrentStage {
		nextStage := state.GetNextStageForMode(stage, st.QuickMode)
		if nextStage != "" {
			fmt.Printf("Advanced to stage: %s\n", nextStage)
		} else {
			fmt.Printf("🏁 All stages completed!\n")
		}
	}

	return nil
}

func handleReject(root, stage, reason string, st *state.State) error {
	gate := st.Gates[stage]
	if gate == nil {
		return fmt.Errorf("stage %s not found", stage)
	}

	if gate.Status != "pending-review" {
		return fmt.Errorf("gate %s is %s, can only reject pending-review gates", stage, gate.Status)
	}

	if err := st.RejectGate(stage, reason); err != nil {
		return err
	}

	if err := state.Save(root, st); err != nil {
		return err
	}

	red := color.New(color.FgRed, color.Bold)
	red.Printf("✗ Gate %s rejected\n", stage)
	if reason != "" {
		fmt.Printf("Reason: %s\n", reason)
	}
	fmt.Printf("Status reset to 'open' for rework\n")

	return nil
}

func init() {
	gateCmd.Flags().Bool("approve", false, "Manually approve a pending gate")
	gateCmd.Flags().Bool("reject", false, "Reject a pending gate")
	gateCmd.Flags().String("reason", "", "Reason for rejection")
	gateCmd.Flags().String("reviewer", "", "Set gate reviewer (auto|human)")
	rootCmd.AddCommand(gateCmd)
}