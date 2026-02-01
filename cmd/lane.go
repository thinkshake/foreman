package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/thinkshake/foreman/internal/brief"
	"github.com/thinkshake/foreman/internal/lane"
	"github.com/thinkshake/foreman/internal/project"
)

var laneCmd = &cobra.Command{
	Use:   "lane",
	Short: "Manage project lanes (workstreams)",
}

var laneAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Create a new lane",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		summary, _ := cmd.Flags().GetString("summary")
		afterStr, _ := cmd.Flags().GetString("after")

		var deps []string
		if afterStr != "" {
			for _, d := range strings.Split(afterStr, ",") {
				d = strings.TrimSpace(d)
				if d != "" {
					deps = append(deps, d)
				}
			}
		}

		wd, _ := os.Getwd()
		root, err := project.FindRoot(wd)
		if err != nil {
			return err
		}

		l, err := lane.Add(root, name, summary, deps)
		if err != nil {
			return err
		}

		green := color.New(color.FgGreen, color.Bold)
		green.Print("✓ ")
		fmt.Printf("Created lane %02d-%s\n", l.Order, l.Name)

		if summary != "" {
			dim := color.New(color.Faint)
			dim.Printf("  %s\n", summary)
		}
		if len(deps) > 0 {
			dim := color.New(color.Faint)
			dim.Printf("  depends on: %s\n", strings.Join(deps, ", "))
		}

		return nil
	},
}

var laneListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all lanes with status and summary",
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, _ := os.Getwd()
		root, err := project.FindRoot(wd)
		if err != nil {
			return err
		}

		statusFilter, _ := cmd.Flags().GetString("status")

		lanes, err := lane.ListAll(root)
		if err != nil {
			return err
		}

		if len(lanes) == 0 {
			dim := color.New(color.Faint)
			dim.Println("No lanes defined. Use `foreman lane add <name>` to create one.")
			return nil
		}

		bold := color.New(color.Bold)
		dim := color.New(color.Faint)

		bold.Println("Lanes:")
		fmt.Println()

		shown := 0
		for _, l := range lanes {
			if statusFilter != "" && l.Status != statusFilter {
				continue
			}
			shown++

			sc := statusColorFn(l.Status)
			fmt.Printf("  %02d. ", l.Order)
			sc.Printf("[%-11s]", l.Status)
			fmt.Printf("  %s", l.Name)
			if l.Summary != "" {
				dim.Printf(" — %s", l.Summary)
			}
			fmt.Println()
			if len(l.Dependencies) > 0 {
				dim.Printf("       deps: %s\n", strings.Join(l.Dependencies, ", "))
			}
		}

		if shown == 0 && statusFilter != "" {
			dim.Printf("No lanes with status %q.\n", statusFilter)
		}

		return nil
	},
}

var laneShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Print lane spec (.md file)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, _ := os.Getwd()
		root, err := project.FindRoot(wd)
		if err != nil {
			return err
		}

		spec, err := lane.GetSpec(root, args[0])
		if err != nil {
			return err
		}
		fmt.Print(spec)
		return nil
	},
}

var laneSetCmd = &cobra.Command{
	Use:   "set <name>",
	Short: "Read lane spec from stdin and save to .md file",
	Args:  cobra.ExactArgs(1),
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

		if err := lane.SetSpec(root, args[0], string(data)); err != nil {
			return err
		}

		green := color.New(color.FgGreen, color.Bold)
		green.Print("✓ ")
		fmt.Printf("Lane spec updated for %q.\n", args[0])
		return nil
	},
}

var laneStatusCmd = &cobra.Command{
	Use:   "status <name> <status>",
	Short: "Update lane status",
	Long:  fmt.Sprintf("Valid statuses: %s", strings.Join(lane.ValidStatuses, ", ")),
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, _ := os.Getwd()
		root, err := project.FindRoot(wd)
		if err != nil {
			return err
		}

		if err := lane.SetStatus(root, args[0], args[1]); err != nil {
			return err
		}

		green := color.New(color.FgGreen, color.Bold)
		green.Print("✓ ")
		sc := statusColorFn(args[1])
		fmt.Printf("Lane %q → ", args[0])
		sc.Println(args[1])
		return nil
	},
}

var laneBriefCmd = &cobra.Command{
	Use:   "brief <name>",
	Short: "Generate a self-contained brief for a coding agent",
	Long: `Compiles all project context, design, dependencies, and the lane spec
into a single document that a coding agent can use to work independently.

This is the primary output of foreman — the bridge between planning and execution.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, _ := os.Getwd()
		root, err := project.FindRoot(wd)
		if err != nil {
			return err
		}

		output, err := brief.Generate(root, args[0])
		if err != nil {
			return err
		}

		fmt.Print(output)
		return nil
	},
}

func init() {
	laneAddCmd.Flags().String("summary", "", "One-line summary of the lane")
	laneAddCmd.Flags().String("after", "", "Comma-separated list of dependency lane names")
	laneListCmd.Flags().String("status", "", "Filter lanes by status")

	laneCmd.AddCommand(laneAddCmd)
	laneCmd.AddCommand(laneListCmd)
	laneCmd.AddCommand(laneShowCmd)
	laneCmd.AddCommand(laneSetCmd)
	laneCmd.AddCommand(laneStatusCmd)
	laneCmd.AddCommand(laneBriefCmd)
	rootCmd.AddCommand(laneCmd)
}
