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
		cmd.Println("- `faz close <id>` - Complete work")
	},
}

func init() {
	rootCmd.AddCommand(onboardCmd)
}
