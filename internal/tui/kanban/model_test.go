package kanban

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/rpcarvs/faz/internal/model"
)

// stubService supplies deterministic kanban reads for tests.
type stubService struct {
	issues        []model.Issue
	dependencies  map[string][]model.Issue
	dependents    map[string][]model.Issue
	dependencyErr error
	dependentErr  error
}

// List returns the configured issue set.
func (s stubService) List(filter model.ListFilter) ([]model.Issue, error) {
	return s.issues, nil
}

// Dependencies returns the configured blockers for an issue.
func (s stubService) Dependencies(publicID string) ([]model.Issue, error) {
	if s.dependencyErr != nil {
		return nil, s.dependencyErr
	}
	return s.dependencies[publicID], nil
}

// Dependents returns the configured dependents for an issue.
func (s stubService) Dependents(publicID string) ([]model.Issue, error) {
	if s.dependentErr != nil {
		return nil, s.dependentErr
	}
	return s.dependents[publicID], nil
}

func TestModelLoadDetailsCmdLoadsDependenciesAndDependents(t *testing.T) {
	now := time.Now()
	issueID := "proj-e1.0"
	svc := stubService{
		dependencies: map[string][]model.Issue{
			issueID: []model.Issue{
				{ID: "proj-e1.1", Title: "Blocker task", CreatedAt: now, UpdatedAt: now},
			},
		},
		dependents: map[string][]model.Issue{
			issueID: []model.Issue{
				{ID: "proj-e1.2", Title: "Dependent task", CreatedAt: now, UpdatedAt: now},
			},
		},
	}

	model := NewModel(svc)
	msg := model.loadDetailsCmd(issueID)()
	loaded, ok := msg.(detailsLoadedMsg)
	if !ok {
		t.Fatalf("expected detailsLoadedMsg, got %T", msg)
	}
	if loaded.issueID != issueID {
		t.Fatalf("expected issue ID %q, got %q", issueID, loaded.issueID)
	}
	if len(loaded.details.Dependencies) != 1 || loaded.details.Dependencies[0].ID != "proj-e1.1" {
		t.Fatalf("unexpected dependencies: %#v", loaded.details.Dependencies)
	}
	if len(loaded.details.Dependents) != 1 || loaded.details.Dependents[0].ID != "proj-e1.2" {
		t.Fatalf("unexpected dependents: %#v", loaded.details.Dependents)
	}
}

func TestRenderDetailsShowsDependenciesAndDependents(t *testing.T) {
	now := time.Now()
	parentID := "proj-e1"
	issueID := parentID + ".0"
	svc := stubService{
		issues: []model.Issue{
			{
				ID:          parentID,
				Title:       "E1: Epic",
				Type:        "epic",
				Status:      "open",
				Description: "Epic desc",
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			{
				ID:          issueID,
				Title:       "Selected task",
				Type:        "task",
				Status:      "open",
				Description: "Task desc",
				ParentID:    &parentID,
				CreatedAt:   now.Add(time.Minute),
				UpdatedAt:   now.Add(time.Minute),
			},
		},
		dependencies: map[string][]model.Issue{
			issueID: []model.Issue{
				{ID: parentID + ".1", Title: "First blocker", CreatedAt: now, UpdatedAt: now},
			},
		},
		dependents: map[string][]model.Issue{
			issueID: []model.Issue{
				{ID: parentID + ".2", Title: "First dependent", CreatedAt: now, UpdatedAt: now},
			},
		},
	}

	catalog, err := LoadCatalog(svc)
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}

	model := NewModel(svc)
	model.catalog = catalog
	model.ready = true
	model.width = 120
	model.height = 40
	model.details = map[string]issueDetails{}

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected details load command")
	}
	model = updated.(Model)

	msg := cmd()
	updated, _ = model.Update(msg)
	model = updated.(Model)

	view := model.renderDetails()
	if !strings.Contains(view, "Blocked by:") {
		t.Fatalf("expected blockers section in details view: %s", view)
	}
	if !strings.Contains(view, "Blocks:") {
		t.Fatalf("expected dependents section in details view: %s", view)
	}
	if !strings.Contains(view, "First blocker") {
		t.Fatalf("expected blocker title in details view: %s", view)
	}
	if !strings.Contains(view, "First dependent") {
		t.Fatalf("expected dependent title in details view: %s", view)
	}
}

