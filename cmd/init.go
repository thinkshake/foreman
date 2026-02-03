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
	Long:  "Creates a .foreman/ directory with initial project files and v2 structure.",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		dir, _ := cmd.Flags().GetString("dir")

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

		root, err := project.Init(abs, name)
		if err != nil {
			return err
		}

		green := color.New(color.FgGreen, color.Bold)
		green.Printf("✓ ")
		fmt.Printf("Initialized foreman v2 project in %s\n", root)
		fmt.Println()

		dim := color.New(color.Faint)
		dim.Println("Created:")
		dim.Println("  .foreman/config.yaml")
		dim.Println("  .foreman/state.yaml")
		dim.Println("  .foreman/requirements.md")
		dim.Println("  .foreman/designs/")
		dim.Println("  .foreman/phases/")
		dim.Println("  .foreman/briefs/")
		fmt.Println()

		cyan := color.New(color.FgCyan)
		cyan.Println("Next steps:")
		fmt.Println("  foreman status              # Check current stage")
		fmt.Println("  # Edit .foreman/requirements.md with your project requirements")
		fmt.Println("  foreman gate requirements   # Validate and advance past requirements")

		return nil
	},
}

func init() {
	initCmd.Flags().String("name", "", "Project name (defaults to directory name)")
	initCmd.Flags().String("dir", "", "Directory to initialize (defaults to cwd)")
	rootCmd.AddCommand(initCmd)
}