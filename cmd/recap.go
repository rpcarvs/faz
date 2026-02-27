package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var recapCmd = &cobra.Command{
	Use:   "recap",
	Short: "Show complete command recap with examples",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("faz recap")
		fmt.Println()
		fmt.Println("Purpose")
		fmt.Println("  Local task tracking for AI agents and humans with no remote integration")
		fmt.Println()
		fmt.Println("Core flow")
		fmt.Println("  faz create \"Checkout revamp\" --type epic --priority 1 --description \"Improve checkout flow\"")
		fmt.Println("  faz create \"Add address validation\" --type task --priority 1 --parent faz-ab12 --description \"Client and server checks\"")
		fmt.Println("  faz dep add faz-ab12.0 faz-ab12")
		fmt.Println("  faz list --status open")
		fmt.Println("  faz children faz-ab12")
		fmt.Println("  faz ready")
		fmt.Println("  faz claim faz-ab12.0")
		fmt.Println("  faz close faz-ab12.0")
		fmt.Println("  faz info")
		fmt.Println()
		fmt.Println("Commands")
		fmt.Println("  onboard  Quick intro")
		fmt.Println("  info     Open count and latest 5 completed")
		fmt.Println("  create   Add issue")
		fmt.Println("  claim    Claim issue and set in_progress with lease")
		fmt.Println("  list     List issues with filters")
		fmt.Println("  children List direct child issues for a parent")
		fmt.Println("  ready    Show unblocked open work")
		fmt.Println("  show     Inspect issue with children and dependencies")
		fmt.Println("  update   Change issue fields")
		fmt.Println("  close    Mark issues closed")
		fmt.Println("  reopen   Reopen closed issues")
		fmt.Println("  delete   Permanently remove issues")
		fmt.Println("  dep      Manage dependencies")
	},
}

func init() {
	rootCmd.AddCommand(recapCmd)
}
