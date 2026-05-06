package installskills

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestInstallCodexCommandInstallsGlobalIntegration covers the simplified command.
func TestInstallCodexCommandInstallsGlobalIntegration(t *testing.T) {
	tmp := t.TempDir()
	codexHome := filepath.Join(tmp, "codex-home")
	t.Setenv("CODEX_HOME", codexHome)
	t.Setenv("HOME", tmp)

	output := runInstallCommand(t, func() (string, error) { return "", nil }, "codex")

	if !strings.Contains(output, "Installed Codex global integration") {
		t.Fatalf("unexpected output:\n%s", output)
	}
	assertPathExists(t, filepath.Join(codexHome, "skills", "task-management-with-faz", "SKILL.md"))
	assertPathExists(t, filepath.Join(codexHome, "AGENTS.md"))
	assertPathExists(t, filepath.Join(codexHome, "hooks.json"))
	assertPathExists(t, filepath.Join(codexHome, "config.toml"))
}

// TestInstallClaudeLocalCommandUsesProjectRoot verifies local install paths.
func TestInstallClaudeLocalCommandUsesProjectRoot(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", t.TempDir())

	output := runInstallCommand(t, func() (string, error) { return root, nil }, "claude", "--local")

	if !strings.Contains(output, "Installed Claude local integration") {
		t.Fatalf("unexpected output:\n%s", output)
	}
	assertPathExists(t, filepath.Join(root, "AGENTS.md"))
	assertPathExists(t, filepath.Join(root, "CLAUDE.md"))
	assertPathExists(t, filepath.Join(root, ".claude", "settings.json"))
	assertPathExists(t, filepath.Join(root, ".claude", "skills", "task-management-with-faz", "SKILL.md"))
}

// runInstallCommand executes the install command with test-local IO.
func runInstallCommand(t *testing.T, root ProjectRootFunc, args ...string) string {
	t.Helper()

	cmd := NewCommand(root)
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute install %v: %v\n%s", args, err, output.String())
	}
	return output.String()
}

// assertPathExists fails when a required install artifact is missing.
func assertPathExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected path %s: %v", path, err)
	}
}
