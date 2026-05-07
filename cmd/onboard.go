package cmd

import "github.com/spf13/cobra"

var onboardCmd = &cobra.Command{
	Use:   "onboard",
	Short: "Show a quick faz introduction",
	Run: func(cmd *cobra.Command, args []string) {
		stdoutPrintln(cmd, "## Issue Tracking")
		stdoutPrintln(cmd)
		stdoutPrintln(cmd, "This project uses **faz** for issue tracking.")
		stdoutPrintln(cmd, "Run `faz recap` for a fuller command reference.")
		stdoutPrintln(cmd, "Faz resolves the current Git repository root, so commands can run from any subdirectory.")
		stdoutPrintln(cmd)
		stdoutPrintln(cmd, "**Agent quick start:**")
		stdoutPrintln(cmd, "- `faz info` - Show open count and latest completed tasks")
		stdoutPrintln(cmd, "- `faz ready` - Find unblocked work")
		stdoutPrintln(cmd, "- `faz create \"Title\" --type task --priority 2` - Create issue")
		stdoutPrintln(cmd, "- `faz dep add <task-id> <blocker-id>` - Mark blocked work")
		stdoutPrintln(cmd, "- `faz claim <id>` - Claim work and mark in_progress")
		stdoutPrintln(cmd, "- `faz show <id>` - Read details, children, and dependencies")
		stdoutPrintln(cmd, "- `faz close <id>` - Complete work")
		stdoutPrintln(cmd)
		stdoutPrintln(cmd, "**Rules of use:**")
		stdoutPrintln(cmd, "- Create atomic issues with rich descriptions before starting work")
		stdoutPrintln(cmd, "- Classify each issue type: task, bug, feature, chore, decision")
		stdoutPrintln(cmd, "- Use epics as containers and child issues as executable work")
		stdoutPrintln(cmd, "- Set blockers explicitly with `faz dep` when work depends on other work")
		stdoutPrintln(cmd, "- Keep scope current with `faz update`; close finished work with `faz close`")
		stdoutPrintln(cmd)
		stdoutPrintln(cmd, "**Claim safety:**")
		stdoutPrintln(cmd, "- Claim exactly one non-epic issue before coding")
		stdoutPrintln(cmd, "- `in_progress` is set by `faz claim`, not by `faz update --status`")
		stdoutPrintln(cmd, "- If `faz claim` says the issue is already claimed, pick another ready issue")
	},
}

// init wires command flags and registration.
func init() {
	rootCmd.AddCommand(onboardCmd)
}
