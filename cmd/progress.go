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

var progressCmd = &cobra.Command{
	Use:   "progress",
	Short: "Show progress dashboard",
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, _ := os.Getwd()
		root, err := project.FindRoot(wd)
		if err != nil {
			return err
		}

		// Refresh progress first
		if err := lane.UpdateProgress(root); err != nil {
			return err
		}

		prog, err := project.LoadProgress(root)
		if err != nil {
			return err
		}

		proj, err := project.Load(root)
		if err != nil {
			return err
		}

		bold := color.New(color.Bold)
		cyan := color.New(color.FgCyan, color.Bold)
		dim := color.New(color.Faint)

		// Header
		cyan.Printf("📊 Progress: %s\n", proj.Name)
		fmt.Println()

		if prog.TotalLanes == 0 {
			dim.Println("No lanes defined yet.")
			return nil
		}

		// Status breakdown
		bold.Println("Status Breakdown:")
		statuses := []string{"planned", "ready", "in-progress", "review", "done"}
		icons := map[string]string{
			"planned":     "○",
			"ready":       "◎",
			"in-progress": "●",
			"review":      "◉",
			"done":        "✓",
		}
		for _, s := range statuses {
			count := prog.ByStatus[s]
			sc := statusColorFn(s)
			sc.Printf("  %s %-12s", icons[s], s)
			fmt.Printf(" %d", count)
			if count > 0 && prog.TotalLanes > 0 {
				bar := buildBar(count, prog.TotalLanes, 20)
				fmt.Printf("  %s", bar)
			}
			fmt.Println()
		}

		// Overall
		fmt.Println()
		done := prog.ByStatus["done"]
		pct := 0
		if prog.TotalLanes > 0 {
			pct = (done * 100) / prog.TotalLanes
		}

		bold.Printf("Overall: ")
		if pct == 100 {
			green := color.New(color.FgGreen, color.Bold)
			green.Printf("🎉 %d/%d lanes done (100%%)\n", done, prog.TotalLanes)
		} else {
			fmt.Printf("%d/%d lanes done (%d%%)\n", done, prog.TotalLanes, pct)
		}

		// Progress bar
		bar := buildProgressBar(pct, 30)
		fmt.Printf("  %s\n", bar)

		// Lane list
		fmt.Println()
		bold.Println("Lanes:")
		for _, l := range prog.Lanes {
			sc := statusColorFn(l.Status)
			sc.Printf("  %-12s", l.Status)
			fmt.Printf("  %s\n", l.Name)
		}

		return nil
	},
}

func buildBar(count, total, width int) string {
	filled := (count * width) / total
	if filled == 0 && count > 0 {
		filled = 1
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

func buildProgressBar(pct, width int) string {
	filled := (pct * width) / 100
	empty := width - filled
	return fmt.Sprintf("[%s%s] %d%%", strings.Repeat("█", filled), strings.Repeat("░", empty), pct)
}

func init() {
	rootCmd.AddCommand(progressCmd)
}
