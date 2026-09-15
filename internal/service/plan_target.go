package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidatePlanTarget confirms an explicitly assigned plan resolves to a regular document inside faz-specs.
func ValidatePlanTarget(projectRoot, planID string) error {
	if err := validatePlanID(planID); err != nil {
		return err
	}

	resolvedRoot, err := filepath.EvalSymlinks(projectRoot)
	if err != nil {
		return fmt.Errorf("resolve project root: %w", err)
	}
	resolvedSpecsDir, err := filepath.EvalSymlinks(filepath.Join(resolvedRoot, "faz-specs"))
	if err != nil {
		return fmt.Errorf("resolve faz-specs directory: %w", err)
	}
	if !pathWithin(resolvedRoot, resolvedSpecsDir) {
		return fmt.Errorf("faz-specs directory escapes project root")
	}

	resolvedPlan, err := filepath.EvalSymlinks(filepath.Join(resolvedSpecsDir, planID+".md"))
	if err != nil {
		return fmt.Errorf("resolve plan document %q: %w", planID, err)
	}
	if !pathWithin(resolvedSpecsDir, resolvedPlan) {
		return fmt.Errorf("plan document %q escapes faz-specs", planID)
	}
	info, err := os.Stat(resolvedPlan)
	if err != nil {
		return fmt.Errorf("inspect plan document %q: %w", planID, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("plan document %q is not a regular file", planID)
	}
	return nil
}

// pathWithin reports whether candidate is root itself or lies below it after symlink resolution.
func pathWithin(root, candidate string) bool {
	relative, err := filepath.Rel(root, candidate)
	if err != nil {
		return false
	}
	return relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)))
}
