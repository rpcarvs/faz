package skillinstaller

import (
	"encoding/json"
	"fmt"
	"io/fs"
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

	assertInstalledBundles(t, ProviderCodex, result)
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
	assertInstalledBundles(t, ProviderClaude, result)
}

// TestInstallProviderInstallsBundlesForEveryProviderScope covers supported destinations.
func TestInstallProviderInstallsBundlesForEveryProviderScope(t *testing.T) {
	for _, testCase := range []struct {
		name     string
		provider Provider
		local    bool
	}{
		{name: "codex global", provider: ProviderCodex},
		{name: "codex local", provider: ProviderCodex, local: true},
		{name: "claude global", provider: ProviderClaude},
		{name: "claude local", provider: ProviderClaude, local: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			tmp := t.TempDir()
			t.Setenv("HOME", tmp)
			t.Setenv("CODEX_HOME", filepath.Join(tmp, "codex-home"))
			options := InstallOptions{Provider: testCase.provider, Local: testCase.local}
			if testCase.local {
				options.LocalRoot = filepath.Join(tmp, "project")
			}

			result, err := InstallProvider(options)
			if err != nil {
				t.Fatalf("install provider: %v", err)
			}
			assertInstalledBundles(t, testCase.provider, result)
			assertExplicitInvocationMetadata(t, testCase.provider, result)
			if filepath.Base(result.SkillPath) != "faz-task-management" {
				t.Fatalf("unexpected task skill directory: %s", result.SkillPath)
			}
			taskSkill, err := os.ReadFile(filepath.Join(result.SkillPath, "SKILL.md"))
			if err != nil {
				t.Fatalf("read task skill: %v", err)
			}
			metadata := frontmatter(t, string(taskSkill))
			if !strings.Contains(metadata, "\nname: faz-task-management\n") {
				t.Fatalf("unexpected task skill name:\n%s", metadata)
			}
			if strings.Contains(metadata, "disable-model-invocation:") {
				t.Fatalf("normal task skill must retain automatic invocation:\n%s", metadata)
			}
			assertFileContains(t, result.ContextPath, "`faz-task-management` SKILL")
		})
	}
}

// TestInstallProviderRerunPreservesCustomFilesAndForceReplacesBundles verifies rerun behavior.
func TestInstallProviderRerunPreservesCustomFilesAndForceReplacesBundles(t *testing.T) {
	for _, provider := range []Provider{ProviderCodex, ProviderClaude} {
		for _, local := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/local=%t", provider, local), func(t *testing.T) {
				root := t.TempDir()
				t.Setenv("HOME", root)
				t.Setenv("CODEX_HOME", filepath.Join(root, "codex-home"))
				options := InstallOptions{Provider: provider, Local: local, LocalRoot: filepath.Join(root, "project")}
				result, err := InstallProvider(options)
				if err != nil {
					t.Fatalf("first install: %v", err)
				}
				for _, path := range result.SkillPaths {
					if err := os.WriteFile(filepath.Join(path, "custom.txt"), []byte("custom"), 0o644); err != nil {
						t.Fatalf("write custom skill file: %v", err)
					}
					if err := os.WriteFile(filepath.Join(path, "SKILL.md"), []byte("stale"), 0o644); err != nil {
						t.Fatalf("change installed skill: %v", err)
					}
				}
				if err := os.WriteFile(filepath.Join(result.SkillPaths[1], "assets", "SPECS.md"), []byte("stale"), 0o644); err != nil {
					t.Fatalf("change installed asset: %v", err)
				}

				result, err = InstallProvider(options)
				if err != nil {
					t.Fatalf("rerun install: %v", err)
				}
				for _, path := range result.SkillPaths {
					content, err := os.ReadFile(filepath.Join(path, "custom.txt"))
					if err != nil || string(content) != "custom" {
						t.Fatalf("rerun changed custom file in %s: %v", path, err)
					}
				}
				assertInstalledBundles(t, provider, result)
				assertExplicitInvocationMetadata(t, provider, result)

				options.Force = true
				result, err = InstallProvider(options)
				if err != nil {
					t.Fatalf("force install: %v", err)
				}
				for _, path := range result.SkillPaths {
					if _, err := os.Stat(filepath.Join(path, "custom.txt")); !os.IsNotExist(err) {
						t.Fatalf("force install retained custom file in %s: %v", path, err)
					}
				}
				assertInstalledBundles(t, provider, result)
				assertExplicitInvocationMetadata(t, provider, result)
			})
		}
	}
}

