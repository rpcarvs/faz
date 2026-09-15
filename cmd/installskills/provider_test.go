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
	assertPathExists(t, filepath.Join(codexHome, "skills", "faz-task-management", "SKILL.md"))
	assertPathExists(t, filepath.Join(codexHome, "skills", "faz-spec-driven", "SKILL.md"))
	assertPathExists(t, filepath.Join(codexHome, "skills", "faz-spec-driven", "agents", "openai.yaml"))
	if count := strings.Count(output, "  Skill: "); count != 2 {
		t.Fatalf("expected two installed skills, got %d:\n%s", count, output)
	}
	assertPathExists(t, filepath.Join(codexHome, "AGENTS.md"))
	assertPathExists(t, filepath.Join(codexHome, "hooks.json"))
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
	assertPathExists(t, filepath.Join(root, ".claude", "skills", "faz-task-management", "SKILL.md"))
	assertPathExists(t, filepath.Join(root, ".claude", "skills", "faz-spec-driven", "SKILL.md"))
	if count := strings.Count(output, "  Skill: "); count != 2 {
		t.Fatalf("expected two installed skills, got %d:\n%s", count, output)
	}
}

// TestInstallHookStatuses verifies force propagation and output for both providers and scopes.
func TestInstallHookStatuses(t *testing.T) {
	for _, provider := range []string{"codex", "claude"} {
		for _, local := range []bool{false, true} {
			scope := "global"
			if local {
				scope = "local"
			}
			t.Run(provider+"/"+scope, func(t *testing.T) {
				tmp := t.TempDir()
				root := filepath.Join(tmp, "project")
				t.Setenv("HOME", tmp)
				t.Setenv("CODEX_HOME", filepath.Join(tmp, ".codex"))
				args := []string{provider}
				base := tmp
				if local {
					args = append(args, "--local")
					base = root
				}
				file := "hooks.json"
				if provider == "claude" {
					file = "settings.json"
				}
				path := filepath.Join(base, "."+provider, file)
				resolve := func() (string, error) { return root, nil }
				for _, status := range []string{"created", "unchanged"} {
					output := runInstallCommand(t, resolve, args...)
					if !strings.Contains(output, "Hook ("+status+"):") {
						t.Fatalf("missing %s hook status:\n%s", status, output)
					}
				}
				current, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				custom := strings.Replace(string(current), `"timeout": 5`, `"timeout": 10`, 1)
				if custom == string(current) {
					t.Fatal("timeout fixture did not change")
				}
				if err := os.WriteFile(path, []byte(custom), 0o600); err != nil {
					t.Fatal(err)
				}
				output := runInstallCommand(t, resolve, args...)
				if !strings.Contains(output, "Hook (skipped") || !strings.Contains(output, "use --force") {
					t.Fatalf("missing skip guidance:\n%s", output)
				}
				preserved, err := os.ReadFile(path)
				if err != nil || string(preserved) != custom {
					t.Fatalf("non-force install changed hook: %v", err)
				}
				output = runInstallCommand(t, resolve, append(args, "--force")...)
				if !strings.Contains(output, "Hook (updated):") {
					t.Fatalf("missing updated hook status:\n%s", output)
				}
				updated, err := os.ReadFile(path)
				if err != nil || string(updated) != string(current) {
					t.Fatalf("force did not restore intended hook: %s, %v", updated, err)
				}
			})
		}
	}
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
