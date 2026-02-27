package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

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
			fmt.Println("No ready work")
			return nil
		}
		printIssueTable(issues)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(readyCmd)
}
