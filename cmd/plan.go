package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/thinkshake/foreman/internal/project"
)

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Manage project plan",
}

var planShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Print plan.md",
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, _ := os.Getwd()
		root, err := project.FindRoot(wd)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(project.PlanPath(root))
		if err != nil {
			return fmt.Errorf("failed to read plan.md: %w", err)
		}
		fmt.Print(string(data))
		return nil
	},
}

var planSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Read plan from stdin and save to plan.md",
	Long:  "Reads plan text from stdin. Example: echo '# My Plan' | foreman plan set",
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, _ := os.Getwd()
		root, err := project.FindRoot(wd)
		if err != nil {
			return err
		}

		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}

		if err := os.WriteFile(project.PlanPath(root), data, 0644); err != nil {
			return err
		}

		green := color.New(color.FgGreen, color.Bold)
		green.Print("✓ ")
		fmt.Println("Plan updated.")
		return nil
	},
}

func init() {
	planCmd.AddCommand(planShowCmd)
	planCmd.AddCommand(planSetCmd)
	rootCmd.AddCommand(planCmd)
}
