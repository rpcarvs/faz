package db

import (
	"database/sql"
	"fmt"
)

// Migrate creates and updates required tables for the faz database.
func Migrate(db *sql.DB) error {
	statements := []string{
		`PRAGMA foreign_keys = ON;`,
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
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("run migration statement: %w", err)
		}
	}

	if err := ensureIssuesColumns(db); err != nil {
		return err
	}

	if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_issues_public_id ON issues(public_id)`); err != nil {
		return fmt.Errorf("create public_id index: %w", err)
	}

	if _, err := db.Exec(`UPDATE issues SET public_id = 'legacy-' || id WHERE public_id IS NULL OR public_id = ''`); err != nil {
		return fmt.Errorf("backfill public IDs: %w", err)
	}

	return nil
}

func ensureIssuesColumns(db *sql.DB) error {
	rows, err := db.Query(`PRAGMA table_info(issues);`)
	if err != nil {
		return fmt.Errorf("inspect issues table columns: %w", err)
	}
	defer func() { _ = rows.Close() }()

	hasPublicID := false
	hasClaimedAt := false
	hasClaimExpiresAt := false
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
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate issues table metadata: %w", err)
	}

	if !hasPublicID {
		if _, err := db.Exec(`ALTER TABLE issues ADD COLUMN public_id TEXT`); err != nil {
			return fmt.Errorf("add public_id column: %w", err)
		}
	}
	if !hasClaimedAt {
		if _, err := db.Exec(`ALTER TABLE issues ADD COLUMN claimed_at DATETIME`); err != nil {
			return fmt.Errorf("add claimed_at column: %w", err)
		}
	}
	if !hasClaimExpiresAt {
		if _, err := db.Exec(`ALTER TABLE issues ADD COLUMN claim_expires_at DATETIME`); err != nil {
			return fmt.Errorf("add claim_expires_at column: %w", err)
		}
	}

	if _, err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_issues_public_id_unique ON issues(public_id)`); err != nil {
		return fmt.Errorf("create unique public_id index: %w", err)
	}
	if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_issues_claim_expires_at ON issues(claim_expires_at)`); err != nil {
		return fmt.Errorf("create claim_expires_at index: %w", err)
	}

	return nil
}
