package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/thinkshake/foreman/internal/config"
	"github.com/thinkshake/foreman/internal/project"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage project configuration",
	Long:  "View and modify project configuration settings.",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  "Display the current project configuration from config.yaml.",
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

		// Pretty print the configuration
		data, err := yaml.Marshal(cfg)
		if err != nil {
			return err
		}

		fmt.Println("Current configuration:")
		fmt.Println()
		fmt.Print(string(data))

		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value using dot notation.

Examples:
  foreman config set name "my-project"
  foreman config set description "A sample project"
  foreman config set reviewers.default human
  foreman config set reviewers.overrides.requirements auto

Supported keys:
  name                          - Project name
  description                   - Project description  
  tech_stack                    - Technology stack (comma-separated)
  reviewers.default             - Default reviewer (auto|human)
  reviewers.overrides.<stage>   - Stage-specific reviewer override`,
	Args: cobra.ExactArgs(2),
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

		key := args[0]
		value := args[1]

		// Handle different configuration keys
		switch key {
		case "name":
			cfg.Name = value
		case "description":
			cfg.Description = value
		case "tech_stack":
			// Split comma-separated values
			if value == "" {
				cfg.TechStack = []string{}
			} else {
				cfg.TechStack = strings.Split(value, ",")
				// Trim spaces
				for i, tech := range cfg.TechStack {
					cfg.TechStack[i] = strings.TrimSpace(tech)
				}
			}
		case "reviewers.default":
			if value != "auto" && value != "human" {
				return fmt.Errorf("reviewers.default must be 'auto' or 'human', got: %s", value)
			}
			cfg.Reviewers.Default = value
		default:
			// Handle reviewers.overrides.<stage>
			if strings.HasPrefix(key, "reviewers.overrides.") {
				stage := strings.TrimPrefix(key, "reviewers.overrides.")
				if value != "auto" && value != "human" {
					return fmt.Errorf("reviewer must be 'auto' or 'human', got: %s", value)
				}
				cfg.Reviewers.SetReviewer(stage, value)
			} else {
				return fmt.Errorf("unknown configuration key: %s", key)
			}
		}

		// Save the updated configuration
		if err := config.Save(root, cfg); err != nil {
			return err
		}

		green := color.New(color.FgGreen)
		green.Printf("✓ ")
		fmt.Printf("Set %s = %s\n", key, formatValue(value))

		return nil
	},
}

func formatValue(value string) string {
	// Add quotes around string values for display
	if strings.Contains(value, " ") || value == "" {
		return fmt.Sprintf("%q", value)
	}
	return value
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configCmd)
}