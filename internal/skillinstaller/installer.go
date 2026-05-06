package skillinstaller

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

const skillDirName = "task-management-with-faz"
const bundledSkillPath = "bundled/task-management-with-faz/SKILL.md"
const sessionStartCommand = "git rev-parse --show-toplevel >/dev/null 2>&1 && faz init && faz onboard"

// bundledFiles contains built-in skill files to install for supported tools.
//
//go:embed bundled/task-management-with-faz/SKILL.md
var bundledFiles embed.FS

type Provider string

const (
	ProviderCodex  Provider = "codex"
	ProviderClaude Provider = "claude"
)

// InstallOptions configures a provider-level agent installation.
type InstallOptions struct {
	Provider  Provider
	Local     bool
	LocalRoot string
	Force     bool
}

// InstallResult reports all paths touched by a provider install.
type InstallResult struct {
	SkillPath           string
	ContextPath         string
	ContextAction       string
	HookPath            string
	HookAction          string
	CodexConfigPath     string
	CodexConfigAction   string
	ClaudePointerPath   string
	ClaudePointerAction string
}

// InstallProvider installs skill, context, and hooks for one supported agent.
func InstallProvider(options InstallOptions) (InstallResult, error) {
	if err := validateInstallOptions(options); err != nil {
		return InstallResult{}, err
	}

	skillRoot, err := skillsRoot(options)
	if err != nil {
		return InstallResult{}, err
	}
	contextPath, err := contextPath(options)
	if err != nil {
		return InstallResult{}, err
	}
	hookPath, err := hookConfigPath(options)
	if err != nil {
		return InstallResult{}, err
	}

	skillPath, err := installBundledSkill(skillRoot, options.Force)
	if err != nil {
		return InstallResult{}, err
	}
	contextAction, err := InstallContextAtPath(contextPath)
	if err != nil {
		return InstallResult{}, err
	}
	hookAction, err := InstallHookConfigAtPath(hookPath)
	if err != nil {
		return InstallResult{}, err
	}

	result := InstallResult{
		SkillPath:     skillPath,
		ContextPath:   contextPath,
		ContextAction: contextAction,
		HookPath:      hookPath,
		HookAction:    hookAction,
	}

	if options.Provider == ProviderCodex {
		configPath, action, err := EnsureCodexHooksEnabled(codexConfigPath(options))
		if err != nil {
			return InstallResult{}, err
		}
		result.CodexConfigPath = configPath
		result.CodexConfigAction = action
	}

	if options.Provider == ProviderClaude && options.Local {
		pointerPath := filepath.Join(options.LocalRoot, "CLAUDE.md")
		action, err := InstallClaudePointerAtPath(pointerPath)
		if err != nil {
			return InstallResult{}, err
		}
		result.ClaudePointerPath = pointerPath
		result.ClaudePointerAction = action
	}

	return result, nil
}

// InstallCodexSkill installs the bundled faz skill into Codex skills directory.
func InstallCodexSkill(force bool) (string, error) {
	root, err := codexSkillsRoot()
	if err != nil {
		return "", err
	}
	return installBundledSkill(root, force)
}

