package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"faz/internal/db"
	"faz/internal/model"
	"faz/internal/repo"
	"faz/internal/service"
)

const (
	ansiReset  = "\033[0m"
	ansiGray   = "\033[90m"
	ansiGreen  = "\033[32m"
	ansiRed    = "\033[31m"
	ansiOrange = "\033[38;5;208m"
	ansiAmber  = "\033[38;5;214m"
	ansiYellow = "\033[33m"
)

func currentProjectDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve working directory: %w", err)
	}
	return cwd, nil
}

func openService() (*service.IssueService, *sql.DB, error) {
	projectDir, err := currentProjectDir()
	if err != nil {
		return nil, nil, err
	}

	sqlDB, _, err := db.OpenProjectDB(projectDir)
	if err != nil {
		if err == db.ErrNotInitialized {
			return nil, nil, fmt.Errorf("project is not initialized. Run `faz init`")
		}
		return nil, nil, err
	}
	if err := db.Migrate(sqlDB); err != nil {
		_ = sqlDB.Close()
		return nil, nil, err
	}

	projectName := filepath.Base(projectDir)
	issueRepo := repo.NewIssueRepo(sqlDB)
	return service.NewIssueService(issueRepo, projectName), sqlDB, nil
}

func printIssueTable(issues []model.Issue) {
	if len(issues) == 0 {
		fmt.Println("No issues found")
		return
	}

	writer := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	_, _ = fmt.Fprintln(writer, "ID\tTYPE\tPRIORITY\tSTATUS\tTITLE")
	for _, issue := range issues {
		_, _ = fmt.Fprintf(writer, "%s\t%s\tP%d\t%s\t%s\n", issue.ID, issue.Type, issue.Priority, issue.Status, issue.Title)
	}
	_ = writer.Flush()
}

func printIssueList(issues []model.Issue) {
	if len(issues) == 0 {
		fmt.Println("No issues found")
		return
	}
	for _, issue := range issues {
		symbol := statusSymbol(issue.Status)

		priority := colorizePriority(issue.Priority)
		typeLabel := "[" + issue.Type + "]"
		title := issue.Title
		if issue.Status == "closed" {
			line := fmt.Sprintf("%s %s [P%d] %s - %s", symbol, issue.ID, issue.Priority, typeLabel, title)
			fmt.Println(ansiGray + line + ansiReset)
			continue
		}
		if issue.Type == "epic" {
			typeLabel = colorizeEpic(typeLabel)
			title = colorizeEpic(title)
		}

		fmt.Printf("%s %s %s %s - %s\n", symbol, issue.ID, priority, typeLabel, title)
	}
}

func statusSymbol(status string) string {
	switch status {
	case "open":
		return "○"
	case "in_progress":
		return "◐"
	case "closed":
		return "✓"
	default:
		return "?"
	}
}

func colorizeEpic(v string) string {
	return ansiGreen + v + ansiReset
}

func colorizePriority(priority int) string {
	label := fmt.Sprintf("[P%d]", priority)
	switch priority {
	case 0:
		return ansiRed + label + ansiReset
	case 1:
		return ansiOrange + label + ansiReset
	case 2:
		return ansiAmber + label + ansiReset
	default:
		return ansiYellow + label + ansiReset
	}
}

func parseIDs(args []string) ([]string, error) {
	ids := make([]string, 0, len(args))
	for _, raw := range args {
		id, err := service.NormalizeIssueID(raw)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func defaultDescription(input string) string {
	return strings.TrimSpace(input)
}

func fazPaths(projectDir string) (string, string) {
	fazDir := filepath.Join(projectDir, db.DirName)
	return fazDir, filepath.Join(fazDir, db.DBFileName)
}