func TestRenderDetailsShowsLoadingStateBeforeDependencyFetchCompletes(t *testing.T) {
	now := time.Now()
	parentID := "proj-e1"
	issueID := parentID + ".0"
	catalog := buildCatalog([]model.Issue{
		{
			ID:        parentID,
			Title:     "E1: Epic",
			Type:      "epic",
			Status:    "open",
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:          issueID,
			Title:       "Selected task",
			Type:        "task",
			Status:      "open",
			Description: "Task desc",
			ParentID:    &parentID,
			CreatedAt:   now.Add(time.Minute),
			UpdatedAt:   now.Add(time.Minute),
		},
	})

	kanbanModel := NewModel(stubService{})
	kanbanModel.catalog = catalog
	kanbanModel.ready = true
	kanbanModel.width = 120
	kanbanModel.height = 40
	kanbanModel.showDetails = true
	kanbanModel.inspectedIssueID = issueID
	kanbanModel.inspectedIssue = model.Issue{
		ID:          issueID,
		Title:       "Selected task",
		Type:        "task",
		Status:      "open",
		Description: "Task desc",
		ParentID:    &parentID,
		CreatedAt:   now.Add(time.Minute),
		UpdatedAt:   now.Add(time.Minute),
	}
	kanbanModel.details = map[string]issueDetails{
		issueID: {Loading: true},
	}

	view := kanbanModel.renderDetails()
	if !strings.Contains(view, "Loading dependencies...") {
		t.Fatalf("expected loading state in details view: %s", view)
	}
	if strings.Contains(view, "Blocked by:\n  None") || strings.Contains(view, "Blocks:\n  None") {
		t.Fatalf("expected loading state without empty dependency sections: %s", view)
	}
}

func TestBoardColumnLayoutFitsWithinAvailableWidth(t *testing.T) {
	widths, gap := boardColumnLayout(60)
	total := widths[0] + widths[1] + widths[2] + 2*gap
	if total > 60 {
		t.Fatalf("expected board width to fit in 60 columns, got %d", total)
	}
}

func TestViewRendersHeaderAndColumnsAtNarrowWidth(t *testing.T) {
	now := time.Now()
	parentID := "proj-e1"
	svc := stubService{
		issues: []model.Issue{
			{
				ID:        parentID,
				Title:     "E1: Epic",
				Type:      "epic",
				Status:    "open",
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				ID:          parentID + ".0",
				Title:       "First task for narrow layout coverage",
				Type:        "task",
				Status:      "open",
				Description: "Task desc",
				ParentID:    &parentID,
				CreatedAt:   now.Add(time.Minute),
				UpdatedAt:   now.Add(time.Minute),
			},
		},
	}

	catalog, err := LoadCatalog(svc)
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}

	model := NewModel(svc)
	model.catalog = catalog
	model.ready = true
	model.width = 60
	model.height = 24

	view := model.View()
	if !strings.Contains(view, "faz kanban") {
		t.Fatalf("expected header in narrow view: %s", view)
	}
	if !strings.Contains(view, "TO DO") || !strings.Contains(view, "CLAIMED") || !strings.Contains(view, "DONE") {
		t.Fatalf("expected board columns in narrow view: %s", view)
	}
	if maxRenderedLineWidth(view) > model.width {
		t.Fatalf("expected rendered view to fit width %d, got %d", model.width, maxRenderedLineWidth(view))
	}
}

func TestRenderCardSelectedStylesTextRowsWithSelectedBackground(t *testing.T) {
	oldProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	defer lipgloss.SetColorProfile(oldProfile)

	now := time.Now()
	issue := model.Issue{
		ID:        "proj-e1.0",
		Title:     "Short title",
		Type:      "task",
		Priority:  1,
		Status:    "open",
		CreatedAt: now,
		UpdatedAt: now,
	}
	card := NewModel(stubService{}).renderCard(issue, true, 32)

	assertMarkerStyleContains(t, card, "Short title", "48;5;240")
	assertMarkerStyleContains(t, card, "P1", "48;5;240")
}

func TestApplyWindowSizeIgnoresTransientZeroDimensions(t *testing.T) {
	model := NewModel(stubService{})

	model.applyWindowSize(80, 24)
	if model.width != 80 || model.height != 24 || !model.ready {
		t.Fatalf("expected initial valid size to be recorded, got %+v", model)
	}

	model.applyWindowSize(0, 0)
	if model.width != 80 || model.height != 24 {
		t.Fatalf("expected zero-size event to preserve last valid dimensions, got %dx%d", model.width, model.height)
	}
}

func TestFooterLinesUsesFullHintsBeforeCondensing(t *testing.T) {
	model := NewModel(stubService{})
	lines, condensed := model.footerLines(72)
	if condensed {
		t.Fatalf("expected non-condensed footer at medium width, got %#v", lines)
	}
	if len(lines) != 2 {
		t.Fatalf("expected wrapped two-line footer at medium width, got %d lines (%#v)", len(lines), lines)
	}
	joined := strings.Join(lines, " ")
	if !strings.Contains(joined, "Shift+Tab previous") {
		t.Fatalf("expected full hints in medium-width footer, got %q", joined)
	}
}

func TestFooterLinesCondensesAtVeryNarrowWidth(t *testing.T) {
	model := NewModel(stubService{})
	lines, condensed := model.footerLines(24)
	if !condensed {
		t.Fatalf("expected condensed footer at narrow width, got %#v", lines)
	}
	joined := strings.Join(lines, " ")
	if !strings.Contains(joined, "o") {
		t.Fatalf("expected condensed footer to advertise help key, got %q", joined)
	}
}

