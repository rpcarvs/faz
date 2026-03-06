package cmd

import "faz/cmd/installskills"

// init wires command flags and registration.
func init() {
	rootCmd.AddCommand(installskills.NewCommand())
}
