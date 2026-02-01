package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "foreman",
	Short: "🏗️  foreman — AI-native project management CLI",
	Long: `foreman is a file-based project management tool for AI assistants.

It handles PM/EM responsibilities: requirements gathering, planning,
high-level design, task breakdown into lanes, and progress tracking.

Lanes are self-contained work units that get handed off to coding agents
via compiled briefs containing all the context needed for independent execution.`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