func TestUpdatePreservesLastValidSizeAcrossTransientZeroResize(t *testing.T) {
	model := NewModel(stubService{})

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 90, Height: 28})
	model = updated.(Model)
	updated, _ = model.Update(tea.WindowSizeMsg{Width: 0, Height: 0})
	model = updated.(Model)

	if model.width != 90 || model.height != 28 {
		t.Fatalf("expected transient zero resize to preserve dimensions, got %dx%d", model.width, model.height)
	}
}

func TestQClosesDetailsModalWithoutQuitting(t *testing.T) {
	now := time.Now()
	parentID := "proj-e1"
	catalog := buildCatalog([]model.Issue{
		{
			ID:        parentID,
			Title:     "E1: Epic",
			Type:      "epic",
			Status:    "open",
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:          parentID + ".0",
			Title:       "Selected task",
			Type:        "task",
			Status:      "open",
			Description: "Task desc",
			ParentID:    &parentID,
			CreatedAt:   now.Add(time.Minute),
			UpdatedAt:   now.Add(time.Minute),
		},
	})

	model := NewModel(stubService{})
	model.catalog = catalog
	model.ready = true
	model.width = 80
	model.height = 24
	model.showDetails = true

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	model = updated.(Model)

	if model.showDetails {
		t.Fatal("expected q to close details modal")
	}
	if cmd != nil {
		t.Fatal("expected q in details modal to avoid quitting the app")
	}
}

func TestQClosesEpicPickerWithoutQuitting(t *testing.T) {
	model := NewModel(stubService{})
	model.ready = true
	model.showPicker = true
	model.catalog = Catalog{Scopes: []Scope{{Key: scopeAll, Title: "All Epics"}}}

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	model = updated.(Model)

	if model.showPicker {
		t.Fatal("expected q to close epic picker")
	}
	if cmd != nil {
		t.Fatal("expected q in epic picker to avoid quitting the app")
	}
}

func TestStartupEpicPickerIgnoresTabBeforeCatalogLoads(t *testing.T) {
	model := NewModel(stubService{}, WithPicker())
	model.ready = true

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(Model)

	if !model.showPicker {
		t.Fatal("expected startup picker to stay open")
	}
	if model.pickerIndex != 0 {
		t.Fatalf("expected picker index to remain 0, got %d", model.pickerIndex)
	}
	if cmd != nil {
		t.Fatal("expected tab before catalog load to avoid commands")
	}
}

func TestEOpensEpicPickerAtTop(t *testing.T) {
	model := NewModel(stubService{})
	model.ready = true
	model.width = 80
	model.height = 24
	model.scopeIndex = 4
	model.pickerIndex = 4
	model.pickerScroll = 3
	model.catalog = Catalog{Scopes: testScopes(8)}

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	model = updated.(Model)

	if !model.showPicker {
		t.Fatal("expected e to open epic picker")
	}
	if model.pickerIndex != 0 || model.pickerScroll != 0 {
		t.Fatalf("expected picker to open at top, got index=%d scroll=%d", model.pickerIndex, model.pickerScroll)
	}
	if cmd != nil {
		t.Fatal("expected no command when opening epic picker")
	}
}

func TestEpicPickerScrollsWhenSelectionMovesPastVisibleRows(t *testing.T) {
	model := NewModel(stubService{})
	model.ready = true
	model.width = 80
	model.height = 14
	model.showPicker = true
	model.catalog = Catalog{Scopes: testScopes(12)}

	for i := 0; i < 6; i++ {
		updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyDown})
		model = updated.(Model)
		if cmd != nil {
			t.Fatal("expected no command while moving picker selection")
		}
	}

	if model.pickerIndex != 6 {
		t.Fatalf("expected picker index 6, got %d", model.pickerIndex)
	}
	if model.pickerScroll == 0 {
		t.Fatal("expected picker to scroll after moving past visible rows")
	}

	view := model.renderPicker()
	if strings.Contains(view, "Epic 01") {
		t.Fatalf("expected first epic to scroll out of view: %s", view)
	}
	if !strings.Contains(view, "> Epic 07") {
		t.Fatalf("expected selected epic to remain visible: %s", view)
	}
}

func TestEpicPickerFitsTerminalHeightWithManyEpics(t *testing.T) {
	model := NewModel(stubService{})
	model.ready = true
	model.width = 80
	model.height = 14
	model.showPicker = true
	model.catalog = Catalog{Scopes: testScopes(40)}

	view := model.View()
	if got := renderedLineCount(view); got > model.height {
		t.Fatalf("expected picker view height <= %d, got %d\n%s", model.height, got, view)
	}
}

