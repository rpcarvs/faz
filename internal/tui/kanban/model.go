package kanban

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rpcarvs/faz/internal/model"
)

const refreshInterval = 10 * time.Second

type catalogLoadedMsg struct {
	catalog Catalog
	err     error
}

type refreshTickMsg time.Time

// Model manages the read-only kanban TUI state.
type Model struct {
	svc Service

	width  int
	height int
	ready  bool

	catalog Catalog
	err     error

	scopeIndex  int
	selectedCol int
	selectedRow int
	scrollRow   int

	showPicker  bool
	pickerIndex int

	showDetails bool
}

// NewModel builds a new kanban TUI model.
func NewModel(svc Service) Model {
	return Model{svc: svc}
}

// Init starts the first data load and background refresh loop.
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.loadCatalogCmd(), tickCmd())
}

// Update handles input, resizing, refreshes, and modal state transitions.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.ensureSelection()
		return m, nil

	case catalogLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.catalog = msg.catalog
		if m.scopeIndex >= len(m.catalog.Scopes) {
			m.scopeIndex = 0
		}
		m.ensureSelection()
		return m, nil

	case refreshTickMsg:
		return m, tea.Batch(m.loadCatalogCmd(), tickCmd())

	case tea.KeyMsg:
		if m.showDetails {
			switch msg.String() {
			case "enter", "esc":
				m.showDetails = false
			case "q", "ctrl+c":
				return m, tea.Quit
			}
			return m, nil
		}
		if m.showPicker {
			return m.updatePicker(msg)
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.cycleScope(1)
			return m, nil
		case "shift+tab":
			m.cycleScope(-1)
			return m, nil
		case "a":
			m.selectScopeByKey(scopeAll)
			return m, nil
		case "e":
			m.showPicker = true
			m.pickerIndex = m.scopeIndex
			return m, nil
		case "left", "h":
			m.moveCol(-1)
			return m, nil
		case "right", "l":
			m.moveCol(1)
			return m, nil
		case "up", "k":
			m.moveRow(-1)
			return m, nil
		case "down", "j":
			m.moveRow(1)
			return m, nil
		case "enter":
			if m.currentIssue() != nil {
				m.showDetails = true
			}
			return m, nil
		case "r":
			return m, m.loadCatalogCmd()
		}
	}

	return m, nil
}

// View renders the kanban layout, modals, and help footer.
func (m Model) View() string {
	if !m.ready {
		return "Loading kanban..."
	}
	if m.err != nil {
		return fmt.Sprintf("Failed to load kanban: %v", m.err)
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		m.renderHeader(),
		m.renderBoard(),
		m.renderFooter(),
	)

	if m.showPicker {
		return m.overlay(content, m.renderPicker())
	}
	if m.showDetails {
		return m.overlay(content, m.renderDetails())
	}
	return content
}

func (m Model) updatePicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.showPicker = false
		return m, nil
	case "up", "k":
		if m.pickerIndex > 0 {
			m.pickerIndex--
		}
		return m, nil
	case "down", "j":
		if m.pickerIndex < len(m.catalog.Scopes)-1 {
			m.pickerIndex++
		}
		return m, nil
	case "enter":
		m.scopeIndex = m.pickerIndex
		m.showPicker = false
		m.selectedCol = 0
		m.selectedRow = 0
		m.scrollRow = 0
		m.ensureSelection()
		return m, nil
	case "a":
		m.selectScopeByKey(scopeAll)
		m.showPicker = false
		return m, nil
	case "tab":
		m.pickerIndex = (m.pickerIndex + 1) % len(m.catalog.Scopes)
		return m, nil
	}
	return m, nil
}

func (m *Model) cycleScope(step int) {
	if len(m.catalog.Scopes) == 0 {
		return
	}
	m.scopeIndex = (m.scopeIndex + step + len(m.catalog.Scopes)) % len(m.catalog.Scopes)
	m.selectedCol = 0
	m.selectedRow = 0
	m.scrollRow = 0
	m.ensureSelection()
}

func (m *Model) selectScopeByKey(key string) {
	for i, scope := range m.catalog.Scopes {
		if scope.Key != key {
			continue
		}
		m.scopeIndex = i
		m.selectedCol = 0
		m.selectedRow = 0
		m.scrollRow = 0
		m.ensureSelection()
		return
	}
}

func (m *Model) moveCol(delta int) {
	next := m.selectedCol + delta
	if next < 0 || next > 2 {
		return
	}
	m.selectedCol = next
	m.ensureSelection()
}

func (m *Model) moveRow(delta int) {
	column := m.currentColumn()
	if len(column) == 0 {
		m.selectedRow = 0
		m.scrollRow = 0
		return
	}
	next := m.selectedRow + delta
	if next < 0 {
		next = 0
	}
	if next >= len(column) {
		next = len(column) - 1
	}
	m.selectedRow = next
	m.adjustScroll()
}

