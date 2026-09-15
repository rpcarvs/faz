package skillinstaller

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestHookInstallationReconcilesOnlyFaz covers legacy metadata and grouped hooks.
func TestHookInstallationReconcilesOnlyFaz(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(map[string]any)
	}{
		{"timeout", func(entry map[string]any) { entry["hooks"].([]any)[0].(map[string]any)["timeout"] = 10 }},
		{"matcher", func(entry map[string]any) { entry["matcher"] = "startup" }},
		{"command", func(entry map[string]any) {
			entry["hooks"].([]any)[0].(map[string]any)["command"] = "faz onboard --custom"
		}},
		{"missing label", func(entry map[string]any) { delete(entry["hooks"].([]any)[0].(map[string]any), "statusMessage") }},
		{"group metadata", func(entry map[string]any) { entry["custom"] = true }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			entry := managedHookEntry()
			tc.change(entry)
			otherCommand := map[string]any{"type": "command", "command": "echo keep", "timeout": 30}
			entry["hooks"] = append(entry["hooks"].([]any), otherCommand)
			originalGroup, err := marshalIndent(entry)
			if err != nil {
				t.Fatal(err)
			}
			config := hookTestConfig(entry)
			original, err := marshalIndent(config)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), "hooks.json")
			if err := os.WriteFile(path, original, 0o600); err != nil {
				t.Fatal(err)
			}
			assertHookNoWrite(t, path, false, "skipped")
			action, err := InstallHookConfigAtPath(path, true)
			if err != nil || action != "updated" {
				t.Fatalf("force: action=%q error=%v", action, err)
			}
			updated := readHookTestConfig(t, path)
			groups := updated["hooks"].(map[string]any)["SessionStart"].([]any)
			if len(groups) != 2 {
				t.Fatalf("expected preserved group plus Faz group: %v", groups)
			}
			var expectedGroup map[string]any
			if err := json.Unmarshal(originalGroup, &expectedGroup); err != nil {
				t.Fatal(err)
			}
			expectedGroup["hooks"] = []any{otherCommand}
			if canonicalJSON(groups[0]) != canonicalJSON(expectedGroup) || canonicalJSON(groups[1]) != canonicalJSON(managedHookEntry()) {
				t.Fatalf("unexpected groups after force: %v", groups)
			}
			assertUnrelatedHookSettings(t, config, updated)
			assertHookNoWrite(t, path, true, "unchanged")
		})
	}
}

// TestHookInstallationTimeoutDuplicates reproduces the reported Faz/treelines config.
func TestHookInstallationTimeoutDuplicates(t *testing.T) {
	old := managedHookEntry()
	old["hooks"].([]any)[0].(map[string]any)["timeout"] = 10
	treelines := managedHookEntry()
	treelinesHook := treelines["hooks"].([]any)[0].(map[string]any)
	treelinesHook["command"] = "git rev-parse --show-toplevel >/dev/null 2>&1 && treelines init && treelines onboard"
	treelinesHook["statusMessage"] = "Loading treelines codebase context"
	treelinesHook["timeout"] = 10
	config := hookTestConfig(old, treelines, managedHookEntry())
	data, err := marshalIndent(config)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "hooks.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	assertHookNoWrite(t, path, false, "skipped")
	if action, err := InstallHookConfigAtPath(path, true); err != nil || action != "updated" {
		t.Fatalf("force: action=%q error=%v", action, err)
	}
	updated := readHookTestConfig(t, path)
	groups := updated["hooks"].(map[string]any)["SessionStart"].([]any)
	if len(groups) != 2 || canonicalJSON(groups[0]) != canonicalJSON(treelines) || canonicalJSON(groups[1]) != canonicalJSON(managedHookEntry()) {
		t.Fatalf("expected preserved treelines hook and one current Faz hook: %v", groups)
	}
	assertUnrelatedHookSettings(t, config, updated)
	assertHookNoWrite(t, path, false, "unchanged")
}

