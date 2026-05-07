package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rpcarvs/faz/internal/model"
)

// TestOnboardAndRecapWriteToStdout verifies static help text stays off stderr.
func TestOnboardAndRecapWriteToStdout(t *testing.T) {
	for _, tc := range []struct {
		name string
		run  func(*testing.T, *bytes.Buffer, *bytes.Buffer)
	}{
		{
			name: "onboard",
			run: func(t *testing.T, stdout, stderr *bytes.Buffer) {
				onboardCmd.SetOut(stdout)
				onboardCmd.SetErr(stderr)
				onboardCmd.Run(onboardCmd, nil)
			},
		},
		{
			name: "recap",
			run: func(t *testing.T, stdout, stderr *bytes.Buffer) {
				recapCmd.SetOut(stdout)
				recapCmd.SetErr(stderr)
				recapCmd.Run(recapCmd, nil)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
			tc.run(t, stdout, stderr)
			if stdout.Len() == 0 {
				t.Fatalf("%s stdout is empty", tc.name)
			}
			if stderr.Len() != 0 {
				t.Fatalf("%s stderr = %q", tc.name, stderr.String())
			}
		})
	}
}

// TestInfoAndReadyWriteToStdout verifies repo-backed informational output uses stdout.
func TestInfoAndReadyWriteToStdout(t *testing.T) {
	root := initGitRepo(t)
	restore := chdir(t, root)
	defer restore()

	runInitForTest(t)

	for _, tc := range []struct {
		name string
		run  func(*bytes.Buffer, *bytes.Buffer) error
		want string
	}{
		{
			name: "info",
			run: func(stdout, stderr *bytes.Buffer) error {
				infoCmd.SetOut(stdout)
				infoCmd.SetErr(stderr)
				return infoCmd.RunE(infoCmd, nil)
			},
			want: "Open issues:",
		},
		{
			name: "ready",
			run: func(stdout, stderr *bytes.Buffer) error {
				readyCmd.SetOut(stdout)
				readyCmd.SetErr(stderr)
				return readyCmd.RunE(readyCmd, nil)
			},
			want: "No ready work",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
			if err := tc.run(stdout, stderr); err != nil {
				t.Fatalf("%s failed: %v", tc.name, err)
			}
			if !strings.Contains(stdout.String(), tc.want) {
				t.Fatalf("%s stdout = %q, want substring %q", tc.name, stdout.String(), tc.want)
			}
			if stderr.Len() != 0 {
				t.Fatalf("%s stderr = %q", tc.name, stderr.String())
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

	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	claimCmd.SetOut(stdout)
	claimCmd.SetErr(stderr)
	if err := claimCmd.RunE(claimCmd, []string{issueID}); err != nil {
		t.Fatalf("claim failed: %v", err)
	}

	for _, want := range []string{
		"Claimed issue:",
		"Status: in_progress",
		"Lease TTL:",
		"Title: Claim me",
		"Description:",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("claim stdout = %q, want substring %q", stdout.String(), want)
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("claim stderr = %q", stderr.String())
	}
}

// runInitForTest initializes faz in the current repository and fails fast on errors.
func runInitForTest(t *testing.T) {
	t.Helper()

	var stdout bytes.Buffer
	initCmd.SetOut(&stdout)
	initCmd.SetErr(&bytes.Buffer{})
	if err := initCmd.RunE(initCmd, nil); err != nil {
		t.Fatalf("run init: %v", err)
	}
}
