package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rpcarvs/faz/internal/db"
	"github.com/rpcarvs/faz/internal/model"
	"github.com/rpcarvs/faz/internal/service"
)

// TestSDDAssociationCommandsKeepLinksOptional verifies explicit links, exact filters, clears, and conditional context.
func TestSDDAssociationCommandsKeepLinksOptional(t *testing.T) {
	root := initGitRepo(t)
	specsDir := filepath.Join(root, "faz-specs")
	if err := os.Mkdir(specsDir, 0o755); err != nil {
		t.Fatalf("create faz-specs: %v", err)
	}
	for _, planID := range []string{"PLAN01", "PLAN02"} {
		if err := os.WriteFile(filepath.Join(specsDir, planID+".md"), []byte("# "+planID), 0o644); err != nil {
			t.Fatalf("write %s: %v", planID, err)
		}
	}
	restore := chdir(t, root)
	defer restore()
	runInitForTest(t)

	if _, _, err := executeRootCommand(t, "create", "Plan one W01", "--plan", "PLAN01", "--work", "W01"); err != nil {
		t.Fatalf("create PLAN01 W01: %v", err)
	}
	subdir := filepath.Join(root, "nested")
	if err := os.Mkdir(subdir, 0o755); err != nil {
		t.Fatalf("create subdir: %v", err)
	}
	insideSubdir := chdir(t, subdir)
	if _, _, err := executeRootCommand(t, "create", "Plan two W1", "--plan", "PLAN02", "--work", "W1"); err != nil {
		t.Fatalf("create from subdir: %v", err)
	}
	if _, _, err := executeRootCommand(t, "create", "Plan one W1", "--plan", "PLAN01", "--work", "W1"); err != nil {
		t.Fatalf("create PLAN01 W1: %v", err)
	}
	insideSubdir()

	issues := issuesByTitle(t)
	planOneW01 := issues["Plan one W01"]
	planOneW1 := issues["Plan one W1"]
	if planOneW01.ID == "" || planOneW1.ID == "" {
		t.Fatalf("created issues were not found: %#v", issues)
	}

	stdout, stderr, err := executeRootCommand(t, "list", "--plan", "PLAN01", "--work", "W1")
	if err != nil || stderr != "" || !strings.Contains(stdout, planOneW1.ID) || strings.Contains(stdout, planOneW01.ID) {
		t.Fatalf("combined exact filter stdout=%q stderr=%q err=%v", stdout, stderr, err)
	}
	for _, args := range [][]string{{"list", "--plan", ""}, {"list", "--work", ""}} {
		if _, _, err := executeRootCommand(t, args...); err == nil {
			t.Fatalf("empty filter %v succeeded unexpectedly", args)
		}
	}
	if err := issuesByTitleService(t).Close(planOneW1.ID); err != nil {
		t.Fatalf("close filtered issue: %v", err)
	}
	stdout, _, err = executeRootCommand(t, "list", "--plan", "PLAN01", "--work", "W1")
	if err != nil || strings.Contains(stdout, planOneW1.ID) {
		t.Fatalf("open filter retained closed issue: stdout=%q err=%v", stdout, err)
	}
	stdout, _, err = executeRootCommand(t, "list", "--all", "--plan", "PLAN01", "--work", "W1")
	if err != nil || !strings.Contains(stdout, planOneW1.ID) {
		t.Fatalf("all filter omitted closed issue: stdout=%q err=%v", stdout, err)
	}

	if _, _, err := executeRootCommand(t, "update", planOneW01.ID, "--plan", "PLAN01", "--clear-plan"); err == nil {
		t.Fatalf("conflicting plan update succeeded")
	}
	if _, _, err := executeRootCommand(t, "update", planOneW01.ID, "--work", "W01", "--clear-work"); err == nil {
		t.Fatalf("conflicting work update succeeded")
	}
	if _, _, err := executeRootCommand(t, "update", planOneW01.ID, "--clear-plan", "--clear-work"); err != nil {
		t.Fatalf("clear associations: %v", err)
	}
	cleared, err := issuesByTitleService(t).Get(planOneW01.ID)
	if err != nil || cleared.PlanID != nil || cleared.WorkID != nil {
		t.Fatalf("cleared issue = %#v, err=%v", cleared, err)
	}
	showOut, showErr, err := executeRootCommand(t, "show", planOneW01.ID)
	if err != nil || showErr != "" || strings.Contains(showOut, "SDD Plan:") || strings.Contains(showOut, "SDD Work:") {
		t.Fatalf("unlinked show output=%q stderr=%q err=%v", showOut, showErr, err)
	}

	associatedID, err := issuesByTitleService(t).Create(model.Issue{Title: "Claim linked", Type: "task", Priority: 1, PlanID: stringPointer("PLAN01"), WorkID: stringPointer("W01")})
	if err != nil {
		t.Fatalf("create linked claim issue: %v", err)
	}
	showOut, _, err = executeRootCommand(t, "show", associatedID)
	if err != nil || !strings.Contains(showOut, "SDD Plan: PLAN01 (faz-specs/PLAN01.md)") || !strings.Contains(showOut, "SDD Work: W01") {
		t.Fatalf("linked show output=%q err=%v", showOut, err)
	}
	claimOut, claimErr, err := executeRootCommand(t, "claim", associatedID)
	if err != nil || claimErr != "" || !strings.Contains(claimOut, "SDD Plan: PLAN01 (faz-specs/PLAN01.md)") || !strings.Contains(claimOut, "SDD Work: W01") {
		t.Fatalf("linked claim output=%q stderr=%q err=%v", claimOut, claimErr, err)
	}
}

