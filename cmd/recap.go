package cmd

import "github.com/spf13/cobra"

var recapCmd = &cobra.Command{
	Use:   "recap",
	Short: "Show complete command recap with examples",
	Run: func(cmd *cobra.Command, args []string) {
		stdoutPrintln(cmd, "faz recap")
		stdoutPrintln(cmd)
		stdoutPrintln(cmd, "Purpose")
		stdoutPrintln(cmd, "  Local task tracking for AI agents and humans with no remote integration")
		stdoutPrintln(cmd)
		stdoutPrintln(cmd, "Core flow")
		stdoutPrintln(cmd, "  faz create \"Checkout revamp\" --type epic --priority 1 --description \"Improve checkout flow\"")
		stdoutPrintln(cmd, "  faz create \"Add address validation\" --type task --priority 1 --parent faz-ab12 --description \"Client and server checks\"")
		stdoutPrintln(cmd, "  faz dep add faz-ab12.0 faz-ab12")
		stdoutPrintln(cmd, "  faz list --status open")
		stdoutPrintln(cmd, "  faz children faz-ab12")
		stdoutPrintln(cmd, "  faz ready")
		stdoutPrintln(cmd, "  faz claim faz-ab12.0")
		stdoutPrintln(cmd, "  faz close faz-ab12.0")
		stdoutPrintln(cmd, "  faz info")
		stdoutPrintln(cmd)
		stdoutPrintln(cmd, "Agent install")
		stdoutPrintln(cmd, "  faz install codex         Install Codex skill, context, and SessionStart hook")
		stdoutPrintln(cmd, "  faz install claude        Install Claude skill, context, and SessionStart hook")
		stdoutPrintln(cmd, "  faz install codex --local Install into current Git repository")
		stdoutPrintln(cmd, "  faz install claude --local Install into current Git repository")
		stdoutPrintln(cmd)
		stdoutPrintln(cmd, "Commands")
		stdoutPrintln(cmd, "  onboard  Quick intro")
		stdoutPrintln(cmd, "  info     Open count and latest 5 completed")
		stdoutPrintln(cmd, "  create   Add issue")
		stdoutPrintln(cmd, "  claim    Claim issue and set in_progress with lease")
		stdoutPrintln(cmd, "  list     List issues with filters")
		stdoutPrintln(cmd, "  children List direct child issues for a parent")
		stdoutPrintln(cmd, "  ready    Show unblocked open work")
		stdoutPrintln(cmd, "  show     Inspect issue with children and dependencies")
		stdoutPrintln(cmd, "  update   Change issue fields")
		stdoutPrintln(cmd, "  close    Mark issues closed")
		stdoutPrintln(cmd, "  reopen   Reopen closed issues")
		stdoutPrintln(cmd, "  delete   Permanently remove issues")
		stdoutPrintln(cmd, "  dep      Manage dependencies")
		stdoutPrintln(cmd, "  install  Install Codex or Claude integration")
		stdoutPrintln(cmd, "  completion Generate shell completions")
		stdoutPrintln(cmd)
		stdoutPrintln(cmd, "Version")
		stdoutPrintln(cmd, "  faz -v")
		stdoutPrintln(cmd)
		stdoutPrintln(cmd, "Type reminder")
		stdoutPrintln(cmd, "  Classify each issue type: task, bug, feature, chore, decision")
		stdoutPrintln(cmd)
		stdoutPrintln(cmd, "Claim rules")
		stdoutPrintln(cmd, "  - in_progress is set by `faz claim`, not `faz update --status`")
		stdoutPrintln(cmd, "  - Epics are not claimable")
		stdoutPrintln(cmd, "  - Already-claimed tasks return non-zero from `faz claim`")
	},
}

// init wires command flags and registration.
func init() {
	rootCmd.AddCommand(recapCmd)
}
