package cmd

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/rpcarvs/faz/internal/db"
	"github.com/rpcarvs/faz/internal/model"
	"github.com/spf13/cobra"
)

var (
	monitorAll     bool
	monitorClaimed bool
)

const (
	monitorEventDebounceInterval = 100 * time.Millisecond
	monitorRefreshInterval       = 2 * time.Second
)

// monitorService defines the issue reads required by the monitor command.
type monitorService interface {
	List(model.ListFilter) ([]model.Issue, error)
	Get(string) (model.Issue, error)
}

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
		defer func() {
			if closeErr := watcher.Close(); closeErr != nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "close watcher: %v\n", closeErr)
			}
		}()
		if err := watcher.Add(filepath.Join(projectDir, db.DirName)); err != nil {
			return err
		}

		filter := model.ListFilter{All: monitorAll}
		if monitorClaimed {
			filter.Status = "in_progress"
		}

		refreshTicker := time.NewTicker(monitorRefreshInterval)
		defer refreshTicker.Stop()

		return runMonitor(
			svc,
			filter,
			monitorClaimed,
			cmd.OutOrStdout(),
			watcher.Events,
			watcher.Errors,
			refreshTicker.C,
			monitorEventDebounceInterval,
		)
	},
}

// init wires command flags and registration.
func init() {
	monitorCmd.Flags().BoolVar(&monitorAll, "all", false, "Include closed issues")
	monitorCmd.Flags().BoolVar(&monitorClaimed, "claimed", false, "Show only claimed issues (in_progress)")
	rootCmd.AddCommand(monitorCmd)
}

// runMonitor renders issue state and refreshes it after settled filesystem events or fallback ticks.
func runMonitor(
	svc monitorService,
	filter model.ListFilter,
	includeClaimedParents bool,
	writer io.Writer,
	events <-chan fsnotify.Event,
	watcherErrors <-chan error,
	refreshTicks <-chan time.Time,
	debounceInterval time.Duration,
) error {
	var lastFrame string
	var rendered bool
	render := func() error {
		frame, err := buildMonitorFrame(svc, filter, includeClaimedParents)
		if err != nil {
			return err
		}
		if rendered && frame == lastFrame {
			return nil
		}
		if _, err := fmt.Fprint(writer, "\033[H\033[2J", frame); err != nil {
			return fmt.Errorf("render monitor: %w", err)
		}
		lastFrame = frame
		rendered = true
		return nil
	}

	if err := render(); err != nil {
		return err
	}

	var debounceTimer *time.Timer
	var debounce <-chan time.Time
	stopDebounce := func() {
		if debounceTimer == nil {
			return
		}
		if !debounceTimer.Stop() {
			select {
			case <-debounceTimer.C:
			default:
			}
		}
		debounce = nil
	}
	defer stopDebounce()

	for {
		select {
		case _, ok := <-events:
			if !ok {
				return nil
			}
			stopDebounce()
			if debounceTimer == nil {
				debounceTimer = time.NewTimer(debounceInterval)
			} else {
				debounceTimer.Reset(debounceInterval)
			}
			debounce = debounceTimer.C
		case <-debounce:
			debounce = nil
			if err := render(); err != nil {
				return err
			}
		case _, ok := <-refreshTicks:
			if !ok {
				refreshTicks = nil
				continue
			}
			if err := render(); err != nil {
				return err
			}
		case err, ok := <-watcherErrors:
			if !ok {
				return nil
			}
			return err
		}
	}
}

// buildMonitorFrame returns the current filtered issue list as terminal output.
func buildMonitorFrame(svc monitorService, filter model.ListFilter, includeClaimedParents bool) (string, error) {
	issues, err := svc.List(filter)
	if err != nil {
		return "", err
	}
	if includeClaimedParents {
		issues, err = claimedIssuesWithParents(svc, issues)
		if err != nil {
			return "", err
		}
	}
	var frame bytes.Buffer
	printIssueList(&frame, issues)
	return frame.String(), nil
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