// TestIdenticalHookLeavesBytesUntouched checks formatting, timestamps, and grouped commands.
func TestIdenticalHookLeavesBytesUntouched(t *testing.T) {
	entry := managedHookEntry()
	entry["hooks"] = append(entry["hooks"].([]any), map[string]any{"type": "command", "command": "echo keep"})
	path := filepath.Join(t.TempDir(), "hooks.json")
	if err := os.WriteFile(path, []byte(canonicalJSON(hookTestConfig(entry))), 0o600); err != nil {
		t.Fatal(err)
	}
	assertHookNoWrite(t, path, false, "unchanged")
	assertHookNoWrite(t, path, true, "unchanged")
}

// TestHookInstallationRejectsInvalidStructure prevents silent loss of malformed settings.
func TestHookInstallationRejectsInvalidStructure(t *testing.T) {
	for _, content := range []string{
		`null`, `[]`, `{`, `{} {}`, `{"hooks":null}`, `{"hooks":[]}`,
		`{"hooks":{"SessionStart":{}}}`, `{"hooks":{"SessionStart":null}}`,
		`{"hooks":{"SessionStart":[null]}}`, `{"hooks":{"SessionStart":[{"hooks":{}}]}}`,
		`{"hooks":{"SessionStart":[{"hooks":[null]}]}}`,
	} {
		t.Run(content, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "hooks.json")
			if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
				t.Fatal(err)
			}
			for _, force := range []bool{false, true} {
				if _, err := InstallHookConfigAtPath(path, force); err == nil {
					t.Fatal("expected invalid configuration error")
				}
				data, err := os.ReadFile(path)
				if err != nil || string(data) != content {
					t.Fatalf("invalid config was changed: %s, %v", data, err)
				}
			}
		})
	}
}

// TestFazHookIdentity avoids treating unrelated commands mentioning Faz as managed hooks.
func TestFazHookIdentity(t *testing.T) {
	for _, command := range []string{sessionStartCommand, "faz init && faz onboard", "faz onboard"} {
		if !isFazHook(map[string]any{"type": "command", "command": command}) {
			t.Fatalf("missing known Faz command %q", command)
		}
	}
	for _, command := range []string{"echo faz onboard", "treelines onboard", "faz list", "my-faz onboard"} {
		if isFazHook(map[string]any{"type": "command", "command": command}) {
			t.Fatalf("unexpected managed command %q", command)
		}
	}
}

// hookTestConfig adds unrelated settings whose values must survive installation.
func hookTestConfig(entries ...any) map[string]any {
	return map[string]any{
		"custom": json.Number("9007199254740993"),
		"hooks": map[string]any{
			"SessionStart": entries,
			"Stop":         []any{map[string]any{"hooks": []any{map[string]any{"type": "command", "command": "echo stopped"}}}},
		},
	}
}

// readHookTestConfig decodes JSON without rounding unrelated numeric settings.
func readHookTestConfig(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.UseNumber()
	var config map[string]any
	if err := decoder.Decode(&config); err != nil {
		t.Fatal(err)
	}
	return config
}

// assertUnrelatedHookSettings verifies preservation of other events and root settings.
func assertUnrelatedHookSettings(t *testing.T, original, updated map[string]any) {
	t.Helper()
	if canonicalJSON(original["custom"]) != canonicalJSON(updated["custom"]) ||
		canonicalJSON(original["hooks"].(map[string]any)["Stop"]) != canonicalJSON(updated["hooks"].(map[string]any)["Stop"]) {
		t.Fatalf("unrelated settings changed: %v", updated)
	}
}

// assertHookNoWrite checks that no-op installation preserves exact bytes and file metadata.
func assertHookNoWrite(t *testing.T, path string, force bool, expectedAction string) {
	t.Helper()
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	stamp := time.Unix(1000000000, 0)
	if err := os.Chtimes(path, stamp, stamp); err != nil {
		t.Fatal(err)
	}
	action, err := InstallHookConfigAtPath(path, force)
	if err != nil || !strings.HasPrefix(action, expectedAction) {
		t.Fatalf("action=%q error=%v; expected %s", action, err, expectedAction)
	}
	after, err := os.ReadFile(path)
	if err != nil || string(after) != string(before) {
		t.Fatalf("no-op changed bytes: %s, %v", after, err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !info.ModTime().Equal(stamp) || info.Mode().Perm() != 0o600 {
		t.Fatalf("no-op changed metadata: %v", info)
	}
}
