package cmd

import (
	"fmt"
	"strings"

	"github.com/rpcarvs/faz/internal/model"
	"github.com/spf13/cobra"
)

var (
	listType     string
	listStatus   string
	listPriority int
	listParent   string
	listPlan     string
	listWork     string
	listAll      bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List issues with optional filters",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, sqlDB, err := openService()
		if err != nil {
			return err
		}
		defer func() { _ = sqlDB.Close() }()

		filter := model.ListFilter{
			Type:   listType,
			Status: listStatus,
			All:    listAll,
		}
		if cmd.Flags().Changed("priority") {
			filter.Priority = &listPriority
		}
		if cmd.Flags().Changed("parent") {
			normalized, err := parseIDs([]string{listParent})
			if err != nil {
				return err
			}
			filter.ParentID = normalized[0]
		}
		if cmd.Flags().Changed("plan") {
			if strings.TrimSpace(listPlan) == "" {
				return fmt.Errorf("plan filter cannot be empty")
			}
			filter.PlanID = listPlan
		}
		if cmd.Flags().Changed("work") {
			if strings.TrimSpace(listWork) == "" {
				return fmt.Errorf("work filter cannot be empty")
			}
			filter.WorkID = listWork
		}

		issues, err := svc.List(filter)
		if err != nil {
			return err
		}
		printIssueList(cmd.OutOrStdout(), issues)
		return nil
	},
}

// init wires command flags and registration.
func init() {
	listCmd.Flags().StringVar(&listType, "type", "", "Filter by issue type")
	listCmd.Flags().StringVar(&listStatus, "status", "", "Filter by issue status")
	listCmd.Flags().IntVar(&listPriority, "priority", 2, "Filter by priority (0-3)")
	listCmd.Flags().StringVar(&listParent, "parent", "", "Filter by parent ID")
	listCmd.Flags().StringVar(&listPlan, "plan", "", "Filter by SDD plan identifier")
	listCmd.Flags().StringVar(&listWork, "work", "", "Filter by SDD work identifier")
	listCmd.Flags().BoolVar(&listAll, "all", false, "Include closed issues")
	rootCmd.AddCommand(listCmd)
}
