package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show open count and latest completed tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, sqlDB, err := openService()
		if err != nil {
			return err
		}
		defer func() { _ = sqlDB.Close() }()

		openCount, completed, err := svc.Info()
		if err != nil {
			return err
		}

		fmt.Printf("Open issues: %d\n", openCount)
		fmt.Println()
		fmt.Println("Latest completed (max 5):")
		if len(completed) == 0 {
			fmt.Println("  none")
			return nil
		}
		for _, issue := range completed {
			closedAt := "unknown"
			if issue.ClosedAt != nil {
				closedAt = issue.ClosedAt.Format("2006-01-02 15:04")
			}
			fmt.Printf("  %s [%s P%d] %s (closed %s)\n", issue.ID, issue.Type, issue.Priority, issue.Title, closedAt)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
