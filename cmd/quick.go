package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/thinkshake/foreman/internal/brief"
	"github.com/thinkshake/foreman/internal/config"
	"github.com/thinkshake/foreman/internal/state"
)

var quickCmd = &cobra.Command{
	Use:   "quick <task>",
	Short: "Quick build: skip design/phases, go straight to code",
	Long: `Quick mode for rapid development without full ceremony.

Skips design and phases stages, going directly from requirements to implementation.
Perfect for small scripts, quick fixes, and nightly builds.

Examples:
  foreman quick "build a CLI that fetches weather data"
  foreman quick "add rate limiting to the API" --dir ./myproject

This creates a minimal .foreman/ with:
  - requirements.md (populated with your task)
  - briefs/impl.md (ready for coding agent handoff)`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		task := args[0]
		dir, _ := cmd.Flags().GetString("dir")
		name, _ := cmd.Flags().GetString("name")
		autoAdvance, _ := cmd.Flags().GetInt("auto-advance")
		generateBrief, _ := cmd.Flags().GetBool("brief")

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

		if name == "" {
			name = filepath.Base(abs)
		}

		// Check if .foreman already exists
		foremanDir := filepath.Join(abs, ".foreman")
		if _, err := os.Stat(foremanDir); err == nil {
			return fmt.Errorf(".foreman/ already exists in %s\nUse 'foreman status' to check existing project", abs)
		}

		// Create directory structure
		dirs := []string{
			foremanDir,
			filepath.Join(foremanDir, "briefs"),
		}
		for _, d := range dirs {
			if err := os.MkdirAll(d, 0755); err != nil {
				return fmt.Errorf("failed to create %s: %w", d, err)
			}
		}

		// Default auto-advance for quick mode
		if autoAdvance == 0 {
			autoAdvance = 70
		}

		// Create config
		cfg := &config.Config{
			Name:        name,
			Description: task,
			TechStack:   []string{},
			Reviewers: config.Reviewers{
				Default:   "auto",
				Overrides: make(map[string]string),
			},
			Preset:      "nightly",
			AutoAdvance: autoAdvance,
		}
		if err := config.Save(abs, cfg); err != nil {
			return fmt.Errorf("failed to create config.yaml: %w", err)
		}

		// Create state with task
		st := state.NewQuickMode(task, autoAdvance)
		if err := state.Save(abs, st); err != nil {
			return fmt.Errorf("failed to create state.yaml: %w", err)
		}

		// Create requirements.md with task
		reqContent := fmt.Sprintf(`# Task: %s

## Goal
%s

## Features
_To be filled in based on the task above._

## Tech Stack
_Specify preferred technologies._

## Success Criteria
- Task is complete and working
- Code is clean and tested
`, name, task)

		reqPath := filepath.Join(foremanDir, "requirements.md")
		if err := os.WriteFile(reqPath, []byte(reqContent), 0644); err != nil {
			return fmt.Errorf("failed to create requirements.md: %w", err)
		}

		// Output
		green := color.New(color.FgGreen, color.Bold)
		green.Printf("✓ ")
		fmt.Printf("Quick project initialized: %s\n", name)
		fmt.Println()

		dim := color.New(color.Faint)
		dim.Printf("Task: %s\n", task)
		dim.Printf("Mode: quick (requirements → implementation)\n")
		dim.Printf("Auto-advance: %d%% confidence\n", autoAdvance)
		fmt.Println()

		// Generate brief if requested
		if generateBrief {
			// Auto-approve requirements gate for quick brief generation
			st.Gates["requirements"].Status = "approved"
			st.CurrentStage = "implementation"
			st.Gates["implementation"] = &state.Gate{Status: "open"}
			if err := state.Save(abs, st); err != nil {
				return fmt.Errorf("failed to update state: %w", err)
			}

			briefContent, err := brief.GenerateQuickBrief(abs, task)
			if err != nil {
				return fmt.Errorf("failed to generate brief: %w", err)
			}

			briefPath := filepath.Join(foremanDir, "briefs", "impl.md")
			if err := os.WriteFile(briefPath, []byte(briefContent), 0644); err != nil {
				return fmt.Errorf("failed to write brief: %w", err)
			}

			green.Printf("✓ ")
			fmt.Printf("Generated brief: .foreman/briefs/impl.md\n")
			fmt.Println()

			cyan := color.New(color.FgCyan)
			cyan.Println("Brief content:")
			fmt.Println(strings.Repeat("=", 60))
			fmt.Print(briefContent)
		} else {
			cyan := color.New(color.FgCyan)
			cyan.Println("Next steps:")
			fmt.Println("  # Review/edit .foreman/requirements.md")
			fmt.Println("  foreman gate requirements   # Approve requirements")
			fmt.Println("  foreman brief impl          # Generate brief for coding agent")
			fmt.Println()
			fmt.Println("Or add --brief to generate brief immediately:")
			fmt.Printf("  foreman quick %q --brief\n", task)
		}

		return nil
	},
}

func init() {
	quickCmd.Flags().String("dir", "", "Directory to initialize (defaults to cwd)")
	quickCmd.Flags().String("name", "", "Project name (defaults to directory name)")
	quickCmd.Flags().Int("auto-advance", 70, "Auto-advance confidence threshold (0-100)")
	quickCmd.Flags().Bool("brief", false, "Generate brief immediately (skip manual approval)")
	rootCmd.AddCommand(quickCmd)
}
