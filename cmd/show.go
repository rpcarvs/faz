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

		cmd.Printf("ID: %s\n", issue.ID)
		cmd.Printf("Title: %s\n", issue.Title)
		cmd.Printf("Type: %s\n", issue.Type)
		cmd.Printf("Priority: P%d\n", issue.Priority)
		cmd.Printf("Status: %s\n", issue.Status)
		if issue.ClaimedAt != nil {
			cmd.Printf("Claimed at: %s\n", issue.ClaimedAt.Format("2006-01-02 15:04:05"))
		}
		if issue.ClaimExpiresAt != nil {
			cmd.Printf("Claim expires: %s\n", issue.ClaimExpiresAt.Format("2006-01-02 15:04:05"))
		}
		if issue.ParentID != nil {
			cmd.Printf("Parent: %s\n", *issue.ParentID)
		}
		cmd.Printf("Created: %s\n", issue.CreatedAt.Format("2006-01-02 15:04:05"))
		cmd.Printf("Updated: %s\n", issue.UpdatedAt.Format("2006-01-02 15:04:05"))
		if issue.ClosedAt != nil {
			cmd.Printf("Closed: %s\n", issue.ClosedAt.Format("2006-01-02 15:04:05"))
		}
		if strings.TrimSpace(issue.Description) != "" {
			cmd.Println()
			cmd.Println("Description:")
			cmd.Println(issue.Description)
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

		cmd.Println()
		cmd.Println("Children:")
		if len(children) == 0 {
			cmd.Println("  none")
		} else {
			for _, child := range children {
				cmd.Printf("  %s [%s P%d %s] %s\n", child.ID, child.Type, child.Priority, child.Status, child.Title)
			}
		}

		cmd.Println("Dependencies:")
		if len(deps) == 0 {
			cmd.Println("  none")
		} else {
			for _, dep := range deps {
				cmd.Printf("  %s [%s]\n", dep.ID, dep.Title)
			}
		}

		cmd.Println("Dependents:")
		if len(dependents) == 0 {
			cmd.Println("  none")
		} else {
			for _, dependent := range dependents {
				cmd.Printf("  %s [%s]\n", dependent.ID, dependent.Title)
			}
		}

		return nil
	},
}

// init wires command flags and registration.
func init() {
	rootCmd.AddCommand(showCmd)
}
