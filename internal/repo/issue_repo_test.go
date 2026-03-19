package repo

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/rpcarvs/faz/internal/db"
	"github.com/rpcarvs/faz/internal/model"
)

func TestReadyIssuesRespectsOpenBlockers(t *testing.T) {
	projectDir := t.TempDir()
	dbPath, err := db.EnsureProjectFiles(projectDir)
	if err != nil {
		t.Fatalf("ensure project files: %v", err)
	}

	sqlDB, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()

	if err := db.Migrate(sqlDB); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	repo := NewIssueRepo(sqlDB)

	blockerID, err := repo.CreateIssue(model.Issue{ID: "faz-ab12", Title: "Blocker", Type: "task", Priority: 1, Status: "open"})
	if err != nil {
		t.Fatalf("create blocker: %v", err)
	}
	blockedID, err := repo.CreateIssue(model.Issue{ID: "faz-ab12.0", Title: "Blocked", Type: "task", Priority: 1, Status: "open"})
	if err != nil {
		t.Fatalf("create blocked: %v", err)
	}
	readyID, err := repo.CreateIssue(model.Issue{ID: "faz-cd34", Title: "Ready", Type: "task", Priority: 2, Status: "open"})
	if err != nil {
		t.Fatalf("create ready: %v", err)
	}

	if err := repo.AddDependency(blockedID, blockerID); err != nil {
		t.Fatalf("add dependency: %v", err)
	}

	ready, err := repo.ReadyIssues()
	if err != nil {
		t.Fatalf("query ready: %v", err)
	}
	if len(ready) != 2 {
		t.Fatalf("expected 2 ready issues, got %d", len(ready))
	}

	if err := repo.CloseIssue(blockerID); err != nil {
		t.Fatalf("close blocker: %v", err)
	}

	ready, err = repo.ReadyIssues()
	if err != nil {
		t.Fatalf("query ready after close: %v", err)
	}

	foundBlocked := false
	foundReady := false
	for _, issue := range ready {
		if issue.ID == blockedID {
			foundBlocked = true
		}
		if issue.ID == readyID {
			foundReady = true
		}
	}
	if !foundBlocked || !foundReady {
		t.Fatalf("expected both blocked and ready issues in ready list after blocker close")
	}

}

func TestClaimIssueSetsInProgressAndPreventsSecondClaim(t *testing.T) {
	projectDir := t.TempDir()
	dbPath, err := db.EnsureProjectFiles(projectDir)
	if err != nil {
		t.Fatalf("ensure project files: %v", err)
	}

	sqlDB, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()

	if err := db.Migrate(sqlDB); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	repo := NewIssueRepo(sqlDB)
	issueID, err := repo.CreateIssue(model.Issue{
		ID:       "faz-ab12",
		Title:    "Claimable task",
		Type:     "task",
		Priority: 1,
		Status:   "open",
	})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}

	if err := repo.ClaimIssue(issueID, time.Hour); err != nil {
		t.Fatalf("first claim should succeed: %v", err)
	}

	claimed, err := repo.GetIssue(issueID)
	if err != nil {
		t.Fatalf("get claimed issue: %v", err)
	}
	if claimed.Status != "in_progress" {
		t.Fatalf("expected in_progress after claim, got %s", claimed.Status)
	}
	if claimed.ClaimedAt == nil || claimed.ClaimExpiresAt == nil {
		t.Fatalf("expected claim timestamps to be set")
	}

	if err := repo.ClaimIssue(issueID, time.Hour); err == nil {
		t.Fatalf("second claim should fail")
	}
}

