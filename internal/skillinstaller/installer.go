package skillinstaller

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const taskManagementSkillDirName = "faz-task-management"
const specDrivenSkillDirName = "faz-spec-driven"
const skillDirName = taskManagementSkillDirName
const bundledSkillPath = "bundled/faz-task-management/SKILL.md"
const sessionStartCommand = "git rev-parse --show-toplevel >/dev/null 2>&1 && faz init && faz onboard"

// bundledFiles contains built-in skill files to install for supported tools.
//
//go:embed bundled/faz-task-management/SKILL.md bundled/faz-spec-driven
var bundledFiles embed.FS

type bundledSkill struct {
	directory string
	root      string
}

var bundledSkills = []bundledSkill{
	{directory: taskManagementSkillDirName, root: "bundled/faz-task-management"},
	{directory: specDrivenSkillDirName, root: "bundled/faz-spec-driven"},
}

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
	// SkillPath is retained for compatibility and identifies the task-management skill.
	SkillPath           string
	SkillPaths          []string
	ContextPath         string
	ContextAction       string
	HookPath            string
	HookAction          string
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

	skillPaths, err := installBundledSkills(skillRoot, options.Provider, options.Force)
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
		SkillPath: skillPaths[taskManagementSkillDirName],
		SkillPaths: []string{
			skillPaths[taskManagementSkillDirName],
			skillPaths[specDrivenSkillDirName],
		},
		ContextPath:   contextPath,
		ContextAction: contextAction,
		HookPath:      hookPath,
		HookAction:    hookAction,
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
	return installBundledSkill(root, bundledSkills[0], ProviderCodex, force)
}

// InstallClaudeSkill installs the bundled faz skill into Claude skills directory.
func InstallClaudeSkill(force bool) (string, error) {
	root, err := claudeSkillsRoot()
	if err != nil {
		return "", err
	}
	return installBundledSkill(root, bundledSkills[0], ProviderClaude, force)
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

// installBundledSkills writes all bundled skills and returns their target directories.
func installBundledSkills(root string, provider Provider, force bool) (map[string]string, error) {
	for _, skill := range bundledSkills {
		if err := rejectSymlink(filepath.Join(root, skill.directory)); err != nil {
			return nil, err
		}
	}

	paths := make(map[string]string, len(bundledSkills))
	for _, skill := range bundledSkills {
		path, err := installBundledSkill(root, skill, provider, force)
		if err != nil {
			return nil, err
		}
		paths[skill.directory] = path
	}
	return paths, nil
}

// installBundledSkill writes one embedded skill and its provider-specific files.
func installBundledSkill(root string, skill bundledSkill, provider Provider, force bool) (string, error) {
	target := filepath.Join(root, skill.directory)
	if err := ensureSkillTarget(target, force); err != nil {
		return "", err
	}

	err := fs.WalkDir(bundledFiles, skill.root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !shouldInstallBundledFile(provider, skill, path) {
			return nil
		}

		relativePath, err := filepath.Rel(skill.root, path)
		if err != nil {
			return fmt.Errorf("resolve bundled skill path: %w", err)
		}
		content, err := bundledFiles.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read bundled skill file %s: %w", path, err)
		}
		if provider == ProviderClaude && skill.directory == specDrivenSkillDirName && relativePath == "SKILL.md" {
			content, err = claudeSkillContent(content)
			if err != nil {
				return err
			}
		}

		destination := filepath.Join(target, filepath.FromSlash(relativePath))
		if err := ensureBundledDirectory(target, filepath.Dir(destination)); err != nil {
			return err
		}
		if err := rejectSymlink(destination); err != nil {
			return err
		}
		if err := os.WriteFile(destination, content, 0o644); err != nil {
			return fmt.Errorf("write bundled skill file %s: %w", destination, err)
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	return target, nil
}

// ensureBundledDirectory creates a destination directory without traversing symlinks.
func ensureBundledDirectory(root, directory string) error {
	relativePath, err := filepath.Rel(root, directory)
	if err != nil {
		return fmt.Errorf("resolve bundled skill directory: %w", err)
	}
	if relativePath == "." {
		return nil
	}

	current := root
	for _, component := range strings.Split(relativePath, string(filepath.Separator)) {
		current = filepath.Join(current, component)
		info, err := os.Lstat(current)
		if os.IsNotExist(err) {
			if err := os.Mkdir(current, 0o755); err != nil {
				return fmt.Errorf("create bundled skill directory %s: %w", current, err)
			}
			continue
		}
		if err != nil {
			return fmt.Errorf("check bundled skill directory %s: %w", current, err)
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return fmt.Errorf("bundled skill directory %s must be a directory, not a symlink", current)
		}
	}
	return nil
}

// rejectSymlink prevents installation from overwriting files outside the skill target.
func rejectSymlink(path string) error {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("check bundled skill file %s: %w", path, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("bundled skill file %s must not be a symlink", path)
	}
	return nil
}

// shouldInstallBundledFile excludes Codex-only metadata from other providers.
func shouldInstallBundledFile(provider Provider, skill bundledSkill, path string) bool {
	return provider == ProviderCodex || !strings.HasPrefix(path, skill.root+"/agents/")
}

// claudeSkillContent adds Claude's explicit-only invocation setting to SDD metadata.
func claudeSkillContent(content []byte) ([]byte, error) {
	const frontmatterBoundary = "---\n"
	if !strings.HasPrefix(string(content), frontmatterBoundary) {
		return nil, fmt.Errorf("SDD skill is missing YAML frontmatter")
	}
	return []byte(strings.Replace(string(content), frontmatterBoundary, frontmatterBoundary+"disable-model-invocation: true\n", 1)), nil
}

// ensureSkillTarget creates the skill directory and optionally clears it first.
func ensureSkillTarget(target string, force bool) error {
	info, err := os.Lstat(target)
	if err == nil && info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("skill target %s must not be a symlink", target)
	} else if err == nil && force {
		if err := os.RemoveAll(target); err != nil {
			return fmt.Errorf("remove existing skill at %s: %w", target, err)
		}
	} else if err == nil && !info.IsDir() {
		return fmt.Errorf("skill target %s must be a directory", target)
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
