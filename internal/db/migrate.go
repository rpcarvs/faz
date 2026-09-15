package db

import (
	"context"
	"database/sql"
	"fmt"
)

// schemaVersion covers the current tables, association columns, indexes and triggers.
const schemaVersion = 1

// Migrate upgrades older databases and leaves current schemas read-only.
func Migrate(db *sql.DB) error {
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("acquire migration connection: %w", err)
	}
	defer func() { _ = conn.Close() }()
	needed, err := needsMigration(ctx, conn)
	if err != nil || !needed {
		return err
	}
	if _, err := conn.ExecContext(ctx, `PRAGMA foreign_keys = ON`); err != nil {
		return fmt.Errorf("enable migration foreign keys: %w", err)
	}

	if _, err := conn.ExecContext(ctx, `BEGIN IMMEDIATE`); err != nil {
		return fmt.Errorf("begin migration: %w", err)
	}
	transactionOpen := true
	defer func() {
		if transactionOpen {
			_, _ = conn.ExecContext(ctx, `ROLLBACK`)
		}
	}()
	// Another opener may have completed the upgrade while this connection waited.
	needed, err = needsMigration(ctx, conn)
	if err != nil || !needed {
		return err
	}
	if err := migrateConnection(ctx, conn); err != nil {
		return err
	}
	if _, err := conn.ExecContext(ctx, fmt.Sprintf(`PRAGMA user_version = %d`, schemaVersion)); err != nil {
		return fmt.Errorf("record schema version: %w", err)
	}
	if _, err := conn.ExecContext(ctx, `COMMIT`); err != nil {
		return fmt.Errorf("commit migration: %w", err)
	}
	transactionOpen = false
	return nil
}

// needsMigration reads the schema version without requesting a writer lock.
func needsMigration(ctx context.Context, conn *sql.Conn) (bool, error) {
	var version int
	if err := conn.QueryRowContext(ctx, `PRAGMA user_version`).Scan(&version); err != nil {
		return false, fmt.Errorf("read schema version: %w", err)
	}
	if version < 0 || version > schemaVersion {
		return false, fmt.Errorf("unsupported database schema version %d (supported: %d)", version, schemaVersion)
	}
	return version < schemaVersion, nil
}

// migrationExecutor is the single pinned connection used for a migration transaction.
type migrationExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

// migrateConnection applies all schema changes while a SQLite immediate transaction prevents concurrent upgrades.
func migrateConnection(ctx context.Context, db migrationExecutor) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS issues (
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
			plan_id TEXT,
			work_id TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			closed_at DATETIME,
			FOREIGN KEY(parent_id) REFERENCES issues(id) ON DELETE SET NULL
		);`,
		`CREATE TABLE IF NOT EXISTS dependencies (
			issue_id INTEGER NOT NULL,
			depends_on_id INTEGER NOT NULL,
			PRIMARY KEY (issue_id, depends_on_id),
			FOREIGN KEY(issue_id) REFERENCES issues(id) ON DELETE CASCADE,
			FOREIGN KEY(depends_on_id) REFERENCES issues(id) ON DELETE CASCADE,
			CHECK (issue_id != depends_on_id)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_issues_status ON issues(status);`,
		`CREATE INDEX IF NOT EXISTS idx_issues_parent ON issues(parent_id);`,
		`CREATE INDEX IF NOT EXISTS idx_issues_closed_at ON issues(closed_at);`,
		`CREATE TRIGGER IF NOT EXISTS trg_issues_updated_at
		AFTER UPDATE ON issues
		FOR EACH ROW
		BEGIN
			UPDATE issues SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
		END;`,
	}

	for _, stmt := range statements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("run migration statement: %w", err)
		}
	}

	if err := ensureIssuesColumns(ctx, db); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, `UPDATE issues SET public_id = 'legacy-' || id WHERE public_id IS NULL OR public_id = ''`); err != nil {
		return fmt.Errorf("backfill public IDs: %w", err)
	}

	return nil
}

// ensureIssuesColumns adds missing issues columns and indexes for upgrades.
func ensureIssuesColumns(ctx context.Context, db migrationExecutor) error {
	rows, err := db.QueryContext(ctx, `PRAGMA table_info(issues);`)
	if err != nil {
		return fmt.Errorf("inspect issues table columns: %w", err)
	}
	defer func() { _ = rows.Close() }()

	hasPublicID := false
	hasClaimedAt := false
	hasClaimExpiresAt := false
	hasPlanID := false
	hasWorkID := false
	for rows.Next() {
		var cid int
		var name string
		var columnType string
		var notNull int
		var defaultValue sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &pk); err != nil {
			return fmt.Errorf("scan issues table metadata: %w", err)
		}
		if name == "public_id" {
			hasPublicID = true
		}
		if name == "claimed_at" {
			hasClaimedAt = true
		}
		if name == "claim_expires_at" {
			hasClaimExpiresAt = true
		}
		if name == "plan_id" {
			hasPlanID = true
		}
		if name == "work_id" {
			hasWorkID = true
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate issues table metadata: %w", err)
	}

	if !hasPublicID {
		if _, err := db.ExecContext(ctx, `ALTER TABLE issues ADD COLUMN public_id TEXT`); err != nil {
			return fmt.Errorf("add public_id column: %w", err)
		}
	}
	if !hasClaimedAt {
		if _, err := db.ExecContext(ctx, `ALTER TABLE issues ADD COLUMN claimed_at DATETIME`); err != nil {
			return fmt.Errorf("add claimed_at column: %w", err)
		}
	}
	if !hasClaimExpiresAt {
		if _, err := db.ExecContext(ctx, `ALTER TABLE issues ADD COLUMN claim_expires_at DATETIME`); err != nil {
			return fmt.Errorf("add claim_expires_at column: %w", err)
		}
	}
	if !hasPlanID {
		if _, err := db.ExecContext(ctx, `ALTER TABLE issues ADD COLUMN plan_id TEXT`); err != nil {
			return fmt.Errorf("add plan_id column: %w", err)
		}
	}
	if !hasWorkID {
		if _, err := db.ExecContext(ctx, `ALTER TABLE issues ADD COLUMN work_id TEXT`); err != nil {
			return fmt.Errorf("add work_id column: %w", err)
		}
	}

	if _, err := db.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS idx_issues_public_id_unique ON issues(public_id)`); err != nil {
		return fmt.Errorf("create unique public_id index: %w", err)
	}
	if _, err := db.ExecContext(ctx, `DROP INDEX IF EXISTS idx_issues_public_id`); err != nil {
		return fmt.Errorf("drop redundant public_id index: %w", err)
	}
	if _, err := db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_issues_claim_expires_at ON issues(claim_expires_at)`); err != nil {
		return fmt.Errorf("create claim_expires_at index: %w", err)
	}
	if _, err := db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_issues_work_id ON issues(work_id)`); err != nil {
		return fmt.Errorf("create work_id index: %w", err)
	}
	if _, err := db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_issues_plan_work_id ON issues(plan_id, work_id)`); err != nil {
		return fmt.Errorf("create plan/work index: %w", err)
	}

	return nil
}
