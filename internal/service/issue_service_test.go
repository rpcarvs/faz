package service

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/rpcarvs/faz/internal/db"
	"github.com/rpcarvs/faz/internal/model"
	"github.com/rpcarvs/faz/internal/repo"
)

func testStringPtr(value string) *string {
	return &value
}

// TestCreateInheritsAndPreservesSDDAssociations verifies creation-time inheritance without later propagation.
func TestCreateInheritsAndPreservesSDDAssociations(t *testing.T) {
	projectDir := t.TempDir()
	dbPath, err := db.EnsureProjectFiles(projectDir)
	if err != nil {
		t.Fatalf("ensure project files: %v", err)
	}
	sqlDB, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()
	if err := db.Migrate(sqlDB); err != nil {
		t.Fatalf("migrate database: %v", err)
	}

	svc := NewIssueService(repo.NewIssueRepo(sqlDB), "faz")
	parentID, err := svc.Create(model.Issue{
		Title: "Associated parent", Type: "epic", Priority: 1,
		PlanID: testStringPtr("PLAN01"), WorkID: testStringPtr("W01"),
	})
	if err != nil {
		t.Fatalf("create parent: %v", err)
	}
	inheritedID, err := svc.Create(model.Issue{
		Title: "Inherited child", Type: "task", Priority: 1, ParentID: &parentID,
	})
	if err != nil {
		t.Fatalf("create inherited child: %v", err)
	}
	overriddenID, err := svc.Create(model.Issue{
		Title: "Overridden child", Type: "task", Priority: 1, ParentID: &parentID,
		WorkID: testStringPtr("W1"),
	})
	if err != nil {
		t.Fatalf("create overridden child: %v", err)
	}
	planOverrideID, err := svc.Create(model.Issue{
		Title: "Plan override child", Type: "task", Priority: 1, ParentID: &parentID,
		PlanID: testStringPtr("PLAN02"),
	})
	if err != nil {
		t.Fatalf("create plan override child: %v", err)
	}
	bothOverrideID, err := svc.Create(model.Issue{
		Title: "Both overrides child", Type: "task", Priority: 1, ParentID: &parentID,
		PlanID: testStringPtr("PLAN04"), WorkID: testStringPtr("W04"),
	})
	if err != nil {
		t.Fatalf("create both overrides child: %v", err)
	}
	unlinkedParentID, err := svc.Create(model.Issue{Title: "Unlinked parent", Type: "epic", Priority: 1})
	if err != nil {
		t.Fatalf("create unlinked parent: %v", err)
	}
	unlinkedChildID, err := svc.Create(model.Issue{Title: "Unlinked child", Type: "task", Priority: 1, ParentID: &unlinkedParentID})
	if err != nil {
		t.Fatalf("create unlinked child: %v", err)
	}

	inherited, err := svc.Get(inheritedID)
	if err != nil || inherited.PlanID == nil || *inherited.PlanID != "PLAN01" || inherited.WorkID == nil || *inherited.WorkID != "W01" {
		t.Fatalf("inherited associations = %#v, err=%v", inherited, err)
	}
	overridden, err := svc.Get(overriddenID)
	if err != nil || overridden.PlanID == nil || *overridden.PlanID != "PLAN01" || overridden.WorkID == nil || *overridden.WorkID != "W1" {
		t.Fatalf("overridden associations = %#v, err=%v", overridden, err)
	}
	planOverride, err := svc.Get(planOverrideID)
	if err != nil || planOverride.PlanID == nil || *planOverride.PlanID != "PLAN02" || planOverride.WorkID == nil || *planOverride.WorkID != "W01" {
		t.Fatalf("plan override associations = %#v, err=%v", planOverride, err)
	}
	bothOverride, err := svc.Get(bothOverrideID)
	if err != nil || bothOverride.PlanID == nil || *bothOverride.PlanID != "PLAN04" || bothOverride.WorkID == nil || *bothOverride.WorkID != "W04" {
		t.Fatalf("both override associations = %#v, err=%v", bothOverride, err)
	}
	unlinkedChild, err := svc.Get(unlinkedChildID)
	if err != nil || unlinkedChild.PlanID != nil || unlinkedChild.WorkID != nil {
		t.Fatalf("unlinked child associations = %#v, err=%v", unlinkedChild, err)
	}

	if err := svc.Update(parentID, map[string]any{"plan_id": "PLAN02", "work_id": testStringPtr("W02")}); err != nil {
		t.Fatalf("update parent associations: %v", err)
	}
	newParentID, err := svc.Create(model.Issue{
		Title: "Replacement parent", Type: "epic", Priority: 1,
		PlanID: testStringPtr("PLAN03"), WorkID: testStringPtr("W03"),
	})
	if err != nil {
		t.Fatalf("create replacement parent: %v", err)
	}
	if err := svc.Update(inheritedID, map[string]any{"parent_public_id": &newParentID}); err != nil {
		t.Fatalf("reparent child: %v", err)
	}
	afterParentChange, err := svc.Get(inheritedID)
	if err != nil || afterParentChange.PlanID == nil || *afterParentChange.PlanID != "PLAN01" || afterParentChange.WorkID == nil || *afterParentChange.WorkID != "W01" {
		t.Fatalf("child associations changed after parent update or reparenting: %#v, err=%v", afterParentChange, err)
	}

	if err := svc.Claim(inheritedID, time.Minute); err != nil {
		t.Fatalf("claim child: %v", err)
	}
	if err := svc.Close(inheritedID); err != nil {
		t.Fatalf("close child: %v", err)
	}
	if err := svc.Reopen(inheritedID); err != nil {
		t.Fatalf("reopen child: %v", err)
	}
	lifecycleIssue, err := svc.Get(inheritedID)
	if err != nil || lifecycleIssue.PlanID == nil || *lifecycleIssue.PlanID != "PLAN01" || lifecycleIssue.WorkID == nil || *lifecycleIssue.WorkID != "W01" {
		t.Fatalf("lifecycle operations lost associations: %#v, err=%v", lifecycleIssue, err)
	}

	if err := svc.Update(overriddenID, map[string]any{"plan_id": nil, "work_id": (*string)(nil)}); err != nil {
		t.Fatalf("clear child associations: %v", err)
	}
	cleared, err := svc.Get(overriddenID)
	if err != nil || cleared.PlanID != nil || cleared.WorkID != nil {
		t.Fatalf("cleared associations = %#v, err=%v", cleared, err)
	}
}

