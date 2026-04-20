package kanban

import (
	"sort"
	"strings"
	"time"

	"github.com/rpcarvs/faz/internal/model"
)

const (
	scopeAll    = "__all__"
	scopeNoEpic = "__no_epic__"
)

// Service defines the read-only operations needed by the kanban TUI.
type Service interface {
	List(filter model.ListFilter) ([]model.Issue, error)
}

// Scope identifies one kanban grouping target for the TUI.
type Scope struct {
	Key       string
	Title     string
	CreatedAt time.Time
}

type scopeColumns struct {
	Todo    []model.Issue
	Claimed []model.Issue
	Done    []model.Issue
}

// Catalog stores pre-grouped tasks for each selectable scope.
type Catalog struct {
	Scopes     []Scope
	Columns    map[string]scopeColumns
	EpicTitles map[string]string
}

// LoadCatalog fetches all issues and groups them into kanban scopes.
func LoadCatalog(svc Service) (Catalog, error) {
	issues, err := svc.List(model.ListFilter{All: true})
	if err != nil {
		return Catalog{}, err
	}
	return buildCatalog(issues), nil
}

// buildCatalog derives epic scopes and kanban columns from a flat issue list.
func buildCatalog(issues []model.Issue) Catalog {
	catalog := Catalog{
		Scopes: []Scope{
			{Key: scopeAll, Title: "All Epics"},
			{Key: scopeNoEpic, Title: "Tasks with No Epic"},
		},
		Columns: map[string]scopeColumns{
			scopeAll:    {},
			scopeNoEpic: {},
		},
		EpicTitles: make(map[string]string),
	}

	epicScopes := make([]Scope, 0)
	for _, issue := range issues {
		if issue.Type != "epic" {
			continue
		}
		epicScopes = append(epicScopes, Scope{
			Key:       issue.ID,
			Title:     issue.Title,
			CreatedAt: issue.CreatedAt,
		})
		catalog.Columns[issue.ID] = scopeColumns{}
		catalog.EpicTitles[issue.ID] = issue.Title
	}
	sort.Slice(epicScopes, func(i, j int) bool {
		return epicScopes[i].CreatedAt.After(epicScopes[j].CreatedAt)
	})
	catalog.Scopes = append(catalog.Scopes, epicScopes...)

	for _, issue := range issues {
		if issue.Type == "epic" {
			continue
		}
		columnTarget := scopeColumnForStatus(issue.Status)
		columns := catalog.Columns[scopeAll]
		assignIssue(&columns, columnTarget, issue)
		catalog.Columns[scopeAll] = columns
		if issue.ParentID == nil {
			noEpicColumns := catalog.Columns[scopeNoEpic]
			assignIssue(&noEpicColumns, columnTarget, issue)
			catalog.Columns[scopeNoEpic] = noEpicColumns
			continue
		}
		if _, ok := catalog.Columns[*issue.ParentID]; ok {
			parentColumns := catalog.Columns[*issue.ParentID]
			assignIssue(&parentColumns, columnTarget, issue)
			catalog.Columns[*issue.ParentID] = parentColumns
		}
	}

	for key, columns := range catalog.Columns {
		sortColumnsNewestFirst(&columns)
		catalog.Columns[key] = columns
	}

	return catalog
}

func assignIssue(columns *scopeColumns, status string, issue model.Issue) {
	switch status {
	case "closed":
		columns.Done = append(columns.Done, issue)
	case "in_progress":
		columns.Claimed = append(columns.Claimed, issue)
	default:
		columns.Todo = append(columns.Todo, issue)
	}
}

func scopeColumnForStatus(status string) string {
	if strings.TrimSpace(status) == "" {
		return "open"
	}
	return status
}

// sortColumnsNewestFirst keeps each kanban column ordered by newest issue first.
func sortColumnsNewestFirst(columns *scopeColumns) {
	sortIssuesNewestFirst(columns.Todo)
	sortIssuesNewestFirst(columns.Claimed)
	sortIssuesNewestFirst(columns.Done)
}

// sortIssuesNewestFirst orders issues by creation time descending with ID fallback.
func sortIssuesNewestFirst(issues []model.Issue) {
	sort.SliceStable(issues, func(i, j int) bool {
		if issues[i].CreatedAt.Equal(issues[j].CreatedAt) {
			return issues[i].ID > issues[j].ID
		}
		return issues[i].CreatedAt.After(issues[j].CreatedAt)
	})
}
