package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "faz",
	Short: "Simple local task tracking for AI agent workflows",
	Long: `faz is a lightweight task tracker.
It stores tasks in a project-local SQLite database and keeps epics, tasks,
and dependencies in a simple graph model without external integrations.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.SilenceUsage = true
}
