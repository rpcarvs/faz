package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const (
	DirName    = ".faz"
	DBFileName = "faz.db"
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
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite database: %w", err)
	}

	return db, nil
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
