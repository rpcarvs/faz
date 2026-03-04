package skillinstaller

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallCodexContextAppendsManagedBlock(t *testing.T) {
	tmp := t.TempDir()
	codexHome := filepath.Join(tmp, "codex-home")
	t.Setenv("CODEX_HOME", codexHome)
	t.Setenv("HOME", tmp)

	path, action, err := InstallCodexContext()
	if err != nil {
		t.Fatalf("install codex context: %v", err)
	}
	if action != "appended" {
		t.Fatalf("expected appended action, got %q", action)
	}
	expectedPath := filepath.Join(codexHome, "AGENTS.md")
	if path != expectedPath {
		t.Fatalf("expected path %q, got %q", expectedPath, path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, contextBlockBegin) || !strings.Contains(content, contextBlockEnd) {
		t.Fatalf("missing managed markers:\n%s", content)
	}
	if !strings.Contains(content, mandatoryContextHeading) {
		t.Fatalf("missing mandatory heading:\n%s", content)
	}
}

func TestInstallContextAtPathTogglesOffWhenBlockExists(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "AGENTS.md")

	firstAction, err := InstallContextAtPath(path)
	if err != nil {
		t.Fatalf("first install: %v", err)
	}
	if firstAction != "appended" {
		t.Fatalf("expected appended action, got %q", firstAction)
	}

	secondAction, err := InstallContextAtPath(path)
	if err != nil {
		t.Fatalf("second install: %v", err)
	}
	if secondAction != "removed" {
		t.Fatalf("expected removed action, got %q", secondAction)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	content := string(data)
	if strings.Contains(content, contextBlockBegin) || strings.Contains(content, contextBlockEnd) {
		t.Fatalf("managed block should be removed:\n%s", content)
	}
	if strings.Contains(content, mandatoryContextHeading) {
		t.Fatalf("mandatory heading should be removed:\n%s", content)
	}
}

func TestInstallClaudeContextPreservesExistingContent(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	path := filepath.Join(tmp, ".claude", "CLAUDE.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create dir: %v", err)
	}
	if err := os.WriteFile(path, []byte("# Existing\n\nKeep this.\n"), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	installedPath, action, err := InstallClaudeContext()
	if err != nil {
		t.Fatalf("install claude context: %v", err)
	}
	if action != "appended" {
		t.Fatalf("expected appended action, got %q", action)
	}
	if installedPath != path {
		t.Fatalf("expected path %q, got %q", path, installedPath)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "# Existing") {
		t.Fatalf("existing content removed unexpectedly:\n%s", content)
	}
	if strings.Count(content, contextBlockBegin) != 1 || strings.Count(content, contextBlockEnd) != 1 {
		t.Fatalf("expected one managed block:\n%s", content)
	}
}

func TestInstallContextAtPathRemovesOnlyManagedBlock(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "CLAUDE.md")

	seed := "# Header\n\n" +
		contextBlockBegin + "\n" + mandatoryContextBody + "\n" + contextBlockEnd + "\n\n" +
		"# Footer\n"
	if err := os.WriteFile(path, []byte(seed), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	action, err := InstallContextAtPath(path)
	if err != nil {
		t.Fatalf("toggle install: %v", err)
	}
	if action != "removed" {
		t.Fatalf("expected removed action, got %q", action)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	content := string(data)
	if strings.Contains(content, contextBlockBegin) || strings.Contains(content, mandatoryContextHeading) {
		t.Fatalf("managed content should be removed:\n%s", content)
	}
	if !strings.Contains(content, "# Header") || !strings.Contains(content, "# Footer") {
		t.Fatalf("surrounding content should stay:\n%s", content)
	}
}