// TestInstallProviderRejectsSymlinkedSkillTarget prevents target directory traversal.
func TestInstallProviderRejectsSymlinkedSkillTarget(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, ".codex", "skills", specDrivenSkillDirName)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatalf("create skill root: %v", err)
	}
	if err := os.Symlink(t.TempDir(), target); err != nil {
		t.Fatalf("create symlink target: %v", err)
	}

	_, err := InstallProvider(InstallOptions{Provider: ProviderCodex, Local: true, LocalRoot: root})
	if err == nil || !strings.Contains(err.Error(), "must not be a symlink") {
		t.Fatalf("expected symlink rejection, got %v", err)
	}
}

// TestInstallProviderRejectsNestedSymlinks preserves files outside the skill target.
func TestInstallProviderRejectsNestedSymlinks(t *testing.T) {
	for _, testCase := range []struct {
		name     string
		prepare  func(t *testing.T, target, outside string)
		expected string
	}{
		{
			name: "asset directory",
			prepare: func(t *testing.T, target, outside string) {
				t.Helper()
				assetsPath := filepath.Join(target, "assets")
				if err := os.RemoveAll(assetsPath); err != nil {
					t.Fatalf("remove installed assets: %v", err)
				}
				if err := os.Symlink(outside, assetsPath); err != nil {
					t.Fatalf("create assets symlink: %v", err)
				}
			},
			expected: "must be a directory",
		},
		{
			name: "skill file",
			prepare: func(t *testing.T, target, outside string) {
				t.Helper()
				skillPath := filepath.Join(target, "SKILL.md")
				if err := os.Remove(skillPath); err != nil {
					t.Fatalf("remove installed skill: %v", err)
				}
				if err := os.Symlink(filepath.Join(outside, "protected.txt"), skillPath); err != nil {
					t.Fatalf("create skill symlink: %v", err)
				}
			},
			expected: "must not be a symlink",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			root := t.TempDir()
			outside := t.TempDir()
			protectedPath := filepath.Join(outside, "protected.txt")
			if err := os.WriteFile(protectedPath, []byte("protected"), 0o644); err != nil {
				t.Fatalf("write protected file: %v", err)
			}

			options := InstallOptions{Provider: ProviderCodex, Local: true, LocalRoot: root}
			result, err := InstallProvider(options)
			if err != nil {
				t.Fatalf("first install: %v", err)
			}
			testCase.prepare(t, result.SkillPaths[1], outside)

			_, err = InstallProvider(options)
			if err == nil || !strings.Contains(err.Error(), testCase.expected) {
				t.Fatalf("expected nested symlink rejection, got %v", err)
			}
			protected, err := os.ReadFile(protectedPath)
			if err != nil {
				t.Fatalf("read protected file: %v", err)
			}
			if string(protected) != "protected" {
				t.Fatalf("outside file changed to %q", protected)
			}
		})
	}
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

	if _, err := InstallHookConfigAtPath(path, false); err != nil {
		t.Fatalf("first hook install: %v", err)
	}
	if _, err := InstallHookConfigAtPath(path, false); err != nil {
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

	_, err := InstallHookConfigAtPath(path, false)
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

func assertInstalledBundles(t *testing.T, provider Provider, result InstallResult) {
	t.Helper()
	if len(result.SkillPaths) != len(bundledSkills) {
		t.Fatalf("expected %d installed skill paths, got %d", len(bundledSkills), len(result.SkillPaths))
	}
	if result.SkillPath != result.SkillPaths[0] {
		t.Fatalf("compatibility skill path %s does not match task skill %s", result.SkillPath, result.SkillPaths[0])
	}
	for i, skill := range bundledSkills {
		target := filepath.Join(filepath.Dir(result.SkillPath), skill.directory)
		if result.SkillPaths[i] != target {
			t.Fatalf("unexpected installed path for %s: %s", skill.directory, result.SkillPaths[i])
		}
		assertInstalledBundle(t, provider, skill, target)
	}
}

func assertInstalledBundle(t *testing.T, provider Provider, skill bundledSkill, target string) {
	t.Helper()
	err := fs.WalkDir(bundledFiles, skill.root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !shouldInstallBundledFile(provider, skill, path) {
			return nil
		}
		relativePath, err := filepath.Rel(skill.root, path)
		if err != nil {
			return err
		}
		expected, err := bundledFiles.ReadFile(path)
		if err != nil {
			return err
		}
		if provider == ProviderClaude && skill.explicitOnly && relativePath == "SKILL.md" {
			expected, err = claudeSkillContent(expected)
			if err != nil {
				return err
			}
		}
		actual, err := os.ReadFile(filepath.Join(target, filepath.FromSlash(relativePath)))
		if err != nil {
			return err
		}
		if string(actual) != string(expected) {
			return fmt.Errorf("installed file %s does not match bundle", relativePath)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("verify installed %s bundle: %v", skill.directory, err)
	}

	agentsPath := filepath.Join(target, "agents")
	if provider == ProviderClaude {
		if _, err := os.Stat(agentsPath); !os.IsNotExist(err) {
			t.Fatalf("expected no Codex metadata directory for Claude, got %v", err)
		}
	}
}

// assertExplicitInvocationMetadata verifies each provider's independent policy.
func assertExplicitInvocationMetadata(t *testing.T, provider Provider, result InstallResult) {
	t.Helper()
	if len(result.SkillPaths) != 3 {
		t.Fatalf("expected three installed skills, got %d", len(result.SkillPaths))
	}
	sddPath := result.SkillPaths[1]
	if filepath.Base(sddPath) != "faz-spec-driven" {
		t.Fatalf("unexpected SDD skill directory: %s", sddPath)
	}
	for _, name := range []string{"SPECS.md", "ARCHITECTURE.md", "PLAN.md"} {
		content, err := os.ReadFile(filepath.Join(sddPath, "assets", name))
		if err != nil {
			t.Fatalf("read installed template %s: %v", name, err)
		}
		if len(content) == 0 {
			t.Fatalf("installed template %s is empty", name)
		}
	}
	for i, name := range []string{"faz-spec-driven", "faz-orchestration"} {
		path := result.SkillPaths[i+1]
		if filepath.Base(path) != name {
			t.Fatalf("unexpected skill directory for %s: %s", name, path)
		}
		assertExplicitSkillMetadata(t, provider, path, name)
	}
}

// assertExplicitSkillMetadata checks installed metadata without relying on the bundle policy flag.
func assertExplicitSkillMetadata(t *testing.T, provider Provider, path, name string) {
	t.Helper()
	skill, err := os.ReadFile(filepath.Join(path, "SKILL.md"))
	if err != nil {
		t.Fatalf("read %s skill: %v", name, err)
	}
	frontmatter := frontmatter(t, string(skill))
	if !strings.Contains(frontmatter, "\nname: "+name+"\n") {
		t.Fatalf("unexpected %s skill name:\n%s", name, frontmatter)
	}
	if provider == ProviderClaude {
		if count := strings.Count(frontmatter, "disable-model-invocation: true\n"); count != 1 {
			t.Fatalf("expected one Claude explicit invocation setting, got %d in:\n%s", count, frontmatter)
		}
		return
	}
	if strings.Contains(frontmatter, "disable-model-invocation:") {
		t.Fatalf("Codex %s skill unexpectedly includes Claude metadata:\n%s", name, frontmatter)
	}
	metadata, err := os.ReadFile(filepath.Join(path, "agents", "openai.yaml"))
	if err != nil {
		t.Fatalf("read Codex metadata: %v", err)
	}
	if !strings.Contains(string(metadata), "allow_implicit_invocation: false") {
		t.Fatalf("Codex metadata does not disable implicit invocation:\n%s", metadata)
	}
}

// frontmatter returns the YAML metadata block from a bundled skill.
func frontmatter(t *testing.T, content string) string {
	t.Helper()
	if !strings.HasPrefix(content, "---\n") {
		t.Fatalf("skill is missing opening frontmatter boundary:\n%s", content)
	}
	end := strings.Index(content[len("---\n"):], "---\n")
	if end < 0 {
		t.Fatalf("skill is missing closing frontmatter boundary:\n%s", content)
	}
	return content[:len("---\n")+end+len("---\n")]
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
