package cmd

import (
	"fmt"
	"time"

	"github.com/rpcarvs/faz/internal/model"
	"github.com/spf13/cobra"
)

var (
	monitorIntervalSeconds int
	monitorAll             bool
	monitorClaimed         bool
)

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Continuously refresh task list output",
	Long:  "Monitor loops over list output with a terminal refresh and a configurable interval.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if monitorIntervalSeconds <= 0 {
			return fmt.Errorf("interval must be greater than 0")
		}
		if monitorAll && monitorClaimed {
			return fmt.Errorf("--all and --claimed cannot be used together")
		}

		svc, sqlDB, err := openService()
		if err != nil {
			return err
		}
		defer func() { _ = sqlDB.Close() }()

		filter := model.ListFilter{All: monitorAll}
		if monitorClaimed {
			filter.Status = "in_progress"
		}
		interval := time.Duration(monitorIntervalSeconds) * time.Second

		for {
			_, _ = fmt.Fprint(cmd.OutOrStdout(), "\033[H\033[2J")

			issues, err := svc.List(filter)
			if err != nil {
				return err
			}
			if monitorClaimed {
				issues, err = claimedIssuesWithParents(svc, issues)
				if err != nil {
					return err
				}
			}
			printIssueList(cmd.OutOrStdout(), issues)

			time.Sleep(interval)
		}
	},
}

// init wires command flags and registration.
func init() {
	monitorCmd.Flags().IntVarP(&monitorIntervalSeconds, "interval", "t", 5, "Refresh interval in seconds")
	monitorCmd.Flags().BoolVar(&monitorAll, "all", false, "Include closed issues")
	monitorCmd.Flags().BoolVar(&monitorClaimed, "claimed", false, "Show only claimed issues (in_progress)")
	rootCmd.AddCommand(monitorCmd)
}

// claimedIssuesWithParents prepends parent epics for claimed tasks to keep context visible.
func claimedIssuesWithParents(svc interface {
	Get(string) (model.Issue, error)
}, issues []model.Issue) ([]model.Issue, error) {
	if len(issues) == 0 {
		return issues, nil
	}

	parentOrder := make([]string, 0)
	parentSeen := make(map[string]struct{})
	for _, issue := range issues {
		if issue.ParentID == nil {
			continue
		}
		parentID := *issue.ParentID
		if _, ok := parentSeen[parentID]; ok {
			continue
		}
		parentSeen[parentID] = struct{}{}
		parentOrder = append(parentOrder, parentID)
	}

	if len(parentOrder) == 0 {
		return issues, nil
	}

	parents := make([]model.Issue, 0, len(parentOrder))
	for _, parentID := range parentOrder {
		parent, err := svc.Get(parentID)
		if err != nil {
			return nil, err
		}
		parents = append(parents, parent)
	}

	return append(parents, issues...), nil
}
