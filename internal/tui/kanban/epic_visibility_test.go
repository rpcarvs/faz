package kanban

import (
	"reflect"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rpcarvs/faz/internal/model"
)

// epicVisibilityIssues includes unfinished epics, a closed epic and standalone work.
func epicVisibilityIssues() []model.Issue {
	openID, closedID, busyID := "proj-open", "proj-shut", "proj-busy"
	return []model.Issue{
		{ID: openID, Title: "Unfinished epic", Type: "epic", Status: "open", CreatedAt: time.Unix(1, 0)},
		{ID: closedID, Title: "Concluded epic", Type: "epic", Status: "closed", CreatedAt: time.Unix(2, 0)},
		{ID: busyID, Title: "Claimed epic", Type: "epic", Status: "in_progress", CreatedAt: time.Unix(3, 0)},
		{ID: openID + ".0", Type: "bug", Status: "closed", ParentID: &openID},
		{ID: busyID + ".0", Type: "task", Status: "in_progress", ParentID: &busyID},
		{ID: closedID + ".0", Type: "task", Status: "open", ParentID: &closedID},
		{ID: closedID + ".1", Type: "bug", Status: "in_progress", ParentID: &closedID},
		{ID: closedID + ".2", Type: "bug", Status: "closed", ParentID: &closedID},
		{ID: "proj-solo", Type: "bug", Status: "closed"},
	}
}

// visibilityModel loads data through the normal update path.
func visibilityModel(issues []model.Issue) Model {
	m := NewModel(stubService{issues: issues})
	m.applyWindowSize(100, 30)
	updated, _ := m.Update(m.loadCatalogCmd()())
	return updated.(Model)
}

// visibilityKey sends a board or modal key through the normal update path.
func visibilityKey(m Model, key tea.KeyMsg) Model {
	updated, _ := m.Update(key)
	return updated.(Model)
}

// TestCatalogClosedEpicVisibility retains unfinished work and leaves the source snapshot intact.
func TestCatalogClosedEpicVisibility(t *testing.T) {
	all := buildCatalog(epicVisibilityIssues())
	visible := all.withClosedEpics(false)
	var keys []string
	for _, scope := range visible.Scopes {
		keys = append(keys, scope.Key)
	}
	if want := []string{scopeAll, scopeNoEpic, "proj-busy", "proj-open"}; !reflect.DeepEqual(keys, want) {
		t.Fatalf("visible scopes=%v, want %v", keys, want)
	}
	columns := visible.Columns[scopeAll]
	if len(columns.Todo) != 0 || len(columns.Claimed) != 1 || len(columns.Done) != 2 {
		t.Fatalf("unexpected aggregate columns: %+v", columns)
	}
	if len(visible.Columns[scopeNoEpic].Done) != 1 || len(visible.Columns["proj-open"].Done) != 1 {
		t.Fatal("standalone work or completed work under unfinished epic disappeared")
	}
	if _, exists := visible.Columns["proj-shut"]; exists {
		t.Fatal("closed epic scope remained visible")
	}
	if len(all.Scopes) != 5 || len(all.Columns[scopeAll].Todo) != 1 {
		t.Fatal("visibility filtering modified the original catalog")
	}
	if restored := all.withClosedEpics(true); !reflect.DeepEqual(restored, all) {
		t.Fatal("showing closed epics did not restore the complete catalog")
	}
}

