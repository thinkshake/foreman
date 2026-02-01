package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/thinkshake/foreman/internal/lane"
	"github.com/thinkshake/foreman/internal/project"
)

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Show compact project summary",
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

		lanes, err := lane.ListAll(root)
		if err != nil {
			return err
		}

		bold := color.New(color.Bold)
		cyan := color.New(color.FgCyan, color.Bold)
		dim := color.New(color.Faint)

		cyan.Printf("📋 %s", proj.Name)
		if proj.Description != "" {
			fmt.Printf(" — %s", proj.Description)
		}
		fmt.Println()

		if len(proj.TechStack) > 0 {
			dim.Printf("   Tech: %s\n", strings.Join(proj.TechStack, ", "))
		}
		if len(proj.Tags) > 0 {
			dim.Printf("   Tags: %s\n", strings.Join(proj.Tags, ", "))
		}
		fmt.Println()

		if len(lanes) == 0 {
			dim.Println("   No lanes defined yet.")
			return nil
		}

		bold.Println("Lanes:")
		for _, l := range lanes {
			statusColor := statusColorFn(l.Status)
			fmt.Printf("  %02d. ", l.Order)
			statusColor.Printf("%-12s", l.Status)
			fmt.Printf(" %s", l.Name)
			if l.Summary != "" {
				dim.Printf(" — %s", l.Summary)
			}
			fmt.Println()
		}

		// Progress summary
		done := 0
		for _, l := range lanes {
			if l.Status == "done" {
				done++
			}
		}
		fmt.Println()
		pct := 0
		if len(lanes) > 0 {
			pct = (done * 100) / len(lanes)
		}
		bold.Printf("Progress: %d/%d lanes done (%d%%)\n", done, len(lanes), pct)

		return nil
	},
}

func statusColorFn(status string) *color.Color {
	switch status {
	case "planned":
		return color.New(color.FgWhite, color.Faint)
	case "ready":
		return color.New(color.FgCyan)
	case "in-progress":
		return color.New(color.FgYellow, color.Bold)
	case "review":
		return color.New(color.FgMagenta)
	case "done":
		return color.New(color.FgGreen, color.Bold)
	default:
		return color.New(color.FgWhite)
	}
}

func init() {
	rootCmd.AddCommand(summaryCmd)
}
