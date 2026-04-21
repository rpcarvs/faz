package cmd

import "github.com/spf13/cobra"

var onboardCmd = &cobra.Command{
	Use:   "onboard",
	Short: "Show a quick faz introduction",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("## Issue Tracking")
		cmd.Println("")
		cmd.Println("This project uses **faz** for issue tracking.")
		cmd.Println("Run `faz recap` to a quick recap on faz commands and tools.")
		cmd.Println("")
		cmd.Println("**Quick reference:**")
		cmd.Println("- `faz info` - Show open count and latest completed tasks")
		cmd.Println("- `faz ready` - Find unblocked work")
		cmd.Println("- `faz create \"Title\" --type task --priority 2` - Create issue")
			cmd.Println("- `faz claim <id>` - Claim work and mark in_progress")
			cmd.Println("- `faz close <id>` - Complete work")
			cmd.Println("- Classify each issue type: task, bug, feature, chore, decision")
			cmd.Println("")
			cmd.Println("Claim rules:")
		cmd.Println("- `in_progress` is set by `faz claim`, not by `faz update --status`")
		cmd.Println("- Epics are containers and cannot be claimed")
		cmd.Println("- If a task is already claimed, `faz claim` exits non-zero")
	},
}

// init wires command flags and registration.
func init() {
	rootCmd.AddCommand(onboardCmd)
}
