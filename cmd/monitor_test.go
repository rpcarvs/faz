package cmd

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/rpcarvs/faz/internal/model"
)

const monitorTestTimeout = time.Second

type monitorStubService struct {
	mu        sync.Mutex
	responses [][]model.Issue
	filters   []model.ListFilter
	listCalls chan struct{}
}

// List returns the next configured issue snapshot and records the requested filter.
func (s *monitorStubService) List(filter model.ListFilter) ([]model.Issue, error) {
	s.mu.Lock()
	callIndex := len(s.filters)
	s.filters = append(s.filters, filter)
	responseIndex := min(callIndex, len(s.responses)-1)
	issues := append([]model.Issue(nil), s.responses[responseIndex]...)
	s.mu.Unlock()

	s.listCalls <- struct{}{}
	return issues, nil
}

// listCallCount returns the number of monitor list reads observed by the stub.
func (s *monitorStubService) listCallCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.filters)
}

// requestedFilters returns a copy of the filters observed by the stub.
func (s *monitorStubService) requestedFilters() []model.ListFilter {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]model.ListFilter(nil), s.filters...)
}

func TestRunMonitorCoalescesFilesystemEventBursts(t *testing.T) {
	epic := model.Issue{ID: "proj-epic", Title: "Epic", Type: "epic", Priority: 1, Status: "open"}
	bug := model.Issue{ID: "proj-epic.0", Title: "Bug", Type: "bug", Priority: 1, Status: "open"}
	svc := newMonitorStubService([][]model.Issue{{epic}, {epic, bug}})
	events := make(chan fsnotify.Event, 3)
	errorsChannel := make(chan error)
	refreshTicks := make(chan time.Time)
	var output bytes.Buffer
	done := runMonitorForTest(
		svc,
		model.ListFilter{},
		&output,
		events,
		errorsChannel,
		refreshTicks,
		25*time.Millisecond,
	)

	waitForMonitorListCall(t, svc)
	events <- fsnotify.Event{}
	events <- fsnotify.Event{}
	events <- fsnotify.Event{}
	waitForMonitorListCall(t, svc)
	time.Sleep(75 * time.Millisecond)

	if got := svc.listCallCount(); got != 2 {
		t.Fatalf("list calls after one event burst = %d, want 2", got)
	}
	close(events)
	waitForMonitorResult(t, done, nil)

	frame := finalMonitorFrame(output.String())
	if !strings.Contains(frame, epic.ID) || !strings.Contains(frame, bug.ID) {
		t.Fatalf("final monitor frame did not contain committed issues: %q", frame)
	}
}

func TestRunMonitorFallbackRecoversFromStaleEventRefresh(t *testing.T) {
	epic := model.Issue{ID: "proj-epic", Title: "Epic", Type: "epic", Priority: 1, Status: "open"}
	bug := model.Issue{ID: "proj-epic.0", Title: "Bug", Type: "bug", Priority: 1, Status: "open"}
	svc := newMonitorStubService([][]model.Issue{{epic}, {epic}, {epic, bug}})
	events := make(chan fsnotify.Event)
	errorsChannel := make(chan error)
	refreshTicks := make(chan time.Time)
	var output bytes.Buffer
	done := runMonitorForTest(
		svc,
		model.ListFilter{},
		&output,
		events,
		errorsChannel,
		refreshTicks,
		time.Millisecond,
	)

	waitForMonitorListCall(t, svc)
	events <- fsnotify.Event{}
	waitForMonitorListCall(t, svc)

	refreshTicks <- time.Now()
	waitForMonitorListCall(t, svc)
	close(events)
	waitForMonitorResult(t, done, nil)

	if frame := finalMonitorFrame(output.String()); !strings.Contains(frame, bug.ID) {
		t.Fatalf("fallback refresh did not contain newly created bug: %q", frame)
	}
}

