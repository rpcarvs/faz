package db

import (
	"strings"
	"testing"
)

// TestMigrateAddsOptionalSDDAssociationsToLegacyIssues verifies upgrades preserve existing issue state.
func TestMigrateAddsOptionalSDDAssociationsToLegacyIssues(t *testing.T) {
	dbPath := t.TempDir() + "/legacy.db"
	sqlDB, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()

	if _, err := sqlDB.Exec(`CREATE TABLE issues (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id TEXT,
		title TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		type TEXT NOT NULL,
		priority INTEGER NOT NULL DEFAULT 2,
		status TEXT NOT NULL DEFAULT 'open',
		claimed_at DATETIME,
		claim_expires_at DATETIME,
		parent_id INTEGER,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		closed_at DATETIME
	)`); err != nil {
		t.Fatalf("create legacy issues table: %v", err)
	}
	if _, err := sqlDB.Exec(`CREATE TABLE dependencies (issue_id INTEGER NOT NULL, depends_on_id INTEGER NOT NULL)`); err != nil {
		t.Fatalf("create legacy dependencies table: %v", err)
	}
	if _, err := sqlDB.Exec(`INSERT INTO issues(id, public_id, title, type, status, claimed_at, claim_expires_at, created_at, updated_at) VALUES(1, 'faz-legacy', 'parent', 'epic', 'in_progress', '2026-01-01', '2026-01-02', '2025-01-01', '2025-01-02')`); err != nil {
		t.Fatalf("insert legacy parent: %v", err)
	}
	if _, err := sqlDB.Exec(`INSERT INTO issues(id, title, type, parent_id) VALUES(2, 'child', 'task', 1)`); err != nil {
		t.Fatalf("insert legacy child: %v", err)
	}
	if _, err := sqlDB.Exec(`INSERT INTO dependencies(issue_id, depends_on_id) VALUES(2, 1)`); err != nil {
		t.Fatalf("insert legacy dependency: %v", err)
	}

	if err := Migrate(sqlDB); err != nil {
		t.Fatalf("migrate legacy database: %v", err)
	}
	var publicID string
	var planID, workID *string
	var parentID *int64
	if err := sqlDB.QueryRow(`SELECT public_id, plan_id, work_id, parent_id FROM issues WHERE id = 2`).Scan(&publicID, &planID, &workID, &parentID); err != nil {
		t.Fatalf("read migrated child: %v", err)
	}
	if publicID != "legacy-2" || parentID == nil || *parentID != 1 {
		t.Fatalf("migrated child = public_id %q, parent_id %v", publicID, parentID)
	}
	if planID != nil || workID != nil {
		t.Fatalf("optional associations = (%v, %v), want NULL values", planID, workID)
	}
	if _, err := sqlDB.Exec(`UPDATE issues SET plan_id = 'PLAN01', work_id = 'W01' WHERE id = 2`); err != nil {
		t.Fatalf("set migrated associations: %v", err)
	}
	if err := Migrate(sqlDB); err != nil {
		t.Fatalf("repeat migration: %v", err)
	}
	if err := sqlDB.QueryRow(`SELECT plan_id, work_id FROM issues WHERE id = 2`).Scan(&planID, &workID); err != nil {
		t.Fatalf("read repeated migration associations: %v", err)
	}
	if planID == nil || *planID != "PLAN01" || workID == nil || *workID != "W01" {
		t.Fatalf("repeat migration associations = (%v, %v)", planID, workID)
	}
	var claimedAt, claimExpiresAt *string
	if err := sqlDB.QueryRow(`SELECT claimed_at, claim_expires_at FROM issues WHERE id = 1`).Scan(&claimedAt, &claimExpiresAt); err != nil {
		t.Fatalf("read migrated claims: %v", err)
	}
	if claimedAt == nil || claimExpiresAt == nil {
		t.Fatalf("legacy claims were not preserved")
	}
	var createdAt, updatedAt string
	if err := sqlDB.QueryRow(`SELECT created_at, updated_at FROM issues WHERE id = 1`).Scan(&createdAt, &updatedAt); err != nil {
		t.Fatalf("read migrated timestamps: %v", err)
	}
	if !strings.HasPrefix(createdAt, "2025-01-01") || !strings.HasPrefix(updatedAt, "2025-01-02") {
		t.Fatalf("legacy timestamps = (%q, %q)", createdAt, updatedAt)
	}
	var dependencyCount int
	if err := sqlDB.QueryRow(`SELECT COUNT(*) FROM dependencies WHERE issue_id = 2 AND depends_on_id = 1`).Scan(&dependencyCount); err != nil {
		t.Fatalf("read migrated dependency: %v", err)
	}
	if dependencyCount != 1 {
		t.Fatalf("legacy dependency count = %d, want 1", dependencyCount)
	}

	for _, indexName := range []string{"idx_issues_work_id", "idx_issues_plan_work_id"} {
		var unique int
		if err := sqlDB.QueryRow(`SELECT "unique" FROM pragma_index_list('issues') WHERE name = ?`, indexName).Scan(&unique); err != nil {
			t.Fatalf("inspect %s: %v", indexName, err)
		}
		if unique != 0 {
			t.Fatalf("%s must be nonunique", indexName)
		}
	}
}