func TestEpicPickerUsesExpandedMaximumVisibleRows(t *testing.T) {
	model := NewModel(stubService{})
	model.ready = true
	model.width = 80
	model.height = 40
	model.showPicker = true
	model.catalog = Catalog{Scopes: testScopes(20)}

	view := model.renderPicker()
	for i := 1; i <= maxPickerRows; i++ {
		title := fmt.Sprintf("Epic %02d", i)
		if !strings.Contains(view, title) {
			t.Fatalf("expected picker to show %q: %s", title, view)
		}
	}
	if strings.Contains(view, "Epic 14") {
		t.Fatalf("expected picker to cap visible rows at %d: %s", maxPickerRows, view)
	}
}

func TestOOpensHelpAndQClosesHelpWithoutQuitting(t *testing.T) {
	model := NewModel(stubService{})
	model.ready = true
	model.width = 80
	model.height = 24
	model.catalog = Catalog{Scopes: []Scope{{Key: scopeAll, Title: "All Epics"}}}

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")})
	model = updated.(Model)
	if !model.showHelp {
		t.Fatal("expected o to open help modal")
	}
	if cmd != nil {
		t.Fatal("expected no command when opening help modal")
	}

	updated, cmd = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	model = updated.(Model)
	if model.showHelp {
		t.Fatal("expected q to close help modal")
	}
	if cmd != nil {
		t.Fatal("expected q in help modal to avoid quitting the app")
	}
}

func TestFOpensTypePickerWithDefaultAllSelection(t *testing.T) {
	model := NewModel(stubService{})
	model.ready = true
	model.width = 80
	model.height = 24
	model.catalog = Catalog{Scopes: []Scope{{Key: scopeAll, Title: "All Epics"}}}
	model.typeFilter = "feature"
	model.typeIndex = 3

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})
	model = updated.(Model)
	if !model.showType {
		t.Fatal("expected f to open type picker")
	}
	if model.typeIndex != 0 {
		t.Fatalf("expected f to reset picker to all, got index %d", model.typeIndex)
	}
	if cmd != nil {
		t.Fatal("expected no command when opening type picker")
	}
}

func TestTypePickerEnterAppliesTypeFilterToBoardColumns(t *testing.T) {
	now := time.Now()
	parentID := "proj-e1"
	catalog := buildCatalog([]model.Issue{
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
			Title:     "Todo bug",
			Type:      "bug",
			Status:    "open",
			ParentID:  &parentID,
			CreatedAt: now.Add(1 * time.Minute),
			UpdatedAt: now.Add(1 * time.Minute),
		},
		{
			ID:        parentID + ".1",
			Title:     "Todo feature",
			Type:      "feature",
			Status:    "open",
			ParentID:  &parentID,
			CreatedAt: now.Add(2 * time.Minute),
			UpdatedAt: now.Add(2 * time.Minute),
		},
		{
			ID:        parentID + ".2",
			Title:     "Claimed bug",
			Type:      "bug",
			Status:    "in_progress",
			ParentID:  &parentID,
			CreatedAt: now.Add(3 * time.Minute),
			UpdatedAt: now.Add(3 * time.Minute),
		},
		{
			ID:        parentID + ".3",
			Title:     "Done bug",
			Type:      "bug",
			Status:    "closed",
			ParentID:  &parentID,
			CreatedAt: now.Add(4 * time.Minute),
			UpdatedAt: now.Add(4 * time.Minute),
			ClosedAt:  ptrTime(now.Add(5 * time.Minute)),
		},
	})

	model := NewModel(stubService{})
	model.catalog = catalog
	model.ready = true
	model.width = 100
	model.height = 30
	model.scopeIndex = 0
	model.showType = true
	model.typeIndex = 2 // bug

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	if model.showType {
		t.Fatal("expected enter to close type picker")
	}
	if model.typeFilter != "bug" {
		t.Fatalf("expected selected type filter bug, got %q", model.typeFilter)
	}
	if cmd != nil {
		t.Fatal("expected no command when selecting type filter")
	}

	columns := model.currentColumns()
	if len(columns.Todo) != 1 || columns.Todo[0].Type != "bug" {
		t.Fatalf("expected only bug todo items, got %#v", columns.Todo)
	}
	if len(columns.Claimed) != 1 || columns.Claimed[0].Type != "bug" {
		t.Fatalf("expected only bug claimed items, got %#v", columns.Claimed)
	}
	if len(columns.Done) != 1 || columns.Done[0].Type != "bug" {
		t.Fatalf("expected only bug done items, got %#v", columns.Done)
	}
}

func TestQClosesTypePickerWithoutQuitting(t *testing.T) {
	model := NewModel(stubService{})
	model.ready = true
	model.showType = true

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	model = updated.(Model)
	if model.showType {
		t.Fatal("expected q to close type picker")
	}
	if cmd != nil {
		t.Fatal("expected q in type picker to avoid quitting the app")
	}
}

