package cmd

import "github.com/spf13/cobra"

var onboardCmd = &cobra.Command{
	Use:   "onboard",
	Short: "Show a quick faz introduction",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("## Issue Tracking")
		cmd.Println("")
		cmd.Println("This project uses **faz** for issue tracking.")
		cmd.Println("Run `faz recap` for a fuller command reference.")
		cmd.Println("Faz resolves the current Git repository root, so commands can run from any subdirectory.")
		cmd.Println("")
		cmd.Println("**Agent quick start:**")
		cmd.Println("- `faz info` - Show open count and latest completed tasks")
		cmd.Println("- `faz ready` - Find unblocked work")
		cmd.Println("- `faz create \"Title\" --type task --priority 2` - Create issue")
		cmd.Println("- `faz dep add <task-id> <blocker-id>` - Mark blocked work")
		cmd.Println("- `faz claim <id>` - Claim work and mark in_progress")
		cmd.Println("- `faz show <id>` - Read details, children, and dependencies")
		cmd.Println("- `faz close <id>` - Complete work")
		cmd.Println("")
		cmd.Println("**Rules of use:**")
		cmd.Println("- Create atomic issues with rich descriptions before starting work")
		cmd.Println("- Classify each issue type: task, bug, feature, chore, decision")
		cmd.Println("- Use epics as containers and child issues as executable work")
		cmd.Println("- Set blockers explicitly with `faz dep` when work depends on other work")
		cmd.Println("- Keep scope current with `faz update`; close finished work with `faz close`")
		cmd.Println("")
		cmd.Println("**Claim safety:**")
		cmd.Println("- Claim exactly one non-epic issue before coding")
		cmd.Println("- `in_progress` is set by `faz claim`, not by `faz update --status`")
		cmd.Println("- If `faz claim` says the issue is already claimed, pick another ready issue")
	},
}

// init wires command flags and registration.
func init() {
	rootCmd.AddCommand(onboardCmd)
}
