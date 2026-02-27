package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var closeCmd = &cobra.Command{
	Use:   "close <id> [id...]",
	Short: "Close one or more issues",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, sqlDB, err := openService()
		if err != nil {
			return err
		}
		defer func() { _ = sqlDB.Close() }()

		ids, err := parseIDs(args)
		if err != nil {
			return err
		}

		for _, id := range ids {
			if err := svc.Close(id); err != nil {
				return err
			}
			fmt.Printf("Closed issue: %s\n", id)
			fmt.Printf("  Status: closed\n")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(closeCmd)
}
