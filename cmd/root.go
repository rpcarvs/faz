package cmd

import (
	"context"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "faz",
	Short: "Simple local task tracking for AI agent workflows",
	Long: `faz is a lightweight task tracker.
It stores tasks in a project-local SQLite database and keeps epics, tasks,
and dependencies in a simple graph model without external integrations.`,
}

// Execute runs the root command and exits with non-zero on failure.
func Execute() {
	if err := fang.Execute(context.Background(), rootCmd, fang.WithoutManpage()); err != nil {
		os.Exit(1)
	}
}

// init wires command flags and registration.
func init() {
	rootCmd.SilenceUsage = true
}