// TestClosedEpicToggleControlsBoardAndPicker covers both cycling directions and kind filtering.
func TestClosedEpicToggleControlsBoardAndPicker(t *testing.T) {
	m := visibilityModel(epicVisibilityIssues())
	if m.showClosedEpics || !strings.Contains(m.renderHeader(), "Closed epics: hidden") {
		t.Fatal("closed epics should be hidden by default")
	}
	for _, want := range []string{scopeNoEpic, "proj-busy", "proj-open", scopeAll} {
		m = visibilityKey(m, tea.KeyMsg{Type: tea.KeyTab})
		if m.currentScope().Key != want {
			t.Fatalf("forward cycle=%q, want %q", m.currentScope().Key, want)
		}
	}
	m = visibilityKey(m, tea.KeyMsg{Type: tea.KeyShiftTab})
	if m.currentScope().Key != "proj-open" {
		t.Fatal("reverse cycling should skip the closed epic")
	}
	m = visibilityKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	if strings.Contains(m.renderPicker(), "Concluded epic") {
		t.Fatal("picker displayed a hidden epic")
	}
	m = visibilityKey(m, tea.KeyMsg{Type: tea.KeyShiftTab})
	if m.catalog.Scopes[m.pickerIndex].Key != "proj-open" {
		t.Fatal("picker reverse cycling did not wrap")
	}
	m = visibilityKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	if !m.showPicker || !m.showClosedEpics || !strings.Contains(m.renderPicker(), "Concluded epic") {
		t.Fatal("toggle in picker did not reveal closed epics")
	}
	m = visibilityKey(m, tea.KeyMsg{Type: tea.KeyTab})
	if m.pickerIndex != 0 {
		t.Fatal("picker forward cycling did not wrap")
	}
	m = visibilityKey(m, tea.KeyMsg{Type: tea.KeyEsc})
	m.selectScopeByKey("proj-shut")
	m.typeFilter = "bug"
	if columns := m.currentColumns(); len(columns.Todo) != 0 || len(columns.Claimed) != 1 || len(columns.Done) != 1 {
		t.Fatalf("kind filter did not compose with closed epic scope: %+v", columns)
	}
	m.selectedCol, m.selectedRow, m.scrollRow = 2, 5, 4
	m = visibilityKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	if m.currentScope().Key != scopeAll || m.typeFilter != "bug" || m.selectedCol != 0 || m.selectedRow != 0 || m.scrollRow != 0 {
		t.Fatal("hiding the selected epic did not safely fall back while preserving kind filter")
	}
	if columns := m.currentColumns(); len(columns.Done) != 2 || len(columns.Claimed) != 0 {
		t.Fatalf("aggregate kind filter includes hidden work: %+v", columns)
	}
	for _, key := range []string{scopeNoEpic, "proj-open"} {
		m.selectScopeByKey(key)
		m = visibilityKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
		if m.currentScope().Key != key {
			t.Fatalf("toggle changed a retained scope %q", key)
		}
	}
}

// TestEpicVisibilityRefreshPreservesKeys handles reordering, closure, reopening and removal.
func TestEpicVisibilityRefreshPreservesKeys(t *testing.T) {
	issues := epicVisibilityIssues()
	m := visibilityModel(issues)
	m.selectScopeByKey("proj-open")
	m.showPicker = true
	m.pickerIndex = m.scopeIndex
	issues = append(issues, model.Issue{ID: "proj-new", Type: "epic", Status: "open", CreatedAt: time.Unix(4, 0)})
	updated, _ := m.Update(catalogLoadedMsg{catalog: buildCatalog(issues)})
	m = updated.(Model)
	if m.currentScope().Key != "proj-open" || m.catalog.Scopes[m.pickerIndex].Key != "proj-open" {
		t.Fatal("refresh changed board or picker selection when scopes reordered")
	}
	issues[0].Status = "closed"
	updated, _ = m.Update(catalogLoadedMsg{catalog: buildCatalog(issues)})
	m = updated.(Model)
	if m.currentScope().Key != scopeAll || m.pickerIndex != 0 || m.pickerScroll != 0 {
		t.Fatal("closure did not reconcile hidden board and picker selection")
	}
	issues[0].Status = "open"
	updated, _ = m.Update(catalogLoadedMsg{catalog: buildCatalog(issues)})
	m = updated.(Model)
	m.selectScopeByKey("proj-open")
	if m.currentScope().Key != "proj-open" {
		t.Fatal("reopened epic did not reappear")
	}
	updated, _ = m.Update(catalogLoadedMsg{catalog: buildCatalog(nil)})
	m = updated.(Model)
	if m.currentScope().Key != scopeAll || len(m.catalog.Scopes) != 2 {
		t.Fatal("removing selected epic did not preserve aggregate and standalone scopes")
	}
}

// TestEpicVisibilityEmptyAndModalStates checks startup safety and key isolation.
func TestEpicVisibilityEmptyAndModalStates(t *testing.T) {
	for _, issues := range [][]model.Issue{nil, {{ID: "proj-shut", Type: "epic", Status: "closed"}}} {
		m := visibilityModel(issues)
		if len(m.catalog.Scopes) != 2 {
			t.Fatal("empty or all-closed catalog lost default scopes")
		}
	}
	m := NewModel(stubService{}, WithPicker())
	m = visibilityKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	updated, _ := m.Update(catalogLoadedMsg{catalog: buildCatalog(epicVisibilityIssues())})
	m = updated.(Model)
	if len(m.catalog.Scopes) != 5 || !m.showClosedEpics {
		t.Fatal("toggle before initial load was not retained")
	}
	for _, modal := range []string{"details", "epic", "kind", "help"} {
		t.Run(modal, func(t *testing.T) {
			m := visibilityModel(epicVisibilityIssues())
			m.showDetails = modal == "details"
			m.showEpic = modal == "epic"
			m.showType = modal == "kind"
			m.showHelp = modal == "help"
			m = visibilityKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
			if m.showClosedEpics {
				t.Fatal("toggle leaked through a modal")
			}
		})
	}
}
