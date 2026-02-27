package service

import "testing"

func TestNormalizeIssueID(t *testing.T) {
	if _, err := NormalizeIssueID("abc"); err == nil {
		t.Fatalf("expected parse error for invalid issue ID")
	}

	id, err := NormalizeIssueID("FAZ-Ab12")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if id != "faz-ab12" {
		t.Fatalf("expected faz-ab12, got %s", id)
	}

	child, err := NormalizeIssueID("faz-ab12.3")
	if err != nil {
		t.Fatalf("unexpected child parse error: %v", err)
	}
	if child != "faz-ab12.3" {
		t.Fatalf("expected faz-ab12.3, got %s", child)
	}
}
