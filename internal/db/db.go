package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

const (
	DirName    = ".faz"
	DBFileName = "taskstore.db"

	maxOpenAttempts = 8
	baseOpenBackoff = 20 * time.Millisecond
)

var ErrNotInitialized = errors.New("faz project is not initialized")

// EnsureProjectFiles creates the faz directory and database path.
func EnsureProjectFiles(projectDir string) (string, error) {
	fazDir := filepath.Join(projectDir, DirName)
	if err := os.MkdirAll(fazDir, 0o755); err != nil {
		return "", fmt.Errorf("create .faz directory: %w", err)
	}

	return filepath.Join(fazDir, DBFileName), nil
}

// Open opens a SQLite database from the given path.
func Open(dbPath string) (*sql.DB, error) {
	var lastRetryErr error
	for attempt := 0; attempt < maxOpenAttempts; attempt++ {
		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			return nil, fmt.Errorf("open sqlite database: %w", err)
		}
		db.SetMaxOpenConns(1)
		db.SetMaxIdleConns(1)

		if err := db.Ping(); err != nil {
			_ = db.Close()
			wrapped := fmt.Errorf("ping sqlite database: %w", err)
			if isRetryableOpenError(wrapped) {
				lastRetryErr = wrapped
				time.Sleep(openBackoff(attempt))
				continue
			}
			return nil, wrapped
		}
		if err := applyConnectionPragmas(db); err != nil {
			_ = db.Close()
			if isRetryableOpenError(err) {
				lastRetryErr = err
				time.Sleep(openBackoff(attempt))
				continue
			}
			return nil, err
		}

		return db, nil
	}

	if lastRetryErr != nil {
		return nil, fmt.Errorf("open sqlite database failed after %d attempts: %w", maxOpenAttempts, lastRetryErr)
	}
	return nil, fmt.Errorf("open sqlite database failed after %d attempts", maxOpenAttempts)
}

// OpenProjectDB opens a project database and errors if init has not been run.
func OpenProjectDB(projectDir string) (*sql.DB, string, error) {
	dbPath := filepath.Join(projectDir, DirName, DBFileName)
	if _, err := os.Stat(dbPath); errors.Is(err, os.ErrNotExist) {
		return nil, "", ErrNotInitialized
	}

	db, err := Open(dbPath)
	if err != nil {
		return nil, "", err
	}

	return db, dbPath, nil
}

// applyConnectionPragmas enables lock waiting and concurrent read/write mode.
func applyConnectionPragmas(db *sql.DB) error {
	statements := []string{
		`PRAGMA busy_timeout = 5000;`,
		`PRAGMA foreign_keys = ON;`,
	}
	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("apply sqlite pragma %q: %w", stmt, err)
		}
	}
	return nil
}

// isRetryableOpenError reports whether opening the DB should retry.
func isRetryableOpenError(err error) bool {
	if err == nil {
		return false
	}
	for current := err; current != nil; current = unwrap(current) {
		text := strings.ToLower(current.Error())
		if strings.Contains(text, "sqlite_busy") || strings.Contains(text, "database is locked") {
			return true
		}
	}
	return false
}

// openBackoff computes bounded exponential delay between open retries.
func openBackoff(attempt int) time.Duration {
	if attempt < 0 {
		return baseOpenBackoff
	}
	multiplier := 1 << attempt
	if multiplier > 16 {
		multiplier = 16
	}
	return time.Duration(multiplier) * baseOpenBackoff
}

// unwrap returns the next wrapped error when available.
func unwrap(err error) error {
	type unwrapper interface {
		Unwrap() error
	}
	w, ok := err.(unwrapper)
	if !ok {
		return nil
	}
	return w.Unwrap()
}