func (m *Model) ensureSelection() {
	column := m.currentColumn()
	if len(column) == 0 {
		m.selectedRow = 0
		m.scrollRow = 0
		return
	}
	if m.selectedRow >= len(column) {
		m.selectedRow = len(column) - 1
	}
	if m.selectedRow < 0 {
		m.selectedRow = 0
	}
	m.adjustScroll()
}

func (m *Model) adjustScroll() {
	rowsVisible := m.visibleRows()
	if rowsVisible < 1 {
		m.scrollRow = 0
		return
	}
	if m.selectedRow < m.scrollRow {
		m.scrollRow = m.selectedRow
	}
	if m.selectedRow >= m.scrollRow+rowsVisible {
		m.scrollRow = m.selectedRow - rowsVisible + 1
	}
	if m.scrollRow < 0 {
		m.scrollRow = 0
	}
}

func (m Model) visibleRows() int {
	boardHeight := m.height - 9
	if boardHeight < 8 {
		return 1
	}
	cardHeight := 5
	rowUnit := cardHeight + 1
	rows := boardHeight / rowUnit
	if rows < 1 {
		return 1
	}
	return rows
}

func (m Model) currentScope() Scope {
	if len(m.catalog.Scopes) == 0 || m.scopeIndex >= len(m.catalog.Scopes) {
		return Scope{Key: scopeAll, Title: "All Epics"}
	}
	return m.catalog.Scopes[m.scopeIndex]
}

func (m Model) currentColumns() scopeColumns {
	scope := m.currentScope()
	return m.catalog.Columns[scope.Key]
}

func (m Model) currentColumn() []model.Issue {
	columns := m.currentColumns()
	switch m.selectedCol {
	case 1:
		return columns.Claimed
	case 2:
		return columns.Done
	default:
		return columns.Todo
	}
}

func (m Model) currentIssue() *model.Issue {
	column := m.currentColumn()
	if len(column) == 0 || m.selectedRow < 0 || m.selectedRow >= len(column) {
		return nil
	}
	issue := column[m.selectedRow]
	return &issue
}

func (m Model) loadCatalogCmd() tea.Cmd {
	return func() tea.Msg {
		catalog, err := LoadCatalog(m.svc)
		return catalogLoadedMsg{catalog: catalog, err: err}
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(refreshInterval, func(t time.Time) tea.Msg {
		return refreshTickMsg(t)
	})
}

func (m Model) renderHeader() string {
	scope := m.currentScope()
	headerStyle := lipgloss.NewStyle().
		Width(m.width).
		Padding(0, 1).
		Bold(true).
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("24"))

	subtitleStyle := lipgloss.NewStyle().
		Width(m.width).
		Padding(0, 1).
		Foreground(lipgloss.Color("245"))

	title := "faz kanban"
	scopeTitle := "Epic: " + scope.Title
	subtitle := "Auto-refresh every 10s."
	return lipgloss.JoinVertical(
		lipgloss.Left,
		headerStyle.Render(fmt.Sprintf("%s  |  %s", title, scopeTitle)),
		subtitleStyle.Render(subtitle),
	)
}

func (m Model) renderBoard() string {
	columns := m.currentColumns()
	boardWidth := m.width
	gap := 2
	colWidth := (boardWidth - 2*gap) / 3
	if colWidth < 18 {
		colWidth = 18
	}

	todo := m.renderColumn("TO DO", columns.Todo, 0, colWidth, lipgloss.Color("178"))
	claimed := m.renderColumn("CLAIMED", columns.Claimed, 1, colWidth, lipgloss.Color("39"))
	done := m.renderColumn("DONE", columns.Done, 2, colWidth, lipgloss.Color("71"))

	row := lipgloss.JoinHorizontal(lipgloss.Top, todo, strings.Repeat(" ", gap), claimed, strings.Repeat(" ", gap), done)
	return lipgloss.PlaceHorizontal(m.width, lipgloss.Center, row)
}