// InstallClaudeSkill installs the bundled faz skill into Claude skills directory.
func InstallClaudeSkill(force bool) (string, error) {
	root, err := claudeSkillsRoot()
	if err != nil {
		return "", err
	}
	return installBundledSkill(root, force)
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

// installBundledSkill writes embedded skill files into the destination directory.
func installBundledSkill(root string, force bool) (string, error) {
	target := filepath.Join(root, skillDirName)
	if err := ensureSkillTarget(target, force); err != nil {
		return "", err
	}

	content, err := bundledFiles.ReadFile(bundledSkillPath)
	if err != nil {
		return "", fmt.Errorf("read bundled skill: %w", err)
	}
	if err := os.WriteFile(filepath.Join(target, "SKILL.md"), content, 0o644); err != nil {
		return "", fmt.Errorf("write bundled skill: %w", err)
	}

	return target, nil
}

// ensureSkillTarget creates the skill directory and optionally clears it first.
func ensureSkillTarget(target string, force bool) error {
	_, err := os.Stat(target)
	if err == nil && force {
		if err := os.RemoveAll(target); err != nil {
			return fmt.Errorf("remove existing skill at %s: %w", target, err)
		}
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("check target %s: %w", target, err)
	}

	if err := os.MkdirAll(target, 0o755); err != nil {
		return fmt.Errorf("create skill directory %s: %w", target, err)
	}
	return nil
}

// validateInstallOptions verifies provider and scope before writing files.
func validateInstallOptions(options InstallOptions) error {
	switch options.Provider {
	case ProviderCodex, ProviderClaude:
	default:
		return fmt.Errorf("unsupported install provider %q", options.Provider)
	}
	if options.Local && options.LocalRoot == "" {
		return fmt.Errorf("local install requires repository root")
	}
	return nil
}

// skillsRoot resolves the target skill root for a provider install.
func skillsRoot(options InstallOptions) (string, error) {
	if options.Local {
		switch options.Provider {
		case ProviderCodex:
			return filepath.Join(options.LocalRoot, ".codex", "skills"), nil
		case ProviderClaude:
			return filepath.Join(options.LocalRoot, ".claude", "skills"), nil
		}
	}

	switch options.Provider {
	case ProviderCodex:
		return codexSkillsRoot()
	case ProviderClaude:
		return claudeSkillsRoot()
	default:
		return "", fmt.Errorf("unsupported install provider %q", options.Provider)
	}
}

// contextPath resolves where the managed task context should be installed.
func contextPath(options InstallOptions) (string, error) {
	if options.Local {
		return filepath.Join(options.LocalRoot, "AGENTS.md"), nil
	}

	switch options.Provider {
	case ProviderCodex:
		return CodexContextPath()
	case ProviderClaude:
		return ClaudeContextPath()
	default:
		return "", fmt.Errorf("unsupported install provider %q", options.Provider)
	}
}

// hookConfigPath resolves where provider hook configuration should be installed.
func hookConfigPath(options InstallOptions) (string, error) {
	if options.Local {
		switch options.Provider {
		case ProviderCodex:
			return filepath.Join(options.LocalRoot, ".codex", "hooks.json"), nil
		case ProviderClaude:
			return filepath.Join(options.LocalRoot, ".claude", "settings.json"), nil
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	switch options.Provider {
	case ProviderCodex:
		codexHome := os.Getenv("CODEX_HOME")
		if codexHome == "" {
			codexHome = filepath.Join(home, ".codex")
		}
		return filepath.Join(codexHome, "hooks.json"), nil
	case ProviderClaude:
		return filepath.Join(home, ".claude", "settings.json"), nil
	default:
		return "", fmt.Errorf("unsupported install provider %q", options.Provider)
	}
}

// codexConfigPath resolves the Codex config.toml path for hook feature flags.
func codexConfigPath(options InstallOptions) string {
	if options.Local {
		return filepath.Join(options.LocalRoot, ".codex", "config.toml")
	}
	codexHome := os.Getenv("CODEX_HOME")
	if codexHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return filepath.Join(".codex", "config.toml")
		}
		codexHome = filepath.Join(home, ".codex")
	}
	return filepath.Join(codexHome, "config.toml")
}

// writeFileIfChanged writes content and reports whether the file changed.
func writeFileIfChanged(path string, content []byte, mode fs.FileMode) (string, error) {
	existing, err := os.ReadFile(path)
	if err == nil && string(existing) == string(content) {
		return "unchanged", nil
	}
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("create directory %s: %w", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, content, mode); err != nil {
		return "", fmt.Errorf("write %s: %w", path, err)
	}
	if os.IsNotExist(err) {
		return "created", nil
	}
	return "updated", nil
}
