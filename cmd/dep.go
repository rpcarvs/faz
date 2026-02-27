package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var depCmd = &cobra.Command{
	Use:   "dep",
	Short: "Manage issue dependencies",
}

var depAddCmd = &cobra.Command{
	Use:   "add <issue-id> <depends-on-id>",
	Short: "Add dependency",
	Args:  cobra.ExactArgs(2),
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
		if err := svc.AddDependency(ids[0], ids[1]); err != nil {
			return err
		}
		fmt.Printf("Added dependency:\n")
		fmt.Printf("  %s depends on %s\n", ids[0], ids[1])
		return nil
	},
}

var depRemoveCmd = &cobra.Command{
	Use:   "remove <issue-id> <depends-on-id>",
	Short: "Remove dependency",
	Args:  cobra.ExactArgs(2),
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
		if err := svc.RemoveDependency(ids[0], ids[1]); err != nil {
			return err
		}
		fmt.Printf("Removed dependency:\n")
		fmt.Printf("  %s no longer depends on %s\n", ids[0], ids[1])
		return nil
	},
}

var depListDirection string

var depListCmd = &cobra.Command{
	Use:   "list <id>",
	Short: "List dependencies or dependents",
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

		if depListDirection == "up" {
			items, err := svc.Dependents(id[0])
			if err != nil {
				return err
			}
			fmt.Printf("Issues blocked by %s:\n", id[0])
			printIssueTable(items)
			return nil
		}

		items, err := svc.Dependencies(id[0])
		if err != nil {
			return err
		}
		fmt.Printf("Dependencies for %s:\n", id[0])
		printIssueTable(items)
		return nil
	},
}

func init() {
	depListCmd.Flags().StringVar(&depListDirection, "direction", "down", "Direction: down (dependencies) or up (dependents)")
	depCmd.AddCommand(depAddCmd)
	depCmd.AddCommand(depRemoveCmd)
	depCmd.AddCommand(depListCmd)
	rootCmd.AddCommand(depCmd)
}
