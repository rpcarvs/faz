package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
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

		initializedGit, err := ensureGitRepo(projectDir)
		if err != nil {
			return err
		}
		if err := ensureGitExclude(projectDir); err != nil {
			return err
		}

		fmt.Println("faz initialized")
		fmt.Println("Directory:", fazDir)
		fmt.Println("Database:", dbPath)
		if initializedGit {
			fmt.Println("Git:", "initialized")
		} else {
			fmt.Println("Git:", "already initialized")
		}
		fmt.Println("Exclude:", ".git/info/exclude updated")
		return nil
	},
}

func ensureGitRepo(projectDir string) (bool, error) {
	gitDir := filepath.Join(projectDir, ".git")
	if stat, err := os.Stat(gitDir); err == nil && stat.IsDir() {
		return false, nil
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = projectDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return false, fmt.Errorf("git init failed: %v (%s)", err, strings.TrimSpace(string(output)))
	}

	return true, nil
}

func ensureGitExclude(projectDir string) error {
	excludePath := filepath.Join(projectDir, ".git", "info", "exclude")
	if err := os.MkdirAll(filepath.Dir(excludePath), 0o755); err != nil {
		return fmt.Errorf("create .git/info directory: %w", err)
	}

	existing := make(map[string]struct{})
	file, err := os.OpenFile(excludePath, os.O_CREATE|os.O_RDONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open .git/info/exclude: %w", err)
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
		return fmt.Errorf("read .git/info/exclude: %w", err)
	}
	_ = file.Close()

	required := []string{
		".faz/",
		".claude/",
		"*CLAUDE.md",
		"*AGENTS.md",
		"*PLAN.md",
	}

	toAppend := make([]string, 0)
	for _, line := range required {
		if _, ok := existing[line]; !ok {
			toAppend = append(toAppend, line)
		}
	}
	if len(toAppend) == 0 {
		return nil
	}

	appendFile, err := os.OpenFile(excludePath, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("append .git/info/exclude: %w", err)
	}
	defer func() { _ = appendFile.Close() }()

	for _, line := range toAppend {
		if _, err := appendFile.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("write .git/info/exclude: %w", err)
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(initCmd)
}
