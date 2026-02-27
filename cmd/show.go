package cmd

import (
	"fmt"
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

		fmt.Printf("ID: %s\n", issue.ID)
		fmt.Printf("Title: %s\n", issue.Title)
		fmt.Printf("Type: %s\n", issue.Type)
		fmt.Printf("Priority: P%d\n", issue.Priority)
		fmt.Printf("Status: %s\n", issue.Status)
		if issue.ParentID != nil {
			fmt.Printf("Parent: %s\n", *issue.ParentID)
		}
		fmt.Printf("Created: %s\n", issue.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Updated: %s\n", issue.UpdatedAt.Format("2006-01-02 15:04:05"))
		if issue.ClosedAt != nil {
			fmt.Printf("Closed: %s\n", issue.ClosedAt.Format("2006-01-02 15:04:05"))
		}
		if strings.TrimSpace(issue.Description) != "" {
			fmt.Println()
			fmt.Println("Description:")
			fmt.Println(issue.Description)
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

		fmt.Println()
		fmt.Println("Children:")
		if len(children) == 0 {
			fmt.Println("  none")
		} else {
			for _, child := range children {
				fmt.Printf("  %s [%s P%d %s] %s\n", child.ID, child.Type, child.Priority, child.Status, child.Title)
			}
		}

		fmt.Println("Dependencies:")
		if len(deps) == 0 {
			fmt.Println("  none")
		} else {
			for _, dep := range deps {
				fmt.Printf("  %s [%s]\n", dep.ID, dep.Title)
			}
		}

		fmt.Println("Dependents:")
		if len(dependents) == 0 {
			fmt.Println("  none")
		} else {
			for _, dependent := range dependents {
				fmt.Printf("  %s [%s]\n", dependent.ID, dependent.Title)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
}