func (m Model) renderColumn(title string, issues []model.Issue, colIndex, width int, accent lipgloss.Color) string {
	header := lipgloss.NewStyle().
		Width(width).
		Bold(true).
		Padding(0, 1).
		Foreground(lipgloss.Color("230")).
		Background(accent).
		Render(fmt.Sprintf("%s (%d)", title, len(issues)))

	rowsVisible := m.visibleRows()
	start := 0
	if colIndex == m.selectedCol {
		start = m.scrollRow
	}
	if start > len(issues) {
		start = len(issues)
	}
	end := start + rowsVisible
	if end > len(issues) {
		end = len(issues)
	}

	body := make([]string, 0, rowsVisible)
	if len(issues) == 0 {
		empty := lipgloss.NewStyle().
			Width(maxInt(1, width-4)).
			Height(5).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238")).
			Padding(1, 1).
			Foreground(lipgloss.Color("244")).
			Render("No tasks")
		body = append(body, lipgloss.PlaceHorizontal(width, lipgloss.Center, empty))
	} else {
		for rowIndex := start; rowIndex < end; rowIndex++ {
			body = append(body, m.renderCard(issues[rowIndex], colIndex == m.selectedCol && rowIndex == m.selectedRow, width))
		}
	}
	return lipgloss.JoinVertical(lipgloss.Left, append([]string{header}, body...)...)
}

func (m Model) renderCard(issue model.Issue, selected bool, width int) string {
	borderColor := lipgloss.Color("238")
	bgColor := lipgloss.Color("235")
	titleColor := lipgloss.Color("252")
	if selected {
		borderColor = lipgloss.Color("246")
		bgColor = lipgloss.Color("240")
		titleColor = lipgloss.Color("230")
	}

	meta := fmt.Sprintf("P%d • %s", issue.Priority, issue.Type)
	if m.currentScope().Key == scopeAll {
		if issue.ParentID != nil {
			meta = fmt.Sprintf("%s • %s", meta, truncateLine(m.catalog.EpicTitles[*issue.ParentID], width-6))
		} else {
			meta = fmt.Sprintf("%s • No Epic", meta)
		}
	}
	title := fitLines(issue.Title, width-4, 2)

	cardStyle := lipgloss.NewStyle().
		Width(maxInt(1, width-4)).
		Height(5).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Background(bgColor).
		Padding(0, 1)

	titleStyle := lipgloss.NewStyle().Foreground(titleColor).Bold(selected)
	metaColor := lipgloss.Color("243")
	if selected {
		metaColor = lipgloss.Color("251")
	}
	metaStyle := lipgloss.NewStyle().Foreground(metaColor)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render(title),
		"",
		metaStyle.Render(truncateLine(meta, width-4)),
	)
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, cardStyle.Render(content))
}

func (m Model) renderFooter() string {
	footer := "Tab next epic • Shift+Tab previous • e epic list • a all • arrows move • Enter details • r refresh • q quit"
	return lipgloss.NewStyle().
		Width(m.width).
		Padding(0, 1).
		Foreground(lipgloss.Color("244")).
		Render(truncateLine(footer, m.width-2))
}

func (m Model) renderPicker() string {
	width := minInt(60, maxInt(40, m.width-10))
	lines := []string{"Select scope", ""}
	for i, scope := range m.catalog.Scopes {
		prefix := "  "
		if i == m.pickerIndex {
			prefix = "> "
		}
		lines = append(lines, prefix+scope.Title)
	}
	lines = append(lines, "", "Enter select • Esc close • a all")
	box := lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("69")).
		Padding(1, 2).
		Background(lipgloss.Color("235")).
		Render(strings.Join(lines, "\n"))
	return box
}

func (m Model) renderDetails() string {
	issue := m.currentIssue()
	if issue == nil {
		return ""
	}
	width := minInt(84, maxInt(54, m.width-12))
	parentTitle := "None"
	if issue.ParentID != nil {
		parentTitle = m.catalog.EpicTitles[*issue.ParentID]
	}
	lines := []string{
		issue.Title,
		"",
		fmt.Sprintf("ID: %s", issue.ID),
		fmt.Sprintf("Type: %s", issue.Type),
		fmt.Sprintf("Priority: P%d", issue.Priority),
		fmt.Sprintf("Status: %s", issue.Status),
		fmt.Sprintf("Epic: %s", parentTitle),
		"",
		issue.Description,
		"",
		"Enter or Esc closes this view.",
	}
	box := lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("69")).
		Padding(1, 2).
		Background(lipgloss.Color("235")).
		Render(strings.Join(lines, "\n"))
	return box
}

func (m Model) overlay(base, modal string) string {
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal)
}

func fitLines(text string, width, maxLines int) string {
	if width < 8 {
		width = 8
	}
	lines := lipgloss.NewStyle().Width(width).MaxWidth(width).Render(text)
	parts := strings.Split(lines, "\n")
	if len(parts) <= maxLines {
		return strings.Join(parts, "\n")
	}
	parts = parts[:maxLines]
	last := truncateLine(parts[len(parts)-1], width)
	parts[len(parts)-1] = strings.TrimRight(last, " ") + "…"
	return strings.Join(parts, "\n")
}

func truncateLine(text string, width int) string {
	if width <= 1 {
		return ""
	}
	runes := []rune(text)
	if len(runes) <= width {
		return text
	}
	return string(runes[:width-1]) + "…"
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
