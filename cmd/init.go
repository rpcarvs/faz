package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"faz/internal/db"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize faz in the current project",
	RunE: func(cmd *cobra.Command, args []string) error {
		projectDir, err := currentProjectDir()
		if err != nil {
			return err
		}

		fazDir, _ := fazPaths(projectDir)
		dbPath, err := db.EnsureProjectFiles(projectDir)
		if err != nil {
			return err
		}

		sqlDB, err := db.Open(dbPath)
		if err != nil {
			return err
		}
		defer func() { _ = sqlDB.Close() }()

		if err := db.Migrate(sqlDB); err != nil {
			return err
		}

		addedGitIgnore, err := ensureGitIgnoreEntry(projectDir, ".faz/")
		if err != nil {
			return err
		}

		fmt.Println("faz initialized")
		fmt.Println("Directory:", fazDir)
		fmt.Println("Database:", dbPath)
		if addedGitIgnore {
			fmt.Println("Gitignore:", ".faz/ added")
		} else {
			fmt.Println("Gitignore:", ".faz/ already present")
		}
		return nil
	},
}

func ensureGitIgnoreEntry(projectDir, entry string) (bool, error) {
	gitIgnorePath := filepath.Join(projectDir, ".gitignore")
	existing := make(map[string]struct{})
	file, err := os.OpenFile(gitIgnorePath, os.O_CREATE|os.O_RDONLY, 0o644)
	if err != nil {
		return false, fmt.Errorf("open .gitignore: %w", err)
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		existing[line] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		_ = file.Close()
		return false, fmt.Errorf("read .gitignore: %w", err)
	}
	_ = file.Close()

	if _, ok := existing[entry]; ok {
		return false, nil
	}

	appendFile, err := os.OpenFile(gitIgnorePath, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return false, fmt.Errorf("append .gitignore: %w", err)
	}
	defer func() { _ = appendFile.Close() }()

	if _, err := appendFile.WriteString(entry + "\n"); err != nil {
		return false, fmt.Errorf("write .gitignore: %w", err)
	}

	return true, nil
}

func init() {
	rootCmd.AddCommand(initCmd)
}