func TestDOpensEpicDetailsInEpicScopeAndQCloses(t *testing.T) {
	now := time.Now()
	parentID := "proj-e1"
	catalog := buildCatalog([]model.Issue{
		{
			ID:          parentID,
			Title:       "E1: Epic",
			Type:        "epic",
			Status:      "open",
			Priority:    1,
			Description: "Epic description text",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:        parentID + ".0",
			Title:     "Epic task",
			Type:      "task",
			Status:    "open",
			ParentID:  &parentID,
			CreatedAt: now.Add(time.Minute),
			UpdatedAt: now.Add(time.Minute),
		},
	})

	kanbanModel := NewModel(stubService{})
	kanbanModel.ready = true
	kanbanModel.width = 100
	kanbanModel.height = 30
	kanbanModel.catalog = catalog
	kanbanModel.scopeIndex = 2

	updated, cmd := kanbanModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	kanbanModel = updated.(Model)
	if !kanbanModel.showEpic {
		t.Fatal("expected d to open epic details modal")
	}
	if cmd != nil {
		t.Fatal("expected no command when opening epic details modal")
	}

	view := kanbanModel.renderEpicDetails()
	if !strings.Contains(view, "E1: Epic") || !strings.Contains(view, "Epic description text") {
		t.Fatalf("expected epic details content, got %s", view)
	}
	if strings.Contains(view, "Epic: ") {
		t.Fatalf("expected epic modal without redundant Epic prefix, got %s", view)
	}

	updated, cmd = kanbanModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	kanbanModel = updated.(Model)
	if kanbanModel.showEpic {
		t.Fatal("expected q to close epic modal")
	}
	if cmd != nil {
		t.Fatal("expected q in epic modal to avoid quitting the app")
	}
}

func TestEpicDetailsInAllScopeUsesSelectedTaskParentEpic(t *testing.T) {
	now := time.Now()
	parentID := "proj-e1"
	catalog := buildCatalog([]model.Issue{
		{
			ID:          parentID,
			Title:       "E1: Epic",
			Type:        "epic",
			Status:      "open",
			Priority:    1,
			Description: "Epic from parent",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:        parentID + ".0",
			Title:     "Task with parent",
			Type:      "task",
			Status:    "open",
			ParentID:  &parentID,
			CreatedAt: now.Add(time.Minute),
			UpdatedAt: now.Add(time.Minute),
		},
	})

	kanbanModel := NewModel(stubService{})
	kanbanModel.ready = true
	kanbanModel.width = 100
	kanbanModel.height = 30
	kanbanModel.catalog = catalog
	kanbanModel.scopeIndex = 0

	view := kanbanModel.renderEpicDetails()
	if !strings.Contains(view, "Epic from parent") {
		t.Fatalf("expected parent epic details from all scope selection, got %s", view)
	}
}

func TestEpicDetailsInNoEpicScopeShowsGracefulMessage(t *testing.T) {
	now := time.Now()
	catalog := buildCatalog([]model.Issue{
		{
			ID:        "proj-r001",
			Title:     "Root task",
			Type:      "task",
			Status:    "open",
			CreatedAt: now,
			UpdatedAt: now,
		},
	})

	kanbanModel := NewModel(stubService{})
	kanbanModel.ready = true
	kanbanModel.width = 100
	kanbanModel.height = 30
	kanbanModel.catalog = catalog
	kanbanModel.scopeIndex = 1

	view := kanbanModel.renderEpicDetails()
	if !strings.Contains(view, "No epic is associated with this scope.") {
		t.Fatalf("expected graceful no-epic message, got %s", view)
	}
}

func TestDetailsModalDownMovesSelectionWithinEpicScopeAndStaysOpen(t *testing.T) {
	now := time.Now()
	parentID := "proj-e1"
	catalog := buildCatalog([]model.Issue{
		{
			ID:        parentID,
			Title:     "E1: Epic",
			Type:      "epic",
			Status:    "open",
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:          parentID + ".0",
			Title:       "Older open task",
			Type:        "task",
			Status:      "open",
			Description: "Task desc",
			ParentID:    &parentID,
			CreatedAt:   now.Add(-2 * time.Minute),
			UpdatedAt:   now.Add(-2 * time.Minute),
		},
		{
			ID:          parentID + ".1",
			Title:       "Newest open task",
			Type:        "task",
			Status:      "open",
			Description: "Task desc",
			ParentID:    &parentID,
			CreatedAt:   now.Add(-1 * time.Minute),
			UpdatedAt:   now.Add(-1 * time.Minute),
		},
	})

	model := NewModel(stubService{})
	model.catalog = catalog
	model.ready = true
	model.width = 80
	model.height = 24
	model.scopeIndex = 2
	model.details = map[string]issueDetails{}

	if issue := model.currentIssue(); issue == nil || issue.ID != parentID+".1" {
		t.Fatalf("expected newest epic task selected first, got %#v", issue)
	}

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	if cmd == nil {
		t.Fatal("expected details load command when opening modal")
	}

	updated, cmd = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(Model)

	if !model.showDetails {
		t.Fatal("expected details modal to remain open")
	}
	if issue := model.currentIssue(); issue == nil || issue.ID != parentID+".0" {
		t.Fatalf("expected down to move to next epic task, got %#v", issue)
	}
	if model.inspectedIssueID != parentID+".0" {
		t.Fatalf("expected inspected issue to move with modal navigation, got %q", model.inspectedIssueID)
	}
	if cmd == nil {
		t.Fatal("expected details reload command for the new task")
	}
	if !model.details[parentID+".0"].Loading {
		t.Fatal("expected next task details to enter loading state")
	}
}

