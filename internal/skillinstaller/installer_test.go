package skillinstaller

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallCodexSkillCreatesFiles(t *testing.T) {
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
	openAIPath := filepath.Join(installedPath, "agents", "openai.yaml")

	if _, err := os.Stat(skillPath); err != nil {
		t.Fatalf("missing SKILL.md: %v", err)
	}
	if _, err := os.Stat(openAIPath); err != nil {
		t.Fatalf("missing openai.yaml: %v", err)
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

func TestInstallCodexSkillExistingWithoutForceFails(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("CODEX_HOME", filepath.Join(tmp, "codex-home"))
	t.Setenv("HOME", tmp)

	if _, err := InstallCodexSkill(false); err != nil {
		t.Fatalf("first install should succeed: %v", err)
	}

	_, err := InstallCodexSkill(false)
	if err == nil {
		t.Fatal("expected error when skill already exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected already exists error, got: %v", err)
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
	if _, err := os.Stat(filepath.Join(installedPath, "agents")); !os.IsNotExist(err) {
		t.Fatalf("expected no agents directory, got err=%v", err)
	}
}
