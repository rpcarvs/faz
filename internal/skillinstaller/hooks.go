package skillinstaller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// InstallHookConfigAtPath merges the faz SessionStart hook into a JSON config.
func InstallHookConfigAtPath(path string) (string, error) {
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("read hook config %s: %w", path, err)
	}

	updated, action, err := upsertHookConfig(existing, hookConfigJSON())
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("create hook config directory %s: %w", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, updated, 0o644); err != nil {
		return "", fmt.Errorf("write hook config %s: %w", path, err)
	}
	return action, nil
}

// upsertHookConfig merges managed hooks with existing hook JSON without duplicates.
func upsertHookConfig(existing []byte, managed []byte) ([]byte, string, error) {
	current := make(map[string]any)
	if len(existing) > 0 {
		if err := json.Unmarshal(existing, &current); err != nil {
			return nil, "", fmt.Errorf("parse current hook config: %w", err)
		}
	}

	var expected map[string]any
	if err := json.Unmarshal(managed, &expected); err != nil {
		return nil, "", fmt.Errorf("parse managed hook config: %w", err)
	}

	before := canonicalJSON(current)
	mergeHookMaps(current, expected)
	after, err := marshalIndent(current)
	if err != nil {
		return nil, "", fmt.Errorf("marshal hook config: %w", err)
	}

	action := "updated"
	if len(existing) == 0 {
		action = "created"
	} else if before == canonicalJSON(current) {
		action = "unchanged"
	}
	return after, action, nil
}

// hookConfigJSON returns the managed SessionStart hook configuration.
func hookConfigJSON() []byte {
	config := map[string]any{
		"hooks": map[string]any{
			"SessionStart": []any{
				map[string]any{
					"matcher": "startup|resume|clear|compact",
					"hooks": []any{
						map[string]any{
							"type":          "command",
							"command":       sessionStartCommand,
							"statusMessage": "Loading faz task context",
							"timeout":       5,
						},
					},
				},
			},
		},
	}
	encoded, _ := json.Marshal(config)
	return encoded
}

// mergeHookMaps merges hook arrays while preserving unrelated settings.
func mergeHookMaps(dst map[string]any, src map[string]any) {
	for key, value := range src {
		if key != "hooks" {
			if _, exists := dst[key]; !exists {
				dst[key] = value
			}
			continue
		}

		srcHooks, ok := value.(map[string]any)
		if !ok {
			continue
		}
		dstHooks, ok := dst[key].(map[string]any)
		if !ok || dstHooks == nil {
			dstHooks = make(map[string]any, len(srcHooks))
			dst[key] = dstHooks
		}

		for hookName, rawEntries := range srcHooks {
			entries, ok := rawEntries.([]any)
			if !ok {
				continue
			}
			existing, _ := dstHooks[hookName].([]any)
			dstHooks[hookName] = mergeJSONArray(existing, entries)
		}
	}
}

// mergeJSONArray appends managed JSON entries that are not already present.
func mergeJSONArray(existing []any, managed []any) []any {
	merged := append([]any(nil), existing...)
	seen := make(map[string]struct{}, len(existing))
	for _, entry := range existing {
		seen[canonicalJSON(entry)] = struct{}{}
	}
	for _, entry := range managed {
		key := canonicalJSON(entry)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		merged = append(merged, entry)
	}
	return merged
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
