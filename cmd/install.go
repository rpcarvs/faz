package cmd

import "faz/cmd/installskills"

func init() {
	rootCmd.AddCommand(installskills.NewCommand())
}
