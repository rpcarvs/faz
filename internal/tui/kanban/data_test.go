package kanban

import (
	"testing"
	"time"

	"github.com/rpcarvs/faz/internal/model"
)

func TestBuildCatalogGroupsEpicAndNoEpicScopes(t *testing.T) {
	parentID := "proj-e111"
	now := time.Now()

	issues := []model.Issue{
		{
			ID:        parentID,
			Title:     "E1: First epic",
			Type:      "epic",
			Status:    "open",
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        parentID + ".0",
			Title:     "Epic open task",
			Type:      "task",
			Status:    "open",
			ParentID:  &parentID,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        parentID + ".1",
			Title:     "Epic claimed task",
			Type:      "task",
			Status:    "in_progress",
			ParentID:  &parentID,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "proj-r123",
			Title:     "Standalone done task",
			Type:      "task",
			Status:    "closed",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	catalog := buildCatalog(issues)

	if len(catalog.Scopes) != 3 {
		t.Fatalf("expected 3 scopes, got %d", len(catalog.Scopes))
	}
	if catalog.Scopes[0].Key != scopeAll {
		t.Fatalf("expected first scope to be all, got %q", catalog.Scopes[0].Key)
	}
	if catalog.Scopes[1].Key != scopeNoEpic {
		t.Fatalf("expected second scope to be no epic, got %q", catalog.Scopes[1].Key)
	}
	if catalog.Scopes[2].Key != parentID {
		t.Fatalf("expected third scope to be epic %q, got %q", parentID, catalog.Scopes[2].Key)
	}

	all := catalog.Columns[scopeAll]
	if len(all.Todo) != 1 || all.Todo[0].ID != parentID+".0" {
		t.Fatalf("unexpected all/todo contents: %#v", all.Todo)
	}
	if len(all.Claimed) != 1 || all.Claimed[0].ID != parentID+".1" {
		t.Fatalf("unexpected all/claimed contents: %#v", all.Claimed)
	}
	if len(all.Done) != 1 || all.Done[0].ID != "proj-r123" {
		t.Fatalf("unexpected all/done contents: %#v", all.Done)
	}

	noEpic := catalog.Columns[scopeNoEpic]
	if len(noEpic.Todo) != 0 || len(noEpic.Claimed) != 0 {
		t.Fatalf("expected no open or claimed no-epic tasks, got todo=%d claimed=%d", len(noEpic.Todo), len(noEpic.Claimed))
	}
	if len(noEpic.Done) != 1 || noEpic.Done[0].ID != "proj-r123" {
		t.Fatalf("unexpected no-epic done contents: %#v", noEpic.Done)
	}

	epic := catalog.Columns[parentID]
	if len(epic.Todo) != 1 || epic.Todo[0].ID != parentID+".0" {
		t.Fatalf("unexpected epic/todo contents: %#v", epic.Todo)
	}
	if len(epic.Claimed) != 1 || epic.Claimed[0].ID != parentID+".1" {
		t.Fatalf("unexpected epic/claimed contents: %#v", epic.Claimed)
	}
	if len(epic.Done) != 0 {
		t.Fatalf("expected no epic done tasks, got %#v", epic.Done)
	}
}

func TestBuildCatalogOrdersEpicScopesNewestFirstAfterDefaults(t *testing.T) {
	now := time.Now()
	issues := []model.Issue{
		{
			ID:        "proj-e111",
			Title:     "E1: Older epic",
			Type:      "epic",
			Status:    "open",
			CreatedAt: now.Add(-2 * time.Hour),
			UpdatedAt: now,
		},
		{
			ID:        "proj-e222",
			Title:     "E2: Newer epic",
			Type:      "epic",
			Status:    "open",
			CreatedAt: now.Add(-1 * time.Hour),
			UpdatedAt: now,
		},
	}

	catalog := buildCatalog(issues)
	if len(catalog.Scopes) != 4 {
		t.Fatalf("expected 4 scopes, got %d", len(catalog.Scopes))
	}
	if catalog.Scopes[0].Key != scopeAll || catalog.Scopes[1].Key != scopeNoEpic {
		t.Fatalf("expected default scopes first, got %#v", catalog.Scopes[:2])
	}
	if catalog.Scopes[2].Key != "proj-e222" || catalog.Scopes[3].Key != "proj-e111" {
		t.Fatalf("expected newest epic first after defaults, got %#v", catalog.Scopes[2:])
	}
}

func TestBuildCatalogOrdersColumnTasksNewestFirst(t *testing.T) {
	parentID := "proj-e111"
	now := time.Now()

	issues := []model.Issue{
		{
			ID:        parentID,
			Title:     "E1: Epic",
			Type:      "epic",
			Status:    "open",
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        parentID + ".0",
			Title:     "Older open task",
			Type:      "task",
			Status:    "open",
			ParentID:  &parentID,
			CreatedAt: now.Add(-2 * time.Hour),
			UpdatedAt: now,
		},
		{
			ID:        parentID + ".1",
			Title:     "Newer open task",
			Type:      "task",
			Status:    "open",
			ParentID:  &parentID,
			CreatedAt: now.Add(-1 * time.Hour),
			UpdatedAt: now,
		},
		{
			ID:        parentID + ".2",
			Title:     "Older claimed task",
			Type:      "task",
			Status:    "in_progress",
			ParentID:  &parentID,
			CreatedAt: now.Add(-3 * time.Hour),
			UpdatedAt: now,
		},
		{
			ID:        parentID + ".3",
			Title:     "Newer claimed task",
			Type:      "task",
			Status:    "in_progress",
			ParentID:  &parentID,
			CreatedAt: now.Add(-30 * time.Minute),
			UpdatedAt: now,
		},
	}

	catalog := buildCatalog(issues)
	epic := catalog.Columns[parentID]

	if len(epic.Todo) != 2 || epic.Todo[0].ID != parentID+".1" || epic.Todo[1].ID != parentID+".0" {
		t.Fatalf("expected open tasks newest first, got %#v", epic.Todo)
	}
	if len(epic.Claimed) != 2 || epic.Claimed[0].ID != parentID+".3" || epic.Claimed[1].ID != parentID+".2" {
		t.Fatalf("expected claimed tasks newest first, got %#v", epic.Claimed)
	}
}