func TestReadyIssuesIncludesExpiredInProgressClaims(t *testing.T) {
	projectDir := t.TempDir()
	dbPath, err := db.EnsureProjectFiles(projectDir)
	if err != nil {
		t.Fatalf("ensure project files: %v", err)
	}

	sqlDB, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()

	if err := db.Migrate(sqlDB); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	repo := NewIssueRepo(sqlDB)
	issueID, err := repo.CreateIssue(model.Issue{
		ID:       "faz-r111",
		Title:    "Reclaimable task",
		Type:     "task",
		Priority: 1,
		Status:   "open",
	})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}

	if err := repo.ClaimIssue(issueID, 50*time.Millisecond); err != nil {
		t.Fatalf("claim issue: %v", err)
	}

	time.Sleep(80 * time.Millisecond)

	ready, err := repo.ReadyIssues()
	if err != nil {
		t.Fatalf("query ready: %v", err)
	}

	found := false
	for _, issue := range ready {
		if issue.ID == issueID {
			found = true
			if issue.Status != "in_progress" {
				t.Fatalf("expected in_progress status to remain until reclaim, got %s", issue.Status)
			}
			break
		}
	}
	if !found {
		t.Fatalf("expected expired in_progress issue %q to be in ready list", issueID)
	}
}

func TestCloseIssueClearsClaimFields(t *testing.T) {
	projectDir := t.TempDir()
	dbPath, err := db.EnsureProjectFiles(projectDir)
	if err != nil {
		t.Fatalf("ensure project files: %v", err)
	}

	sqlDB, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()

	if err := db.Migrate(sqlDB); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	repo := NewIssueRepo(sqlDB)
	issueID, err := repo.CreateIssue(model.Issue{
		ID:       "faz-cd34",
		Title:    "Closable claimed task",
		Type:     "task",
		Priority: 1,
		Status:   "open",
	})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}

	if err := repo.ClaimIssue(issueID, time.Hour); err != nil {
		t.Fatalf("claim issue: %v", err)
	}
	if err := repo.CloseIssue(issueID); err != nil {
		t.Fatalf("close issue: %v", err)
	}

	closed, err := repo.GetIssue(issueID)
	if err != nil {
		t.Fatalf("get closed issue: %v", err)
	}
	if closed.Status != "closed" {
		t.Fatalf("expected closed status, got %s", closed.Status)
	}
	if closed.ClaimedAt != nil || closed.ClaimExpiresAt != nil {
		t.Fatalf("expected claim fields to be cleared on close")
	}
}

func TestClaimIssueRejectsEpic(t *testing.T) {
	projectDir := t.TempDir()
	dbPath, err := db.EnsureProjectFiles(projectDir)
	if err != nil {
		t.Fatalf("ensure project files: %v", err)
	}

	sqlDB, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()

	if err := db.Migrate(sqlDB); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	repo := NewIssueRepo(sqlDB)
	epicID, err := repo.CreateIssue(model.Issue{
		ID:       "faz-z111",
		Title:    "Epic container",
		Type:     "epic",
		Priority: 1,
		Status:   "open",
	})
	if err != nil {
		t.Fatalf("create epic: %v", err)
	}

	err = repo.ClaimIssue(epicID, time.Hour)
	if !errors.Is(err, ErrIssueTypeNotClaimable) {
		t.Fatalf("expected ErrIssueTypeNotClaimable, got %v", err)
	}
}

func TestNextChildIndexReusesDeletedGap(t *testing.T) {
	projectDir := t.TempDir()
	dbPath, err := db.EnsureProjectFiles(projectDir)
	if err != nil {
		t.Fatalf("ensure project files: %v", err)
	}

	sqlDB, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()

	if err := db.Migrate(sqlDB); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	repo := NewIssueRepo(sqlDB)
	parentID, err := repo.CreateIssue(model.Issue{
		ID:       "faz-p111",
		Title:    "Parent",
		Type:     "epic",
		Priority: 1,
		Status:   "open",
	})
	if err != nil {
		t.Fatalf("create parent: %v", err)
	}
	if _, err := repo.CreateIssue(model.Issue{
		ID:       "faz-p111.0",
		Title:    "Child 0",
		Type:     "task",
		Priority: 1,
		Status:   "open",
		ParentID: &parentID,
	}); err != nil {
		t.Fatalf("create child 0: %v", err)
	}
	if _, err := repo.CreateIssue(model.Issue{
		ID:       "faz-p111.1",
		Title:    "Child 1",
		Type:     "task",
		Priority: 1,
		Status:   "open",
		ParentID: &parentID,
	}); err != nil {
		t.Fatalf("create child 1: %v", err)
	}

	if err := repo.DeleteIssue("faz-p111.0"); err != nil {
		t.Fatalf("delete child 0: %v", err)
	}

	next, err := repo.NextChildIndex(parentID)
	if err != nil {
		t.Fatalf("next child index: %v", err)
	}
	if next != 0 {
		t.Fatalf("expected reused index 0, got %d", next)
	}
}

