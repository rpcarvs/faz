package cmd

import "github.com/spf13/cobra"

var recapCmd = &cobra.Command{
	Use:   "recap",
	Short: "Show complete command recap with examples",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("faz recap")
		cmd.Println()
		cmd.Println("Purpose")
		cmd.Println("  Local task tracking for AI agents and humans with no remote integration")
		cmd.Println()
		cmd.Println("Core flow")
		cmd.Println("  faz create \"Checkout revamp\" --type epic --priority 1 --description \"Improve checkout flow\"")
		cmd.Println("  faz create \"Add address validation\" --type task --priority 1 --parent faz-ab12 --description \"Client and server checks\"")
		cmd.Println("  faz dep add faz-ab12.0 faz-ab12")
		cmd.Println("  faz list --status open")
		cmd.Println("  faz children faz-ab12")
		cmd.Println("  faz ready")
		cmd.Println("  faz claim faz-ab12.0")
		cmd.Println("  faz close faz-ab12.0")
		cmd.Println("  faz info")
		cmd.Println()
		cmd.Println("Commands")
		cmd.Println("  onboard  Quick intro")
		cmd.Println("  info     Open count and latest 5 completed")
		cmd.Println("  create   Add issue")
		cmd.Println("  claim    Claim issue and set in_progress with lease")
		cmd.Println("  list     List issues with filters")
		cmd.Println("  children List direct child issues for a parent")
		cmd.Println("  ready    Show unblocked open work")
		cmd.Println("  show     Inspect issue with children and dependencies")
		cmd.Println("  update   Change issue fields")
		cmd.Println("  close    Mark issues closed")
		cmd.Println("  reopen   Reopen closed issues")
		cmd.Println("  delete   Permanently remove issues")
		cmd.Println("  dep      Manage dependencies")
		cmd.Println("  completion Generate shell completions")
		cmd.Println()
		cmd.Println("Version")
		cmd.Println("  faz -v")
		cmd.Println()
		cmd.Println("Claim rules")
		cmd.Println("  - in_progress is set by `faz claim`, not `faz update --status`")
		cmd.Println("  - Epics are not claimable")
		cmd.Println("  - Already-claimed tasks return non-zero from `faz claim`")
	},
}

// init wires command flags and registration.
func init() {
	rootCmd.AddCommand(recapCmd)
}
