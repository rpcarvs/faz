package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rpcarvs/faz/internal/model"
)

// TestRootCommandsWriteNormalOutputToStdout verifies runtime command execution keeps stderr clean.
func TestRootCommandsWriteNormalOutputToStdout(t *testing.T) {
	root := initGitRepo(t)
	restore := chdir(t, root)
	defer restore()

	runInitForTest(t)

	tests := []struct {
		name    string
		args    []string
		wantOut string
	}{
		{name: "onboard", args: []string{"onboard"}, wantOut: "## Issue Tracking"},
		{name: "recap", args: []string{"recap"}, wantOut: "faz recap"},
		{name: "info", args: []string{"info"}, wantOut: "Open issues:"},
		{name: "ready", args: []string{"ready"}, wantOut: "No ready work"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, err := executeRootCommand(t, tc.args...)
			if err != nil {
				t.Fatalf("execute %s: %v", tc.name, err)
			}
			if !strings.Contains(stdout, tc.wantOut) {
				t.Fatalf("%s stdout = %q, want substring %q", tc.name, stdout, tc.wantOut)
			}
			if stderr != "" {
				t.Fatalf("%s stderr = %q", tc.name, stderr)
			}
		})
	}
}

// TestClaimWritesFullSuccessResponseToStdout verifies claim does not split streams.
func TestClaimWritesFullSuccessResponseToStdout(t *testing.T) {
	root := initGitRepo(t)
	restore := chdir(t, root)
	defer restore()

	runInitForTest(t)

	svc, sqlDB, err := openService()
	if err != nil {
		t.Fatalf("open service: %v", err)
	}
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Fatalf("close db: %v", err)
		}
	})

	issueID, err := svc.Create(model.Issue{
		Title:       "Claim me",
		Description: "Rich description for claim output.",
		Type:        "task",
		Priority:    1,
		Status:      "open",
	})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}

	stdout, stderr, err := executeRootCommand(t, "claim", issueID)
	if err != nil {
		t.Fatalf("claim failed: %v", err)
	}

	for _, want := range []string{
		"Claimed issue:",
		"Status: in_progress",
		"Lease TTL:",
		"Title: Claim me",
		"Description:",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("claim stdout = %q, want substring %q", stdout, want)
		}
	}
	if stderr != "" {
		t.Fatalf("claim stderr = %q", stderr)
	}
}

// executeRootCommand runs the root command with captured stdout and stderr.
func executeRootCommand(t *testing.T, args ...string) (string, string, error) {
	t.Helper()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	previousOut := rootCmd.OutOrStdout()
	previousErr := rootCmd.ErrOrStderr()

	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs(args)

	_, err := rootCmd.ExecuteC()

	rootCmd.SetOut(previousOut)
	rootCmd.SetErr(previousErr)
	rootCmd.SetArgs(nil)

	return stdout.String(), stderr.String(), err
}

// runInitForTest initializes faz in the current repository and fails fast on errors.
func runInitForTest(t *testing.T) {
	t.Helper()

	stdout, stderr, err := executeRootCommand(t, "init")
	if err != nil {
		t.Fatalf("run init: %v", err)
	}
	if stdout == "" {
		t.Fatalf("init stdout is empty")
	}
	if stderr != "" {
		t.Fatalf("init stderr = %q", stderr)
	}
}