func TestListIssuesOrdersByPublicIDNaturalChildOrder(t *testing.T) {
	projectDir := t.TempDir()
	dbPath, err := db.EnsureProjectFiles(projectDir)
	if err != nil {
		t.Fatalf("ensure project files: %v", err)
	}

	sqlDB, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()

	if err := db.Migrate(sqlDB); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	repo := NewIssueRepo(sqlDB)

	parentA, err := repo.CreateIssue(model.Issue{
		ID:       "faz-a111",
		Title:    "Parent A",
		Type:     "epic",
		Priority: 3,
		Status:   "open",
	})
	if err != nil {
		t.Fatalf("create parent A: %v", err)
	}
	parentB, err := repo.CreateIssue(model.Issue{
		ID:       "faz-b111",
		Title:    "Parent B",
		Type:     "epic",
		Priority: 1,
		Status:   "open",
	})
	if err != nil {
		t.Fatalf("create parent B: %v", err)
	}

	if _, err := repo.CreateIssue(model.Issue{
		ID:       "faz-a111.0",
		Title:    "A child low priority",
		Type:     "task",
		Priority: 3,
		Status:   "open",
		ParentID: &parentA,
	}); err != nil {
		t.Fatalf("create a child 0: %v", err)
	}
	if _, err := repo.CreateIssue(model.Issue{
		ID:       "faz-a111.1",
		Title:    "A child high priority",
		Type:     "task",
		Priority: 1,
		Status:   "open",
		ParentID: &parentA,
	}); err != nil {
		t.Fatalf("create a child 1: %v", err)
	}
	if _, err := repo.CreateIssue(model.Issue{
		ID:       "faz-a111.10",
		Title:    "A child ten",
		Type:     "task",
		Priority: 2,
		Status:   "open",
		ParentID: &parentA,
	}); err != nil {
		t.Fatalf("create a child 10: %v", err)
	}
	if _, err := repo.CreateIssue(model.Issue{
		ID:       "faz-a111.2",
		Title:    "A child two",
		Type:     "task",
		Priority: 2,
		Status:   "open",
		ParentID: &parentA,
	}); err != nil {
		t.Fatalf("create a child 2: %v", err)
	}
	if _, err := repo.CreateIssue(model.Issue{
		ID:       "faz-b111.0",
		Title:    "B child",
		Type:     "task",
		Priority: 2,
		Status:   "open",
		ParentID: &parentB,
	}); err != nil {
		t.Fatalf("create b child: %v", err)
	}

	issues, err := repo.ListIssues(model.ListFilter{All: true})
	if err != nil {
		t.Fatalf("list issues: %v", err)
	}

	gotIDs := make([]string, 0, len(issues))
	for _, issue := range issues {
		gotIDs = append(gotIDs, issue.ID)
	}
	wantIDs := []string{
		"faz-a111",
		"faz-a111.0",
		"faz-a111.1",
		"faz-a111.2",
		"faz-a111.10",
		"faz-b111",
		"faz-b111.0",
	}
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Fatalf("unexpected list ordering, got=%v want=%v", gotIDs, wantIDs)
	}
}