// TestUpdateMigratesLegacyProjectDatabase verifies normal CLI assignment upgrades an initialized old database.
func TestUpdateMigratesLegacyProjectDatabase(t *testing.T) {
	root := initGitRepo(t)
	specsDir := filepath.Join(root, "faz-specs")
	if err := os.Mkdir(specsDir, 0o755); err != nil {
		t.Fatalf("create faz-specs: %v", err)
	}
	if err := os.WriteFile(filepath.Join(specsDir, "PLAN01.md"), []byte("# PLAN01"), 0o644); err != nil {
		t.Fatalf("write plan: %v", err)
	}
	createLegacyCLIProjectDB(t, root)
	restore := chdir(t, root)
	defer restore()

	stdout, stderr, err := executeRootCommand(t, "update", "faz-old1", "--plan", "PLAN01", "--work", "W01")
	if err != nil || stderr != "" || !strings.Contains(stdout, "Updated issue: faz-old1") {
		t.Fatalf("update legacy database stdout=%q stderr=%q err=%v", stdout, stderr, err)
	}
	legacy, err := issuesByTitleService(t).Get("faz-old1")
	if err != nil || legacy.PlanID == nil || *legacy.PlanID != "PLAN01" || legacy.WorkID == nil || *legacy.WorkID != "W01" {
		t.Fatalf("updated legacy issue = %#v, err=%v", legacy, err)
	}
}

// TestExplicitPlanAssignmentRejectsInvalidTargetsButAllowsInheritedMissingDocuments verifies lifecycle compatibility.
func TestExplicitPlanAssignmentRejectsInvalidTargetsButAllowsInheritedMissingDocuments(t *testing.T) {
	root := initGitRepo(t)
	specsDir := filepath.Join(root, "faz-specs")
	if err := os.Mkdir(specsDir, 0o755); err != nil {
		t.Fatalf("create faz-specs: %v", err)
	}
	if err := os.WriteFile(filepath.Join(specsDir, "PLAN01.md"), []byte("# PLAN01"), 0o644); err != nil {
		t.Fatalf("write plan: %v", err)
	}
	restore := chdir(t, root)
	defer restore()
	runInitForTest(t)

	for _, args := range [][]string{
		{"create", "Missing plan", "--plan", "PLAN02"},
		{"create", "Unsafe plan", "--plan", "../PLAN01"},
	} {
		if _, _, err := executeRootCommand(t, args...); err == nil {
			t.Fatalf("%v succeeded unexpectedly", args)
		}
	}
	if _, _, err := executeRootCommand(t, "create", "Linked parent", "--type", "epic", "--plan", "PLAN01", "--work", "W01"); err != nil {
		t.Fatalf("create linked parent: %v", err)
	}
	parent := issuesByTitle(t)["Linked parent"]
	if err := os.Remove(filepath.Join(specsDir, "PLAN01.md")); err != nil {
		t.Fatalf("remove plan document: %v", err)
	}
	if _, _, err := executeRootCommand(t, "create", "Inherited child", "--parent", parent.ID); err != nil {
		t.Fatalf("create inherited missing-plan child: %v", err)
	}
	child := issuesByTitle(t)["Inherited child"]
	if child.PlanID == nil || *child.PlanID != "PLAN01" || child.WorkID == nil || *child.WorkID != "W01" {
		t.Fatalf("inherited child associations = %#v", child)
	}
}

// issuesByTitle returns all current issues keyed by their distinct test titles.
func issuesByTitle(t *testing.T) map[string]model.Issue {
	t.Helper()
	svc := issuesByTitleService(t)
	issues, err := svc.List(model.ListFilter{All: true})
	if err != nil {
		t.Fatalf("list issues: %v", err)
	}
	byTitle := make(map[string]model.Issue, len(issues))
	for _, issue := range issues {
		byTitle[issue.Title] = issue
	}
	return byTitle
}

// issuesByTitleService opens the current test project service and closes its database at cleanup.
func issuesByTitleService(t *testing.T) *service.IssueService {
	t.Helper()
	svc, sqlDB, err := openService()
	if err != nil {
		t.Fatalf("open service: %v", err)
	}
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Fatalf("close database: %v", err)
		}
	})
	return svc
}

func stringPointer(value string) *string {
	return &value
}

// createLegacyCLIProjectDB writes an initialized issue database from before SDD columns existed.
func createLegacyCLIProjectDB(t *testing.T, projectDir string) {
	t.Helper()
	dbPath, err := db.EnsureProjectFiles(projectDir)
	if err != nil {
		t.Fatalf("create legacy project files: %v", err)
	}
	sqlDB, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("open legacy database: %v", err)
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			t.Fatalf("close legacy database: %v", err)
		}
	}()
	if _, err := sqlDB.Exec(`CREATE TABLE issues (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id TEXT UNIQUE,
		title TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		type TEXT NOT NULL,
		priority INTEGER NOT NULL DEFAULT 2,
		status TEXT NOT NULL DEFAULT 'open',
		claimed_at DATETIME,
		claim_expires_at DATETIME,
		parent_id INTEGER,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		closed_at DATETIME
	)`); err != nil {
		t.Fatalf("create legacy issues table: %v", err)
	}
	if _, err := sqlDB.Exec(`INSERT INTO issues(public_id, title, type) VALUES('faz-old1', 'Legacy CLI task', 'task')`); err != nil {
		t.Fatalf("insert legacy issue: %v", err)
	}
}
