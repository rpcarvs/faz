package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var childrenCmd = &cobra.Command{
	Use:   "children <parent-id>",
	Short: "List direct children for an issue",
	Args:  cobra.ExactArgs(1),
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

		children, err := svc.Children(ids[0])
		if err != nil {
			return err
		}

		fmt.Printf("Children of %s:\n", ids[0])
		printIssueList(children)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(childrenCmd)
}
