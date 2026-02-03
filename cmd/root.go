package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "foreman",
	Short: "🏗️  foreman v2 — AI-native project management CLI",
	Long: `foreman v2 is a structured project workflow management tool for AI assistants.

It enforces a staged development workflow: requirements → design → phases → implementation.
Each stage has gates that validate completion before advancing to the next stage.

Phases are self-contained implementation units that get handed off to coding agents
via compiled briefs containing all the context needed for independent execution.`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