func TestRunMonitorFallbackDoesNotCancelPendingDebounce(t *testing.T) {
	epic := model.Issue{ID: "proj-epic", Title: "Epic", Type: "epic", Priority: 1, Status: "open"}
	bug := model.Issue{ID: "proj-epic.0", Title: "Bug", Type: "bug", Priority: 1, Status: "open"}
	svc := newMonitorStubService([][]model.Issue{{epic}, {epic}, {epic, bug}})
	events := make(chan fsnotify.Event)
	refreshTicks := make(chan time.Time)
	var output bytes.Buffer
	done := runMonitorForTest(
		svc,
		model.ListFilter{},
		&output,
		events,
		make(chan error),
		refreshTicks,
		50*time.Millisecond,
	)

	waitForMonitorListCall(t, svc)
	events <- fsnotify.Event{}
	refreshTicks <- time.Now()
	waitForMonitorListCall(t, svc)
	waitForMonitorListCall(t, svc)
	close(events)
	waitForMonitorResult(t, done, nil)

	if frame := finalMonitorFrame(output.String()); !strings.Contains(frame, bug.ID) {
		t.Fatalf("pending debounce did not refresh after fallback tick: %q", frame)
	}
}

func TestRunMonitorFallbackRefreshesEveryIssueType(t *testing.T) {
	issueTypes := []string{"epic", "task", "bug", "feature", "chore", "decision"}
	openIssues := make([]model.Issue, 0, len(issueTypes))
	closedIssues := make([]model.Issue, 0, len(issueTypes))
	for index, issueType := range issueTypes {
		issueID := "proj-" + issueType
		openIssues = append(openIssues, model.Issue{
			ID:       issueID,
			Title:    "Issue",
			Type:     issueType,
			Priority: index % 4,
			Status:   "open",
		})
		closedIssues = append(closedIssues, model.Issue{
			ID:       issueID,
			Title:    "Issue",
			Type:     issueType,
			Priority: index % 4,
			Status:   "closed",
		})
	}

	svc := newMonitorStubService([][]model.Issue{openIssues, closedIssues})
	events := make(chan fsnotify.Event)
	refreshTicks := make(chan time.Time)
	var output bytes.Buffer
	done := runMonitorForTest(
		svc,
		model.ListFilter{All: true},
		&output,
		events,
		make(chan error),
		refreshTicks,
		time.Millisecond,
	)

	waitForMonitorListCall(t, svc)
	refreshTicks <- time.Now()
	waitForMonitorListCall(t, svc)
	close(events)
	waitForMonitorResult(t, done, nil)

	frame := finalMonitorFrame(output.String())
	for _, issue := range closedIssues {
		if !strings.Contains(frame, "✓ "+issue.ID) {
			t.Errorf("final monitor frame did not show %s issue as closed: %q", issue.Type, frame)
		}
	}
}

func TestRunMonitorDoesNotRedrawUnchangedFallbackFrame(t *testing.T) {
	issue := model.Issue{ID: "proj-task", Title: "Task", Type: "task", Priority: 2, Status: "open"}
	svc := newMonitorStubService([][]model.Issue{{issue}})
	events := make(chan fsnotify.Event)
	refreshTicks := make(chan time.Time)
	var output bytes.Buffer
	done := runMonitorForTest(
		svc,
		model.ListFilter{},
		&output,
		events,
		make(chan error),
		refreshTicks,
		time.Millisecond,
	)

	waitForMonitorListCall(t, svc)
	refreshTicks <- time.Now()
	waitForMonitorListCall(t, svc)
	close(events)
	waitForMonitorResult(t, done, nil)

	if redraws := strings.Count(output.String(), "\033[H\033[2J"); redraws != 1 {
		t.Fatalf("unchanged fallback redraw count = %d, want 1", redraws)
	}
}

