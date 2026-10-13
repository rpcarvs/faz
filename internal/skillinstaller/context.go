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
	contextBlockBegin = "<!-- FAZ-TASK-MANAGEMENT:BEGIN -->"
	contextBlockEnd   = "<!-- FAZ-TASK-MANAGEMENT:END -->"
	pointerBlockBegin = "<!-- FAZ-CLAUDE-POINTER:BEGIN -->"
	pointerBlockEnd   = "<!-- FAZ-CLAUDE-POINTER:END -->"
)

var managedContextBlockPattern = regexp.MustCompile(`(?s)` + regexp.QuoteMeta(contextBlockBegin) + `.*?` + regexp.QuoteMeta(contextBlockEnd))
var managedPointerBlockPattern = regexp.MustCompile(`(?s)` + regexp.QuoteMeta(pointerBlockBegin) + `.*?` + regexp.QuoteMeta(pointerBlockEnd))

const mandatoryContextBody = `# MANDATORY Task Management

These rules govern implementation and its execution plan. Read-only questions, research, file inspection, codebase exploration, reviews, and diagnosis need no investigation task, claim, or Faz approval. Identified bugs must still be filed. Missing Faz or a repository must not block read-only assistance; report any inability to file bugs without initializing a repository solely for tracking.

- Use the ` + "`faz-task-management`" + ` SKILL and the ` + "`faz`" + ` CLI for all task management, tracking, and status updates.
- Remember to add a meaningful and rich descriptions for the tasks so other agents will know what needs to be done by just reading the task.
- After preparing tasks for planned changes, wait for user approval before implementation. Bug filing requires no approval or claim.
- Instead of aggregating a lot of work in one task, prefer creating several task with concise short work.
- Group epics by domain or deliverable, not by the whole project; small work may need one or none.
- Plan for parallel work with non-conflicting file ownership. Record every real prerequisite, but never add blockers merely to impose an order.
- Type reminder: Classify each issue type: task, bug, feature, chore, decision.
- Automatically create or reuse a bug issue when a bug is identified, including during read-only reviews. Record other discovered work during tracked implementation. Filing does not authorize fixing.
- When creating or updating files under ` + "`faz-specs/`" + `, exclude Faz issue IDs from every section, including decision and revision notes. Committed documents must be understandable without the uncommitted ` + "`.faz/`" + ` database. Preserve substantive information and document references; keep issue provenance in Faz. This applies during ordinary implementation as well as SDD.
- **Never take decisions autonomously!** If you need a decision, immediately create a ` + "`decision`" + ` issue with FAZ and continue your work if you can, communicating the need to the user at the end. If the decision is needed immediately, stop and communicate to the user.`

const claudeLocalPointerBody = "See [AGENTS.md](./AGENTS.md)"

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

// InstallCodexContext installs or updates the managed context block in Codex AGENTS.md.
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

// InstallClaudeContext installs or updates the managed context block in Claude CLAUDE.md.
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

	updated, action := upsertContextBlock(string(existing))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("create context directory %s: %w", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return "", fmt.Errorf("write context file %s: %w", path, err)
	}
	return action, nil
}

// InstallClaudePointerAtPath appends or updates the managed Claude pointer block.
func InstallClaudePointerAtPath(path string) (string, error) {
	existing, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("read Claude pointer file %s: %w", path, err)
	}

	updated, action := upsertClaudePointerBlock(string(existing))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("create Claude pointer directory %s: %w", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return "", fmt.Errorf("write Claude pointer file %s: %w", path, err)
	}
	return action, nil
}

// upsertContextBlock keeps exactly one managed block with the latest text.
func upsertContextBlock(content string) (string, string) {
	block := contextBlockBegin + "\n" + mandatoryContextBody + "\n" + contextBlockEnd
	if managedContextBlockPattern.MatchString(content) {
		replaced := managedContextBlockPattern.ReplaceAllString(content, block)
		replaced = strings.TrimRight(replaced, "\n\t ")
		if replaced == "" {
			return block + "\n", "updated"
		}
		return replaced + "\n", "updated"
	}

	base := strings.TrimRight(content, "\n\t ")
	if base == "" {
		return block + "\n", "appended"
	}
	return base + "\n\n" + block + "\n", "appended"
}

// upsertClaudePointerBlock keeps exactly one managed Claude pointer block.
func upsertClaudePointerBlock(content string) (string, string) {
	block := pointerBlockBegin + "\n" + claudeLocalPointerBody + "\n" + pointerBlockEnd
	if managedPointerBlockPattern.MatchString(content) {
		replaced := managedPointerBlockPattern.ReplaceAllString(content, block)
		replaced = strings.TrimRight(replaced, "\n\t ")
		if replaced == "" {
			return block + "\n", "updated"
		}
		return replaced + "\n", "updated"
	}

	base := strings.TrimRight(content, "\n\t ")
	if base == "" {
		return block + "\n", "appended"
	}
	return base + "\n\n" + block + "\n", "appended"
}
