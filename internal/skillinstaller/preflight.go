package skillinstaller

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// validateInstallInputs checks existing configuration and destination shapes without writing files.
func validateInstallInputs(options InstallOptions, skillRoot, contextPath, hookPath string) error {
	existing, err := readOptionalInstallFile(hookPath)
	if err != nil {
		return err
	}
	if _, _, err := upsertHookConfig(existing, options.Force); err != nil {
		return err
	}
	paths := []string{contextPath}
	if options.Provider == ProviderClaude && options.Local {
		paths = append(paths, filepath.Join(options.LocalRoot, "CLAUDE.md"))
	}
	for _, path := range paths {
		if _, err := readOptionalInstallFile(path); err != nil {
			return err
		}
	}
	if err := validateDirectoryPath(skillRoot); err != nil {
		return err
	}
	for _, skill := range bundledSkills {
		target := filepath.Join(skillRoot, skill.directory)
		if err := rejectSymlink(target); err != nil {
			return err
		}
		if options.Force {
			// The target will be replaced, so its existing contents are irrelevant.
			continue
		}
		if err := validateBundledTarget(target, skill, options.Provider); err != nil {
			return err
		}
	}
	return nil
}

// readOptionalInstallFile validates parent directories and reads an existing regular file.
func readOptionalInstallFile(path string) ([]byte, error) {
	if err := validateDirectoryPath(filepath.Dir(path)); err != nil {
		return nil, err
	}
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("inspect install file %s: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("install file %s must be a regular file", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read install file %s: %w", path, err)
	}
	return data, nil
}

// validateDirectoryPath allows missing directories but rejects non-directory ancestors.
func validateDirectoryPath(path string) error {
	for {
		info, err := os.Stat(path)
		if err == nil {
			if !info.IsDir() {
				return fmt.Errorf("install directory %s must be a directory", path)
			}
			return nil
		}
		if !os.IsNotExist(err) {
			return fmt.Errorf("inspect install directory %s: %w", path, err)
		}
		parent := filepath.Dir(path)
		if parent == path {
			return fmt.Errorf("install directory %s has no existing ancestor", path)
		}
		path = parent
	}
}

// validateBundledTarget checks every destination used by a non-forced bundle installation.
func validateBundledTarget(target string, skill bundledSkill, provider Provider) error {
	return fs.WalkDir(bundledFiles, skill.root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() && !shouldInstallBundledFile(provider, skill, path+"/") {
			return fs.SkipDir
		}
		if !shouldInstallBundledFile(provider, skill, path) {
			return nil
		}
		relative, err := filepath.Rel(skill.root, path)
		if err != nil {
			return err
		}
		destination := filepath.Join(target, relative)
		info, err := os.Lstat(destination)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("inspect bundled target %s: %w", destination, err)
		}
		if entry.IsDir() {
			if !info.IsDir() {
				return fmt.Errorf("bundled skill directory %s must be a directory, not a symlink", destination)
			}
			return nil
		}
		if err := rejectSymlink(destination); err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("bundled skill file %s must be a regular file", destination)
		}
		return nil
	})
}
