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
	Long:  "Creates a .foreman/ directory with initial project files.",
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
		fmt.Printf("Initialized foreman project in %s\n", root)
		fmt.Println()

		dim := color.New(color.Faint)
		dim.Println("Created:")
		dim.Println("  .foreman/project.yaml")
		dim.Println("  .foreman/plan.md")
		dim.Println("  .foreman/design.md")
		dim.Println("  .foreman/progress.yaml")
		dim.Println("  .foreman/lanes/")
		fmt.Println()

		cyan := color.New(color.FgCyan)
		cyan.Println("Next steps:")
		fmt.Println("  foreman req set      # Set project requirements")
		fmt.Println("  foreman plan set     # Define the plan")
		fmt.Println("  foreman design set   # Document the design")
		fmt.Println("  foreman lane add     # Add work lanes")

		return nil
	},
}

func init() {
	initCmd.Flags().String("name", "", "Project name (defaults to directory name)")
	initCmd.Flags().String("dir", "", "Directory to initialize (defaults to cwd)")
	rootCmd.AddCommand(initCmd)
}
