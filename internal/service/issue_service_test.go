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
