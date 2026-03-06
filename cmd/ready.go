package cmd

import "github.com/spf13/cobra"

var readyCmd = &cobra.Command{
	Use:   "ready",
	Short: "List unblocked open work",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, sqlDB, err := openService()
		if err != nil {
			return err
		}
		defer func() { _ = sqlDB.Close() }()

		issues, err := svc.Ready()
		if err != nil {
			return err
		}

		if len(issues) == 0 {
			cmd.Println("No ready work")
			return nil
		}
		printIssueList(cmd.OutOrStdout(), issues)
		return nil
	},
}

// init wires command flags and registration.
func init() {
	rootCmd.AddCommand(readyCmd)
}
