package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/thinkshake/foreman/internal/config"
	"github.com/thinkshake/foreman/internal/project"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new foreman project",
	Long: `Creates a .foreman/ directory with initial project files.

Presets (v2.1):
  --preset minimal   Script/hotfix: no gates, straight to implementation
  --preset light     Small tool: requirements gate only, no design phase
  --preset full      Product: full workflow with design and phases

Legacy aliases (v3 compat):
  --preset nightly   Alias for minimal
  --preset product   Alias for full

TDD Integration:
  --tdd              Enable test-driven development mode`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		dir, _ := cmd.Flags().GetString("dir")
		preset, _ := cmd.Flags().GetString("preset")
		quick, _ := cmd.Flags().GetBool("quick")
		tdd, _ := cmd.Flags().GetBool("tdd")

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

		// --quick implies minimal preset
		if quick && preset == "" {
			preset = config.PresetMinimal
		}

		opts := project.InitOptions{
			Name:   name,
			Preset: preset,
			TDD:    tdd,
		}

		root, err := project.InitWithOptions(abs, opts)
		if err != nil {
			return err
		}

		green := color.New(color.FgGreen, color.Bold)
		green.Printf("✓ ")

		version := "v2.1"
		normalizedPreset := config.NormalizePreset(preset)
		mode := ""
		switch normalizedPreset {
		case config.PresetMinimal:
			mode = " (minimal mode)"
		case config.PresetLight:
			mode = " (light mode)"
		case config.PresetFull:
			mode = " (full workflow)"
		}
		if tdd {
			mode += " + TDD"
		}
		fmt.Printf("Initialized foreman %s project%s in %s\n", version, mode, root)
		fmt.Println()

		dim := color.New(color.Faint)
		dim.Println("Created:")
		dim.Println("  .foreman/config.yaml")
		dim.Println("  .foreman/state.yaml")
		dim.Println("  .foreman/requirements.md")
		if normalizedPreset == config.PresetFull {
			dim.Println("  .foreman/designs/")
			dim.Println("  .foreman/phases/")
		}
		dim.Println("  .foreman/briefs/")
		fmt.Println()

		cyan := color.New(color.FgCyan)
		cyan.Println("Next steps:")
		switch normalizedPreset {
		case config.PresetMinimal:
			fmt.Println("  foreman status              # Check current stage")
			fmt.Println("  # Edit .foreman/requirements.md with your task")
			fmt.Println("  foreman brief impl          # Generate brief and start building")
		case config.PresetLight:
			fmt.Println("  foreman status              # Check current stage")
			fmt.Println("  # Edit .foreman/requirements.md with your task")
			fmt.Println("  foreman gate requirements   # Advance to implementation")
			fmt.Println("  foreman brief impl          # Generate brief for coding agent")
		default:
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
	initCmd.Flags().String("preset", "", "Workflow preset: minimal, light, full (or aliases: nightly, product)")
	initCmd.Flags().Bool("quick", false, "Shorthand for --preset minimal")
	initCmd.Flags().Bool("tdd", false, "Enable test-driven development mode")
	rootCmd.AddCommand(initCmd)
}