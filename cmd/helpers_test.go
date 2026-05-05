package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestCurrentProjectDirUsesGitRootFromSubdir verifies commands share one repo DB.
func TestCurrentProjectDirUsesGitRootFromSubdir(t *testing.T) {
	root := initGitRepo(t)
	subdir := filepath.Join(root, "internal", "nested")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatalf("create subdir: %v", err)
	}

	restore := chdir(t, subdir)
	defer restore()

	projectDir, err := currentProjectDir()
	if err != nil {
		t.Fatalf("resolve project dir: %v", err)
	}
	if projectDir != root {
		t.Fatalf("project dir = %q, want %q", projectDir, root)
	}
}

// initGitRepo creates a temporary Git repository for command path tests.
func initGitRepo(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	cmd := exec.Command("git", "init")
	cmd.Dir = root
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v\n%s", err, output)
	}
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatalf("resolve temp repo path: %v", err)
	}
	return resolvedRoot
}

// chdir changes the process working directory and returns a restore function.
func chdir(t *testing.T, dir string) func() {
	t.Helper()

	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("change cwd: %v", err)
	}
	return func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}
}