func TestRunMonitorClaimedViewShowsOnlyClaimedRows(t *testing.T) {
	parentID := "proj-epic"
	child := model.Issue{
		ID:       parentID + ".0",
		Title:    "Claimed bug",
		Type:     "bug",
		Priority: 1,
		Status:   "in_progress",
		ParentID: &parentID,
	}
	standalone := model.Issue{ID: "proj-task", Title: "Claimed task", Type: "task", Status: "in_progress"}
	svc := newMonitorStubService([][]model.Issue{{child, standalone}, {}})
	events := make(chan fsnotify.Event)
	refreshTicks := make(chan time.Time)
	var output bytes.Buffer
	done := runMonitorForTest(
		svc,
		model.ListFilter{Status: "in_progress"},
		&output,
		events,
		make(chan error),
		refreshTicks,
		time.Millisecond,
	)

	waitForMonitorListCall(t, svc)
	refreshTicks <- time.Now()
	waitForMonitorListCall(t, svc)
	close(events)
	waitForMonitorResult(t, done, nil)

	frames := strings.Split(output.String(), "\033[H\033[2J")
	if len(frames) != 3 {
		t.Fatalf("monitor frames = %d, want initial and post-closure frames", len(frames)-1)
	}
	rows := strings.Split(strings.TrimSpace(frames[1]), "\n")
	if len(rows) != 2 || !strings.Contains(rows[0], child.ID) || !strings.Contains(rows[1], standalone.ID) {
		t.Fatalf("claimed frame should contain only the two claimed rows: %q", frames[1])
	}
	if frame := finalMonitorFrame(output.String()); frame != "No issues found\n" {
		t.Fatalf("post-closure claimed frame = %q, want no issues", frame)
	}
	for _, filter := range svc.requestedFilters() {
		if filter.Status != "in_progress" || filter.All {
			t.Fatalf("claimed monitor filter = %#v, want only in_progress status", filter)
		}
	}
}

func TestRunMonitorReturnsWatcherError(t *testing.T) {
	svc := newMonitorStubService([][]model.Issue{{}})
	events := make(chan fsnotify.Event)
	errorsChannel := make(chan error)
	watcherErr := errors.New("watch failed")
	done := runMonitorForTest(
		svc,
		model.ListFilter{},
		&bytes.Buffer{},
		events,
		errorsChannel,
		make(chan time.Time),
		time.Millisecond,
	)

	waitForMonitorListCall(t, svc)
	errorsChannel <- watcherErr
	waitForMonitorResult(t, done, watcherErr)
}

// newMonitorStubService builds a monitor service with deterministic list snapshots.
func newMonitorStubService(responses [][]model.Issue) *monitorStubService {
	return &monitorStubService{
		responses: responses,
		listCalls: make(chan struct{}, 16),
	}
}

// runMonitorForTest starts the monitor loop and returns its eventual result.
func runMonitorForTest(
	svc monitorService,
	filter model.ListFilter,
	output *bytes.Buffer,
	events <-chan fsnotify.Event,
	watcherErrors <-chan error,
	refreshTicks <-chan time.Time,
	debounceInterval time.Duration,
) <-chan error {
	done := make(chan error, 1)
	go func() {
		done <- runMonitor(
			context.Background(),
			svc,
			filter,
			output,
			nil,
			events,
			watcherErrors,
			refreshTicks,
			debounceInterval,
		)
	}()
	return done
}

// waitForMonitorListCall waits for one list read or fails on timeout.
func waitForMonitorListCall(t *testing.T, svc *monitorStubService) {
	t.Helper()
	select {
	case <-svc.listCalls:
	case <-time.After(monitorTestTimeout):
		t.Fatal("timed out waiting for monitor list call")
	}
}

// waitForMonitorResult waits for the monitor to stop and verifies its result.
func waitForMonitorResult(t *testing.T, done <-chan error, want error) {
	t.Helper()
	select {
	case err := <-done:
		if !errors.Is(err, want) {
			t.Fatalf("monitor result = %v, want %v", err, want)
		}
	case <-time.After(monitorTestTimeout):
		t.Fatal("timed out waiting for monitor to stop")
	}
}

// finalMonitorFrame returns the most recently rendered terminal frame.
func finalMonitorFrame(output string) string {
	frames := strings.Split(output, "\033[H\033[2J")
	return frames[len(frames)-1]
}
