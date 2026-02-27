package repo

import (
	"testing"

	"faz/internal/db"
	"faz/internal/model"
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
