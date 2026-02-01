package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/thinkshake/foreman/internal/project"
)

var designCmd = &cobra.Command{
	Use:   "design",
	Short: "Manage project design",
}

var designShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Print design.md",
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, _ := os.Getwd()
		root, err := project.FindRoot(wd)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(project.DesignPath(root))
		if err != nil {
			return fmt.Errorf("failed to read design.md: %w", err)
		}
		fmt.Print(string(data))
		return nil
	},
}

var designSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Read design from stdin and save to design.md",
	Long:  "Reads design text from stdin. Example: echo '# Architecture' | foreman design set",
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

		if err := os.WriteFile(project.DesignPath(root), data, 0644); err != nil {
			return err
		}

		green := color.New(color.FgGreen, color.Bold)
		green.Print("✓ ")
		fmt.Println("Design updated.")
		return nil
	},
}

func init() {
	designCmd.AddCommand(designShowCmd)
	designCmd.AddCommand(designSetCmd)
	rootCmd.AddCommand(designCmd)
}
