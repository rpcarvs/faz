package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/rpcarvs/faz/internal/db"
	"github.com/rpcarvs/faz/internal/model"
	"github.com/spf13/cobra"
)

var (
	monitorAll     bool
	monitorClaimed bool
)

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Continuously refresh task list output",
	Long:  "Monitor watches the task database and refreshes the list whenever it changes.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if monitorAll && monitorClaimed {
			return fmt.Errorf("--all and --claimed cannot be used together")
		}

		svc, sqlDB, err := openService()
		if err != nil {
			return err
		}
		defer func() { _ = sqlDB.Close() }()

		projectDir, err := currentProjectDir()
		if err != nil {
			return err
		}
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return err
		}
		defer watcher.Close()
		if err := watcher.Add(filepath.Join(projectDir, db.DirName)); err != nil {
			return err
		}

		filter := model.ListFilter{All: monitorAll}
		if monitorClaimed {
			filter.Status = "in_progress"
		}

		render := func() error {
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
			return nil
		}

		if err := render(); err != nil {
			return err
		}

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return nil
				}
				if event.Has(fsnotify.Write) {
					if err := render(); err != nil {
						return err
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return nil
				}
				return err
			}
		}
	},
}

// init wires command flags and registration.
func init() {
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
