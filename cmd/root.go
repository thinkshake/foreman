package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "foreman",
	Short: "🏗️  foreman v3 — AI-native project management CLI",
	Long: `foreman v3 is a structured project workflow management tool for AI assistants.

FULL MODE (default):
  Enforces a staged workflow: requirements → design → phases → implementation.
  Each stage has gates that validate completion before advancing.

QUICK MODE (v3):
  Streamlined workflow: requirements → implementation.
  Use 'foreman quick "<task>"' or 'foreman init --preset nightly'.
  Perfect for scripts, quick fixes, and nightly builds.

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