func TestDetailsModalUpMovesWithinAllTasksDoneColumnUsingClosedAtOrder(t *testing.T) {
	now := time.Now()
	olderClosed := now.Add(-20 * time.Minute)
	newerClosed := now.Add(-5 * time.Minute)
	catalog := buildCatalog([]model.Issue{
		{
			ID:        "proj-e1",
			Title:     "E1: Epic",
			Type:      "epic",
			Status:    "open",
			CreatedAt: now.Add(-3 * time.Hour),
			UpdatedAt: now.Add(-3 * time.Hour),
		},
		{
			ID:        "proj-e1.0",
			Title:     "Done first on screen",
			Type:      "task",
			Status:    "closed",
			CreatedAt: now.Add(-1 * time.Hour),
			UpdatedAt: now.Add(-1 * time.Hour),
			ClosedAt:  &newerClosed,
		},
		{
			ID:        "proj-e1.1",
			Title:     "Done second on screen",
			Type:      "task",
			Status:    "closed",
			CreatedAt: now.Add(-10 * time.Minute),
			UpdatedAt: now.Add(-10 * time.Minute),
			ClosedAt:  &olderClosed,
		},
	})

	model := NewModel(stubService{})
	model.catalog = catalog
	model.ready = true
	model.width = 80
	model.height = 24
	model.selectedCol = 2
	model.selectedRow = 1
	model.details = map[string]issueDetails{}

	if issue := model.currentIssue(); issue == nil || issue.ID != "proj-e1.1" {
		t.Fatalf("expected second done task selected, got %#v", issue)
	}

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	if cmd == nil {
		t.Fatal("expected details load command when opening modal")
	}

	updated, cmd = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	model = updated.(Model)

	if !model.showDetails {
		t.Fatal("expected details modal to remain open")
	}
	if issue := model.currentIssue(); issue == nil || issue.ID != "proj-e1.0" {
		t.Fatalf("expected up to move to newer-closed done task, got %#v", issue)
	}
	if model.inspectedIssueID != "proj-e1.0" {
		t.Fatalf("expected inspected issue to move to newer-closed task, got %q", model.inspectedIssueID)
	}
	if cmd == nil {
		t.Fatal("expected details reload command for the new done task")
	}
	if !model.details["proj-e1.0"].Loading {
		t.Fatal("expected done task details to enter loading state")
	}
}

func TestDetailsModalRightMovesToAdjacentColumnWithoutCycling(t *testing.T) {
	now := time.Now()
	parentID := "proj-e1"
	catalog := buildCatalog([]model.Issue{
		{
			ID:        parentID,
			Title:     "E1: Epic",
			Type:      "epic",
			Status:    "open",
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:          parentID + ".0",
			Title:       "Todo task",
			Type:        "task",
			Status:      "open",
			Description: "Task desc",
			ParentID:    &parentID,
			CreatedAt:   now.Add(2 * time.Minute),
			UpdatedAt:   now.Add(2 * time.Minute),
		},
		{
			ID:          parentID + ".1",
			Title:       "Claimed task",
			Type:        "task",
			Status:      "in_progress",
			Description: "Task desc",
			ParentID:    &parentID,
			CreatedAt:   now.Add(time.Minute),
			UpdatedAt:   now.Add(time.Minute),
		},
		{
			ID:          parentID + ".2",
			Title:       "Done task",
			Type:        "task",
			Status:      "closed",
			Description: "Task desc",
			ParentID:    &parentID,
			CreatedAt:   now,
			UpdatedAt:   now,
			ClosedAt:    ptrTime(now.Add(3 * time.Minute)),
		},
	})

	model := NewModel(stubService{})
	model.catalog = catalog
	model.ready = true
	model.width = 90
	model.height = 24

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	if cmd == nil {
		t.Fatal("expected details load command when opening modal")
	}

	updated, cmd = model.Update(tea.KeyMsg{Type: tea.KeyRight})
	model = updated.(Model)
	if !model.showDetails {
		t.Fatal("expected details modal to remain open")
	}
	if model.inspectedIssueID != parentID+".1" {
		t.Fatalf("expected right to move to claimed task, got %q", model.inspectedIssueID)
	}
	if model.selectedCol != 1 {
		t.Fatalf("expected selected column to move to claimed, got %d", model.selectedCol)
	}
	if cmd == nil {
		t.Fatal("expected details reload command for adjacent-column task")
	}

	updated, cmd = model.Update(tea.KeyMsg{Type: tea.KeyRight})
	model = updated.(Model)
	if model.inspectedIssueID != parentID+".2" {
		t.Fatalf("expected second right to move to done task, got %q", model.inspectedIssueID)
	}
	if model.selectedCol != 2 {
		t.Fatalf("expected selected column to move to done, got %d", model.selectedCol)
	}
	if cmd == nil {
		t.Fatal("expected details reload command for done task")
	}

	updated, cmd = model.Update(tea.KeyMsg{Type: tea.KeyRight})
	model = updated.(Model)
	if model.inspectedIssueID != parentID+".2" {
		t.Fatalf("expected right at board edge to remain on done task, got %q", model.inspectedIssueID)
	}
	if cmd != nil {
		t.Fatal("expected no command when moving right past the last column")
	}
}

