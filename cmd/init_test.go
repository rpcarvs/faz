package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestInitFromSubdirInitializesGitRoot verifies init never creates nested stores.
func TestInitFromSubdirInitializesGitRoot(t *testing.T) {
	root := initGitRepo(t)
	subdir := filepath.Join(root, "cmd")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatalf("create subdir: %v", err)
	}

	restore := chdir(t, subdir)
	defer restore()

	var output bytes.Buffer
	initCmd.SetOut(&output)
	t.Cleanup(func() { initCmd.SetOut(nil) })

	if err := initCmd.RunE(initCmd, nil); err != nil {
		t.Fatalf("run init: %v", err)
	}

	rootDB := filepath.Join(root, ".faz", "taskstore.db")
	nestedDB := filepath.Join(subdir, ".faz", "taskstore.db")
	if _, err := os.Stat(rootDB); err != nil {
		t.Fatalf("expected root database: %v", err)
	}
	if _, err := os.Stat(nestedDB); !os.IsNotExist(err) {
		t.Fatalf("nested database exists or stat failed unexpectedly: %v", err)
	}
}

// TestInitReportsAlreadyInitialized verifies repeated init is explicit and safe.
func TestInitReportsAlreadyInitialized(t *testing.T) {
	root := initGitRepo(t)
	restore := chdir(t, root)
	defer restore()

	var output bytes.Buffer
	initCmd.SetOut(&output)
	t.Cleanup(func() { initCmd.SetOut(nil) })

	if err := initCmd.RunE(initCmd, nil); err != nil {
		t.Fatalf("first init: %v", err)
	}
	output.Reset()

	if err := initCmd.RunE(initCmd, nil); err != nil {
		t.Fatalf("second init: %v", err)
	}
	if !strings.Contains(output.String(), "faz is already initialized") {
		t.Fatalf("second init output = %q", output.String())
	}
}
