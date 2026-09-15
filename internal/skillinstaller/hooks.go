package skillinstaller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// InstallHookConfigAtPath installs the Faz hook, replacing a differing integration only with force.
func InstallHookConfigAtPath(path string, force bool) (string, error) {
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("read hook config %s: %w", path, err)
	}

	updated, action, err := upsertHookConfig(existing, force)
	if err != nil {
		return "", err
	}
	if bytes.Equal(existing, updated) {
		return action, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("create hook config directory %s: %w", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, updated, 0o644); err != nil {
		return "", fmt.Errorf("write hook config %s: %w", path, err)
	}
	return action, nil
}

// upsertHookConfig reconciles only the identified Faz SessionStart integration.
func upsertHookConfig(existing []byte, force bool) ([]byte, string, error) {
	current := make(map[string]any)
	if len(existing) > 0 {
		decoder := json.NewDecoder(bytes.NewReader(existing))
		decoder.UseNumber()
		if err := decoder.Decode(&current); err != nil {
			return nil, "", fmt.Errorf("parse current hook config: %w", err)
		}
		if err := decoder.Decode(new(any)); err != io.EOF {
			return nil, "", fmt.Errorf("parse current hook config: expected one JSON object")
		}
		if current == nil {
			return nil, "", fmt.Errorf("parse current hook config: expected an object, not null")
		}
	}

	hooks := make(map[string]any)
	if raw, exists := current["hooks"]; exists {
		var ok bool
		hooks, ok = raw.(map[string]any)
		if !ok || hooks == nil {
			return nil, "", fmt.Errorf("parse current hook config: hooks must be an object")
		}
	}
	var entries []any
	if raw, exists := hooks["SessionStart"]; exists {
		var ok bool
		entries, ok = raw.([]any)
		if !ok {
			return nil, "", fmt.Errorf("parse current hook config: SessionStart must be an array")
		}
	}

	expected := managedHookEntry()
	expectedHook := expected["hooks"].([]any)[0]
	matches, identical := 0, false
	var preserved []any
	for _, raw := range entries {
		entry, ok := raw.(map[string]any)
		if !ok {
			return nil, "", fmt.Errorf("parse current hook config: SessionStart entries must be objects")
		}
		commands, ok := entry["hooks"].([]any)
		if !ok {
			return nil, "", fmt.Errorf("parse current hook config: SessionStart entry hooks must be an array")
		}
		remaining := make([]any, 0, len(commands))
		for _, rawCommand := range commands {
			command, ok := rawCommand.(map[string]any)
			if !ok {
				return nil, "", fmt.Errorf("parse current hook config: hook entries must be objects")
			}
			if !isFazHook(command) {
				remaining = append(remaining, rawCommand)
				continue
			}
			matches++
			// Other commands in the same group do not change the Faz integration.
			identical = len(entry) == 2 && entry["matcher"] == expected["matcher"] && canonicalJSON(command) == canonicalJSON(expectedHook)
		}
		if len(remaining) == len(commands) {
			preserved = append(preserved, entry)
		} else if len(remaining) > 0 {
			entry["hooks"] = remaining
			preserved = append(preserved, entry)
		}
	}
	if matches == 1 && identical {
		return existing, "unchanged", nil
	}
	if matches > 0 && !force {
		return existing, "skipped (existing Faz hook differs; use --force to update)", nil
	}

	hooks["SessionStart"] = append(preserved, expected)
	current["hooks"] = hooks
	after, err := marshalIndent(current)
	if err != nil {
		return nil, "", fmt.Errorf("marshal hook config: %w", err)
	}
	action := "created"
	if matches > 0 {
		action = "updated"
	}
	return after, action, nil
}

// managedHookEntry returns the intended Faz SessionStart group.
func managedHookEntry() map[string]any {
	return map[string]any{
		"matcher": "startup|resume|clear|compact",
		"hooks": []any{
			map[string]any{
				"type":          "command",
				"command":       sessionStartCommand,
				"statusMessage": "Loading faz task context",
				"timeout":       5,
			},
		},
	}
}

// isFazHook recognizes the installer label and known commands without matching arbitrary mentions of Faz.
func isFazHook(hook map[string]any) bool {
	if hook["type"] != "command" {
		return false
	}
	command, _ := hook["command"].(string)
	return hook["statusMessage"] == "Loading faz task context" || command == sessionStartCommand ||
		command == "faz init && faz onboard" || command == "faz onboard"
}

// canonicalJSON returns a stable JSON string for duplicate detection.
func canonicalJSON(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(encoded)
}

// marshalIndent encodes JSON without escaping shell operators.
func marshalIndent(value any) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