func TestDetailsModalLeftRightStayStableWhenAdjacentColumnIsEmpty(t *testing.T) {
	now := time.Now()
	parentID := "proj-e1"
	catalog := buildCatalog([]model.Issue{
		{
			ID:        parentID,
			Title:     "E1: Epic",
			Type:      "epic",
			Status:    "open",
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:          parentID + ".0",
			Title:       "Only claimed task",
			Type:        "task",
			Status:      "in_progress",
			Description: "Task desc",
			ParentID:    &parentID,
			CreatedAt:   now.Add(time.Minute),
			UpdatedAt:   now.Add(time.Minute),
		},
	})

	model := NewModel(stubService{})
	model.catalog = catalog
	model.ready = true
	model.width = 90
	model.height = 24
	model.selectedCol = 1

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	if cmd == nil {
		t.Fatal("expected details load command when opening modal")
	}

	updated, cmd = model.Update(tea.KeyMsg{Type: tea.KeyLeft})
	model = updated.(Model)
	if model.inspectedIssueID != parentID+".0" || model.selectedCol != 1 {
		t.Fatalf("expected left toward empty column to stay stable, got id=%q col=%d", model.inspectedIssueID, model.selectedCol)
	}
	if cmd != nil {
		t.Fatal("expected no command when adjacent column is empty")
	}

	updated, cmd = model.Update(tea.KeyMsg{Type: tea.KeyRight})
	model = updated.(Model)
	if model.inspectedIssueID != parentID+".0" || model.selectedCol != 1 {
		t.Fatalf("expected right toward empty column to stay stable, got id=%q col=%d", model.inspectedIssueID, model.selectedCol)
	}
	if cmd != nil {
		t.Fatal("expected no command when adjacent column is empty")
	}
}

func TestRenderDetailsShowsCurrentColumnLabelAboveModal(t *testing.T) {
	now := time.Now()
	parentID := "proj-e1"
	issueID := parentID + ".0"
	catalog := buildCatalog([]model.Issue{
		{
			ID:        parentID,
			Title:     "E1: Epic",
			Type:      "epic",
			Status:    "open",
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:          issueID,
			Title:       "Claimed task",
			Type:        "task",
			Status:      "in_progress",
			Description: "Task desc",
			ParentID:    &parentID,
			CreatedAt:   now.Add(time.Minute),
			UpdatedAt:   now.Add(time.Minute),
		},
	})

	model := NewModel(stubService{})
	model.catalog = catalog
	model.ready = true
	model.width = 100
	model.height = 30
	model.showDetails = true
	model.inspectedIssueID = issueID

	view := model.renderDetails()
	if !strings.Contains(view, "CLAIMED") {
		t.Fatalf("expected current column label above modal, got %s", view)
	}
}

func TestDetailsModalStaysOpenWhenRefreshMovesIssueToAnotherColumn(t *testing.T) {
	now := time.Now()
	parentID := "proj-e1"
	issueID := parentID + ".0"

	initialCatalog := buildCatalog([]model.Issue{
		{
			ID:        parentID,
			Title:     "E1: Epic",
			Type:      "epic",
			Status:    "open",
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:          issueID,
			Title:       "Task that gets claimed",
			Type:        "task",
			Status:      "open",
			Description: "Task desc",
			ParentID:    &parentID,
			CreatedAt:   now.Add(time.Minute),
			UpdatedAt:   now.Add(time.Minute),
		},
	})

	claimedCatalog := buildCatalog([]model.Issue{
		{
			ID:        parentID,
			Title:     "E1: Epic",
			Type:      "epic",
			Status:    "open",
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:          issueID,
			Title:       "Task that gets claimed",
			Type:        "task",
			Status:      "in_progress",
			Description: "Task desc",
			ParentID:    &parentID,
			CreatedAt:   now.Add(time.Minute),
			UpdatedAt:   now.Add(2 * time.Minute),
		},
	})

	model := NewModel(stubService{})
	model.catalog = initialCatalog
	model.ready = true
	model.width = 100
	model.height = 30

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	if cmd == nil {
		t.Fatal("expected details load command when opening modal")
	}
	if !model.showDetails || model.inspectedIssueID != issueID {
		t.Fatalf("expected modal to target %q, got show=%v id=%q", issueID, model.showDetails, model.inspectedIssueID)
	}

	updated, _ = model.Update(catalogLoadedMsg{catalog: claimedCatalog})
	model = updated.(Model)

	if !model.showDetails {
		t.Fatal("expected details modal to remain open after refresh")
	}
	if issue := model.currentIssue(); issue == nil || issue.ID != issueID {
		t.Fatalf("expected board selection to follow inspected task after refresh, got %#v", issue)
	}
	if model.selectedCol != 1 {
		t.Fatalf("expected refreshed board selection to move to claimed column, got %d", model.selectedCol)
	}

	view := model.renderDetails()
	if !strings.Contains(view, "Task that gets claimed") {
		t.Fatalf("expected modal to keep showing the inspected task, got %s", view)
	}
	if !strings.Contains(view, "Status: in_progress") {
		t.Fatalf("expected modal to reflect refreshed claimed status, got %s", view)
	}
}

