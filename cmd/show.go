package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show issue details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, sqlDB, err := openService()
		if err != nil {
			return err
		}
		defer func() { _ = sqlDB.Close() }()

		id, err := parseIDs(args)
		if err != nil {
			return err
		}

		issue, err := svc.Get(id[0])
		if err != nil {
			return err
		}

		stdoutPrintf(cmd, "ID: %s\n", issue.ID)
		stdoutPrintf(cmd, "Title: %s\n", issue.Title)
		stdoutPrintf(cmd, "Type: %s\n", issue.Type)
		stdoutPrintf(cmd, "Priority: P%d\n", issue.Priority)
		stdoutPrintf(cmd, "Status: %s\n", issue.Status)
		if issue.ClaimedAt != nil {
			stdoutPrintf(cmd, "Claimed at: %s\n", issue.ClaimedAt.Format("2006-01-02 15:04:05"))
		}
		if issue.ClaimExpiresAt != nil {
			stdoutPrintf(cmd, "Claim expires: %s\n", issue.ClaimExpiresAt.Format("2006-01-02 15:04:05"))
		}
		if issue.ParentID != nil {
			stdoutPrintf(cmd, "Parent: %s\n", *issue.ParentID)
		}
		stdoutPrintf(cmd, "Created: %s\n", issue.CreatedAt.Format("2006-01-02 15:04:05"))
		stdoutPrintf(cmd, "Updated: %s\n", issue.UpdatedAt.Format("2006-01-02 15:04:05"))
		if issue.ClosedAt != nil {
			stdoutPrintf(cmd, "Closed: %s\n", issue.ClosedAt.Format("2006-01-02 15:04:05"))
		}
		if strings.TrimSpace(issue.Description) != "" {
			stdoutPrintln(cmd)
			stdoutPrintln(cmd, "Description:")
			stdoutPrintln(cmd, issue.Description)
		}

		children, err := svc.Children(issue.ID)
		if err != nil {
			return err
		}
		deps, err := svc.Dependencies(issue.ID)
		if err != nil {
			return err
		}
		dependents, err := svc.Dependents(issue.ID)
		if err != nil {
			return err
		}

		stdoutPrintln(cmd)
		stdoutPrintln(cmd, "Children:")
		if len(children) == 0 {
			stdoutPrintln(cmd, "  none")
		} else {
			for _, child := range children {
				stdoutPrintf(cmd, "  %s [%s P%d %s] %s\n", child.ID, child.Type, child.Priority, child.Status, child.Title)
			}
		}

		stdoutPrintln(cmd, "Dependencies:")
		if len(deps) == 0 {
			stdoutPrintln(cmd, "  none")
		} else {
			for _, dep := range deps {
				stdoutPrintf(cmd, "  %s [%s]\n", dep.ID, dep.Title)
			}
		}

		stdoutPrintln(cmd, "Dependents:")
		if len(dependents) == 0 {
			stdoutPrintln(cmd, "  none")
		} else {
			for _, dependent := range dependents {
				stdoutPrintf(cmd, "  %s [%s]\n", dependent.ID, dependent.Title)
			}
		}

		return nil
	},
}

// init wires command flags and registration.
func init() {
	rootCmd.AddCommand(showCmd)
}
