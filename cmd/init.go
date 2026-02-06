package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/thinkshake/foreman/internal/project"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new foreman project",
	Long: `Creates a .foreman/ directory with initial project files.

Presets (v3):
  --preset nightly   Quick mode, auto-advance gates, minimal ceremony
  --preset product   Full workflow, human review for requirements/design`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		dir, _ := cmd.Flags().GetString("dir")
		preset, _ := cmd.Flags().GetString("preset")
		quick, _ := cmd.Flags().GetBool("quick")

		if dir == "" {
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			dir = wd
		}

		abs, err := filepath.Abs(dir)
		if err != nil {
			return err
		}

		// --quick implies nightly preset
		if quick && preset == "" {
			preset = "nightly"
		}

		opts := project.InitOptions{
			Name:   name,
			Preset: preset,
		}

		root, err := project.InitWithOptions(abs, opts)
		if err != nil {
			return err
		}

		green := color.New(color.FgGreen, color.Bold)
		green.Printf("✓ ")
		
		version := "v3"
		mode := ""
		if preset == "nightly" {
			mode = " (quick mode)"
		} else if preset == "product" {
			mode = " (full workflow)"
		}
		fmt.Printf("Initialized foreman %s project%s in %s\n", version, mode, root)
		fmt.Println()

		dim := color.New(color.Faint)
		dim.Println("Created:")
		dim.Println("  .foreman/config.yaml")
		dim.Println("  .foreman/state.yaml")
		dim.Println("  .foreman/requirements.md")
		if preset != "nightly" {
			dim.Println("  .foreman/designs/")
			dim.Println("  .foreman/phases/")
		}
		dim.Println("  .foreman/briefs/")
		fmt.Println()

		cyan := color.New(color.FgCyan)
		cyan.Println("Next steps:")
		if preset == "nightly" {
			fmt.Println("  foreman status              # Check current stage")
			fmt.Println("  # Edit .foreman/requirements.md with your task")
			fmt.Println("  foreman gate requirements   # Advance to implementation")
			fmt.Println("  foreman brief impl          # Generate brief for coding agent")
		} else {
			fmt.Println("  foreman status              # Check current stage")
			fmt.Println("  # Edit .foreman/requirements.md with your project requirements")
			fmt.Println("  foreman gate requirements   # Validate and advance past requirements")
		}

		return nil
	},
}

func init() {
	initCmd.Flags().String("name", "", "Project name (defaults to directory name)")
	initCmd.Flags().String("dir", "", "Directory to initialize (defaults to cwd)")
	initCmd.Flags().String("preset", "", "Workflow preset: nightly (quick) or product (full)")
	initCmd.Flags().Bool("quick", false, "Shorthand for --preset nightly")
	rootCmd.AddCommand(initCmd)
}