func TestViewRendersFallbackWhenWindowIsUndersized(t *testing.T) {
	model := NewModel(stubService{})
	model.ready = true
	model.width = 30
	model.height = 10

	view := model.View()
	if !strings.Contains(view, "Window too small") {
		t.Fatalf("expected undersized fallback, got %s", view)
	}
	if maxRenderedLineWidth(view) > model.width {
		t.Fatalf("expected undersized fallback to fit width %d, got %d", model.width, maxRenderedLineWidth(view))
	}
}

func TestViewFitsViewportAcrossRepresentativeSizes(t *testing.T) {
	now := time.Now()
	parentID := "proj-e1"
	svc := stubService{
		issues: []model.Issue{
			{
				ID:        parentID,
				Title:     "E1: Epic",
				Type:      "epic",
				Status:    "open",
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				ID:          parentID + ".0",
				Title:       "First task for narrow layout coverage",
				Type:        "task",
				Status:      "open",
				Description: "Task desc",
				ParentID:    &parentID,
				CreatedAt:   now.Add(time.Minute),
				UpdatedAt:   now.Add(time.Minute),
			},
			{
				ID:          parentID + ".1",
				Title:       "Claimed task with longer text to stress layout wrapping behavior",
				Type:        "task",
				Status:      "in_progress",
				Description: "Task desc",
				ParentID:    &parentID,
				CreatedAt:   now.Add(2 * time.Minute),
				UpdatedAt:   now.Add(2 * time.Minute),
			},
			{
				ID:          parentID + ".2",
				Title:       "Closed task with even longer text to stress the done column rendering path and metadata widths",
				Type:        "task",
				Status:      "closed",
				Description: "Task desc",
				ParentID:    &parentID,
				CreatedAt:   now.Add(3 * time.Minute),
				UpdatedAt:   now.Add(3 * time.Minute),
			},
		},
	}

	catalog, err := LoadCatalog(svc)
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}

	for width := 30; width <= 120; width++ {
		for height := 5; height <= 40; height++ {
			model := NewModel(svc)
			model.catalog = catalog
			model.ready = true
			model.width = width
			model.height = height

			view := model.View()
			if got := maxRenderedLineWidth(view); got > width {
				t.Fatalf("expected rendered width <= %d, got %d at %dx%d\n%s", width, got, width, height, view)
			}
			if got := renderedLineCount(view); got > height {
				t.Fatalf("expected rendered height <= %d, got %d at %dx%d\n%s", height, got, width, height, view)
			}
		}
	}
}

func maxRenderedLineWidth(text string) int {
	maxWidth := 0
	for _, line := range strings.Split(text, "\n") {
		lineWidth := lipgloss.Width(line)
		if lineWidth > maxWidth {
			maxWidth = lineWidth
		}
	}
	return maxWidth
}

func renderedLineCount(text string) int {
	return len(strings.Split(text, "\n"))
}

func ptrTime(value time.Time) *time.Time {
	return &value
}

func testScopes(count int) []Scope {
	scopes := make([]Scope, 0, count)
	for i := 0; i < count; i++ {
		scopes = append(scopes, Scope{
			Key:   fmt.Sprintf("scope-%02d", i+1),
			Title: fmt.Sprintf("Epic %02d", i+1),
		})
	}
	return scopes
}

func assertMarkerStyleContains(t *testing.T, rendered, marker, stylePart string) {
	t.Helper()

	markerIndex := strings.Index(rendered, marker)
	if markerIndex < 0 {
		t.Fatalf("expected rendered card to contain %q: %q", marker, rendered)
	}
	styleStart := strings.LastIndex(rendered[:markerIndex], "\x1b[")
	if styleStart < 0 {
		t.Fatalf("expected ANSI style before %q: %q", marker, rendered)
	}
	styleEnd := strings.Index(rendered[styleStart:], "m")
	if styleEnd < 0 {
		t.Fatalf("expected ANSI style terminator before %q: %q", marker, rendered)
	}
	style := rendered[styleStart : styleStart+styleEnd+1]
	if !strings.Contains(style, stylePart) {
		t.Fatalf("expected style before %q to contain %q, got %q", marker, stylePart, style)
	}
}
