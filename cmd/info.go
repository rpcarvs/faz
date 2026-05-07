package cmd

import "github.com/spf13/cobra"

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

		stdoutPrintf(cmd, "Open issues: %d\n", openCount)
		stdoutPrintln(cmd)
		stdoutPrintln(cmd, "Latest completed (max 5):")
		if len(completed) == 0 {
			stdoutPrintln(cmd, "  none")
			return nil
		}
		for _, issue := range completed {
			closedAt := "unknown"
			if issue.ClosedAt != nil {
				closedAt = issue.ClosedAt.Format("2006-01-02 15:04")
			}
			stdoutPrintf(cmd, "  %s [%s P%d] %s (closed %s)\n", issue.ID, issue.Type, issue.Priority, issue.Title, closedAt)
		}

		return nil
	},
}

// init wires command flags and registration.
func init() {
	rootCmd.AddCommand(infoCmd)
}
