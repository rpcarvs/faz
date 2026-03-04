package skillinstaller

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	contextBlockBegin       = "<!-- FAZ-TASK-MANAGEMENT:BEGIN -->"
	contextBlockEnd         = "<!-- FAZ-TASK-MANAGEMENT:END -->"
	mandatoryContextHeading = "# MANDATORY Task Management"
)

var managedContextBlockPattern = regexp.MustCompile(`(?s)` + regexp.QuoteMeta(contextBlockBegin) + `.*?` + regexp.QuoteMeta(contextBlockEnd))

const mandatoryContextBody = `# MANDATORY Task Management

- Use the ` + "`task-management-with-faz`" + ` SKILL and the ` + "`faz`" + ` CLI for all task management, tracking, and status updates.
- Remember to add a meaningful and rich descriptions for the tasks so other agents will know what needs to be done by just reading the task.
- Instead of aggregating a lot of work in one task, prefer creating several task with concise short work.`

// CodexContextPath resolves the global Codex AGENTS.md path.
func CodexContextPath() (string, error) {
	codexHome := os.Getenv("CODEX_HOME")
	if codexHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		codexHome = filepath.Join(home, ".codex")
	}
	return filepath.Join(codexHome, "AGENTS.md"), nil
}

// ClaudeContextPath resolves the global Claude CLAUDE.md path.
func ClaudeContextPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".claude", "CLAUDE.md"), nil
}

// InstallCodexContext installs or toggles the managed context block in Codex AGENTS.md.
func InstallCodexContext() (string, string, error) {
	path, err := CodexContextPath()
	if err != nil {
		return "", "", err
	}
	action, err := InstallContextAtPath(path)
	if err != nil {
		return "", "", err
	}
	return path, action, nil
}

// InstallClaudeContext installs or toggles the managed context block in Claude CLAUDE.md.
func InstallClaudeContext() (string, string, error) {
	path, err := ClaudeContextPath()
	if err != nil {
		return "", "", err
	}
	action, err := InstallContextAtPath(path)
	if err != nil {
		return "", "", err
	}
	return path, action, nil
}

// InstallContextAtPath creates or updates the managed block at the target path.
func InstallContextAtPath(path string) (string, error) {
	existing, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("read context file %s: %w", path, err)
	}

	updated, action := toggleContextBlock(string(existing))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("create context directory %s: %w", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return "", fmt.Errorf("write context file %s: %w", path, err)
	}
	return action, nil
}

// toggleContextBlock removes managed block when heading exists, else appends it.
func toggleContextBlock(content string) (string, string) {
	if strings.Contains(content, mandatoryContextHeading) {
		cleaned := managedContextBlockPattern.ReplaceAllString(content, "")
		cleaned = strings.TrimRight(cleaned, "\n\t ")
		if cleaned == "" {
			return "", "removed"
		}
		return cleaned + "\n", "removed"
	}

	block := contextBlockBegin + "\n" + mandatoryContextBody + "\n" + contextBlockEnd
	base := strings.TrimRight(content, "\n\t ")
	if base == "" {
		return block + "\n", "appended"
	}
	return base + "\n\n" + block + "\n", "appended"
}
