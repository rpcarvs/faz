package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// TestOpenProjectDBMigratesExistingDatabase verifies normal opens upgrade initialized legacy storage.
func TestOpenProjectDBMigratesExistingDatabase(t *testing.T) {
	projectDir := t.TempDir()
	createLegacyProjectDB(t, projectDir)

	sqlDB, dbPath, err := OpenProjectDB(projectDir)
	if err != nil {
		t.Fatalf("open migrated project database: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()
	if dbPath != filepath.Join(projectDir, DirName, DBFileName) {
		t.Fatalf("database path = %q", dbPath)
	}

	var planID, workID *string
	if err := sqlDB.QueryRow(`SELECT plan_id, work_id FROM issues WHERE public_id = 'faz-old1'`).Scan(&planID, &workID); err != nil {
		t.Fatalf("read migrated associations: %v", err)
	}
	if planID != nil || workID != nil {
		t.Fatalf("migrated associations = (%v, %v), want NULL values", planID, workID)
	}
}

// TestOpenProjectDBMigratesConcurrentLegacyOpens verifies simultaneous commands serialize an upgrade safely.
func TestOpenProjectDBMigratesConcurrentLegacyOpens(t *testing.T) {
	projectDir := t.TempDir()
	createLegacyProjectDB(t, projectDir)

	const openers = 8
	errs := make(chan error, openers)
	var group sync.WaitGroup
	for range openers {
		group.Add(1)
		go func() {
			defer group.Done()
			sqlDB, _, err := OpenProjectDB(projectDir)
			if err != nil {
				errs <- err
				return
			}
			defer func() { _ = sqlDB.Close() }()
			var count int
			if err := sqlDB.QueryRow(`SELECT COUNT(plan_id) FROM issues`).Scan(&count); err != nil {
				errs <- err
			}
		}()
	}
	group.Wait()
	close(errs)
	for err := range errs {
		t.Errorf("concurrent project open: %v", err)
	}
}

// TestOpenProjectDBDoesNotInitializeMissingStorage verifies ordinary commands cannot create task storage implicitly.
func TestOpenProjectDBDoesNotInitializeMissingStorage(t *testing.T) {
	projectDir := t.TempDir()
	if _, _, err := OpenProjectDB(projectDir); !errors.Is(err, ErrNotInitialized) {
		t.Fatalf("open uninitialized project error = %v, want ErrNotInitialized", err)
	}
	if _, err := os.Stat(filepath.Join(projectDir, DirName)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("ordinary open created .faz directory: %v", err)
	}
}

// TestMigrateRollsBackFailedUpgradeAndRetries verifies a failed transactional upgrade leaves no partial columns.
func TestMigrateRollsBackFailedUpgradeAndRetries(t *testing.T) {
	sqlDB, err := Open(filepath.Join(t.TempDir(), "retry.db"))
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
		t.Fatalf("create retry schema: %v", err)
	}
	if _, err := sqlDB.Exec(`INSERT INTO issues(public_id, title, type) VALUES('duplicate', 'first', 'task'), ('duplicate', 'second', 'task')`); err != nil {
		t.Fatalf("insert duplicate legacy IDs: %v", err)
	}
	if err := Migrate(sqlDB); err == nil {
		t.Fatalf("migration with duplicate public IDs succeeded unexpectedly")
	}
	if issueColumnExists(t, sqlDB, "plan_id") {
		t.Fatalf("failed migration left plan_id column behind")
	}
	if _, err := sqlDB.Exec(`UPDATE issues SET public_id = 'unique' WHERE id = 2`); err != nil {
		t.Fatalf("repair duplicate public ID: %v", err)
	}
	if err := Migrate(sqlDB); err != nil {
		t.Fatalf("retry migration: %v", err)
	}
	if !issueColumnExists(t, sqlDB, "plan_id") || !issueColumnExists(t, sqlDB, "work_id") {
		t.Fatalf("retry migration did not add SDD columns")
	}
}

// issueColumnExists reports whether a column is present in the issues table.
func issueColumnExists(t *testing.T, sqlDB *sql.DB, wanted string) bool {
	t.Helper()
	rows, err := sqlDB.Query(`PRAGMA table_info(issues)`)
	if err != nil {
		t.Fatalf("inspect issue columns: %v", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var cid int
		var name string
		var columnType string
		var notNull int
		var defaultValue sql.NullString
		var primaryKey int
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			t.Fatalf("scan issue column: %v", err)
		}
		if name == wanted {
			return true
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate issue columns: %v", err)
	}
	return false
}

// createLegacyProjectDB writes a pre-SDD initialized database without plan/work columns.
func createLegacyProjectDB(t *testing.T, projectDir string) {
	t.Helper()
	dbPath, err := EnsureProjectFiles(projectDir)
	if err != nil {
		t.Fatalf("create legacy project files: %v", err)
	}
	sqlDB, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open legacy database: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()
	if _, err := sqlDB.Exec(`CREATE TABLE issues (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id TEXT UNIQUE,
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
		t.Fatalf("create legacy schema: %v", err)
	}
	if _, err := sqlDB.Exec(`INSERT INTO issues(public_id, title, type) VALUES('faz-old1', 'Legacy task', 'task')`); err != nil {
		t.Fatalf("insert legacy issue: %v", err)
	}
}

func TestIsRetryableOpenError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil",
			err:  nil,
			want: false,
		},
		{
			name: "sqlite busy",
			err:  fmt.Errorf("database is locked (5) (SQLITE_BUSY)"),
			want: true,
		},
		{
			name: "wrapped busy",
			err:  fmt.Errorf("outer: %w", fmt.Errorf("ping sqlite database: database is locked")),
			want: true,
		},
		{
			name: "non retryable",
			err:  fmt.Errorf("no such table: issues"),
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isRetryableOpenError(tc.err); got != tc.want {
				t.Fatalf("isRetryableOpenError(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestOpenBackoff(t *testing.T) {
	if got := openBackoff(0); got != 20*time.Millisecond {
		t.Fatalf("attempt 0 backoff = %s, want %s", got, 20*time.Millisecond)
	}
	if got := openBackoff(3); got != 160*time.Millisecond {
		t.Fatalf("attempt 3 backoff = %s, want %s", got, 160*time.Millisecond)
	}
	if got := openBackoff(10); got != 320*time.Millisecond {
		t.Fatalf("attempt 10 backoff = %s, want %s", got, 320*time.Millisecond)
	}
}
