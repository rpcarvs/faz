package skillinstaller

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallCodexSkillCreatesOnlySkillFile(t *testing.T) {
	tmp := t.TempDir()
	codexHome := filepath.Join(tmp, "codex-home")
	t.Setenv("CODEX_HOME", codexHome)
	t.Setenv("HOME", tmp)

	installedPath, err := InstallCodexSkill(false)
	if err != nil {
		t.Fatalf("install codex skill: %v", err)
	}

	expectedPath := filepath.Join(codexHome, "skills", skillDirName)
	if installedPath != expectedPath {
		t.Fatalf("expected %s, got %s", expectedPath, installedPath)
	}

	skillPath := filepath.Join(installedPath, "SKILL.md")
	if _, err := os.Stat(skillPath); err != nil {
		t.Fatalf("missing SKILL.md: %v", err)
	}
	assertInstalledSharedSkill(t, skillPath)
	if _, err := os.Stat(filepath.Join(installedPath, "agents")); !os.IsNotExist(err) {
		t.Fatalf("expected no agents directory, got err=%v", err)
	}
}

func TestInstallCodexSkillFallsBackToHomeWhenCodexHomeUnset(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CODEX_HOME", "")
	t.Setenv("HOME", tmp)

	installedPath, err := InstallCodexSkill(false)
	if err != nil {
		t.Fatalf("install codex skill: %v", err)
	}

	expected := filepath.Join(tmp, ".codex", "skills", skillDirName)
	if installedPath != expected {
		t.Fatalf("expected %s, got %s", expected, installedPath)
	}
}

func TestInstallCodexSkillExistingWithoutForceIsIdempotent(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CODEX_HOME", filepath.Join(tmp, "codex-home"))
	t.Setenv("HOME", tmp)

	firstPath, err := InstallCodexSkill(false)
	if err != nil {
		t.Fatalf("first install should succeed: %v", err)
	}

	secondPath, err := InstallCodexSkill(false)
	if err != nil {
		t.Fatalf("second install should succeed: %v", err)
	}
	if secondPath != firstPath {
		t.Fatalf("expected same install path, got %s and %s", firstPath, secondPath)
	}
}

func TestInstallCodexSkillForceOverwrites(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CODEX_HOME", filepath.Join(tmp, "codex-home"))
	t.Setenv("HOME", tmp)

	installedPath, err := InstallCodexSkill(false)
	if err != nil {
		t.Fatalf("first install should succeed: %v", err)
	}

	customFile := filepath.Join(installedPath, "custom.txt")
	if err := os.WriteFile(customFile, []byte("custom"), 0o644); err != nil {
		t.Fatalf("write custom file: %v", err)
	}

	if _, err := InstallCodexSkill(true); err != nil {
		t.Fatalf("force install should succeed: %v", err)
	}

	if _, err := os.Stat(customFile); err == nil {
		t.Fatal("expected custom file removed by force overwrite")
	}
}

func TestInstallClaudeSkillCreatesOnlySkillFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	installedPath, err := InstallClaudeSkill(false)
	if err != nil {
		t.Fatalf("install claude skill: %v", err)
	}

	expectedPath := filepath.Join(tmp, ".claude", "skills", skillDirName)
	if installedPath != expectedPath {
		t.Fatalf("expected %s, got %s", expectedPath, installedPath)
	}

	skillPath := filepath.Join(installedPath, "SKILL.md")
	if _, err := os.Stat(skillPath); err != nil {
		t.Fatalf("missing SKILL.md: %v", err)
	}
	assertInstalledSharedSkill(t, skillPath)
	if _, err := os.Stat(filepath.Join(installedPath, "agents")); !os.IsNotExist(err) {
		t.Fatalf("expected no agents directory, got err=%v", err)
	}
}

func TestInstallProviderCodexGlobalInstallsSkillContextAndHooks(t *testing.T) {
	tmp := t.TempDir()
	codexHome := filepath.Join(tmp, "codex-home")
	t.Setenv("CODEX_HOME", codexHome)
	t.Setenv("HOME", tmp)

	result, err := InstallProvider(InstallOptions{Provider: ProviderCodex})
	if err != nil {
		t.Fatalf("install codex provider: %v", err)
	}

	assertInstalledSharedSkill(t, filepath.Join(result.SkillPath, "SKILL.md"))
	assertFileContains(t, result.ContextPath, contextBlockBegin)
	assertFileContains(t, result.HookPath, sessionStartCommand)
	if result.ContextPath != filepath.Join(codexHome, "AGENTS.md") {
		t.Fatalf("unexpected context path: %s", result.ContextPath)
	}
	if result.HookPath != filepath.Join(codexHome, "hooks.json") {
		t.Fatalf("unexpected hook path: %s", result.HookPath)
	}
}

func TestInstallProviderClaudeLocalUsesSharedAgentsAndPointer(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", t.TempDir())

	result, err := InstallProvider(InstallOptions{
		Provider:  ProviderClaude,
		Local:     true,
		LocalRoot: root,
	})
	if err != nil {
		t.Fatalf("install claude provider locally: %v", err)
	}

	if result.ContextPath != filepath.Join(root, "AGENTS.md") {
		t.Fatalf("unexpected context path: %s", result.ContextPath)
	}
	if result.ClaudePointerPath != filepath.Join(root, "CLAUDE.md") {
		t.Fatalf("unexpected pointer path: %s", result.ClaudePointerPath)
	}
	assertFileContains(t, result.ContextPath, contextBlockBegin)
	assertFileContains(t, result.ClaudePointerPath, "See [AGENTS.md](./AGENTS.md)")
	assertFileContains(t, result.HookPath, sessionStartCommand)
	assertInstalledSharedSkill(t, filepath.Join(result.SkillPath, "SKILL.md"))
}

