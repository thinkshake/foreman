package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/thinkshake/foreman/internal/project"
)

var reqCmd = &cobra.Command{
	Use:   "req",
	Short: "Manage project requirements",
}

var reqShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Print requirements from project.yaml",
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, _ := os.Getwd()
		root, err := project.FindRoot(wd)
		if err != nil {
			return err
		}
		proj, err := project.Load(root)
		if err != nil {
			return err
		}
		if proj.Requirements == "" {
			dim := color.New(color.Faint)
			dim.Println("No requirements defined. Use `foreman req set` to add them.")
			return nil
		}
		fmt.Print(proj.Requirements)
		return nil
	},
}

var reqSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Read requirements from stdin and save",
	Long:  "Reads requirements text from stdin. Example: echo 'My requirements' | foreman req set",
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

		proj, err := project.Load(root)
		if err != nil {
			return err
		}

		proj.Requirements = string(data)
		if err := project.Save(root, proj); err != nil {
			return err
		}

		green := color.New(color.FgGreen, color.Bold)
		green.Print("✓ ")
		fmt.Println("Requirements updated.")
		return nil
	},
}

func init() {
	reqCmd.AddCommand(reqShowCmd)
	reqCmd.AddCommand(reqSetCmd)
	rootCmd.AddCommand(reqCmd)
}
