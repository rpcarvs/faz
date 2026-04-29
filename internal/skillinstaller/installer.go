package skillinstaller

import (
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const skillDirName = "task-management-with-faz"
const bundledSkillPath = "bundled/task-management-with-faz/SKILL.md"

// bundledFiles contains built-in skill files to install for supported tools.
//
//go:embed bundled/task-management-with-faz/SKILL.md
var bundledFiles embed.FS

// InstallCodexSkill installs the bundled faz skill into Codex skills directory.
func InstallCodexSkill(force bool) (string, error) {
	root, err := codexSkillsRoot()
	if err != nil {
		return "", err
	}

	files := map[string]string{
		"SKILL.md": bundledSkillPath,
	}

	return installBundledSkill(root, files, force)
}

// InstallClaudeSkill installs the bundled faz skill into Claude skills directory.
func InstallClaudeSkill(force bool) (string, error) {
	root, err := claudeSkillsRoot()
	if err != nil {
		return "", err
	}

	files := map[string]string{
		"SKILL.md": bundledSkillPath,
	}

	return installBundledSkill(root, files, force)
}

// codexSkillsRoot resolves the target Codex skills root directory.
func codexSkillsRoot() (string, error) {
	codexHome := os.Getenv("CODEX_HOME")
	if codexHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		codexHome = filepath.Join(home, ".codex")
	}
	return filepath.Join(codexHome, "skills"), nil
}

// claudeSkillsRoot resolves the target Claude skills root directory.
func claudeSkillsRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".claude", "skills"), nil
}

// installBundledSkill writes embedded files into the destination skill directory.
func installBundledSkill(root string, files map[string]string, force bool) (string, error) {
	target := filepath.Join(root, skillDirName)
	if err := ensureTarget(target, force); err != nil {
		return "", err
	}

	for relPath, bundledPath := range files {
		content, err := bundledFiles.ReadFile(bundledPath)
		if err != nil {
			return "", fmt.Errorf("read bundled file %s: %w", bundledPath, err)
		}

		filePath := filepath.Join(target, relPath)
		if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
			return "", fmt.Errorf("create directory %s: %w", filepath.Dir(filePath), err)
		}
		if err := os.WriteFile(filePath, content, 0o644); err != nil {
			return "", fmt.Errorf("write file %s: %w", filePath, err)
		}
	}

	return target, nil
}

// ensureTarget verifies destination state and removes existing data when forced.
func ensureTarget(target string, force bool) error {
	_, err := os.Stat(target)
	if err == nil {
		if !force {
			return fmt.Errorf("skill already exists at %s (use --force to overwrite)", target)
		}
		if err := os.RemoveAll(target); err != nil {
			return fmt.Errorf("remove existing skill at %s: %w", target, err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("check target %s: %w", target, err)
	}

	if err := os.MkdirAll(target, 0o755); err != nil {
		return fmt.Errorf("create skill directory %s: %w", target, err)
	}
	return nil
}