// TestAssociationValidationRejectsUnsafeValues verifies association updates accept only explicit safe strings or nil.
func TestAssociationValidationRejectsUnsafeValues(t *testing.T) {
	projectDir := t.TempDir()
	dbPath, err := db.EnsureProjectFiles(projectDir)
	if err != nil {
		t.Fatalf("ensure project files: %v", err)
	}
	sqlDB, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()
	if err := db.Migrate(sqlDB); err != nil {
		t.Fatalf("migrate database: %v", err)
	}

	svc := NewIssueService(repo.NewIssueRepo(sqlDB), "faz")
	issueID, err := svc.Create(model.Issue{Title: "Task", Type: "task", Priority: 1})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}
	for _, fields := range []map[string]any{
		{"plan_id": ""},
		{"work_id": " "},
		{"plan_id": "../PLAN01"},
		{"work_id": "W01\n"},
		{"work_id": 1},
		{"plan_id": testStringPtr("PLAN01\t")},
	} {
		if err := svc.Update(issueID, fields); err == nil {
			t.Fatalf("update %v succeeded unexpectedly", fields)
		}
	}
	if err := svc.Update(issueID, map[string]any{"plan_id": testStringPtr("PLAN01"), "work_id": "W1"}); err != nil {
		t.Fatalf("update exact identifiers: %v", err)
	}
	issue, err := svc.Get(issueID)
	if err != nil || issue.PlanID == nil || *issue.PlanID != "PLAN01" || issue.WorkID == nil || *issue.WorkID != "W1" {
		t.Fatalf("exact identifiers = %#v, err=%v", issue, err)
	}
}

func TestNormalizeIssueID(t *testing.T) {
	if _, err := NormalizeIssueID("abc"); err == nil {
		t.Fatalf("expected parse error for invalid issue ID")
	}

	id, err := NormalizeIssueID("FAZ-Ab12")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if id != "faz-ab12" {
		t.Fatalf("expected faz-ab12, got %s", id)
	}

	child, err := NormalizeIssueID("faz-ab12.3")
	if err != nil {
		t.Fatalf("unexpected child parse error: %v", err)
	}
	if child != "faz-ab12.3" {
		t.Fatalf("expected faz-ab12.3, got %s", child)
	}
}

func TestIsRetryableCreateError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "sqlite busy",
			err:  fmt.Errorf("insert issue: database is locked (5) (SQLITE_BUSY)"),
			want: true,
		},
		{
			name: "unique public id collision",
			err:  fmt.Errorf("insert issue: constraint failed: UNIQUE constraint failed: issues.public_id (2067)"),
			want: true,
		},
		{
			name: "wrapped retryable error",
			err:  fmt.Errorf("wrap: %w", errors.New("database is locked")),
			want: true,
		},
		{
			name: "non retryable error",
			err:  errors.New("issue \"faz-ab12\" not found"),
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isRetryableCreateError(tc.err); got != tc.want {
				t.Fatalf("isRetryableCreateError() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCreateBackoff(t *testing.T) {
	if got := createBackoff(0); got != 20*time.Millisecond {
		t.Fatalf("attempt 0 backoff = %s, want %s", got, 20*time.Millisecond)
	}
	if got := createBackoff(3); got != 160*time.Millisecond {
		t.Fatalf("attempt 3 backoff = %s, want %s", got, 160*time.Millisecond)
	}
	if got := createBackoff(10); got != 320*time.Millisecond {
		t.Fatalf("attempt 10 backoff = %s, want %s", got, 320*time.Millisecond)
	}
}

func TestUpdateRejectsInProgressStatus(t *testing.T) {
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

	issueRepo := repo.NewIssueRepo(sqlDB)
	svc := NewIssueService(issueRepo, "faz")
	issueID, err := svc.Create(model.Issue{
		Title:       "Task",
		Description: "test",
		Type:        "task",
		Priority:    1,
		Status:      "open",
	})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}

	err = svc.Update(issueID, map[string]any{"status": "in_progress"})
	if err == nil {
		t.Fatalf("expected update to reject in_progress status")
	}
}