func TestInstallProviderLocalCodexAndClaudeShareOneContextBlock(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", t.TempDir())

	if _, err := InstallProvider(InstallOptions{Provider: ProviderCodex, Local: true, LocalRoot: root}); err != nil {
		t.Fatalf("install codex locally: %v", err)
	}
	if _, err := InstallProvider(InstallOptions{Provider: ProviderClaude, Local: true, LocalRoot: root}); err != nil {
		t.Fatalf("install claude locally: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	if count := strings.Count(string(content), contextBlockBegin); count != 1 {
		t.Fatalf("expected one managed context block, got %d", count)
	}
}

func TestInstallProviderClaudeLocalPreservesExistingClaudeFile(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", t.TempDir())

	claudePath := filepath.Join(root, "CLAUDE.md")
	seed := "# Existing Claude File\n\nDo not remove this.\n"
	if err := os.WriteFile(claudePath, []byte(seed), 0o644); err != nil {
		t.Fatalf("seed CLAUDE.md: %v", err)
	}

	result, err := InstallProvider(InstallOptions{
		Provider:  ProviderClaude,
		Local:     true,
		LocalRoot: root,
	})
	if err != nil {
		t.Fatalf("install claude provider locally: %v", err)
	}

	data, err := os.ReadFile(result.ClaudePointerPath)
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "# Existing Claude File") || !strings.Contains(content, "Do not remove this.") {
		t.Fatalf("existing CLAUDE.md content removed unexpectedly:\n%s", content)
	}
	if strings.Count(content, pointerBlockBegin) != 1 || strings.Count(content, pointerBlockEnd) != 1 {
		t.Fatalf("expected one managed pointer block:\n%s", content)
	}
}

func TestInstallProviderCodexLocalPreservesExistingAgentsFile(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", t.TempDir())

	agentsPath := filepath.Join(root, "AGENTS.md")
	seed := "# Existing Agents\n\nKeep these rules.\n"
	if err := os.WriteFile(agentsPath, []byte(seed), 0o644); err != nil {
		t.Fatalf("seed AGENTS.md: %v", err)
	}

	result, err := InstallProvider(InstallOptions{
		Provider:  ProviderCodex,
		Local:     true,
		LocalRoot: root,
	})
	if err != nil {
		t.Fatalf("install codex provider locally: %v", err)
	}

	data, err := os.ReadFile(result.ContextPath)
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "# Existing Agents") || !strings.Contains(content, "Keep these rules.") {
		t.Fatalf("existing AGENTS.md content removed unexpectedly:\n%s", content)
	}
	if strings.Count(content, contextBlockBegin) != 1 || strings.Count(content, contextBlockEnd) != 1 {
		t.Fatalf("expected one managed context block:\n%s", content)
	}
}

func TestInstallHookConfigAtPathMergesWithoutDuplication(t *testing.T) {
	path := filepath.Join(t.TempDir(), "hooks.json")
	existing := []byte(`{"hooks":{"SessionStart":[{"matcher":"startup","hooks":[{"type":"command","command":"echo existing"}]}]}}`)
	if err := os.WriteFile(path, existing, 0o644); err != nil {
		t.Fatalf("seed hooks: %v", err)
	}

	if _, err := InstallHookConfigAtPath(path); err != nil {
		t.Fatalf("first hook install: %v", err)
	}
	if _, err := InstallHookConfigAtPath(path); err != nil {
		t.Fatalf("second hook install: %v", err)
	}

	var config map[string]any
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read hooks: %v", err)
	}
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("parse hooks: %v", err)
	}
	text := string(data)
	if count := strings.Count(text, sessionStartCommand); count != 1 {
		t.Fatalf("expected one faz hook, got %d in %s", count, text)
	}
	if count := strings.Count(text, "echo existing"); count != 1 {
		t.Fatalf("expected existing hook preserved, got %d in %s", count, text)
	}
}

func TestInstallHookConfigAtPathReportsInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "hooks.json")
	if err := os.WriteFile(path, []byte("{invalid"), 0o644); err != nil {
		t.Fatalf("seed hooks: %v", err)
	}

	_, err := InstallHookConfigAtPath(path)
	if err == nil {
		t.Fatal("expected invalid JSON error")
	}
	if !strings.Contains(err.Error(), "parse current hook config") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertInstalledSharedSkill(t *testing.T, skillPath string) {
	t.Helper()

	installedSkill, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("read installed SKILL.md: %v", err)
	}
	expectedSkill, err := bundledFiles.ReadFile(bundledSkillPath)
	if err != nil {
		t.Fatalf("read bundled SKILL.md: %v", err)
	}
	if string(installedSkill) != string(expectedSkill) {
		t.Fatal("expected installed skill content to match bundled shared skill content")
	}
}

func assertFileContains(t *testing.T, path string, expected string) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !strings.Contains(string(content), expected) {
		t.Fatalf("expected %s to contain %q, got:\n%s", path, expected, content)
	}
}
