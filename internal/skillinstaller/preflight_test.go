package skillinstaller

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// TestInvalidHooksPreventAllInstallWrites covers both scopes, providers and force modes.
func TestInvalidHooksPreventAllInstallWrites(t *testing.T) {
	for _, provider := range []Provider{ProviderCodex, ProviderClaude} {
		for _, local := range []bool{false, true} {
			for _, force := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/local=%t/force=%t", provider, local, force), func(t *testing.T) {
					root := t.TempDir()
					t.Setenv("HOME", root)
					t.Setenv("CODEX_HOME", filepath.Join(root, "codex-home"))
					options := InstallOptions{Provider: provider, Local: local, LocalRoot: filepath.Join(root, "project")}
					result, err := InstallProvider(options)
					if err != nil {
						t.Fatal(err)
					}
					for path, content := range map[string]string{
						filepath.Join(result.SkillPaths[0], "SKILL.md"):   "locally edited skill",
						filepath.Join(result.SkillPaths[1], "custom.txt"): "preserve custom file",
						result.ContextPath: "user context",
					} {
						if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
							t.Fatal(err)
						}
					}
					options.Force = force
					for _, invalid := range []string{`{invalid`, `{"hooks":{"SessionStart":{}}}`} {
						if err := os.WriteFile(result.HookPath, []byte(invalid), 0o600); err != nil {
							t.Fatal(err)
						}
						before := snapshotInstallTree(t, root)
						if _, err := InstallProvider(options); err == nil {
							t.Fatal("expected invalid hook configuration error")
						}
						if after := snapshotInstallTree(t, root); !reflect.DeepEqual(before, after) {
							t.Fatal("invalid hook configuration allowed installation writes")
						}
					}
				})
			}
		}
	}
}

// TestInvalidInstallTargetsPreventEarlierWrites checks destinations reached late in installation.
func TestInvalidInstallTargetsPreventEarlierWrites(t *testing.T) {
	for _, target := range []string{"context", "pointer", "bundle file", "bundle symlink"} {
		t.Run(target, func(t *testing.T) {
			root := t.TempDir()
			options := InstallOptions{Provider: ProviderClaude, Local: true, LocalRoot: root}
			result, err := InstallProvider(options)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(result.SkillPaths[0], "SKILL.md"), []byte("do not overwrite"), 0o600); err != nil {
				t.Fatal(err)
			}
			path := result.ContextPath
			switch target {
			case "pointer":
				path = result.ClaudePointerPath
			case "bundle file", "bundle symlink":
				path = filepath.Join(result.SkillPaths[1], "SKILL.md")
			}
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			if target == "bundle symlink" {
				if err := os.Symlink(result.ContextPath, path); err != nil {
					t.Fatal(err)
				}
			} else if err := os.Mkdir(path, 0o700); err != nil {
				t.Fatal(err)
			}
			before := snapshotInstallTree(t, root)
			if _, err := InstallProvider(options); err == nil {
				t.Fatal("expected invalid target error")
			}
			if after := snapshotInstallTree(t, root); !reflect.DeepEqual(before, after) {
				t.Fatal("invalid target allowed earlier installation writes")
			}
		})
	}
}

// snapshotInstallTree records file contents, metadata, directories and symlinks without following links.
func snapshotInstallTree(t *testing.T, root string) map[string]string {
	t.Helper()
	snapshot := make(map[string]string)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		var content string
		switch {
		case info.Mode()&os.ModeSymlink != 0:
			content, err = os.Readlink(path)
		case !entry.IsDir():
			var data []byte
			data, err = os.ReadFile(path)
			content = string(data)
		}
		if err != nil {
			return err
		}
		snapshot[path] = fmt.Sprintf("%v|%d|%s", info.Mode(), info.ModTime().UnixNano(), content)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return snapshot
}
