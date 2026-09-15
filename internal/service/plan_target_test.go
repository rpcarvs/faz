package service

import (
	"os"
	"path/filepath"
	"testing"
)

// TestValidatePlanTargetAcceptsOnlyContainedRegularDocuments verifies explicit plan assignments stay in faz-specs.
func TestValidatePlanTargetAcceptsOnlyContainedRegularDocuments(t *testing.T) {
	root := t.TempDir()
	specsDir := filepath.Join(root, "faz-specs")
	if err := os.Mkdir(specsDir, 0o755); err != nil {
		t.Fatalf("create faz-specs: %v", err)
	}
	if err := os.WriteFile(filepath.Join(specsDir, "PLAN01.md"), []byte("# plan"), 0o644); err != nil {
		t.Fatalf("write plan: %v", err)
	}
	if err := os.Mkdir(filepath.Join(specsDir, "PLAN02.md"), 0o755); err != nil {
		t.Fatalf("create plan directory: %v", err)
	}

	if err := ValidatePlanTarget(root, "PLAN01"); err != nil {
		t.Fatalf("validate regular plan: %v", err)
	}
	if err := os.WriteFile(filepath.Join(specsDir, "PLAN05-target.md"), []byte("# contained"), 0o644); err != nil {
		t.Fatalf("write contained target: %v", err)
	}
	if err := os.Symlink(filepath.Join(specsDir, "PLAN05-target.md"), filepath.Join(specsDir, "PLAN05.md")); err != nil {
		t.Fatalf("create contained symlink: %v", err)
	}
	if err := ValidatePlanTarget(root, "PLAN05"); err != nil {
		t.Fatalf("validate contained symlink: %v", err)
	}
	rootAlias := filepath.Join(t.TempDir(), "project-root")
	if err := os.Symlink(root, rootAlias); err != nil {
		t.Fatalf("create root symlink: %v", err)
	}
	if err := ValidatePlanTarget(rootAlias, "PLAN01"); err != nil {
		t.Fatalf("validate symlinked root: %v", err)
	}
	for _, planID := range []string{"PLAN03", "PLAN02", "../PLAN01", "PLAN/01"} {
		if err := ValidatePlanTarget(root, planID); err == nil {
			t.Fatalf("validate %q succeeded unexpectedly", planID)
		}
	}

	externalPlan := filepath.Join(t.TempDir(), "PLAN04.md")
	if err := os.WriteFile(externalPlan, []byte("# external"), 0o644); err != nil {
		t.Fatalf("write external plan: %v", err)
	}
	if err := os.Symlink(externalPlan, filepath.Join(specsDir, "PLAN04.md")); err != nil {
		t.Fatalf("create escaping symlink: %v", err)
	}
	if err := ValidatePlanTarget(root, "PLAN04"); err == nil {
		t.Fatalf("validate escaping symlink succeeded unexpectedly")
	}
	insideRootOutsideSpecs := filepath.Join(root, "PLAN06.md")
	if err := os.WriteFile(insideRootOutsideSpecs, []byte("# wrong directory"), 0o644); err != nil {
		t.Fatalf("write out-of-specs plan: %v", err)
	}
	if err := os.Symlink(insideRootOutsideSpecs, filepath.Join(specsDir, "PLAN06.md")); err != nil {
		t.Fatalf("create out-of-specs symlink: %v", err)
	}
	if err := ValidatePlanTarget(root, "PLAN06"); err == nil {
		t.Fatalf("validate out-of-specs symlink succeeded unexpectedly")
	}

	escapingRoot := t.TempDir()
	externalSpecs := t.TempDir()
	if err := os.WriteFile(filepath.Join(externalSpecs, "PLAN07.md"), []byte("# external specs"), 0o644); err != nil {
		t.Fatalf("write external specs plan: %v", err)
	}
	if err := os.Symlink(externalSpecs, filepath.Join(escapingRoot, "faz-specs")); err != nil {
		t.Fatalf("create escaping faz-specs symlink: %v", err)
	}
	if err := ValidatePlanTarget(escapingRoot, "PLAN07"); err == nil {
		t.Fatalf("validate escaping faz-specs symlink succeeded unexpectedly")
	}
}
